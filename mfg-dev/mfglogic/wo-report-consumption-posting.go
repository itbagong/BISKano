package mfglogic

import (
	"fmt"
	"io"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type woReportConsumptionPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *mfgmodel.WorkOrderPlanReportConsumption
	trxType string
	jt      *mfgmodel.WorkOrderJournalType

	additionalInventTrxs []*scmmodel.InventTrx // biar dari calculate bisa dipake di approved()
	items                *sebar.MapRecord[*tenantcoremodel.Item]

	wop            *mfgmodel.WorkOrderPlan
	worep          *mfgmodel.WorkOrderPlanReport
	summaryM       map[string]mfgmodel.WorkOrderSummaryMaterial
	requestedItems []ItemRequestLineParam // for Item Request creation
}

func NewWorkOrderPlanReportConsumptionPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*mfgmodel.WorkOrderPlanReportConsumption, mfgmodel.WorkOrderMaterialItem] {
	p := new(woReportConsumptionPosting)
	p.ctx = ctx
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*mfgmodel.WorkOrderPlanReportConsumption, mfgmodel.WorkOrderMaterialItem](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *woReportConsumptionPosting) Header() (*mfgmodel.WorkOrderPlanReportConsumption, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlanReportConsumption), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = lo.Ternary(string(jt.TrxType) != "", string(jt.TrxType), string(mfgmodel.JournalWorkOrderPlan))

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), jt.PostingProfileIDConsumption)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", jt.PostingProfileID)
	}

	wop, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlan), j.WorkOrderPlanID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: work order plan: %s", j.WorkOrderPlanID)
	}

	worep, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlanReport), j.WorkOrderPlanReportID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: work order plan: %s", j.WorkOrderPlanID)
	}

	sums := []mfgmodel.WorkOrderSummaryMaterial{}
	p.opt.Db.GetsByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", j.WorkOrderPlanID), &sums)
	sumM := lo.SliceToMap(sums, func(d mfgmodel.WorkOrderSummaryMaterial) (string, mfgmodel.WorkOrderSummaryMaterial) {
		return fmt.Sprintf("%s||%s", d.ItemID, d.SKU), d
	})

	p.header = j
	p.wop = wop
	p.worep = worep
	p.summaryM = sumM
	p.jt = jt
	return j, pp, nil
}

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *woReportConsumptionPosting) Lines() ([]mfgmodel.WorkOrderMaterialItem, error) {
	// filter hanya plan line items yang ada qty nya, karena sekarang semua line dari Plan otomatis di munculkan
	p.header.Lines = lo.FilterMap(p.header.Lines, func(line mfgmodel.WorkOrderMaterialItem, i int) (mfgmodel.WorkOrderMaterialItem, bool) {
		if sum, exist := p.summaryM[fmt.Sprintf("%s||%s", line.ItemID, line.SKU)]; exist {
			line.LineNo = sum.LineNo
		}
		line.Qty = lo.Ternary(line.Qty > 0, (-1 * line.Qty), line.Qty) // di minuskan biar pas ketika di split waktu Post
		return line, line.Qty != 0
	})

	p.header.AdditionalLines = lo.Map(p.header.AdditionalLines, func(line mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
		line.LineNo = (i + 1)
		return line
	})

	return append(p.header.Lines, p.header.AdditionalLines...), nil
}

func (p *woReportConsumptionPosting) Calculate(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportConsumption, _ []mfgmodel.WorkOrderMaterialItem) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var err error
	preview := tenantcoremodel.PreviewReport{}

	/*
		Line Plan dan Line Additional prosesnya beda:
		- Plan:
			- [V] waktu Submit & kapanpun -> CEK qty sesuai dengan yang sudah di Reserved di Material Summary
			- [V] isi LineNo yang sesuai dari Summary di Lines()
			- [V] split according Qty inputted in Line Daily (karena sudah di Reserve waktu di WO Plan)
			- [V] waktu Post -> Confirmed
		- Additional:
			- [V] waktu Approve -> Reserve
			- [V] waktu POST -> CEK barang ada di ItemBalance.Qty (kalo ga cukup ga bisa POST, dan harus manual bikin IR dulu)
			- [V] langsung dipakai (Confirmed) tanpa split
	*/

	if e := p.validate(); e != nil {
		return nil, nil, 0, e
	}

	// Additional Item
	for _, line := range p.header.AdditionalLines {
		bal := new(scmmodel.ItemBalance)
		bals := GetAvailableStocks(p.opt.Db, GetAvailableStocksParam{
			CompanyID: p.opt.CompanyID,
			InventDim: line.InventDim,
			Items:     []GetAvailableStocksParamItem{{ItemID: line.ItemID, SKU: line.SKU}},
		})
		if len(bals) > 0 {
			bal = bals[0]
		}

		// convert line Required Qty ke default
		it, _ := p.items.Get(line.ItemID)
		requiredDefUnit, e := scmlogic.ConvertUnit(p.opt.Db, line.Qty, line.UnitID, it.DefaultUnitID)
		if e != nil {
			return nil, nil, 0, e
		}

		// compare dengan item balance
		if requiredDefUnit > bal.Qty {
			// butuh request item sejumlah kurangnya
			requestQty := requiredDefUnit - bal.Qty
			inventDim := line.InventDim.Calc()
			p.requestedItems = append(p.requestedItems, ItemRequestLineParam{
				ItemID:       line.ItemID,
				SKU:          line.SKU,
				QtyRequested: requestQty,
				UoM:          it.DefaultUnitID, // request dengan default unit
				InventDimTo:  *inventDim,
				Dimension:    p.wop.Dimension,
				RequestedBy:  line.RequestedBy,
			})
		}

		requiredTrxUnit, e := scmlogic.ConvertUnit(p.opt.Db, requiredDefUnit, it.DefaultUnitID, line.UnitID)
		if e != nil {
			return nil, nil, 0, e
		}

		//calculate cost per fix QTY
		costPerUnit := 0.0
		unitCost := it.CostUnit
		if unitCost == 0 {
			costPerUnit = scmlogic.GetCostPerUnit(p.opt.Db, *it, line.InventDim, &p.header.TrxDate)
			unitCost = costPerUnit * requiredTrxUnit
		} else {
			costPerUnit = unitCost / requiredTrxUnit
		}

		inventTrx := new(scmmodel.InventTrx)
		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.TrxDate = p.header.TrxDate

		inventTrx.Item = *it
		inventTrx.SKU = line.SKU
		inventTrx.Qty = requiredDefUnit
		inventTrx.TrxQty = requiredTrxUnit
		inventTrx.TrxUnitID = line.UnitID
		inventTrx.Status = lo.Ternary(p.opt.Op == ficologic.PostOpApprove, scmmodel.ItemReserved, scmmodel.ItemConfirmed)
		// inventTrx.AmountPhysical = line.CostPerUnit * inventTrx.Qty // TODO: fix AmountPhysical jangan biarkan jadi NaN di ItemBalance

		inventTrx.SourceType = scmmodel.ModulePurchase
		inventTrx.SourceJournalID = p.header.ID
		inventTrx.SourceLineNo = line.LineNo
		inventTrx.SourceTrxType = string(p.trxType)

		inventTrx.InventDim = *scmlogic.NewInventDimHelper(scmlogic.InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&line.InventDim)
		inventTrx.Dimension = p.header.Dimension
		inventTrx.AmountPhysical = unitCost * requiredDefUnit
		inventTrx.AmountPhysical = lo.Ternary(inventTrx.AmountPhysical > 0, inventTrx.AmountPhysical*-1, inventTrx.AmountPhysical)
		inventTrx.AmountFinancial = inventTrx.AmountPhysical

		p.additionalInventTrxs = append(p.additionalInventTrxs, inventTrx)
	}

	// yg dipake cuman additional aja untuk saving di POST
	additionalTrxs := map[string][]orm.DataModel{}
	additionalTrxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(p.additionalInventTrxs)

	return &preview, additionalTrxs, 0, err
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *woReportConsumptionPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportConsumption, lines []mfgmodel.WorkOrderMaterialItem) []ficomodel.JournalLine {
	return lo.Map(lines, func(line mfgmodel.WorkOrderMaterialItem, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID) // get account dari item
		jl.OffsetAccount = p.jt.DefaultOffsiteConsumption
		return *jl
	})
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *woReportConsumptionPosting) Post(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportConsumption, lines []mfgmodel.WorkOrderMaterialItem, additionalTrxsM map[string][]orm.DataModel) (string, error) {
	var res string

	if e := p.validateOnPost(); e != nil {
		return res, e
	}

	e := sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		var err error
		var joinTrxs []*scmmodel.InventTrx
		var confirmTrxs []*scmmodel.InventTrx
		// Plan Items
		spliter := scmlogic.NewInventSplit(tx)
		for _, line := range lines {
			item, _ := p.items.Get(line.ItemID)
			qtyConv, e := scmlogic.ConvertUnit(tx, line.Qty, line.UnitID, item.DefaultUnitID)
			if e != nil {
				return e
			}

			// TODO: apakah ini sudah termasuk validasi ketika item yang di Item Request waktu Plan-Post() sudah available ato belum?
			// split sudah include nge save invent trx
			sourceSplitTrxs, destSplitTrxs, err := spliter.SetOpts(&scmlogic.InventSplitOpts{
				SplitType:       scmlogic.SplitBySource,
				CompanyID:       p.header.CompanyID,
				SourceType:      string(p.trxType),
				SourceJournalID: p.wop.ID, // pake wop id karena yang me-reserve adalah journal wop, yang ini beda journal
				SourceLineNo:    line.LineNo,
				SourceStatus:    string(scmmodel.ItemReserved),
			}).Split(qtyConv, string(scmmodel.ItemConfirmed))
			if err != nil {
				return err
			}

			joinTrxs = tenantcorelogic.CombineSlices(joinTrxs, sourceSplitTrxs, destSplitTrxs)
			confirmTrxs = append(confirmTrxs, destSplitTrxs...)
		}

		// Additional Items
		// TODO: belum ada proses validasi apakah additional yang di Item Request waktu Approved sudah available atau belum
		// clear all additional item invent trxs and save new confirmed one
		tx.DeleteByFilter(new(scmmodel.InventTrx), dbflex.Eqs(
			"CompanyID", p.header.CompanyID,
			"SourceType", scmmodel.ModulePurchase,
			"SourceTrxType", p.trxType,
			"SourceJournalID", p.header.ID),
		)
		additionalTrxs := ficologic.FromDataModels(additionalTrxsM[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
		for _, trx := range additionalTrxs {
			trx.Status = scmmodel.ItemConfirmed
			tx.Save(trx)
			confirmTrxs = append(confirmTrxs, trx)
		}

		joinTrxs = append(joinTrxs, additionalTrxs...)
		ledgers, err := p.buildLedgerTransaction(tx, confirmTrxs, header)
		if err != nil {
			return err
		}

		ledgerTrx := map[string][]orm.DataModel{}

		if len(ledgers) > 0 {
			ledgerTrx[ledgers[0].TableName()] = ficologic.ToDataModels(ledgers)
		}

		res, err = ficologic.PostModelSave(tx, header, "WorkOrderVoucherNo", ledgerTrx) // validate only
		if err != nil {
			return err
		}

		// Balance Sync All Plan + Additional Items
		balanceOpt := scmlogic.ItemBalanceOpt{
			CompanyID:       p.header.CompanyID,
			DisableGrouping: true,
			ConsiderSKU:     true,
			ItemIDs: lo.Map(joinTrxs, func(d *scmmodel.InventTrx, i int) string {
				return d.Item.ID
			}),
		}

		if _, e := scmlogic.NewItemBalanceHub(tx).Sync(nil, balanceOpt); e != nil {
			return e
		}

		// Update Summary
		for _, line := range p.header.Lines {
			if sum, exist := p.summaryM[fmt.Sprintf("%s||%s", line.ItemID, line.SKU)]; exist {
				qtyLinePositive := lo.Ternary(line.Qty < 0, (line.Qty * -1), line.Qty)

				qtyConv, err := scmlogic.ConvertUnit(tx, qtyLinePositive, line.UnitID, sum.UnitID) // convert line Qty ke unit di Summary
				if err != nil {
					return err
				}

				sum.Used = sum.Used + qtyConv
				sum.Reserved = sum.Reserved - qtyConv
				tx.Update(&sum, "Used", "Reserved")
			}
		}

		err = UpdateWOPRWhenAllChildReportPosted(tx, header.WorkOrderPlanReportID, "Consumption")
		if err != nil {
			return err
		}

		// update each unit cost in lines and additional lines
		header.Lines = lo.Map(header.Lines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
			item, _ := p.items.Get(d.ItemID)
			costPerUnit := scmlogic.GetCostPerUnit(tx, *item, d.InventDim, nil)
			d.UnitCost = costPerUnit
			return d
		})
		tx.UpdateField(header, dbflex.Eq("_id", header.ID), "Lines")

		header.AdditionalLines = lo.Map(header.AdditionalLines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
			item, _ := p.items.Get(d.ItemID)
			costPerUnit := scmlogic.GetCostPerUnit(tx, *item, d.InventDim, nil)
			d.UnitCost = costPerUnit
			return d
		})
		tx.UpdateField(header, dbflex.Eq("_id", header.ID), "AdditionalLines")

		// update status in WorkOrderPlanReport
		p.worep.WorkOrderPlanReportConsumptionStatus = string(ficomodel.JournalStatusPosted)
		tx.UpdateField(p.worep, dbflex.Eq("_id", p.worep.ID), "WorkOrderPlanReportConsumptionStatus")

		return nil
	})
	if e != nil {
		return res, e
	}

	return res, nil
}

func (p *woReportConsumptionPosting) Approved() error {
	if len(p.additionalInventTrxs) == 0 {
		return nil // bypass if no additional item is given
	}

	e := sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		// reserve additional items -> minuskan qty and TrxQty
		p.additionalInventTrxs = lo.Map(p.additionalInventTrxs, func(d *scmmodel.InventTrx, i int) *scmmodel.InventTrx {
			d.Status = scmmodel.ItemReserved
			d.Qty = lo.Ternary(d.Qty > 0, (-1 * d.Qty), d.Qty)
			d.TrxQty = lo.Ternary(d.TrxQty > 0, (-1 * d.TrxQty), d.TrxQty)
			return d
		})

		trxs := map[string][]orm.DataModel{}
		trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(p.additionalInventTrxs)

		// validate header including save trxs
		if _, e := ficologic.PostModelSave(tx, p.header, "WorkOrderVoucherNo", trxs); e != nil {
			return e
		}

		savedAdditionalTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))

		syncOpt := scmlogic.ItemBalanceOpt{
			CompanyID: p.header.CompanyID,
			// InventDim: scmmodel.InventDimension{
			// 	WarehouseID: p.wop.InventDim.WarehouseID,
			// 	AisleID:     p.wop.InventDim.AisleID,
			// 	SectionID:   p.wop.InventDim.SectionID,
			// 	BoxID:       p.wop.InventDim.BoxID,
			// },
			DisableGrouping: true,
			ConsiderSKU:     true,
			ItemIDs: lo.Map(savedAdditionalTrxs, func(d *scmmodel.InventTrx, i int) string {
				return d.Item.ID
			}),
		}
		_, e := scmlogic.NewItemBalanceHub(tx).Sync(nil, syncOpt)
		if e != nil {
			return e
		}

		//push ir hanya ketika semua approval sudah approved
		approval, err := datahub.GetByFilter(tx, new(ficomodel.PostingApproval),
			dbflex.And(
				dbflex.Eq("SourceID", p.header.ID),
				dbflex.Eq("Status", ficomodel.JournalStatusApproved),
			),
		)
		if err != nil {
			approval = new(ficomodel.PostingApproval)
		}

		pendingApprovals := lo.Filter(approval.Approvals, func(approvalItem *ficomodel.PostingProfileApprovalItem, _ int) bool {
			return strings.ToLower(approvalItem.Status) != "approved"
		})

		if len(pendingApprovals) == 0 {
			// proses auto item request additional item if any
			requestedItemPerWH := lo.GroupBy(p.requestedItems, func(d ItemRequestLineParam) string {
				// d.InventDimTo = *d.InventDimTo.Calc()
				return d.InventDimTo.WarehouseID
			})

			for _, lineItems := range requestedItemPerWH {
				// create Item Request
				_, e := ItemRequestInsert(p.ctx, &ItemRequestInsertParam{
					Name: fmt.Sprintf("FROM WO Report: %s", p.worep.ID),
					// WOReff:      p.header.ID,
					WOReff:      p.wop.ID,                  // sama mba fanny disuruh diganti biar dari IR tau ini reff ke WO yang mana
					Requestor:   lineItems[0].RequestedBy,  // TODO: sementara requestor samain dengan yang di Plan
					Department:  p.wop.RequestorDepartment, // TODO: sementara requestor samain dengan yang di Plan
					InventDimTo: lineItems[0].InventDimTo,
					Dimension:   p.wop.Dimension,
					Lines:       lineItems,
				}, p.opt.CompanyID, p.opt.UserID)
				if e != nil {
					return e
				}
			}
		}

		return nil
	})
	if e != nil {
		return e
	}

	return nil
}

func (p *woReportConsumptionPosting) Rejected() error {
	return nil
}

func (p *woReportConsumptionPosting) GetAccount() string {
	return ""
}

func (p *woReportConsumptionPosting) validate() error {
	notMatchedItemNames := []string{} // item2 yang ga ada di summary

	// validate Plan Items: ga boleh > summary yang udah di reserve
	for _, line := range p.header.Lines {
		item, _ := p.items.Get(line.ItemID)
		if sum, exist := p.summaryM[fmt.Sprintf("%s||%s", line.ItemID, line.SKU)]; exist {
			// item dan sku ada di summary material
			unitUsed := sum.UnitID
			qtyConv, e := scmlogic.ConvertUnit(p.opt.Db, line.Qty, line.UnitID, unitUsed) // convert line Qty ke unit di Summary
			if e != nil {
				return e
			}

			if qtyConv > sum.Reserved {
				// error ketika Qty melebihi sisa Reserved (yang sudah dikurangi Used di func Post diatas)
				remainingQty := sum.Reserved

				return fmt.Errorf("Item: %s quantity used exceeded | Used Qty: %v %s | Remaining Qty: %v %s", item.Name, qtyConv, unitUsed, remainingQty, unitUsed)
			}
		} else {
			// item dan sku tidak ada di summary material

			notMatchedItemNames = append(notMatchedItemNames, item.Name)
		}
	}

	if len(notMatchedItemNames) > 0 {
		return fmt.Errorf("item not in plan: %s", strings.Join(notMatchedItemNames, ", "))
	}

	return nil
}

func (p *woReportConsumptionPosting) validateOnPost() error {
	// validate Additional Items: ga boleh > ItemBalance.Qty
	for _, line := range p.header.AdditionalLines {
		bal := new(scmmodel.ItemBalance)
		bals := GetAvailableStocks(p.opt.Db, GetAvailableStocksParam{
			CompanyID: p.opt.CompanyID,
			InventDim: *scmlogic.NewInventDimHelper(scmlogic.InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&line.InventDim),
			Items:     []GetAvailableStocksParamItem{{ItemID: line.ItemID, SKU: line.SKU}},
		})
		if len(bals) > 0 {
			bal = bals[0]
		}

		item, _ := p.items.Get(line.ItemID)
		qtyConv, e := scmlogic.ConvertUnit(p.opt.Db, line.Qty, line.UnitID, item.DefaultUnitID)
		if e != nil {
			return e
		}

		if qtyConv > bal.Qty {
			return fmt.Errorf("Item: %s Quantity used exceeded | Used Qty: %v %s | Remaining Qty: %v %s", item.Name, line.Qty, line.UnitID, bal.Qty, item.DefaultUnitID)
		}
	}

	return nil
}

func (p *woReportConsumptionPosting) buildLedgerTransaction(tx *datahub.Hub, trxs []*scmmodel.InventTrx, header *mfgmodel.WorkOrderPlanReportConsumption) ([]*ficomodel.LedgerTransaction, error) {
	ledgers := sebar.NewMapRecordWithORM(tx, new(tenantcoremodel.LedgerAccount))
	itemGroups := sebar.NewMapRecordWithORM(tx, new(tenantcoremodel.ItemGroup))
	journalTypes := sebar.NewMapRecordWithORM(tx, new(mfgmodel.WorkOrderJournalType))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}
	var err error

	lo.ForEach(trxs, func(trx *scmmodel.InventTrx, index int) {
		itemGroup, _ := itemGroups.Get(trx.Item.ItemGroupID)
		ledgerAccount, err := ledgers.Get(trx.Item.LedgerAccountIDStock)
		if err != nil {
			ledgerAccount, err = ledgers.Get(itemGroup.LedgerAccountIDStock)
			if err != nil {
				err = fmt.Errorf("invalid: main inventory account for item %s: %s", trx.Item.ID, err.Error())
				return
			}
		}

		unitCost := trx.Item.CostUnit
		if unitCost == 0 {
			costPerUnit := scmlogic.GetCostPerUnit(p.opt.Db, trx.Item, trx.InventDim, &p.header.TrxDate)
			unitCost = costPerUnit * trx.Qty
		}

		totalCost := unitCost * trx.Qty
		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         trx.Dimension,
			SourceType:        trx.SourceType,
			SourceJournalID:   trx.SourceJournalID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     trx.SourceTrxType,
			SourceLineNo:      trx.SourceLineNo,
			TrxDate:           p.header.TrxDate,
			Text:              trx.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            totalCost,
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", p.header.WorkOrderPlanID).
				Set("WorkOrderElemenType", "Cost").
				Set("WorkOrderCostType", "Material").
				Set("WorkOrderOutputType", "Output"),
		}

		var offsetLedger *tenantcoremodel.LedgerAccount
		jt, _ := journalTypes.Get(p.header.JournalTypeID)
		offsetLedger, err = ledgers.Get(jt.DefaultOffsiteConsumption.AccountID)
		if err != nil {
			return
		}

		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         trx.Dimension,
			SourceType:        trx.SourceType,
			SourceJournalID:   trx.SourceJournalID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     trx.SourceTrxType,
			SourceLineNo:      trx.SourceLineNo,
			TrxDate:           p.header.TrxDate,
			Text:              trx.Text,
			Account:           *offsetLedger,
			Status:            ficomodel.AmountConfirmed,
			Amount:            (-1 * totalCost),
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", p.header.WorkOrderPlanID).
				Set("WorkOrderElemenType", "Cost").
				Set("WorkOrderCostType", "Material").
				Set("WorkOrderOutputType", "Output"),
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	})

	if err != nil {
		return nil, err
	}

	return ledgerTrxs, nil
}

func (p *woReportConsumptionPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
	employee := new(tenantcoremodel.Employee)
	err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("_id", p.opt.UserID),
	))
	if err != nil && err != io.EOF {
		return fmt.Errorf("error when get email employee : %s", err.Error())
	}

	for _, app := range pa.Approvals {
		if app.Line == pa.CurrentStage {
			notification := ficomodel.Notification{
				UserSubmitter:            p.opt.UserID,
				UserSubmitterEmail:       employee.Email,
				JournalID:                p.header.ID,
				JournalType:              p.header.JournalTypeID,
				PostingProfileApprovalID: pa.ID,
				TrxDate:                  p.header.TrxDate,
				Text:                     p.header.Text,
				UserTo:                   app.UserID,
				TrxType:                  string(p.header.TrxType),
				Menu:                     string(p.header.TrxType),
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
			}

			employee := new(tenantcoremodel.Employee)
			err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
				dbflex.Eq("_id", app.UserID),
			))
			if err != nil && err != io.EOF {
				return fmt.Errorf("error when get employee : %s", err.Error())
			}

			notification.UserToEmail = employee.Email

			err = p.opt.Db.Save(&notification)
			if err != nil {
				return fmt.Errorf("error when save notification : %s", err.Error())
			}
		}
	}

	return nil
}

func (p *woReportConsumptionPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
	// get latest notification
	latestNotif := new(ficomodel.Notification)
	err := p.opt.Db.GetByParm(latestNotif, dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("JournalID", p.header.ID),
			dbflex.Eq("UserTo", p.opt.UserID),
		),
	).SetSort("-Created"))
	if err != nil {
		return fmt.Errorf("error when get notification : %s", err.Error())
	}

	if op == ficologic.PostOpApprove {
		latestNotif.Status = string(ficomodel.JournalStatusApproved)
	} else {
		latestNotif.Status = string(ficomodel.JournalStatusRejected)
	}

	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification : %s", err.Error())
	}

	// create notification for submitter user
	latestNotif.ID = primitive.NewObjectID().Hex()
	latestNotif.UserTo = latestNotif.UserSubmitter
	latestNotif.UserToEmail = latestNotif.UserSubmitterEmail
	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification for submitter user : %s", err.Error())
	}

	// get user approval stage
	userApprovals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.UserID == p.opt.UserID
	})

	approvals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.Line == userApprovals[0].Line && a.Status != "PENDING"
	})

	// check if need to send notification
	if len(approvals) >= pa.Approvers[userApprovals[0].Line-1].MinimalApproverCount &&
		userApprovals[0].Line != pa.CurrentStage {
		for _, app := range pa.Approvals {
			if app.Line == pa.CurrentStage {
				latestNotif.ID = primitive.NewObjectID().Hex()
				latestNotif.UserTo = app.UserID
				latestNotif.Status = app.Status

				employee := new(tenantcoremodel.Employee)
				err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
					dbflex.Eq("_id", app.UserID),
				))
				if err != nil && err != io.EOF {
					return fmt.Errorf("error when get employee : %s", err.Error())
				}

				latestNotif.UserToEmail = employee.Email

				err = p.opt.Db.Save(latestNotif)
				if err != nil {
					return fmt.Errorf("error when save notification for next approval : %s", err.Error())
				}
			}
		}
	}

	return nil
}
