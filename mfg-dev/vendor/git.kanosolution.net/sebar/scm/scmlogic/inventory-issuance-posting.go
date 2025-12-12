package scmlogic

import (
	"fmt"
	"io"
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

	p.header = j
	return j, pp, nil
}

func (p *InventoryIssuancePosting) Lines() ([]scmmodel.InventReceiveIssueLine, error) {
	mapItems := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	lines := lo.Map(p.header.Lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
		item, err := mapItems.Get(line.ItemID)
		if err == nil {
			line.Item = *item
		}

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
	journalTypes := sebar.NewMapRecordWithORM(opt.Db, new(scmmodel.InventJournalType))
	ledgers := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LedgerAccount))
	itemGroups := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemGroup))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}

	if e := p.validate(lines); e != nil {
		return &preview, trxs, 0, e
	}

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

		itemGroup, _ := itemGroups.Get(inventTrx.Item.ItemGroupID)
		ledgerAccount, err := ledgers.Get(inventTrx.Item.LedgerAccountIDStock)
		if err != nil {
			ledgerAccount, err = ledgers.Get(itemGroup.LedgerAccountIDStock)
			if err != nil {
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
		jt, _ := journalTypes.Get(p.header.JournalTypeID)
		offsetLedger, err = ledgers.Get(jt.DefaultOffset.AccountID)
		if err != nil {
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

	trxs[ledgerTrxs[0].TableName()] = ficologic.ToDataModels(ledgerTrxs)
	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(inventTrxs)
	amount := lo.SumBy(inventTrxs, func(i *scmmodel.InventTrx) float64 {
		return i.AmountPhysical
	})

	return p.GetPreview(opt, header, lines), trxs, amount, nil
}

func (p *InventoryIssuancePosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.InventReceiveIssueJournal, lines []scmmodel.InventReceiveIssueLine) *tenantcoremodel.PreviewReport {
	pv := new(tenantcoremodel.PreviewReport)

	itemORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Item))
	uomORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.UoM))
	whORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LocationWarehouse))
	siteORM := sebar.NewMapRecordWithORM(opt.Db, new(bagongmodel.Site))
	specORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemSpec))

	site, _ := siteORM.Get(header.Dimension.Get("Site"))

	// signatureRequestor := tenantcoremodel.Signature{
	// 	ID:        header.Requestor,
	// 	Header:    "Pembuat",
	// 	Footer:    requestor.Name,
	// 	Confirmed: "Tgl. " + header.Created.Format("02-January-2006 15:04:05"),
	// }

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.InventIssuance), header.ID)
	// pv.Signature = append(pv.Signature, signatureRequestor)
	pv.Signature = append(pv.Signature, signature...)

	whName := ""
	if len(lines) > 0 {
		wh, _ := whORM.Get(lines[0].InventDim.WarehouseID)
		whName = wh.Name
	}

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
		item, _ := itemORM.Get(line.ItemID)
		unit, _ := uomORM.Get(line.UnitID)
		spec, _ := specORM.Get(line.SKU)

		tenantcorelogic.MWPreAssignItem(line.ItemID+"~~"+line.SKU, false)(p.ctx, &item)

		res := make([]string, 11)
		res[0] = strconv.Itoa(lineCount)
		res[1] = string(line.SourceType)
		res[2] = line.SourceJournalID
		res[3] = whName // TODO: contohnya gini "WH-HO 200224 MALANG" ini maksudnya apa?
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
