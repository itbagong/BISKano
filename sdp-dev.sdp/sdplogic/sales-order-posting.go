package sdplogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type salesOrderPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *sdpmodel.SalesOrder
	trxType string

	ev kaos.EventHub

	lines      []ficomodel.JournalLine
	inventTrxs *ficomodel.CustomerJournal
	items      *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewSalesOrderPosting(opt ficologic.PostingHubCreateOpt, ev kaos.EventHub) *ficologic.PostingHub[*sdpmodel.SalesOrder, sdpmodel.SalesOrderLine] {
	p := new(salesOrderPosting)
	p.opt = &opt
	p.ev = ev
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*sdpmodel.SalesOrder, sdpmodel.SalesOrderLine](p) //-------
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *salesOrderPosting) Header() (*sdpmodel.SalesOrder, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(sdpmodel.SalesOrder), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(sdpmodel.SalesOrderJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = string(jt.TrxType)
	if p.trxType == "" {
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

func (p *salesOrderPosting) Lines() ([]sdpmodel.SalesOrderLine, error) {
	rils := []ficomodel.JournalLine{}
	Lines := []sdpmodel.SalesOrderLine{}
	for _, line := range p.header.Lines {
		Lines = append(Lines, line)
		ril := ficomodel.JournalLine{}
		subLedger := ficomodel.SubledgerAccount{}

		subLedger.AccountID = line.Asset
		subLedger.AccountType = tenantcoremodel.TrxModule(ficomodel.SubledgerAsset)

		ril.TagObjectID1 = subLedger
		ril.Text = line.Item + line.Description // ------
		ril.Qty = float64(line.Qty)
		ril.UnitID = line.UoM
		ril.PriceEach = float64(line.UnitPrice)
		ril.Amount = float64(line.Amount)
		ril.Taxable = line.Taxable
		ril.TaxCodes = line.TaxCodes

		rils = append(rils, ril)
	}
	p.lines = rils

	return Lines, nil
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *salesOrderPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *sdpmodel.SalesOrder, lines []sdpmodel.SalesOrderLine) []ficomodel.JournalLine {
	// return receiveIssuelineToficoLines(p.lines)
	return lo.Map(lines, func(line sdpmodel.SalesOrderLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		// jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.Item)
		// jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}

func (p *salesOrderPosting) Calculate(opt ficologic.PostingHubExecOpt, header *sdpmodel.SalesOrder, lines []sdpmodel.SalesOrderLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}

	// inventTrxs := []orm.DataModel{}
	// for _, line := range p.lines {
	// 	inventTrx := new(scmmodel.InventTrx)
	// 	inventTrx.CompanyID = p.header.CompanyID
	// 	inventTrx.TrxDate = p.header.TrxDate

	// 	inventTrx.Item = line.Item
	// 	inventTrx.InventDim = line.InventDim
	// 	inventTrx.Qty = line.InventQty
	// 	inventTrx.AmountPhysical = line.CostPerUnit * inventTrx.Qty

	// 	inventTrx.SourceType = scmmodel.ModulePurchase
	// 	inventTrx.SourceJournalID = p.header.ID
	// 	inventTrx.SourceLineNo = line.LineNo
	// 	inventTrx.SourceTrxType = string(p.trxType)

	// 	inventTrx.Status = lo.Ternary(inventTrx.Qty > 0, scmmodel.ItemPlanned, scmmodel.ItemReserved)

	// 	// p.inventTrxs = append(p.inventTrxs, inventTrx)
	// 	inventTrxs = append(inventTrxs, inventTrx)
	// }

	// trxs[inventTrxs[0].TableName()] = inventTrxs

	// TODO: set preview (belum nemu contoh)
	// TODO: bener ga kalo disini kita pake inventTrxs instead of ledgTrxs (fixo LedgerTransaction)

	return &preview, trxs, 0, nil
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *salesOrderPosting) Post(opt ficologic.PostingHubExecOpt, header *sdpmodel.SalesOrder, lines []sdpmodel.SalesOrderLine, trxs map[string][]orm.DataModel) (string, error) {
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

	header.Status = ficomodel.JournalStatusPosted
	err = db.Save(header)

	// CustomerJurnal := ficomodel.CustomerJournal{}
	rils := []ficomodel.JournalLine{}
	Lines := []sdpmodel.SalesOrderLine{}
	// linesInvReceive := []scmmodel.InventReceiveIssueLine{}

	Items := []string{}
	for _, line := range p.header.Lines {
		Items = append(Items, line.Item)
	}

	ItemTenants := []tenantcoremodel.Item{}
	err = p.ev.Publish("/v1/tenant/item/find", map[string]any{
		"Where": map[string]any{
			"Op":    "$in",
			"field": "_id",
			"value": Items,
		},
	}, &ItemTenants, nil)
	if err != nil {
		return "", err
	}

	for _, line := range p.header.Lines {
		Lines = append(Lines, line)
		ril := ficomodel.JournalLine{}
		subLedger := ficomodel.SubledgerAccount{}

		subLedger.AccountID = line.Asset
		subLedger.AccountType = tenantcoremodel.TrxModule(ficomodel.SubledgerAsset)

		ril.TagObjectID1 = subLedger
		ril.Text = line.Item + " - " + line.Description // ------
		ril.Qty = float64(line.Qty)
		ril.UnitID = line.UoM
		ril.PriceEach = float64(line.UnitPrice)
		ril.Amount = float64(line.Amount)
		ril.Taxable = line.Taxable

		ril.TaxCodes = line.TaxCodes

		rils = append(rils, ril)

		// item := tenantcoremodel.Item{}
		// for _, ItemTenant := range ItemTenants {
		// 	if ItemTenant.ID == line.Item {
		// 		item = ItemTenant
		// 		break
		// 	}
		// }

		// linesInvReceive = append(linesInvReceive, scmmodel.InventReceiveIssueLine{
		// 	InventJournalLine: scmmodel.InventJournalLine{},
		// 	InventQty:         0,
		// 	CostPerUnit:       float64(line.UnitPrice),
		// 	Item:              item,
		// 	SourceType:        sdpmodel.SalsOrder,
		// 	SourceJournalID:   p.header.SalesOrderNo,
		// 	SourceTrxType:     sdpmodel.SalsOrder.String(),
		// 	SourceLineNo:      index,
		// 	OriginalQty:       float64(line.Qty),
		// 	SettledQty:        0,
		// 	DiscountType:      scmmodel.DiscountType(line.DiscountType),
		// 	DiscountValue:     float64(line.Discount),
		// 	DiscountAmount:    p.header.DiscountAmount,
		// 	TaxCodes:          line.TaxCodes,
		// 	References:        line.References,
		// })
	}

	// invReceive := scmmodel.InventReceiveIssueJournal{
	// 	CompanyID:        p.header.CompanyID,
	// 	TrxType:          "Sales Order",
	// 	Name:             "",
	// 	ReffNo:           []string{},
	// 	WarehouseID:      p.header.WarehouseID,
	// 	TrxDate:          p.header.SalesOrderDate,
	// 	JournalTypeID:    p.header.JournalTypeID,
	// 	PostingProfileID: "",
	// 	SectionID:        "",
	// 	Status:           ficomodel.JournalStatusDraft,
	// 	Dimension:        p.header.Dimension,
	// 	Lines:            linesInvReceive,
	// 	Created:          time.Now(),
	// 	LastUpdate:       time.Now(),
	// 	IsUsedVendor:     true,
	// 	Text:             "",
	// }

	// err = p.ev.Publish("/v1/scm/inventory/receive/insert", &invReceive, nil, nil)

	// CustomerJurnal.Lines = rils
	// CustomerJurnal.TrxDate = time.Now()
	// CustomerJurnal.Text = p.header.Name
	// CustomerJurnal.CustomerID = p.header.CustomerID
	// CustomerJurnal.TaxCodes = p.header.TaxCodes
	// CustomerJurnal.CompanyID = p.header.CompanyID
	// CustomerJurnal.Dimension = p.header.Dimension
	// CustomerJurnal.Status = ficomodel.JournalStatusDraft
	// CustomerJurnal.References.Set("SalesOrderNo", p.header.SalesOrderNo)

	// // err = db.Insert(&CustomerJurnal)
	// if err := db.Save(&CustomerJurnal); err != nil {
	// 	return "", fmt.Errorf("Customer Journal Insert: %s", err.Error())
	// }

	return res, err
}

func (p *salesOrderPosting) Approved() error {
	return nil
}

func (p *salesOrderPosting) Rejected() error {
	return nil
}

func (p *salesOrderPosting) GetAccount() string {
	return ""
}

func (p *salesOrderPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Sales Order",
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

func (p *salesOrderPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
