package scmlogic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type ItemEngine struct{}

type ItemGetsRequest struct {
	CompanyID string // TODO: required this using suim.Validate
	ItemIDs   []string
	DateFrom  *time.Time
	DateTo    *time.Time
	Spec      ItemGetsSpec
	Check     *tenantcoremodel.ItemDimensionCheck

	Skip int
	Take int
}

type ItemGetsSpec struct {
	VariantID string
	Size      string
	Grade     string
}

func (i *ItemGetsSpec) Hash() string {
	if i.VariantID == "" && i.Size == "" && i.Grade == "" {
		return ""
	}
	dim := scmmodel.InventDimension{
		VariantID: i.VariantID,
		Size:      i.Size,
		Grade:     i.Grade,
	}
	return dim.SpecHash()
}

func (o *ItemEngine) Gets(ctx *kaos.Context, p *ItemGetsRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	param := dbflex.NewQueryParam()
	if len(p.ItemIDs) > 0 {
		param = param.SetWhere(dbflex.In("_id", p.ItemIDs...))
	}

	paramSkipTake := &*param
	if p.Skip >= 0 && p.Take > 0 {
		paramSkipTake = param.SetSkip(p.Skip).SetTake(p.Take)
	}

	items := []tenantcoremodel.Item{}
	if e := h.Gets(new(tenantcoremodel.Item), paramSkipTake, &items); e != nil {
		return nil, e
	}

	total, e := datahub.Count(h, new(tenantcoremodel.Item), param)
	if e != nil {
		return nil, e
	}

	itemIDs := lo.Map(items, func(d tenantcoremodel.Item, index int) string {
		return d.ID
	})

	// get each item balances
	fs := []*dbflex.Filter{
		dbflex.In("ItemID", itemIDs...),
	}

	if p.DateFrom != nil {
		fs = append(fs, dbflex.Gte("BalanceDate", *p.DateFrom))
	}

	if p.DateTo != nil {
		fs = append(fs, dbflex.Lte("BalanceDate", *p.DateTo))
	}

	if sh := p.Spec.Hash(); sh != "" {
		fs = append(fs, dbflex.Eq("InventDim.SpecID", sh))
	}

	balances, e := datahub.FindByFilter(h, new(scmmodel.ItemBalance), dbflex.And(fs...))
	if e != nil {
		return nil, e
	}

	// get BalanceSum from Invent Trx
	trxFilters := []*dbflex.Filter{
		dbflex.Gt("TrxDate", p.DateFrom),
		dbflex.In("Item._id", itemIDs...),
	}

	lastBalDate, e := new(ItemBalanceLogic).GetLastDate(h, p.CompanyID, p.DateFrom)
	if e != nil {
		return nil, e
	}

	if lastBalDate != nil {
		trxFilters = append(trxFilters, dbflex.Lte("TrxDate", lastBalDate))
	}

	type InventTrxResult struct {
		ID struct {
			CompanyID string
			ItemID    string
			InventDim struct {
				InventDimID string
			}
			Status string
		} `json:"_id"`
		Qty             float64
		QtyReserved     float64
		QtyPlanned      float64
		AmountFinancial float64
	}

	trxs, e := datahub.FindAnyByParm(h, new(InventTrxResult), new(scmmodel.InventTrx).TableName(),
		dbflex.NewQueryParam().
			SetWhere(dbflex.And(trxFilters...)).
			SetAggr(
				dbflex.NewAggrItem("Qty", dbflex.AggrSum, "Qty"),
				dbflex.NewAggrItem("QtyReserved", dbflex.AggrSum, "QtyReserved"),
				dbflex.NewAggrItem("QtyPlanned", dbflex.AggrSum, "QtyPlanned"),
				dbflex.NewAggrItem("AmountFinancial", dbflex.AggrSum, "AmountFinancial"),
			).
			SetGroupBy(
				"CompanyID",
				"ItemID",
				"InventDim.InventDimID",
				"Status",
			))
	if e != nil {
		return nil, e
	}

	trxM := lo.GroupBy(trxs, func(b *InventTrxResult) string {
		return fmt.Sprintf("%s||%s||%s", b.ID.CompanyID, b.ID.ItemID, b.ID.InventDim.InventDimID)
	})

	// deduct each item balances with invent trx
	balseM := map[string]scmmodel.ItemBalance{}
	for k, ts := range trxM {
		b := new(scmmodel.ItemBalance)
		for _, t := range ts {
			switch t.ID.Status {
			case string(scmmodel.ItemConfirmed):
				b.Qty -= t.Qty

			case string(scmmodel.ItemReserved):
				b.QtyReserved -= t.Qty

			case string(scmmodel.ItemPlanned):
				b.QtyPlanned -= t.Qty
			}

			b.AmountFinancial -= t.AmountFinancial
		}
		balseM[k] = *b
	}

	for _, b := range balances {
		balSub := balseM[fmt.Sprintf("%s||%s||%s", b.CompanyID, b.ItemID, b.InventDim.InventDimID)]
		b.Qty -= balSub.Qty
		b.QtyReserved -= balSub.QtyReserved
		b.QtyPlanned -= balSub.QtyPlanned
		b.AmountFinancial -= balSub.AmountFinancial
		b.Calc()
	}

	// group balances and calculate qty
	addGrouper := func(grouper []string, checked bool, value string) []string {
		if checked {
			return append(grouper, value)
		}
		return grouper
	}

	groupBalances := lo.GroupBy(balances, func(b *scmmodel.ItemBalance) string {
		groupers := []string{b.ItemID}
		if p.Check != nil {
			addGrouper(groupers, p.Check.IsEnabledSpecVariant, b.InventDim.VariantID)
			addGrouper(groupers, p.Check.IsEnabledSpecSize, b.InventDim.Size)
			addGrouper(groupers, p.Check.IsEnabledSpecGrade, b.InventDim.Grade)
			addGrouper(groupers, p.Check.IsEnabledItemBatch, b.InventDim.BatchID)
			addGrouper(groupers, p.Check.IsEnabledItemSerial, b.InventDim.SerialNumber)
			addGrouper(groupers, p.Check.IsEnabledLocationWarehouse, b.InventDim.WarehouseID)
			addGrouper(groupers, p.Check.IsEnabledLocationSection, b.InventDim.SectionID)
			addGrouper(groupers, p.Check.IsEnabledLocationAisle, b.InventDim.AisleID)
			addGrouper(groupers, p.Check.IsEnabledLocationBox, b.InventDim.BoxID)
		}
		return strings.Join(groupers, "|")
	})

	balsItemM := map[string]scmmodel.ItemBalance{}
	for _, bals := range groupBalances {
		nb := new(scmmodel.ItemBalance)
		nb.CompanyID = bals[0].CompanyID
		nb.ItemID = bals[0].ItemID
		nb.InventDim = bals[0].InventDim
		for _, bal := range bals {
			nb.Qty += bal.Qty
			nb.QtyReserved += bal.QtyReserved
			nb.QtyPlanned += bal.QtyPlanned
			nb.AmountFinancial += bal.AmountFinancial
		}
		nb.Calc()
		balsItemM[nb.ItemID] = *nb
	}

	// map balanace sum to each item
	itemGroupORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemGroup))
	datas := lo.Map(items, func(item tenantcoremodel.Item, index int) codekit.M {
		group, _ := itemGroupORM.Get(item.ItemGroupID)
		item.ItemGroupID = group.Name
		
		m, _ := codekit.ToM(item)
		b := balsItemM[item.ID]
		m.Set("Qty", b.Qty)
		m.Set("QtyPlanned", b.QtyPlanned)
		m.Set("QtyReserved", b.QtyReserved)
		m.Set("QtyAvail", b.QtyAvail)
		m.Set("AmountFinancial", b.AmountFinancial)
		m.Set("InventDim", b.InventDim)
		return m
	})

	// res := lo.MapToSlice(groupBalances, func(_ string, bals []*scmmodel.ItemBalance) *scmmodel.ItemBalance {
	// 	nb := new(scmmodel.ItemBalance)
	// 	nb.CompanyID = bals[0].CompanyID
	// 	nb.ItemID = bals[0].ItemID
	// 	nb.InventDim = scmmodel.InventDimension{}
	// 	for _, bal := range bals {
	// 		nb.Qty += bal.Qty
	// 		nb.QtyReserved += bal.QtyReserved
	// 		nb.QtyPlanned += bal.QtyPlanned
	// 	}
	// 	nb.Calc()
	// 	return nb
	// })

	res := codekit.M{
		"data":  datas,
		"count": 0,
		"total": total,
	}

	return res, nil
}
