package mfglogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type woReportOutputPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *mfgmodel.WorkOrderPlanReportOutput
	trxType string
	jt      *mfgmodel.WorkOrderJournalType

	items *sebar.MapRecord[*tenantcoremodel.Item]

	wop        *mfgmodel.WorkOrderPlan
	worep      *mfgmodel.WorkOrderPlanReport
	inventTrxs []*scmmodel.InventTrx
}

func NewWorkOrderPlanReportOutputPosting(opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*mfgmodel.WorkOrderPlanReportOutput, mfgmodel.WorkOrderOutputItem] {
	p := new(woReportOutputPosting)
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*mfgmodel.WorkOrderPlanReportOutput, mfgmodel.WorkOrderOutputItem](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *woReportOutputPosting) Header() (*mfgmodel.WorkOrderPlanReportOutput, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlanReportOutput), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = lo.Ternary(string(jt.TrxType) != "", string(jt.TrxType), "Work Order")

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), jt.PostingProfileIDOutput)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", jt.PostingProfileID)
	}

	woPlan, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlan), j.WorkOrderPlanID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: work order plan: %s", j.WorkOrderPlanID)
	}

	worep, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlanReport), j.WorkOrderPlanReportID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: work order plan: %s", j.WorkOrderPlanID)
	}

	p.header = j
	p.jt = jt
	p.wop = woPlan
	p.worep = worep
	return j, pp, nil
}

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *woReportOutputPosting) Lines() ([]mfgmodel.WorkOrderOutputItem, error) {
	return p.header.Lines, nil
}

func (p *woReportOutputPosting) Calculate(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportOutput, lines []mfgmodel.WorkOrderOutputItem) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var err error
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}
	return &preview, trxs, 0, err
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *woReportOutputPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportOutput, lines []mfgmodel.WorkOrderOutputItem) []ficomodel.JournalLine {
	return lo.Map(lines, func(line mfgmodel.WorkOrderOutputItem, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.InventoryLedgerAccID)
		return *jl
	})
}

func (p *woReportOutputPosting) Post(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportOutput, lines []mfgmodel.WorkOrderOutputItem, trxs map[string][]orm.DataModel) (string, error) {
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

	sums := []mfgmodel.WorkOrderSummaryOutput{}
	err = db.GetsByFilter(new(mfgmodel.WorkOrderSummaryOutput), dbflex.Eq("WorkOrderPlanID", p.header.WorkOrderPlanID), &sums)
	if err != nil {
		return res, err
	}
	sumGroupM := lo.GroupBy(sums, func(d mfgmodel.WorkOrderSummaryOutput) string {
		return fmt.Sprintf("%s||%s||%s", d.Type, d.InventoryLedgerAccID, d.SKU)
	})

	for _, line := range lines {
		if sumMs, exist := sumGroupM[fmt.Sprintf("%s||%s||%s", line.Type, line.InventoryLedgerAccID, line.SKU)]; exist {
			for _, sum := range sumMs {
				sum.AchievedQtyAmount = sum.AchievedQtyAmount + line.Qty
				db.Update(&sum, "AchievedQtyAmount")
			}
		} else {
			db.Save(&mfgmodel.WorkOrderSummaryOutput{
				WorkOrderPlanID:      p.header.WorkOrderPlanID,
				Type:                 line.Type,
				InventoryLedgerAccID: line.InventoryLedgerAccID,
				SKU:                  line.SKU,
				Description:          line.Description,
				Group:                line.GroupID,
				AchievedQtyAmount:    line.Qty,
				UnitID:               line.UnitID,
			})
		}
	}

	header.Status = ficomodel.JournalStatusPosted
	db.Update(header, "Status")

	// update status in WorkOrderPlanReport
	p.worep.WorkOrderPlanReportOutputStatus = string(ficomodel.JournalStatusPosted)
	db.UpdateField(p.worep, dbflex.Eq("_id", p.worep.ID), "WorkOrderPlanReportOutputStatus")

	err = UpdateWOPRWhenAllChildReportPosted(db, header.WorkOrderPlanReportID, "Output")
	if err != nil {
		return res, err
	}

	return res, nil
}

func (p *woReportOutputPosting) Approved() error {
	return nil
}

func (p *woReportOutputPosting) Rejected() error {
	return nil
}

func (p *woReportOutputPosting) GetAccount() string {
	return ""
}

func (p *woReportOutputPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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

func (p *woReportOutputPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
