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

type InventoryTransferPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.InventJournal
	trxType string

	inventTrxLines []scmmodel.InventTrxLine
}

func NewInventoryTransferPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.InventJournal, scmmodel.InventTrxLine] {
	transfer := new(InventoryTransferPosting)
	transfer.ctx = ctx
	transfer.opt = &opt

	pvd := ficologic.PostingProvider[*scmmodel.InventJournal, scmmodel.InventTrxLine](transfer)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *InventoryTransferPosting) GetAccount() string {
	return p.header.Text
}

func (p *InventoryTransferPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.InventTrxLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}

func (p *InventoryTransferPosting) Header() (*scmmodel.InventJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.InventJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(scmmodel.InventJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = string(jt.TransactionType)
	if p.trxType == "" {
		return nil, nil, fmt.Errorf("invalid transaction type")
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

func (p *InventoryTransferPosting) Lines() ([]scmmodel.InventTrxLine, error) {
	var err error
	mapItems := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	lines := make([]scmmodel.InventTrxLine, len(p.header.Lines))
	for idx, line := range p.header.Lines {
		lineTrx := scmmodel.InventTrxLine{}
		lineTrx.InventJournalLine = line
		if lineTrx.Text == "" {
			lineTrx.Text = p.header.Text
		}
		lineTrx.JournalID = p.header.ID
		lineTrx.TrxDate = p.header.TrxDate
		lineTrx.TrxType = p.header.TrxType
		lineTrx.LineNo = idx + 1

		lineTrx.Item, err = mapItems.Get(line.ItemID)
		if err != nil {
			return lines, fmt.Errorf("invalid item: %s: %s", line.ItemID, err.Error())
		}

		lineTrx.Dimension = tenantcorelogic.TernaryDimension(lineTrx.Dimension, p.header.Dimension)
		lineTrx.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&p.header.InventDim, &lineTrx.InventDim)

		lineTrx.InventQty, err = ConvertUnit(p.opt.Db, line.Qty, line.UnitID, lineTrx.Item.DefaultUnitID)
		if err != nil {
			return lines, fmt.Errorf("invalid: coversion: %s", err.Error())
		}
		lineTrx.UnitCost = line.UnitCost
		lineTrx.CostPerUnit = GetCostPerUnit(p.opt.Db, *lineTrx.Item, lineTrx.InventDim, &p.header.TrxDate)
		lines[idx] = lineTrx
	}

	p.inventTrxLines = lines
	return lines, nil
}

func (p *InventoryTransferPosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		err error
	)

	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}
	inventTrxs := []*scmmodel.InventTrx{}

	err = p.Validate()
	if err != nil {
		return &preview, trxs, 0, err
	}

	// issuance
	for index, line := range lines {
		inventTrx, err := trxLineToInventTrx(&line, p.opt.Db)
		if err != nil {
			return &preview, trxs, 0, fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}
		inventTrx.Dimension = p.header.DimensionFrom
		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.SourceTrxType = string(scmmodel.InventIssuance)
		inventTrx.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&p.header.InventDim, &inventTrx.InventDim)
		inventTrx.Qty = lo.Ternary(inventTrx.Qty > 0, (inventTrx.Qty * -1), inventTrx.Qty)             // dibikin minus karena issuance
		inventTrx.TrxQty = lo.Ternary(inventTrx.TrxQty > 0, (inventTrx.TrxQty * -1), inventTrx.TrxQty) // dibikin minus karena issuance
		inventTrx.AmountPhysical = line.UnitCost * inventTrx.Qty
		inventTrx.AmountFinancial = 0.0
		inventTrxs = append(inventTrxs, inventTrx)
	}

	// receive
	for index, line := range lines {
		inventTrx, err := trxLineToInventTrx(&line, p.opt.Db)
		if err != nil {
			return &preview, trxs, 0, fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}

		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.SourceTrxType = string(scmmodel.InventReceive)
		inventTrx.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&p.header.InventDimTo, &inventTrx.InventDim) // dibalik karena biar warehouse id ga ke overwrite dengan yg issuance
		inventTrx.AmountPhysical = line.UnitCost * inventTrx.Qty
		inventTrx.AmountFinancial = 0.0

		inventTrxs = append(inventTrxs, inventTrx)
	}

	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(inventTrxs)
	amount := lo.SumBy(inventTrxs, func(i *scmmodel.InventTrx) float64 {
		return i.AmountPhysical
	})

	return p.GetPreview(opt, header, lines), trxs, amount, err
}

func (p *InventoryTransferPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine) *tenantcoremodel.PreviewReport {
	pv := new(tenantcoremodel.PreviewReport)

	itemORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Item))
	specORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemSpec))
	uomORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.UoM))
	siteORM := sebar.NewMapRecordWithORM(opt.Db, new(bagongmodel.Site))
	site, _ := siteORM.Get(header.Dimension.Get("Site"))

	whORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LocationWarehouse))
	WHFrom, _ := whORM.Get(header.InventDim.WarehouseID)
	WHTo, _ := whORM.Get(header.InventDimTo.WarehouseID)

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.JournalTransfer), header.ID)
	pv.Signature = append(pv.Signature, signature...)

	pv.Header = codekit.M{}.Set("Data", [][]string{
		{"ID:", header.ID, "", "", "", ""},
		{"Name:", header.Text, "", "", "", ""},
		{"Site:", site.Name, "", "", "", ""},
		{"Date:", FormatDate(&header.TrxDate), "", "", "", ""},
		{"Delivery Service:", header.DeliveryService, "", "", "", ""},
		{"\n-", "", "", "", "", ""},
		{"Delivery From:", "", "", "", "Delivery To:", ""},
		{"Warehouse:", WHFrom.Name, "", "", "Warehouse:", WHTo.Name},
	}).Set("Footer", [][]string{})

	pv.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"ID:", header.ID},
			{"Name:", header.Text},
			{"Site:", site.Name},
			{"Date:", FormatDate(&header.TrxDate)},
			{"Delivery Service:", header.DeliveryService},
			{"Warehouse From:", WHFrom.Name},
			{"Warehouse To:", WHTo.Name},
		},
		Footer: [][]string{},
	}

	sectionLine := tenantcoremodel.PreviewSection{
		HideTitle:   false,
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items: [][]string{
			{"No", "Part Number", "Part Description", "Qty", "UoM", "Unit Cost", "Remarks"},
		},
	}

	lineCount := 1
	sectionLine.Items = append(sectionLine.Items, lo.Map(lines, func(line scmmodel.InventTrxLine, index int) []string {
		item, _ := itemORM.Get(line.ItemID)
		spec, _ := specORM.Get(line.SKU)
		unit, _ := uomORM.Get(line.UnitID)

		tenantcorelogic.MWPreAssignItem(line.ItemID+"~~"+line.SKU, false)(p.ctx, &item)

		res := make([]string, 7)
		res[0] = strconv.Itoa(lineCount)
		res[1] = spec.SKU
		res[2] = lo.Ternary(item.ID != "", item.ID, item.Name)
		res[3] = fmt.Sprintf("%.2f", line.Qty)
		res[4] = unit.Name
		res[5] = FormatMoney(line.UnitCost)
		res[6] = line.Remarks

		lineCount++
		return res
	})...)

	pv.Sections = append(pv.Sections, sectionLine)

	return pv
}

func (p *InventoryTransferPosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	db, _ = p.opt.Db.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	res, err = ficologic.PostModelSave(db, p.header, "TransferVoucherNo", trxs)
	if err != nil {
		return res, err
	}

	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
	itemIDs := lo.Uniq(lo.Map(inventTrxs, func(d *scmmodel.InventTrx, i int) string {
		return d.Item.ID
	}))

	// sync balance for warehouse from
	balOpt := ItemBalanceOpt{
		CompanyID:       p.opt.CompanyID,
		ConsiderSKU:     true,
		DisableGrouping: true, // harus true, kalo ga, Balance.InventDim ga akan keisi
		ItemIDs:         itemIDs,
		InventDim:       p.header.InventDim,
	}
	_, err = NewItemBalanceHub(db).Sync(nil, balOpt)
	if err != nil {
		return res, err
	}

	// sync balance for warehouse to
	balOpt = ItemBalanceOpt{
		CompanyID:       p.opt.CompanyID,
		ConsiderSKU:     true,
		DisableGrouping: true, // harus true, kalo ga, Balance.InventDim ga akan keisi
		ItemIDs:         itemIDs,
		InventDim:       p.header.InventDimTo,
	}
	_, err = NewItemBalanceHub(db).Sync(nil, balOpt)
	if err != nil {
		return res, err
	}

	return res, err
}

func (p *InventoryTransferPosting) Validate() error {
	if len(p.inventTrxLines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	for _, line := range p.inventTrxLines {
		if e := ValidateBalance(p.opt.Db, line.Qty, line.UnitID, line.ItemID, p.header.CompanyID, line.InventDim); e != nil {
			return e
		}

		// ibCalc := &InventBalanceCalcOpts{
		// 	CompanyID: p.header.CompanyID,
		// 	ItemID:    []string{line.ItemID},
		// 	InventDim: line.InventDim,
		// 	// BalanceDate: ???,
		// }

		// if ibs, err := NewInventBalanceCalc(p.opt.Db).Get(ibCalc); err != nil {
		// 	return fmt.Errorf("Sorry, Qty not enough, no balance found")
		// } else if len(ibs) > 0 {
		// 	isMore, defUnit, e := MoreThanDefaultUnit(p.opt.Db, line.Qty, line.UnitID, ibs[0].Qty, line.ItemID)
		// 	if e != nil {
		// 		return fmt.Errorf("error check item balance: %s", e.Error())
		// 	}
		// 	if isMore {
		// 		return fmt.Errorf("Sorry, Qty not enough, balance: %v %s", ibs[0].Qty, defUnit.Name)
		// 	}
		// }
	}

	return nil
}

func (p *InventoryTransferPosting) Approved() error {
	return nil
}

func (p *InventoryTransferPosting) Rejected() error {
	return nil
}

func (p *InventoryTransferPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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

func (p *InventoryTransferPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
