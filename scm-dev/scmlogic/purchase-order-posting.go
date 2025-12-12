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
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/golang-module/carbon/v2"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type purchaseOrderPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.PurchaseOrderJournal
	jt      *scmmodel.PurchaseOrderJournalType
	trxType string

	lines      []POInventReceiveIssueLine
	inventTrxs []*scmmodel.InventTrx
	items      *sebar.MapRecord[*tenantcoremodel.Item]
}

type POInventReceiveIssueLine struct {
	ID string // ID dari PurchaseJournalLine.ID
	scmmodel.InventReceiveIssueLine
}

func NewPurchaseOrderPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.PurchaseOrderJournal, scmmodel.PurchaseJournalLine] {
	p := new(purchaseOrderPosting)
	p.opt = &opt
	p.ctx = ctx
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*scmmodel.PurchaseOrderJournal, scmmodel.PurchaseJournalLine](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *purchaseOrderPosting) Header() (*scmmodel.PurchaseOrderJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.PurchaseOrderJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(scmmodel.PurchaseOrderJournalType), j.JournalTypeID)
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
func (p *purchaseOrderPosting) Lines() ([]scmmodel.PurchaseJournalLine, error) {
	rils := []POInventReceiveIssueLine{}

	for idx, line := range p.header.Lines {
		line.LineNo = idx + 1
		line.ID = lo.Ternary(p.header.Lines[idx].ID != "", p.header.Lines[idx].ID, primitive.NewObjectID().Hex())
		line.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&p.header.Location, &line.InventDim)
		if line.Text == "" {
			line.Text = p.header.Text
		}

		p.header.Lines[idx] = line

		ril := scmmodel.InventReceiveIssueLine{}
		ril.InventJournalLine = line.InventJournalLine
		item, _ := p.items.Get(line.ItemID)
		ril.Item = *item
		ril.InventQty, _ = ConvertUnit(p.opt.Db, line.Qty, line.UnitID, item.DefaultUnitID)
		ril.CostPerUnit = line.SubTotal / ril.InventQty
		ril.SourceJournalID = p.header.ID
		ril.SourceTrxType = scmmodel.ModulePurchase.String()
		ril.SourceLineNo = line.LineNo

		rils = append(rils, POInventReceiveIssueLine{ID: line.ID, InventReceiveIssueLine: ril})
	}

	p.lines = rils

	return p.header.Lines, nil
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *purchaseOrderPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseOrderJournal, lines []scmmodel.PurchaseJournalLine) []ficomodel.JournalLine {
	return receiveIssuelineToficoLines(lo.Map(p.lines, func(d POInventReceiveIssueLine, i int) scmmodel.InventReceiveIssueLine {
		return d.InventReceiveIssueLine
	}))
}

func (p *purchaseOrderPosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseOrderJournal, lines []scmmodel.PurchaseJournalLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	trxs := map[string][]orm.DataModel{}

	for _, line := range p.lines {
		inventTrx := new(scmmodel.InventTrx)
		inventTrx.CompanyID = header.CompanyID
		inventTrx.TrxDate = header.TrxDate
		inventTrx.Text = line.Text

		inventTrx.Item = line.Item
		inventTrx.SKU = line.SKU
		inventTrx.Qty = line.InventQty
		inventTrx.TrxQty = line.Qty
		inventTrx.TrxUnitID = line.UnitID
		inventTrx.AmountPhysical = line.CostPerUnit * inventTrx.Qty
		inventTrx.Status = scmmodel.ItemPlanned

		inventTrx.SourceType = scmmodel.ModulePurchase
		inventTrx.SourceJournalID = header.ID
		inventTrx.SourceLineNo = line.LineNo
		inventTrx.SourceTrxType = string(p.trxType)

		inventTrx.InventDim = line.InventDim
		inventTrx.Dimension = line.Dimension
		inventTrx.References = inventTrx.References.Set(string(scmmodel.RefKeyPurchaseOrderID), header.ID)
		inventTrx.References = inventTrx.References.Set(string(scmmodel.RefKeyPOLineID), line.ID)

		p.inventTrxs = append(p.inventTrxs, inventTrx)
	}

	if e := p.validate(p.inventTrxs); e != nil {
		return nil, nil, 0, e
	}

	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(p.inventTrxs)

	if (p.opt.Op == ficologic.PostOpSubmit || p.opt.Op == ficologic.PostOpReject) && len(header.ReffNo) > 0 {
		if e := p.calcRemainingQtyInReff(header); e != nil {
			return nil, nil, 0, e
		}
	}

	return p.GetPreview(opt, header, lines), trxs, header.GrandTotalAmount, nil
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *purchaseOrderPosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseOrderJournal, lines []scmmodel.PurchaseJournalLine, trxs map[string][]orm.DataModel) (string, error) {
	res := ""

	err := sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		inventTrxs, err := InventTrxSingleJournalSave(tx, trxs, header.CompanyID, scmmodel.ModulePurchase, header.ID, scmmodel.ItemPlanned)
		if err != nil {
			return err
		}

		res, err = ficologic.PostModelSave(tx, header, "PurchaseOrderVoucherNo", map[string][]orm.DataModel{}) // for header validation only
		if err != nil {
			return err
		}

		if e := p.updateLineTaxDiscount(tx); e != nil {
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

		vendorCoreORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Vendor))
		vendorCore, _ := vendorCoreORM.Get(header.VendorID)

		header.Status = ficomodel.JournalStatusPosted
		tx.Update(header, "Status")

		go func() {
			if _, e := ficomodel.SendEmailByTemplate("email", "vendor-po-email-template", "en-us", codekit.M{
				"VendorName":  vendorCore.Name,
				"PoNo":        header.ID,
				"Tanggal Doc": carbon.CreateFromStdTime(header.TrxDate).ToDateString(carbon.Local),
				"Address":     header.DeliveryAddress,
			}); e != nil {
				scmconfig.Config.EventHub().Service().Log().Errorf("SendEmailByTemplate: %s", e.Error())
			}
		}()

		return err
	})
	if err != nil {
		return "", err
	}

	return res, err
}

func (p *purchaseOrderPosting) Approved() error {
	return nil
}

func (p *purchaseOrderPosting) Rejected() error {
	return nil
}

func (p *purchaseOrderPosting) GetAccount() string {
	return p.header.Name // TODO: seharusnya return header.Text kalau field Name sudah diganti dengan Text
}

func (p *purchaseOrderPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.PurchaseOrderJournal, lines []scmmodel.PurchaseJournalLine) *tenantcoremodel.PreviewReport {
	preview := tenantcoremodel.PreviewReport{}

	expenseIDs := lo.Map(header.OtherExpenses, func(m scmmodel.OtherExpenses, index int) string {
		return m.Expenses
	})

	expenses := []tenantcoremodel.ExpenseType{}
	err := p.opt.Db.Gets(new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", expenseIDs...),
	), &expenses)
	if err != nil {
		return nil
	}

	mapExpense := lo.Associate(expenses, func(v tenantcoremodel.ExpenseType) (string, string) {
		return v.ID, v.Name
	})

	// build other expenses
	count := len(header.OtherExpenses) + 1
	totalOtherExp := 0.0
	otherExpenseBuild := make([][]string, count)
	otherExpenseBuildMobile := make([][]string, count)
	for i, e := range header.OtherExpenses {
		totalOtherExp += e.Amount
		otherExpenseBuild[i] = []string{"", "", fmt.Sprintf("  %s:", mapExpense[e.Expenses]), FormatMoney(e.Amount)}
		otherExpenseBuildMobile[i] = []string{fmt.Sprintf("   %s:", mapExpense[e.Expenses]), FormatMoney(e.Amount)}
	}
	otherExpenseBuild[count-1] = []string{"", "", "   Sub Total:", FormatMoney(totalOtherExp)}
	otherExpenseBuildMobile[count-1] = []string{"   Sub Total:", FormatMoney(totalOtherExp)}

	empORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Employee))
	vendorCoreORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Vendor))

	requestor, _ := empORM.Get(header.Requestor)
	vendorCore, _ := vendorCoreORM.Get(header.VendorID)

	vendor := new(bagongmodel.BGVendor)
	if e := p.opt.Db.GetByID(vendor, header.VendorID); e != nil {
		p.opt.Db.GetByFilter(vendor, dbflex.Eq("TenantCoreVendorID", header.VendorID))
	}

	signatureRequestor := tenantcoremodel.Signature{
		ID:        header.Requestor,
		Header:    "Pembuat",
		Footer:    requestor.Name,
		Confirmed: "Tgl. " + header.Created.Format("02-January-2006 15:04:05"),
	}
	preview.Signature = append(preview.Signature, signatureRequestor)

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.PurchOrder), header.ID)
	preview.Signature = append(preview.Signature, signature...)

	// build footer
	footer := [][]string{
		{"", "", "Sub Total:", FormatMoney(header.TotalAmount)},
		{"Note:", "", "Discount Line:", FormatMoney(header.TotalDiscountAmount)},
		{header.Note, "", "PPN:", FormatMoney(header.PPN)},
		{"", "", "PPh:", FormatMoney(header.PPH)},
		{"", "", "Other Expenses:", ""},
	}
	footer = append(footer, otherExpenseBuild...)
	footer = append(footer, [][]string{
		{"", "", "Discount Type:", fmt.Sprintf("%v %s", header.Discount.DiscountValue, header.Discount.DiscountType)},
		{"", "", "Discount General:", FormatMoney(header.Discount.DiscountAmount)},
		{"", "", "Total Amount:", FormatMoney(header.GrandTotalAmount)},
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
			{"Date:", FormatDate(&header.TrxDate), "", vendorCore.Name},
			{"Delivery Date:", FormatDate(header.DueDate), "", vendor.VendorAddress},
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
			{"Delivery Date:", FormatDate(header.DueDate)},
			{"To:", vendorCore.Name},
			{"Delivery Address:", ""},
			{header.DeliveryAddress, ""},
		},
		Footer: footerMobile,
	}

	skuIDs := make([]string, len(lines))
	unitIDs := make([]string, len(lines))
	itemIDs := make([]string, len(lines))
	for i, l := range lines {
		skuIDs[i] = l.SKU
		unitIDs[i] = l.UnitID
		itemIDs[i] = l.ItemID
	}

	skus := []tenantcoremodel.ItemSpec{}
	err = p.opt.Db.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", skuIDs...),
	), &skus)
	if err != nil {
		return nil
	}

	mapSku := lo.Associate(skus, func(v tenantcoremodel.ItemSpec) (string, string) {
		return v.ID, v.SKU
	})

	umos := []tenantcoremodel.UoM{}
	err = p.opt.Db.Gets(new(tenantcoremodel.UoM), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", unitIDs...),
	), &umos)
	if err != nil {
		return nil
	}

	mapUom := lo.Associate(umos, func(v tenantcoremodel.UoM) (string, string) {
		return v.ID, v.Name
	})

	items := []tenantcoremodel.Item{}
	err = p.opt.Db.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemIDs...),
	), &items)
	if err != nil {
		return nil
	}

	mapItem := lo.Associate(items, func(v tenantcoremodel.Item) (string, tenantcoremodel.Item) {
		return v.ID, v
	})

	mapAssignItem, err := AssignItem(p.opt.Db, itemIDs, skuIDs)
	if err != nil {
		return nil
	}

	// build lines
	lineCount := 0
	previewLines := [][]string{{"No", "Part Number", "Part Description", "Qty", "UoM", "Unit Cost", "Discount (%)", "Discount", "Sub Total", "Remarks"}}
	prelines := lo.Map(lines, func(d scmmodel.PurchaseJournalLine, i int) []string {
		item := mapItem[d.ItemID]
		item.ID = mapAssignItem[d.ItemID+d.SKU]
		discPercent := lo.Ternary(d.DiscountType == scmmodel.DiscountTypePercent, fmt.Sprintf("%.2f", d.DiscountValue), "-")
		lineCount++
		return []string{strconv.Itoa(lineCount), mapSku[d.SKU], lo.Ternary(item.ID != "", item.ID, item.Name), fmt.Sprintf("%.2f", d.Qty), mapUom[d.UnitID], FormatMoney(d.UnitCost), discPercent, FormatMoney(d.DiscountAmount), FormatMoney(d.SubTotal), d.Remarks}
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

func (p *purchaseOrderPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Purchase Order",
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

func (p *purchaseOrderPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *purchaseOrderPosting) calcRemainingQtyInReff(header *scmmodel.PurchaseOrderJournal) error {
	// proses perhitungan remaining qty bila memiliki ReffNo
	err := sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		if e := PRLinesUpdateRemainingQty(tx, header, lo.Ternary(p.opt.Op == ficologic.PostOpSubmit, Deduct, Increase)); e != nil {
			return e
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *purchaseOrderPosting) updateLineTaxDiscount(h *datahub.Hub) error {
	vendorTaxes := []string{}
	vend, _ := VendorGet(p.header.VendorID)
	if vend.ID != "" {
		vendorTaxes = append(vendorTaxes, vend.Detail.Terms.Taxes1)
		vendorTaxes = append(vendorTaxes, vend.Detail.Terms.Taxes2)
	}

	p.header.Discount.DiscountAmount = lo.Ternary(p.header.Discount.DiscountAmount < 0, (-1 * p.header.Discount.DiscountAmount), p.header.Discount.DiscountAmount) // make it positive

	for lineIdx, line := range p.header.Lines {
		p.header.Lines[lineIdx].DiscountGeneral = p.header.Discount

		if line.Taxable == false {
			p.header.Lines[lineIdx].TaxCodes = []string{}
			continue
		}

		lineTaxes := make([]string, len(vendorTaxes))
		copy(lineTaxes, vendorTaxes)

		item, _ := p.items.Get(line.ItemID)
		if len(item.TaxCodes) > 0 {
			lineTaxes = lo.Filter(lineTaxes, func(d string, i int) bool {
				return lo.Contains(item.TaxCodes, d)
			})
		}

		p.header.Lines[lineIdx].TaxCodes = lineTaxes
	}

	return h.Update(p.header, "Lines", "Discount")
}

func (p *purchaseOrderPosting) validate(lines []*scmmodel.InventTrx) error {
	// check item min max
	if e := ItemMinMaxValidation(p.opt.Db, lines, ""); e != nil {
		return e
	}

	return nil
}
