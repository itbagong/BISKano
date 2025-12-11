package scmlogic

import (
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/suim"
	"github.com/samber/lo"
)

type byItemDimsType struct {
	ID struct {
		ItemID                string `json:"Item__id" bson:"Item__id"`
		InventDim_InventDimID string
		Status                scmmodel.ItemBalanceStatus
	} `json:"_id" bson:"_id"`
	InventDim        scmmodel.InventDimension
	Qty              float64
	AmountPhysical   float64
	AmountFinancial  float64
	AmountAdjustment float64
}

type InventBalanceCalcOpts struct {
	CompanyID   string `form_required:"1"`
	ItemID      []string
	InventDim   scmmodel.InventDimension
	BalanceDate *time.Time
}

func (opts *InventBalanceCalcOpts) Wheres() []*dbflex.Filter {
	wheres := []*dbflex.Filter{dbflex.Eq("CompanyID", opts.CompanyID)}
	if opts.InventDim.InventDimID != "" {
		wheres = append(wheres, dbflex.In("InventDim.InventDimID", opts.InventDim.InventDimID))
	} else if opts.InventDim.SpecID != "" {
		wheres = append(wheres, dbflex.In("InventDim.SpecID", opts.InventDim.SpecID))
	} else {
		if opts.InventDim.WarehouseID != "" {
			wheres = append(wheres, dbflex.In("InventDim.WarehouseID", opts.InventDim.WarehouseID))
		}
		if opts.InventDim.VariantID != "" {
			wheres = append(wheres, dbflex.Eq("InventDim.VariantID", opts.InventDim.VariantID))
		}
		if opts.InventDim.Grade != "" {
			wheres = append(wheres, dbflex.Eq("InventDim.Grade", opts.InventDim.Grade))
		}
	}
	return wheres
}

type inventBalanceCalc struct {
	db *datahub.Hub
}

// func NewInventBalanceCalc(db *datahub.Hub) *InventBalanceCalc {
func NewInventBalanceCalc(db *datahub.Hub) *inventBalanceCalc {
	b := new(inventBalanceCalc)
	b.db = db
	return b
}

func (c *inventBalanceCalc) Update(balance *scmmodel.ItemBalance) (*scmmodel.ItemBalance, error) {
	if err := suim.Validate(balance); err != nil {
		return nil, err
	}

	f := dbflex.Eqs("CompanyID", balance.CompanyID,
		"ItemID", balance.ItemID,
		"BalanceDate", balance.BalanceDate,
		"InventDim.InventDimID", balance.InventDim.InventDimID)
	bal, err := datahub.GetByFilter(c.db, new(scmmodel.ItemBalance), f)
	if err != nil {
		if err != io.EOF {
			return nil, err
		}

		bal = new(scmmodel.ItemBalance)
		*bal = *balance
	} else {
		inventBalanceUpdate(bal, balance)
	}
	bal.Calc()
	if err = c.db.Save(bal); err != nil {
		return nil, err
	}
	return bal, nil
}

func inventBalanceUpdate(dest *scmmodel.ItemBalance, updater *scmmodel.ItemBalance) {
	dest.Qty += updater.Qty
	dest.QtyReserved += updater.QtyReserved
	dest.QtyPlanned += updater.QtyPlanned
	dest.AmountPhysical += updater.AmountPhysical
	dest.AmountFinancial += updater.AmountFinancial
	dest.AmountAdjustment += updater.AmountAdjustment
}

func (c *inventBalanceCalc) Get(opts *InventBalanceCalcOpts) ([]*scmmodel.ItemBalance, error) {
	if opts == nil {
		return nil, errors.New("options is mandatory")
	}
	if err := suim.Validate(opts); err != nil {
		return nil, err
	}

	var (
		snapshotDate *time.Time
	)

	//TODO:
	/*
		cari snapshotdate yg lebih tinggi daripada tanggal, kalau tidak ada gunakan nil
		kurangi value dgn transkasi yang lebih kecil drpd trxdate
	*/

	// get snapshot
	commonWheres := opts.Wheres()
	snapshotBal, _ := datahub.GetByParm(c.db, new(scmmodel.ItemBalance), dbflex.NewQueryParam().
		SetWhere(dbflex.And(dbflex.Eq("CompanyID", opts.CompanyID), dbflex.Gte("BalanceDate", opts.BalanceDate))).
		SetSort("BalanceDate").
		SetSelect("BalanceDate"))
	snapshotDate = snapshotBal.BalanceDate

	// ambil balance terakhir
	balWheres := append(commonWheres, dbflex.Eq("BalanceDate", snapshotDate))
	if len(opts.ItemID) > 0 {
		balWheres = append(balWheres, dbflex.In("ItemID", opts.ItemID...))
	}
	bals, err := datahub.FindByFilter(c.db, new(scmmodel.ItemBalance), dbflex.And(balWheres...))
	if err != nil {
		return nil, err
	}

	// ambil trx yang lebih > balanceDate yg bersesuaian dengan filter di opts
	trxWheres := commonWheres
	if opts.BalanceDate != nil {
		trxWheres = append(trxWheres, dbflex.Gt("TrxDate", opts.BalanceDate))
	}
	if len(opts.ItemID) > 0 {
		trxWheres = append(trxWheres, dbflex.In("Item._id", opts.ItemID...))
	}

	byItemDims := []byItemDimsType{}
	err = c.db.PopulateByParm(new(scmmodel.InventTrx).TableName(), dbflex.NewQueryParam().SetWhere(dbflex.And(trxWheres...)).
		SetGroupBy("CompanyID", "ItemID", "InventDim.InventDimID", "Status").
		SetAggr(
			dbflex.NewAggrItem("Qty", dbflex.AggrSum, "Qty"),
			dbflex.NewAggrItem("AmountPhysical", dbflex.AggrSum, "AmountPhysical"),
			dbflex.NewAggrItem("AmountFinancial", dbflex.AggrSum, "AmountFinancial"),
			dbflex.NewAggrItem("AmountAdjustment", dbflex.AggrSum, "AmountAdjustment"),
		), &byItemDims)
	if err != nil {
		return nil, fmt.Errorf("invalid: access inventory transactions: %s", err.Error())
	}

	for _, dim := range byItemDims {
		dimBals := lo.Filter(bals, func(b *scmmodel.ItemBalance, i int) bool {
			return b.ItemID == dim.ID.ItemID && b.InventDim.InventDimID == dim.ID.InventDim_InventDimID
		})
		if len(dimBals) > 0 {
			dimBal := dimBals[0]
			switch dim.ID.Status {
			case scmmodel.ItemConfirmed:
				dimBal.Qty += dim.Qty
			case scmmodel.ItemPlanned:
				dimBal.QtyPlanned += dim.Qty

			case scmmodel.ItemReserved:
				dimBal.QtyReserved += dim.Qty
			}
			dimBal.AmountPhysical += dim.AmountPhysical
			dimBal.AmountFinancial += dim.AmountFinancial
			dimBal.AmountAdjustment += dim.AmountAdjustment
		}
	}

	for _, bal := range bals {
		bal.Calc()
		if err = c.db.Save(bal); err != nil {
			return nil, fmt.Errorf("invalid: save balance: %s", err.Error())
		}
	}

	return bals, nil
}

func (c *inventBalanceCalc) Sync(sources []*scmmodel.InventTrx) ([]*scmmodel.ItemBalance, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("missing: invent balance calc sources")
	}

	groupedSources := lo.GroupBy(sources, func(s *scmmodel.InventTrx) string {
		return fmt.Sprintf("%s|%s", s.Item.ID, s.InventDim.InventDimID)
	})

	var snapshotDate *time.Time
	snapshot, err := datahub.GetByParm(c.db, new(scmmodel.ItemBalance), dbflex.NewQueryParam().
		SetSort("-BalanceDate").
		SetWhere(dbflex.Eq("CompanyID", sources[0].CompanyID)))
	if err == nil {
		snapshotDate = snapshot.BalanceDate
	}

	// if snapshot is not nil, validate no trx <= snapshotdate
	if snapshotDate != nil {
		lessTrxs := lo.Filter(sources, func(s *scmmodel.InventTrx, i int) bool {
			return !s.TrxDate.After(*snapshotDate)
		})
		if len(lessTrxs) > 0 {
			return nil, fmt.Errorf("invalid: trx less than last snapshot date: %v", snapshotDate)
		}
	}

	bals := []*scmmodel.ItemBalance{}
	for _, srcs := range groupedSources {
		// get existing nil balance
		balWhere := dbflex.Eqs("CompanyID", srcs[0].CompanyID, "ItemID", srcs[0].Item.ID, "InventDim.InventDimID",
			srcs[0].InventDim.InventDimID, "BalanceDate", nil)
		bal, err := datahub.GetByFilter(c.db, new(scmmodel.ItemBalance), balWhere)
		if err != nil {
			bal = new(scmmodel.ItemBalance)
			bal.CompanyID = srcs[0].CompanyID
			bal.ItemID = srcs[0].Item.ID
			bal.InventDim = srcs[0].InventDim
			bal.SKU = srcs[0].SKU
		}

		// get last snapshot balance, if exist set qty
		if snapshotDate == nil {
			bal.Qty = 0
			bal.QtyPlanned = 0
			bal.QtyReserved = 0
			bal.AmountFinancial = 0
			bal.AmountPhysical = 0
			bal.AmountAdjustment = 0
		} else {
			lastBalWhere := dbflex.Eqs("CompanyID", srcs[0].CompanyID, "ItemID", srcs[0].Item.ID, "InventDim.InventDimID",
				srcs[0].InventDim.InventDimID, "BalanceDate", snapshotDate)
			lastBal, err := datahub.GetByFilter(c.db, new(scmmodel.ItemBalance), lastBalWhere)
			if err == nil {
				bal.Qty = lastBal.Qty
				bal.QtyPlanned = lastBal.QtyPlanned
				bal.QtyReserved = lastBal.QtyReserved
				bal.AmountFinancial = lastBal.AmountFinancial
				bal.AmountPhysical = lastBal.AmountPhysical
				bal.AmountAdjustment = lastBal.AmountAdjustment
			}
		}

		// get trxs
		trxWheres := []*dbflex.Filter{
			dbflex.Eqs("CompanyID", srcs[0].CompanyID, "Item._id", srcs[0].Item.ID, "InventDim.InventDimID", srcs[0].InventDim.InventDimID)}
		if snapshotDate != nil {
			trxWheres = append(trxWheres, dbflex.Gt("TrxDate", snapshotDate))
		}
		trxs, err := datahub.FindAnyByParm(c.db, new(byItemDimsType), new(scmmodel.InventTrx).TableName(),
			dbflex.NewQueryParam().
				SetWhere(dbflex.And(trxWheres...)).
				SetGroupBy("Status").
				SetAggr(
					dbflex.NewAggrItem("Qty", dbflex.AggrSum, "Qty"),
					dbflex.NewAggrItem("AmountPhysical", dbflex.AggrSum, "AmountPhysical"),
					dbflex.NewAggrItem("AmountFinancial", dbflex.AggrSum, "AmountFinancial"),
					dbflex.NewAggrItem("AmountAdjustment", dbflex.AggrSum, "AmountAdjustment")))
		if err != nil {
			return nil, fmt.Errorf("invalid: get invent transaction: %s", err.Error())
		}

		for _, src := range trxs {
			switch src.ID.Status {
			case scmmodel.ItemConfirmed:
				bal.Qty += src.Qty
			case scmmodel.ItemPlanned:
				bal.QtyPlanned += src.Qty

			case scmmodel.ItemReserved:
				bal.QtyReserved += src.Qty
			}

			if math.IsNaN(src.AmountPhysical) {
				src.AmountPhysical = 0
			}

			if math.IsNaN(src.AmountFinancial) {
				src.AmountFinancial = 0
			}

			if math.IsNaN(src.AmountAdjustment) {
				src.AmountAdjustment = 0
			}

			bal.AmountPhysical += src.AmountPhysical
			bal.AmountFinancial += src.AmountFinancial
			bal.AmountAdjustment += src.AmountAdjustment
		}
		bal.Calc()
		bals = append(bals, bal)
	}

	for _, bal := range bals {
		if err := c.db.Save(bal); err != nil {
			return nil, err
		}
	}

	return bals, nil
}

func (c *inventBalanceCalc) MakeSnapshot(companyID string, balDate *time.Time) error {
	var (
		snapshotDate *time.Time
		balances     []*scmmodel.ItemBalance
	)

	snapshot, _ := datahub.GetByParm(c.db, new(scmmodel.ItemBalance), dbflex.NewQueryParam().
		SetSort("-BalanceDate").
		SetWhere(dbflex.Eq("CompanyID", companyID)))
	if snapshot.BalanceDate != nil {
		if balDate != nil && !snapshot.BalanceDate.Before(*balDate) {
			return fmt.Errorf("invalid:last snapshot date:%v", snapshot.BalanceDate)
		}
		snapshotDate = snapshot.BalanceDate
		balances, _ = datahub.FindByFilter(c.db, new(scmmodel.ItemBalance), dbflex.Eqs("CompanyID", companyID, "BalanceDate", snapshotDate))
	}

	wheres := []*dbflex.Filter{dbflex.Eq("CompanyID", companyID)}
	if snapshotDate != nil {
		wheres = append(wheres, dbflex.Gt("TrxDate", snapshotDate))
	}
	if balDate != nil {
		wheres = append(wheres, dbflex.Lte("TrxDate", balDate))
	}
	byItemDims := []*byItemDimsType{}

	parm := dbflex.NewQueryParam().
		SetWhere(dbflex.And(wheres...)).
		SetGroupBy("Item._id", "InventDim.InventDimID", "Status").
		SetAggr(
			dbflex.NewAggrItem("InventDim", dbflex.AggrFirst, "InventDim"),
			dbflex.NewAggrItem("Qty", dbflex.AggrSum, "Qty"),
			dbflex.NewAggrItem("AmountPhysical", dbflex.AggrSum, "AmountPhysical"),
			dbflex.NewAggrItem("AmountFinancial", dbflex.AggrSum, "AmountFinancial"),
			dbflex.NewAggrItem("AmountAdjustment", dbflex.AggrSum, "AmountAdjustment"),
		)
	if err := c.db.PopulateByParm(new(scmmodel.InventTrx).TableName(), parm, &byItemDims); err != nil {
		return fmt.Errorf("invalid:get balance summary from trx:%s", err.Error())
	}

	for _, bid := range byItemDims {
		var bal *scmmodel.ItemBalance
		bals := lo.Filter(balances, func(item *scmmodel.ItemBalance, index int) bool {
			return item.ItemID == bid.ID.ItemID && item.InventDim.InventDimID == bid.InventDim.InventDimID
		})
		if len(bals) == 0 {
			bal = new(scmmodel.ItemBalance)
			bal.CompanyID = companyID
			bal.BalanceDate = balDate
			bal.ItemID = bid.ID.ItemID
			bal.InventDim = bid.InventDim
			balances = append(balances, bal)
		} else {
			bal = bals[0]
		}
		switch bid.ID.Status {
		case scmmodel.ItemPlanned:
			bal.QtyPlanned += bid.Qty
		case scmmodel.ItemReserved:
			bal.QtyReserved += bid.Qty
		case scmmodel.ItemConfirmed:
			bal.Qty += bid.Qty
		}

		if math.IsNaN(bid.AmountPhysical) {
			bid.AmountPhysical = 0
		}

		if math.IsNaN(bid.AmountFinancial) {
			bid.AmountFinancial = 0
		}

		if math.IsNaN(bid.AmountAdjustment) {
			bid.AmountAdjustment = 0
		}

		bal.AmountPhysical += bid.AmountPhysical
		bal.AmountFinancial += bid.AmountFinancial
		bal.AmountAdjustment += bid.AmountAdjustment
		bal.Calc()
	}

	c.db.DeleteByFilter(new(scmmodel.ItemBalance), dbflex.Eqs("CompanyID", companyID, "BalanceDate", balDate))
	for _, bal := range balances {
		bal.ID = ""
		bal.BalanceDate = balDate
		c.db.Save(bal)
	}

	return nil
}

func GetItemBalanceByTrxSum__(db *datahub.Hub, companyID string, id string, dim scmmodel.InventDimension, dateFrom time.Time, dateTo *time.Time) *scmmodel.ItemBalance {
	bal := new(scmmodel.ItemBalance)
	filters := []*dbflex.Filter{
		dbflex.Eqs("CompanyID", companyID, "ItemID", id),
		dbflex.Gt("TrxDate", dateFrom),
		dbflex.Eq("InventDim.InventDimensionID", dim.InventDimID),
	}
	if dateTo != nil {
		filters = append(filters, dbflex.Lte("TrxDate", dateTo))
	}
	trxs, _ := datahub.FindAnyByParm(db, new(scmmodel.InventTrx), new(scmmodel.InventTrx).TableName(),
		dbflex.NewQueryParam().
			SetWhere(dbflex.And(filters...)).
			SetAggr(
				dbflex.NewAggrItem("Qty", dbflex.AggrSum, "Qty"),
				dbflex.NewAggrItem("QtyReserved", dbflex.AggrSum, "QtyReserved"),
				dbflex.NewAggrItem("QtyPlanned", dbflex.AggrSum, "QtyPlanned"),
			).
			SetGroupBy("Status"))
	for _, trx := range trxs {
		switch trx.Status {
		case scmmodel.ItemConfirmed:
			bal.Qty -= trx.Qty

		case scmmodel.ItemReserved:
			bal.QtyReserved -= trx.Qty

		case scmmodel.ItemPlanned:
			bal.QtyPlanned -= trx.Qty
		}
	}
	bal.Calc()
	return bal
}
