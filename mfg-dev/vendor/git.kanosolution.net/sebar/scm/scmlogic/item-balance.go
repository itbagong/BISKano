package scmlogic

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type ItemBalanceEngine struct{}

type ItemBalanceGetQtyRequest struct {
	ItemID    string
	CompanyID string
	SKU       string
	InventDim scmmodel.InventDimension
}

type ItemBalanceGetQtyResponse struct {
	ItemID    string
	CompanyID string
	SKU       string
	InventDim scmmodel.InventDimension

	Qty              float64
	QtyReserved      float64
	QtyPlanned       float64
	QtyAvail         float64
	AmountPhysical   float64
	AmountFinancial  float64
	AmountAdjustment float64
}

func (o *ItemBalanceEngine) GetQty(ctx *kaos.Context, payload []ItemBalanceGetQtyRequest) ([]ItemBalanceGetQtyResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	res := []ItemBalanceGetQtyResponse{}
	for _, req := range payload {
		fs := []*dbflex.Filter{}

		if req.ItemID != "" {
			fs = append(fs, dbflex.Eq("ItemID", req.ItemID))
		}

		if req.CompanyID != "" {
			fs = append(fs, dbflex.Eq("CompanyID", req.CompanyID))
		}

		if req.SKU != "" {
			fs = append(fs, dbflex.Eq("SKU", req.SKU))
		}

		if req.InventDim.WarehouseID != "" {
			fs = append(fs, dbflex.Eq("InventDim.WarehouseID", req.InventDim.WarehouseID))
		}

		if req.InventDim.AisleID != "" {
			fs = append(fs, dbflex.Eq("InventDim.AisleID", req.InventDim.AisleID))
		}

		if req.InventDim.SectionID != "" {
			fs = append(fs, dbflex.Eq("InventDim.SectionID", req.InventDim.SectionID))
		}

		if req.InventDim.BoxID != "" {
			fs = append(fs, dbflex.Eq("InventDim.BoxID", req.InventDim.BoxID))
		}

		parm := dbflex.NewQueryParam()
		if len(fs) > 0 {
			parm = parm.SetWhere(dbflex.And(fs...))
		}

		ibs := []scmmodel.ItemBalance{}
		h.Gets(new(scmmodel.ItemBalance), parm, &ibs)

		rs := ItemBalanceGetQtyResponse{}
		for _, ib := range ibs {
			rs.ItemID = ib.ItemID
			rs.CompanyID = ib.CompanyID
			rs.SKU = ib.SKU
			rs.InventDim = ib.InventDim

			rs.Qty += ib.Qty
			rs.QtyReserved += ib.QtyReserved
			rs.QtyPlanned += ib.QtyPlanned

			nib := scmmodel.ItemBalance{
				Qty:         rs.Qty,
				QtyReserved: rs.QtyReserved,
				QtyPlanned:  rs.QtyPlanned,
			}
			nib.Calc()
			rs.QtyAvail = nib.QtyAvail

			rs.AmountPhysical += ib.AmountPhysical
			rs.AmountFinancial += ib.AmountFinancial
			rs.AmountAdjustment += ib.AmountAdjustment
		}
		res = append(res, rs)
	}

	return res, nil
}

type BalGetAvailWarehouseResponse struct {
	ID          string `json:"_id"`
	Text        string // "BLITAR | AISLE 2 | SCM = 200"
	Qty         float64
	QtyReserved float64
	QtyPlanned  float64
	QtyAvail    float64
	InventDim   scmmodel.InventDimension
}

func (o *ItemBalanceEngine) GetAvailableWarehouse(ctx *kaos.Context, _ *interface{}) ([]BalGetAvailWarehouseResponse, error) {
	req := GetURLQueryParams(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	fs := []*dbflex.Filter{}

	if req["ItemID"] != "" {
		fs = append(fs, dbflex.Eq("ItemID", req["ItemID"]))
	}

	if req["CompanyID"] != "" {
		fs = append(fs, dbflex.Eq("CompanyID", req["CompanyID"]))
	}

	if req["SKU"] != "" {
		fs = append(fs, dbflex.Eq("SKU", req["SKU"]))
	}

	f := new(dbflex.Filter)
	if len(fs) > 0 {
		f = dbflex.And(fs...)
	}

	bals := []scmmodel.ItemBalance{}
	if e := h.GetsByFilter(new(scmmodel.ItemBalance), f, &bals); e != nil {
		return nil, e
	}

	// filter based on IR
	if req["ItemRequestID"] != "" && req["FulfillmentType"] != "" {
		irID := req["ItemRequestID"]

		ir, e := datahub.GetByID(h, new(scmmodel.ItemRequest), irID)
		if e != nil {
			return nil, e
		}

		switch req["FulfillmentType"] {
		case string(scmmodel.ItemRequestFulfillmentTypeItemTransfer):
			// "From Warehouse" nya semua Warehouse dari ItemID dan SKU yang sudah dipilih KECUALI WH yang sudah dipilih di tab General
			bals = lo.Filter(bals, func(d scmmodel.ItemBalance, i int) bool {
				return d.InventDim.WarehouseID != ir.InventDimTo.WarehouseID
			})
		case string(scmmodel.ItemRequestFulfillmentTypeMovementOut):
			// "From Warehouse" hanya WH yang sudah dipilih di tab General (tp bisa lebih dari satu karena bisa beda Section dan Aisle dan Box)
			bals = lo.Filter(bals, func(d scmmodel.ItemBalance, i int) bool {
				return d.InventDim.WarehouseID == ir.InventDimTo.WarehouseID
			})
		}
	}

	if req["WarehouseID"] != "" {
		// filter balance hanya yang WarehouseID -> dipakai di WO
		bals = lo.Filter(bals, func(d scmmodel.ItemBalance, i int) bool {
			return d.InventDim.WarehouseID == req["WarehouseID"]
		})
	}

	if req["_id"] != "" {
		// filter balance hanya ID tertentu -> dipakai di WO
		bals = lo.Filter(bals, func(d scmmodel.ItemBalance, i int) bool {
			return d.ID == req["_id"]
		})
	}

	whs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))
	ais := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationAisle))
	scs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationSection))
	bxs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationBox))

	res := lo.FilterMap(bals, func(d scmmodel.ItemBalance, i int) (BalGetAvailWarehouseResponse, bool) {
		r := BalGetAvailWarehouseResponse{
			ID: d.ID,
		}
		reflector.CopyAttributes(&d, &r)

		locs := []string{}
		if r.InventDim.WarehouseID != "" {
			l, _ := whs.Get(r.InventDim.WarehouseID)
			locs = append(locs, l.Name)
		}
		if r.InventDim.AisleID != "" {
			l, _ := ais.Get(r.InventDim.AisleID)
			locs = append(locs, l.Name)
		}
		if r.InventDim.SectionID != "" {
			l, _ := scs.Get(r.InventDim.SectionID)
			locs = append(locs, l.Name)
		}
		if r.InventDim.BoxID != "" {
			l, _ := bxs.Get(r.InventDim.BoxID)
			locs = append(locs, l.Name)
		}

		r.Text = fmt.Sprintf("%s = %v", strings.Join(locs, " | "), r.Qty)
		return r, r.Qty > 0
	})

	return res, nil
}

type GetsByWarehouseAndSectionRequest struct {
	// ItemID string
	// // WarehouseID string
	// // SectionID   string
	// GroupBy     []string
	// OpeningDate time.Time
	// ClosingDate time.Time
	Where struct {
		ID        []string
		DateFrom  *time.Time
		DateTo    *time.Time
		Dimension []string
	}
	Skip int
	Take int
}

type InventTrxTypeGroup struct {
	WarehouseID      string
	SectionID        string
	Status           scmmodel.ItemBalanceStatus
	Qty              float64
	TrxQty           float64
	TotalPlanned     float64
	TotalReserved    float64
	TotalQty         float64
	AmountPhysical   float64
	AmountFinancial  float64
	AmountAdjustment float64
}

type InventTrxGroupByWarehouseSection struct {
	WarehouseID           string
	SectionID             string
	TotalPlanned          float64
	TotalConfirmed        float64
	TotalReserved         float64
	TotalAmountPhysical   float64
	TotalAmountFinancial  float64
	TotalAmountAdjustment float64
}

func (o *ItemBalanceEngine) GetsByWarehouseAndSection(ctx *kaos.Context, payload *GetsByWarehouseAndSectionRequest) (interface{}, error) {
	companyID, _ := GetCompanyIDFromContext(ctx)

	//get connection
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil || len(payload.Where.ID) == 0 {
		return nil, fmt.Errorf("missing: payload")
	}

	opt := ItemBalanceOpt{
		CompanyID: companyID,
		ItemIDs:   payload.Where.ID,
		GroupBy:   payload.Where.Dimension,
	}

	//get opening balance
	var balanceDate *time.Time = nil
	if payload.Where.DateTo != nil {
		balanceDate = payload.Where.DateTo
	}

	itemBalances, err := NewItemBalanceHub(h).
		Get(balanceDate, opt)

	if err != nil {
		return nil, err
	}

	//get closing balance
	balanceDate = nil
	if payload.Where.DateTo != nil {
		balanceDate = payload.Where.DateTo
	}

	itemBalanceClosings, err := NewItemBalanceHub(h).
		Get(balanceDate, opt)
	if err != nil {
		return nil, err
	}

	//convert closing item balance ke map agar mudah di gabungkan dengan opening balance
	balanceClosingMap := lo.SliceToMap(itemBalanceClosings, func(balance *scmmodel.ItemBalance) (string, *scmmodel.ItemBalance) {
		balance.InventDim.Calc()
		return fmt.Sprintf("%s|%s|%s", balance.ItemID, balance.CompanyID, balance.InventDim.InventDimID), balance
	})

	mapWarehouses := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))
	mapAisles := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationAisle))
	mapSections := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationSection))
	mapBoxs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationBox))

	res := lo.Map(itemBalances, func(balance *scmmodel.ItemBalance, index int) *scmmodel.ItemBalanceWithName {
		res := new(scmmodel.ItemBalanceWithName)
		res.ItemBalance = *balance
		warehouse, err := mapWarehouses.Get(balance.InventDim.WarehouseID)
		if err == nil {
			res.WarehouseID = warehouse.Name
		}

		aisle, err := mapAisles.Get(balance.InventDim.AisleID)
		if err == nil {
			res.AisleID = aisle.Name
		}

		section, err := mapSections.Get(balance.InventDim.SectionID)
		if err == nil {
			res.SectionID = section.Name
		}

		box, err := mapBoxs.Get(balance.InventDim.BoxID)
		if err == nil {
			res.BoxID = box.Name
		}

		res.BatchID = balance.InventDim.BatchID
		res.SerialNumber = balance.InventDim.SerialNumber
		res.Size = balance.InventDim.Size
		res.Grade = balance.InventDim.Grade
		res.VariantID = balance.InventDim.VariantID

		res.Qty = balance.Qty
		res.QtyPlanned = balance.QtyPlanned
		res.QtyReserved = balance.QtyReserved
		res.AmountFinancial = balance.AmountFinancial
		res.AmountPhysical = balance.AmountPhysical
		res.AmountAdjustment = balance.AmountAdjustment

		if res.AmountFinancial != 0 {
			res.Amount += (res.AmountFinancial + res.AmountAdjustment)
		}

		//cek apakah punya closing balance
		balance.InventDim.Calc()
		keyBalance := fmt.Sprintf("%s|%s|%s", balance.ItemID, balance.CompanyID, balance.InventDim.InventDimID)
		if closingBalance, ok := balanceClosingMap[keyBalance]; ok {
			res.QtyClosing = closingBalance.Qty
			res.PlannedQtyClosing = closingBalance.QtyPlanned
			res.ReservedQtyClosing = closingBalance.QtyReserved
		}

		return res
	})

	return codekit.M{"count": len(res), "data": res}, nil
}

func (o *ItemBalanceEngine) findLatestBalanceDate(h *datahub.Hub, companyID string, balanceDate time.Time) *time.Time {
	snapshotBal, _ := datahub.GetByParm(h, new(scmmodel.ItemBalance), dbflex.NewQueryParam().
		SetWhere(dbflex.And(dbflex.Eq("CompanyID", companyID), dbflex.Gte("BalanceDate", balanceDate))).
		SetSort("BalanceDate").
		SetSelect("BalanceDate"))
	return snapshotBal.BalanceDate
}
