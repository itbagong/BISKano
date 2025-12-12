package scmlogic

import (
	"fmt"
	"io"
	"math"
	"strconv"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventoryReceivePosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.InventReceiveIssueJournal
	lines   []scmmodel.InventReceiveIssueLine
	trxType string
	jt      *scmmodel.InventJournalType
}

func NewInventoryReceivePosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.InventReceiveIssueJournal, scmmodel.InventReceiveIssueLine] {
	receive := new(InventoryReceivePosting)
	receive.ctx = ctx
	receive.opt = &opt

	pvd := ficologic.PostingProvider[*scmmodel.InventReceiveIssueJournal, scmmodel.InventReceiveIssueLine](receive)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *InventoryReceivePosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) []ficomodel.JournalLine {

	return lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))

		jl.Text = line.Item.Name
		jl.Qty = line.Qty
		jl.UnitID = line.UnitID
		jl.PriceEach = line.UnitCost
		jl.Amount = jl.Qty * jl.PriceEach
		return *jl
	})
}

func (p *InventoryReceivePosting) Header() (*scmmodel.InventReceiveIssueJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.InventReceiveIssueJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(scmmodel.InventJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.jt = jt

	p.trxType = string(jt.TransactionType)
	if p.trxType == "" {
		return nil, nil, fmt.Errorf("missing: journal type - transaction type: %s", j.JournalTypeID)
	}

	j.PostingProfileID = jt.PostingProfileID
	if j.PostingProfileID == "" {
		return nil, nil, fmt.Errorf("missing: posting profile in journal type")
	}

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	p.header = j
	return j, pp, nil
}

func (p *InventoryReceivePosting) Lines() ([]scmmodel.InventReceiveIssueLine, error) {
	lines := lo.Map(p.header.Lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
		if line.Text == "" {
			line.Text = p.header.Text
		}

		return line
	})

	for index, line := range lines {
		unitRatio, err := ConvertUnit(p.opt.Db, 1, line.UnitID, line.Item.DefaultUnitID)
		if err != nil {
			return nil, fmt.Errorf("invalid: unit for item %s: %s", line.Item.ID, err.Error())
		}
		line.InventQty = unitRatio * line.Qty

		if line.SourceTrxType == string(scmmodel.InventReceive) {
			// TODO: kalau sumbernya dari transfer, ambil cost per unit dari GI sebelumnya

			// Validasi kalo GR ini belum ada yang di GI kan, ga boleh di GR dulu
			gis, _ := datahub.Find(p.opt.Db, new(scmmodel.InventTrx),
				dbflex.NewQueryParam().
					SetWhere(
						dbflex.Eqs("CompanyID", p.header.CompanyID, "SourceType", line.SourceType, "SourceJournalID", line.SourceJournalID, "SourceLineNo", line.SourceLineNo, "SourceTrxType", scmmodel.InventIssuance, "Status", scmmodel.ItemConfirmed),
					).
					SetSort("-_id"),
			)
			if len(gis) == 0 {
				return nil, fmt.Errorf("missing: GI for %s, %s", line.Item.ID, line.SKU)
			}

			// Validasi GR tidak hanya berdasarkan sudah pernah GI tetapi juga memperhitungkan Qty yang sudah di GI
			totalIssued := lo.SumBy(gis, func(d *scmmodel.InventTrx) float64 {
				return d.TrxQty
			})

			// Note: tidak perlu lagi mengambil total received dari invent trx karena dari frontend sudah ditambahkan
			// line.SettledQty = totalReceived + Qty yang akan di GR
			// grecs, _ := datahub.Find(p.opt.Db, new(scmmodel.InventTrx),
			// 	dbflex.NewQueryParam().
			// 		SetWhere(
			// 			dbflex.Eqs("CompanyID", p.header.CompanyID, "SourceType", line.SourceType, "SourceJournalID", line.SourceJournalID, "SourceLineNo", line.LineNo, "SourceTrxType", scmmodel.InventReceive, "Status", scmmodel.ItemConfirmed),
			// 		).
			// 		SetSort("-_id"),
			// )
			// totalReceived := lo.SumBy(grecs, func(d *scmmodel.InventTrx) float64 {
			// 	return d.TrxQty
			// })

			if line.SettledQty > math.Abs(totalIssued) {
				return nil, fmt.Errorf("item: %s, exceed issued qty. Issued: %v. Received+Qty: %v", line.Item.Name, totalIssued, line.SettledQty)
			}

			line.CostPerUnit = lo.Ternary(gis[0].AmountFinancial == 0, gis[0].AmountPhysical, gis[0].AmountFinancial+gis[0].AmountAdjustment) / gis[0].Qty
			line.UnitCost = line.CostPerUnit * unitRatio
		} else {
			if line.UnitCost == 0 {
				line.CostPerUnit = GetCostPerUnit(p.opt.Db, line.Item, line.InventDim, &p.header.TrxDate)
				line.UnitCost = line.CostPerUnit * unitRatio
			} else {
				line.CostPerUnit = line.UnitCost / unitRatio
			}
		}
		lines[index] = line
	}
	p.lines = lines

	return lines, nil
}

func (p *InventoryReceivePosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}
	inventTrxs := []*scmmodel.InventTrx{}
	// ledgers := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LedgerAccount))
	// itemGroups := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemGroup))

	itemGroupIDs := make([]interface{}, len(lines))

	lo.ForEach(lines, func(line scmmodel.InventReceiveIssueLine, index int) {
		itemGroupIDs[index] = line.Item.ItemGroupID
	})

	//get item groups
	itemGroups, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.ItemGroup), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", itemGroupIDs...)))
	ledgerIDs := make([]interface{}, len(lines)+len(itemGroups)+1)

	i := 0
	lo.ForEach(lines, func(line scmmodel.InventReceiveIssueLine, _ int) {
		ledgerIDs[i] = line.Item.LedgerAccountIDStock
		i++
	})

	itemGroupsKV := map[string]*tenantcoremodel.ItemGroup{}
	lo.ForEach(itemGroups, func(item *tenantcoremodel.ItemGroup, _ int) {
		ledgerIDs[i] = item.LedgerAccountIDStock
		itemGroupsKV[item.ID] = item
		i++
	})

	ledgerIDs[i] = p.jt.DefaultOffset.AccountID

	//get ledger account
	ledgers, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.LedgerAccount), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", ledgerIDs...)))
	ledgersKV := lo.Associate(ledgers, func(ledger *tenantcoremodel.LedgerAccount) (string, *tenantcoremodel.LedgerAccount) {
		return ledger.ID, ledger
	})

	ledgerTrxs := []*ficomodel.LedgerTransaction{}
	for index, line := range lines {
		inventTrx, err := receiveIssueLineToTrx(p.opt.Db, p.header, line)
		if err != nil {
			return &preview, trxs, 0, fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}

		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.Status = scmmodel.ItemPlanned
		inventTrx.TrxDate = p.header.TrxDate

		inventTrxs = append(inventTrxs, inventTrx)

		itemGroup, _ := itemGroupsKV[inventTrx.Item.ItemGroupID]
		ledgerAccount, ok := ledgersKV[inventTrx.Item.LedgerAccountIDStock]
		if !ok {
			ledgerAccount, ok = ledgersKV[itemGroup.LedgerAccountIDStock]
			if !ok {
				return nil, nil, 0, fmt.Errorf("invalid: main inventory account for item %s: %s", line.ItemID, err.Error())
			}
		}
		totalCost := line.UnitCost * line.Qty
		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        line.SourceType,
			SourceJournalID:   line.SourceJournalID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.SourceTrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            totalCost,
			References:        tenantcoremodel.References{}.Set(string(scmmodel.RefKeyGoodReceive), p.header.ID),
		}

		var offsetLedger *tenantcoremodel.LedgerAccount
		jt := p.jt

		offsetLedger, ok = ledgersKV[jt.DefaultOffset.AccountID]
		if !ok {
			return nil, nil, 0, fmt.Errorf("invalid: offset account for %s, %s: %s", line.SourceType, p.header.JournalTypeID, err.Error())
		}

		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        line.SourceType,
			SourceJournalID:   line.SourceJournalID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.SourceTrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Account:           *offsetLedger,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -totalCost,
			References:        tenantcoremodel.References{}.Set(string(scmmodel.RefKeyGoodReceive), p.header.ID),
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	}

	if e := p.validate(inventTrxs); e != nil {
		return &preview, trxs, 0, e
	}

	trxs[ledgerTrxs[0].TableName()] = ficologic.ToDataModels(ledgerTrxs)
	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(inventTrxs)
	amount := lo.SumBy(inventTrxs, func(i *scmmodel.InventTrx) float64 {
		return i.AmountPhysical
	})

	return p.GetPreview(opt, header, lines), trxs, amount, nil
}

func (p *InventoryReceivePosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) *tenantcoremodel.PreviewReport {
	pv := new(tenantcoremodel.PreviewReport)

	//checkpoint
	// itemORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Item))
	// uomORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.UoM))
	// whORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LocationWarehouse))
	// specORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemSpec))
	// poORM := sebar.NewMapRecordWithORM(opt.Db, new(scmmodel.PurchaseOrderJournal))
	// ivjORM := sebar.NewMapRecordWithORM(opt.Db, new(scmmodel.InventJournal))

	site, _ := datahub.GetByID(opt.Db, new(bagongmodel.Site), header.Dimension.Get("Site"))

	itemIDs := make([]string, len(lines))
	uomIDs := make([]string, len(lines))
	ivjIDs := []string{}
	poIDs := []string{}
	specIDs := make([]string, len(lines))

	lo.ForEach(lines, func(line scmmodel.InventReceiveIssueLine, index int) {
		itemIDs[index] = line.ItemID
		uomIDs[index] = line.UnitID
		if line.SourceTrxType == string(scmmodel.InventReceive) {
			ivjIDs = append(ivjIDs, line.SourceJournalID)
		}

		if line.SourceTrxType == string(scmmodel.PurchOrder) {
			for _, c := range line.References {
				if c.Key == "PurchaseOrderID" {
					purchaseOrderID := c.Value.(string)
					poIDs = append(poIDs, purchaseOrderID)
				}
			}
		}

		specIDs[index] = line.SKU
	})

	//get item
	itemsKV := map[string]*tenantcoremodel.Item{}
	if len(itemIDs) > 0 {
		items, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", itemIDs...)))
		itemsKV = lo.Associate(items, func(item *tenantcoremodel.Item) (string, *tenantcoremodel.Item) {
			return item.ID, item
		})
	}

	uomsKV := map[string]*tenantcoremodel.UoM{}
	if len(uomIDs) > 0 {
		uoms, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.UoM), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", uomIDs...)))
		uomsKV = lo.Associate(uoms, func(uom *tenantcoremodel.UoM) (string, *tenantcoremodel.UoM) {
			return uom.ID, uom
		})
	}

	posKV := map[string]*scmmodel.PurchaseOrderJournal{}
	if len(poIDs) > 0 {
		pos, _ := datahub.Find(p.opt.Db, new(scmmodel.PurchaseOrderJournal), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", poIDs...)))
		posKV = lo.Associate(pos, func(po *scmmodel.PurchaseOrderJournal) (string, *scmmodel.PurchaseOrderJournal) {
			return po.ID, po
		})
	}

	specKV := map[string]*tenantcoremodel.ItemSpec{}
	if len(specIDs) > 0 {
		specs, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", specIDs...)))
		specKV = lo.Associate(specs, func(spec *tenantcoremodel.ItemSpec) (string, *tenantcoremodel.ItemSpec) {
			return spec.ID, spec
		})
	}

	ivjsKV := map[string]*scmmodel.InventJournal{}
	whIDs := []interface{}{}
	if len(ivjIDs) > 0 {
		ivjs, _ := datahub.Find(p.opt.Db, new(scmmodel.InventJournal), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", ivjIDs...)))

		whIDs = make([]interface{}, len(ivjs))
		lo.ForEach(ivjs, func(ivj *scmmodel.InventJournal, index int) {
			ivjsKV[ivj.ID] = ivj
			whIDs[index] = ivj.InventDim.WarehouseID
		})
	}

	whsKV := map[string]*tenantcoremodel.LocationWarehouse{}
	if len(whIDs) > 0 {
		whs, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", whIDs...)))
		whsKV = lo.Associate(whs, func(wh *tenantcoremodel.LocationWarehouse) (string, *tenantcoremodel.LocationWarehouse) {
			return wh.ID, wh
		})
	}

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.InventReceive), header.ID)
	pv.Signature = append(pv.Signature, signature...)

	pv.Header = codekit.M{}.Set("Data", [][]string{
		{"ID:", header.ID, "", "", "", "", "", "", ""},
		{"Name:", header.Name, "", "", "", "", "", "", ""},
		{"Site:", site.Name, "", "", "", "", "", "", ""},
		{"Date:", FormatDate(&header.TrxDate), "", "", "", "", "", "", ""},
	}).Set("Footer", [][]string{})

	pv.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"ID:", header.ID},
			{"Name:", header.Name},
			{"Site:", site.Name},
			{"Date:", FormatDate(&header.TrxDate)},
		},
		Footer: [][]string{},
	}

	sectionLine := tenantcoremodel.PreviewSection{
		HideTitle:   false,
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items: [][]string{
			{"No", "Source Trx Type", "Source Reff No", "Origin", "Part Number", "Part Description", "Original Qty", "Settled Qty", "Trx Qty", "UoM", "Unit Cost"},
		},
	}

	lineCount := 1
	sectionLine.Items = append(sectionLine.Items, lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) []string {
		item, _ := itemsKV[line.ItemID]
		unit, _ := uomsKV[line.UnitID]
		spec, _ := specKV[line.SKU]

		// get origin (vendor name)
		origin := ""

		if line.SourceTrxType == string(scmmodel.PurchOrder) {
			for _, c := range line.References {
				if c.Key == "PurchaseOrderID" {
					purchaseOrderID := c.Value.(string)
					po, ok := posKV[purchaseOrderID]
					if !ok {
						po = new(scmmodel.PurchaseOrderJournal)
					}
					origin = po.VendorName
					break
				}
			}
		} else if line.SourceTrxType == string(scmmodel.InventReceive) {
			ivj, ok := ivjsKV[line.SourceJournalID]
			if !ok {
				ivj = new(scmmodel.InventJournal)
			}

			whName, ok := whsKV[ivj.InventDim.WarehouseID]
			if !ok {
				whName = new(tenantcoremodel.LocationWarehouse)
			}
			origin = whName.Name
		}

		tenantcorelogic.MWPreAssignItem(line.ItemID+"~~"+line.SKU, false)(p.ctx, &item)

		res := make([]string, 11)
		res[0] = strconv.Itoa(lineCount)
		res[1] = line.SourceTrxType
		res[2] = line.SourceJournalID
		res[3] = origin // TODO: dapet vendor dari mana ya?
		res[4] = spec.SKU
		res[5] = lo.Ternary(item.ID != "", item.ID, item.Name)
		res[6] = fmt.Sprintf("%.2f", line.OriginalQty) // Original Qty
		res[7] = fmt.Sprintf("%.2f", line.SettledQty)
		res[8] = fmt.Sprintf("%.2f", line.TrxQty) // Trx Qty
		// res[9] = fmt.Sprintf("%.2f", line.OriginalQty) // Qty
		res[9] = unit.Name
		res[10] = FormatMoney(line.UnitCost)

		lineCount++
		return res
	})...)

	pv.Sections = append(pv.Sections, sectionLine)

	return pv
}

func (p *InventoryReceivePosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		res string
		err error
	)

	err = sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		trxs, err = InventTrxSplitSave(tx, trxs, scmmodel.ItemPlanned)
		if err != nil {
			return err
		}

		inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
		for _, t := range inventTrxs {
			origLine, ok := lo.Find(p.lines, func(l scmmodel.InventReceiveIssueLine) bool {
				return l.Item.ID == t.Item.ID && l.InventDim.SpecID == t.InventDim.SpecID
			})
			if ok {
				t.AmountPhysical = origLine.CostPerUnit * t.Qty
			}
			t.References = t.References.Set(string(scmmodel.RefKeyGoodReceive), p.header.ID)
		}

		res, err = ficologic.PostModelSave(tx, header, "GoodReceiveVoucherNo", trxs)
		if err != nil {
			return err
		}

		if e := p.updateLines(tx); e != nil {
			return e
		}

		_, err = NewItemBalanceHub(tx).Sync(nil, ItemBalanceOpt{
			CompanyID:       header.CompanyID,
			ConsiderSKU:     true,
			DisableGrouping: true,
			ItemIDs: lo.Map(inventTrxs, func(t *scmmodel.InventTrx, index int) string {
				return t.Item.ID
			}),
		})
		if err != nil {
			return err
		}

		items := lo.GroupBy(inventTrxs, func(i *scmmodel.InventTrx) string {
			return fmt.Sprintf("%s_%s", i.Item.ID, i.InventDim.InventDimID)
		})

		for _, trx := range items {
			CalcUnitCost(p.opt.Db, trx[0].Item, trx[0].InventDim, &p.header.TrxDate)
		}

		header.Status = ficomodel.JournalStatusPosted
		err = tx.Update(header, "Status")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return res, err
}

func (p *InventoryReceivePosting) validate(lines []*scmmodel.InventTrx) error {
	if len(p.lines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	for _, line := range p.lines {
		// SettledQty dari FE = Qty + Settled
		if (math.Abs(line.SettledQty)) > math.Abs(line.OriginalQty) {
			return fmt.Errorf("Qty cannot be more than Original Qty")
		}
	}

	// check item min max
	if e := ItemMinMaxValidation(p.opt.Db, lines, string(p.header.TrxType)); e != nil {
		return e
	}

	return nil
}

func (p *InventoryReceivePosting) Approved() error {
	return nil
}

func (p *InventoryReceivePosting) Rejected() error {
	return nil
}

func (p *InventoryReceivePosting) GetAccount() string {
	return p.header.Name // TODO: seharusnya return header.Text kalau field Name sudah diganti dengan Text
}

func (p *InventoryReceivePosting) updateLines(h *datahub.Hub) error {
	poORM := sebar.NewMapRecordWithORM(h, new(scmmodel.PurchaseOrderJournal))

	for lineIdx, line := range p.lines {
		poID := codekit.ToString(line.References.Get(string(scmmodel.RefKeyPurchaseOrderID), ""))
		poLineID := codekit.ToString(line.References.Get(string(scmmodel.RefKeyPOLineID), ""))

		if poID != "" && poLineID != "" {
			po, _ := poORM.Get(poID)
			if po.ID == "" {
				continue // bypass as PO not found
			}

			// TODO: get PO Line data by id
			poLine, found := lo.Find(po.Lines, func(d scmmodel.PurchaseJournalLine) bool {
				return d.ID == poLineID
			})
			if found == false {
				continue // bypass as PO Line not found
			}

			// TODO: fill GR Line with Tax and Discount from PO Line
			line.DiscountType = poLine.DiscountType
			line.DiscountValue = poLine.DiscountValue
			line.DiscountAmount = poLine.DiscountAmount
			line.DiscountGeneral = poLine.DiscountGeneral
			line.TaxCodes = poLine.TaxCodes
			line.OtherExpenses = po.OtherExpenses
		}

		p.lines[lineIdx] = line
	}

	p.header.Lines = p.lines
	return h.Update(p.header, "Lines")
}

func (p *InventoryReceivePosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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

func (p *InventoryReceivePosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
