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
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type purchaseRequestPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.PurchaseRequestJournal
	trxType string

	lines      []scmmodel.InventReceiveIssueLine
	inventTrxs []*scmmodel.InventTrx
	items      *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewPurchaseRequestPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.PurchaseRequestJournal, scmmodel.PurchaseJournalLine] {
	p := new(purchaseRequestPosting)
	p.ctx = ctx
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*scmmodel.PurchaseRequestJournal, scmmodel.PurchaseJournalLine](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *purchaseRequestPosting) Header() (*scmmodel.PurchaseRequestJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.PurchaseRequestJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(scmmodel.PurchaseRequestJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = string(jt.TrxType)
	if p.trxType == "" {
		p.trxType = "Purchase Order"
	}

	j.PostingProfileID = tenantcorelogic.TernaryString(j.PostingProfileID, jt.PostingProfileID)
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

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *purchaseRequestPosting) Lines() ([]scmmodel.PurchaseJournalLine, error) {
	rils := []scmmodel.InventReceiveIssueLine{}
	for idx, line := range p.header.Lines {
		line.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&p.header.Location, &line.InventDim)
		p.header.Lines[idx] = line

		ril := scmmodel.InventReceiveIssueLine{}
		ril.InventJournalLine = line.InventJournalLine

		item, _ := p.items.Get(line.ItemID)
		ril.Item = *item
		ril.InventQty, _ = ConvertUnit(p.opt.Db, line.Qty, line.UnitID, item.DefaultUnitID)
		ril.CostPerUnit = line.SubTotal / ril.InventQty
		ril.SourceJournalID = p.header.ID
		ril.SourceTrxType = scmmodel.ModulePurchase.String()

		rils = append(rils, ril)
	}
	p.lines = rils

	return p.header.Lines, nil
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *purchaseRequestPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseRequestJournal, lines []scmmodel.PurchaseJournalLine) []ficomodel.JournalLine {
	return receiveIssuelineToficoLines(p.lines)
}

func (p *purchaseRequestPosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseRequestJournal, lines []scmmodel.PurchaseJournalLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	trxs := map[string][]orm.DataModel{}

	inventTrxs := []orm.DataModel{}
	for _, line := range p.lines {
		inventTrx := new(scmmodel.InventTrx)
		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.TrxDate = p.header.TrxDate

		inventTrx.Item = line.Item
		inventTrx.SKU = line.SKU
		inventTrx.InventDim = line.InventDim
		inventTrx.Qty = line.InventQty
		inventTrx.AmountPhysical = line.CostPerUnit * inventTrx.Qty

		inventTrx.SourceType = scmmodel.ModulePurchase
		inventTrx.SourceJournalID = p.header.ID
		inventTrx.SourceLineNo = line.LineNo
		inventTrx.SourceTrxType = string(p.trxType)

		inventTrx.Status = lo.Ternary(inventTrx.Qty > 0, scmmodel.ItemPlanned, scmmodel.ItemReserved)

		p.inventTrxs = append(p.inventTrxs, inventTrx)
		inventTrxs = append(inventTrxs, inventTrx)
	}

	trxs[inventTrxs[0].TableName()] = inventTrxs

	return p.GetPreview(opt, header, lines), trxs, p.header.GrandTotalAmount, nil
}

// Post memproses PR jadi PO
func (p *purchaseRequestPosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseRequestJournal, lines []scmmodel.PurchaseJournalLine, trxs map[string][]orm.DataModel) (string, error) {
	return "", nil // do nothing
}

func (p *purchaseRequestPosting) Approved() error {
	return nil
}

func (p *purchaseRequestPosting) Rejected() error {
	return nil
}

func (p *purchaseRequestPosting) GetAccount() string {
	return p.header.Name // TODO: seharusnya return header.Text kalau field Name sudah diganti dengan Text
}

func (p *purchaseRequestPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseRequestJournal, lines []scmmodel.PurchaseJournalLine) *tenantcoremodel.PreviewReport {
	preview := tenantcoremodel.PreviewReport{}

	expTypeORM := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.ExpenseType))
	uomORM := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.UoM))
	empORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Employee))
	specORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemSpec))
	vndORM := sebar.NewMapRecordWithORM(opt.Db, new(bagongmodel.BGVendor))

	requestor, _ := empORM.Get(header.Requestor)
	vendor, _ := vndORM.Get(header.VendorID)

	// build other expenses
	totalOtherExp := float64(0)
	otherExpenseBuild := lo.Map(header.OtherExpenses, func(d scmmodel.OtherExpenses, i int) []string {
		totalOtherExp = totalOtherExp + d.Amount
		exp, _ := expTypeORM.Get(d.Expenses)
		return []string{"", "", "", "", "", fmt.Sprintf("   %s:", exp.Name), FormatMoney(d.Amount)}
	})
	otherExpenseBuild = append(otherExpenseBuild, []string{"", "", "", "", "", "   Sub Total:", FormatMoney(totalOtherExp)})

	otherExpenseBuildMobile := lo.Map(header.OtherExpenses, func(d scmmodel.OtherExpenses, i int) []string {
		totalOtherExp = totalOtherExp + d.Amount
		exp, _ := expTypeORM.Get(d.Expenses)
		return []string{fmt.Sprintf("   %s:", exp.Name), FormatMoney(d.Amount)}
	})
	otherExpenseBuildMobile = append(otherExpenseBuildMobile, []string{"   Sub Total:", FormatMoney(totalOtherExp)})

	signatureRequestor := tenantcoremodel.Signature{
		ID:        header.Requestor,
		Header:    "Pembuat",
		Footer:    requestor.Name,
		Confirmed: "Tgl. " + header.Created.Format("02-January-2006 15:04:05"),
	}
	preview.Signature = append(preview.Signature, signatureRequestor)

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.PurchRequest), header.ID)
	preview.Signature = append(preview.Signature, signature...)

	// build footer
	footer := [][]string{
		{"", "", "", "", "", "Sub Total:", FormatMoney(header.TotalAmount)},
		{"Note:", "", "", "", "", "Discount:", FormatMoney(header.TotalDiscountAmount)},
		{header.Note, "", "", "", "", "PPN:", FormatMoney(header.PPN)},
		{"", "", "", "", "", "PPh:", FormatMoney(header.PPH)},
		{"", "", "", "", "", "Other Expenses:", ""},
	}
	footer = append(footer, otherExpenseBuild...)
	footer = append(footer, [][]string{
		{"", "", "", "", "", "Discount Type:", fmt.Sprintf("%v %s", header.Discount.DiscountValue, header.Discount.DiscountType)},
		{"", "", "", "", "", "Discount General:", FormatMoney(header.Discount.DiscountAmount)},
		{"", "", "", "", "", "Total Amount:", FormatMoney(header.GrandTotalAmount)},
	}...)

	footerMobile := [][]string{
		{"Note:", ""},
		{header.Note, ""},
		{"Sub Total:", FormatMoney(header.TotalAmount)},
		{"Discount:", FormatMoney(header.TotalDiscountAmount)},
		{"PPN:", FormatMoney(header.PPN)},
		{"PPh:", FormatMoney(header.PPH)},
		{"Other Expenses:", ""},
	}
	footerMobile = append(footerMobile, otherExpenseBuildMobile...)
	footerMobile = append(footerMobile, [][]string{
		{"Discount Type:", fmt.Sprintf("%v %s", header.Discount.DiscountValue, header.Discount.DiscountType)},
		{"Discount General:", FormatMoney(header.Discount.DiscountAmount)},
		{"Total Amount:", FormatMoney(header.GrandTotalAmount)},
	}...)

	statusPrinted := "-"
	if header.TotalPrint > 0 {
		statusPrinted = "--"
	}

	// preview header and footer
	preview.Header = codekit.M{
		"Data": [][]string{
			{"No:", header.ID, "", "To:     "},
			{"Date:", FormatDate(&header.TrxDate), "", header.VendorName},
			{"Expected Date:", FormatDate(header.ExpectedDate), "", vendor.VendorAddress},
			{"", "", "", vendor.City + ", " + vendor.Province},
			{"", "", "", vendor.Country},
			{"", "", "", statusPrinted},
			{"", "", "", "Delivery Address:"},
			{"", "", "", header.DeliveryAddress},
		},
		"Footer": footer,
	}

	preview.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"No:", header.ID},
			{"Date:", FormatDate(&header.TrxDate)},
			{"Expected Date:", FormatDate(header.ExpectedDate)},
			{"To:", ""},
			{header.VendorName, ""},
			{vendor.VendorAddress, ""},
			{vendor.City + ", " + vendor.Province, ""},
			{"Delivery Address:", ""},
			{header.DeliveryAddress, ""},
		},
		Footer: footerMobile,
	}

	// build lines
	lineCount := 0
	previewLines := [][]string{{"No", "Part Number", "Part Description", "Qty", "UoM", "Unit Cost", "Discount (%)", "Discount", "Sub Total", "Remarks"}}
	prelines := lo.Map(lines, func(d scmmodel.PurchaseJournalLine, i int) []string {
		spec, _ := specORM.Get(d.SKU)
		item, _ := p.items.Get(d.ItemID)

		tenantcorelogic.MWPreAssignItem(d.ItemID+"~~"+d.SKU, false)(p.ctx, &item)

		u, _ := uomORM.Get(d.UnitID)
		discPercent := lo.Ternary(d.DiscountType == scmmodel.DiscountTypePercent, fmt.Sprintf("%.2f", d.DiscountValue), "-")

		lineCount++
		return []string{strconv.Itoa(lineCount), spec.SKU, lo.Ternary(item.ID != "", item.ID, item.Name), fmt.Sprintf("%.2f", d.Qty), u.Name, FormatMoney(d.UnitCost), discPercent, FormatMoney(d.DiscountAmount), FormatMoney(d.SubTotal), d.Remarks}
	})
	previewLines = append(previewLines, prelines...)

	// preview sections
	preview.Sections = append(preview.Sections, tenantcoremodel.PreviewSection{
		Title:       "",
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       previewLines,
	})

	return &preview
}

func (p *purchaseRequestPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Purchase Request",
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

func (p *purchaseRequestPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
