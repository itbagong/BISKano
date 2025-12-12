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

type InventoryIssuancePosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.InventReceiveIssueJournal
	lines   []scmmodel.InventReceiveIssueLine
	trxType string

	inventTrxs []*scmmodel.InventTrx
	jt         *scmmodel.InventJournalType
}

func NewInventoryIssuancePosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.InventReceiveIssueJournal, scmmodel.InventReceiveIssueLine] {
	issuance := new(InventoryIssuancePosting)
	issuance.ctx = ctx
	issuance.opt = &opt

	pvd := ficologic.PostingProvider[*scmmodel.InventReceiveIssueJournal, scmmodel.InventReceiveIssueLine](issuance)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *InventoryIssuancePosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}

func (p *InventoryIssuancePosting) Header() (*scmmodel.InventReceiveIssueJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.InventReceiveIssueJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	p.trxType = string(j.TrxType)
	if p.trxType == "" {
		p.trxType = string(scmmodel.InventReceive)
	}

	if j.PostingProfileID == "" {
		return nil, nil, fmt.Errorf("missing: posting profile")
	}

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	jt, _ := datahub.GetByID(p.opt.Db, new(scmmodel.InventJournalType), p.header.JournalTypeID)
	if jt == nil {
		jt = new(scmmodel.InventJournalType)
	}

	p.jt = jt

	p.header = j
	return j, pp, nil
}

func (p *InventoryIssuancePosting) Lines() ([]scmmodel.InventReceiveIssueLine, error) {
	lines := lo.Map(p.header.Lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
		if line.Text == "" {
			line.Text = p.header.Text
		}
		line.InventDim.Calc() // to set InventDimID for validate()
		return line
	})

	for index, line := range lines {
		line.Qty = -line.Qty
		unitRatio, err := ConvertUnit(p.opt.Db, 1, line.UnitID, line.Item.DefaultUnitID)
		if err != nil {
			return nil, fmt.Errorf("invalid: unit for item %s: %s", line.Item.ID, err.Error())
		}
		line.InventQty = unitRatio * line.Qty

		line.CostPerUnit = GetCostPerUnit(p.opt.Db, line.Item, line.InventDim, &p.header.TrxDate)
		line.UnitCost = line.CostPerUnit * unitRatio
		lines[index] = line
	}

	p.lines = lines
	return lines, nil
}

func (p *InventoryIssuancePosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}
	inventTrxs := []*scmmodel.InventTrx{}
	// journalTypes := sebar.NewMapRecordWithORM(opt.Db, new(scmmodel.InventJournalType))
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
		inventTrx.Status = scmmodel.ItemConfirmed
		inventTrx.TrxDate = p.header.TrxDate
		inventTrx.Qty = lo.Ternary(inventTrx.Qty > 0, (inventTrx.Qty * -1), inventTrx.Qty)             // dibikin minus karena issuance
		inventTrx.TrxQty = lo.Ternary(inventTrx.TrxQty > 0, (inventTrx.TrxQty * -1), inventTrx.TrxQty) // dibikin minus karena issuance

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
			References:        tenantcoremodel.References{}.Set("GIID", p.header.ID),
		}

		var offsetLedger *tenantcoremodel.LedgerAccount
		offsetLedger, ok = ledgersKV[p.jt.DefaultOffset.AccountID]
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
			References:        tenantcoremodel.References{}.Set("GIID", p.header.ID),
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	}

	p.inventTrxs = inventTrxs

	if e := p.validate(lines); e != nil {
		return &preview, trxs, 0, e
	}

	trxs[ledgerTrxs[0].TableName()] = ficologic.ToDataModels(ledgerTrxs)
	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(inventTrxs)
	amount := lo.SumBy(inventTrxs, func(i *scmmodel.InventTrx) float64 {
		return i.AmountPhysical
	})

	return p.GetPreview(opt, header, lines), trxs, amount, nil
}

func (p *InventoryIssuancePosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) *tenantcoremodel.PreviewReport {
	pv := new(tenantcoremodel.PreviewReport)

	site, _ := datahub.GetByID(opt.Db, new(bagongmodel.Site), header.Dimension.Get("Site"))

	itemIDs := make([]string, len(lines))
	uomIDs := make([]string, len(lines))
	skuIDs := make([]string, len(lines))
	ivjIDs := []string{}

	lo.ForEach(lines, func(line scmmodel.InventReceiveIssueLine, index int) {
		itemIDs[index] = line.ItemID
		uomIDs[index] = line.UnitID
		skuIDs[index] = line.SKU

		if line.SourceJournalID == string(scmmodel.InventIssuance) {
			ivjIDs = append(ivjIDs, line.SourceJournalID)
		}
	})

	//get item
	itemsKV := map[string]*tenantcoremodel.Item{}
	if len(itemIDs) > 0 {
		items, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", itemIDs...)))
		itemsKV = lo.Associate(items, func(item *tenantcoremodel.Item) (string, *tenantcoremodel.Item) {
			return item.ID, item
		})
	}

	//get uom
	uomsKV := map[string]*tenantcoremodel.UoM{}
	if len(uomIDs) > 0 {
		uoms, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.UoM), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", uomIDs...)))
		uomsKV = lo.Associate(uoms, func(uom *tenantcoremodel.UoM) (string, *tenantcoremodel.UoM) {
			return uom.ID, uom
		})
	}

	//get spec
	specsKV := map[string]*tenantcoremodel.ItemSpec{}
	if len(skuIDs) > 0 {
		specs, _ := datahub.Find(p.opt.Db, new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", skuIDs...)))
		specsKV = lo.Associate(specs, func(spec *tenantcoremodel.ItemSpec) (string, *tenantcoremodel.ItemSpec) {
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

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.InventIssuance), header.ID)
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
			{"No", "Source Type", "Source Reff No", "Destination", "Part Number", "Part Description", "Original Qty", "Settled Qty", "Trx Qty", "Qty", "UoM"},
		},
	}

	lineCount := 1
	sectionLine.Items = append(sectionLine.Items, lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) []string {
		item, _ := itemsKV[line.ItemID]
		if item == nil {
			item = new(tenantcoremodel.Item)
		}

		unit, _ := uomsKV[line.UnitID]
		if unit == nil {
			unit = new(tenantcoremodel.UoM)
		}
		spec, _ := specsKV[line.SKU]
		if spec == nil {
			spec = new(tenantcoremodel.ItemSpec)
		}

		tenantcorelogic.MWPreAssignItem(line.ItemID+"~~"+line.SKU, false)(p.ctx, &item)

		dest := ""
		if line.SourceTrxType == string(scmmodel.InventIssuance) {
			j, _ := ivjsKV[line.SourceJournalID]
			if j == nil {
				j = new(scmmodel.InventJournal)
			}

			originDestinationID := j.InventDimTo.WarehouseID
			wh, _ := whsKV[originDestinationID]
			if wh == nil {
				wh = new(tenantcoremodel.LocationWarehouse)
			}

			dest = wh.Name
		}

		res := make([]string, 11)
		res[0] = strconv.Itoa(lineCount)
		res[1] = string(line.SourceType)
		res[2] = line.SourceJournalID
		res[3] = dest
		res[4] = spec.SKU
		res[5] = lo.Ternary(item.ID != "", item.ID, item.Name)
		res[6] = fmt.Sprintf("%.2f", line.Qty)
		res[7] = fmt.Sprintf("%.2f", line.SettledQty)
		res[8] = fmt.Sprintf("%.2f", line.InventQty)
		res[9] = fmt.Sprintf("%.2f", line.OriginalQty)
		res[10] = unit.Name

		// res[7] = FormatMoney(line.UnitCost)

		lineCount++
		return res
	})...)

	pv.Sections = append(pv.Sections, sectionLine)

	return pv
}

func (p *InventoryIssuancePosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		res string
		err error
	)

	err = sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		trxs, err = InventTrxSplitSave(tx, trxs, scmmodel.ItemReserved)
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
			t.References = t.References.Set("GIID", p.header.ID)
		}

		res, err = ficologic.PostModelSave(tx, header, "GoodReceiveVoucherNo", trxs) // for header validation only
		if err != nil {
			return err
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

		return nil
	})
	if err != nil {
		return "", err
	}

	return res, err
}

func (p *InventoryIssuancePosting) validate(lines []scmmodel.InventReceiveIssueLine) error {
	if len(lines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	if p.opt.Op == ficologic.PostOpPreview {
		return nil // bypass validation
	}

	groupedLines := lo.GroupBy(lines, func(l scmmodel.InventReceiveIssueLine) string {
		return fmt.Sprintf("%s|%s", l.ItemID, l.InventDim.InventDimID)
	})

	for _, gls := range groupedLines {
		bal, _ := NewItemBalanceHub(p.opt.Db).Get(nil, ItemBalanceOpt{
			CompanyID: p.header.CompanyID,
			ItemIDs:   []string{gls[0].ItemID},
			InventDim: scmmodel.InventDimension{InventDimID: gls[0].InventDim.InventDimID},
		})

		qtyReserved := lo.SumBy(bal, func(b *scmmodel.ItemBalance) float64 {
			return b.QtyReserved
		})

		qtyConfirmed := lo.SumBy(gls, func(g scmmodel.InventReceiveIssueLine) float64 {
			return g.Qty
		})

		if moreThan(qtyConfirmed, qtyReserved, true) {
			return fmt.Errorf("over qty: %s: issue %2.f, reserved %2.f", gls[0].ItemID, qtyConfirmed, qtyReserved)
		}
	}

	for _, line := range p.lines {
		// SettledQty dari FE = Qty + Settled
		if (math.Abs(line.SettledQty)) > math.Abs(line.OriginalQty) {
			return fmt.Errorf("Qty cannot be more than Original Qty")
		}
	}

	// check item min max
	if e := ItemMinMaxValidation(p.opt.Db, p.inventTrxs, string(p.header.TrxType)); e != nil {
		return e
	}

	return nil
}

func (p *InventoryIssuancePosting) Approved() error {
	return nil
}

func (p *InventoryIssuancePosting) Rejected() error {
	return nil
}

func (p *InventoryIssuancePosting) GetAccount() string {
	return p.header.Name // TODO: seharusnya return header.Text kalau field Name sudah diganti dengan Text
}

func (p *InventoryIssuancePosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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

func (p *InventoryIssuancePosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
