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

type woReportResourcePosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *mfgmodel.WorkOrderPlanReportResource
	trxType string
	jt      *mfgmodel.WorkOrderJournalType

	items *sebar.MapRecord[*tenantcoremodel.Item]

	inventTrxs []*scmmodel.InventTrx
	wop        *mfgmodel.WorkOrderPlan
	worep      *mfgmodel.WorkOrderPlanReport
}

func NewWorkOrderPlanReportResourcePosting(opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*mfgmodel.WorkOrderPlanReportResource, mfgmodel.WorkOrderResourceItem] {
	p := new(woReportResourcePosting)
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*mfgmodel.WorkOrderPlanReportResource, mfgmodel.WorkOrderResourceItem](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *woReportResourcePosting) Header() (*mfgmodel.WorkOrderPlanReportResource, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlanReportResource), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = lo.Ternary(string(jt.TrxType) != "", string(jt.TrxType), "Work Order")

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), jt.PostingProfileIDResource)
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
	p.wop = woPlan
	p.worep = worep
	p.jt = jt
	return j, pp, nil
}

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *woReportResourcePosting) Lines() ([]mfgmodel.WorkOrderResourceItem, error) {
	return p.header.Lines, nil
}

func (p *woReportResourcePosting) Calculate(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportResource, lines []mfgmodel.WorkOrderResourceItem) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var err error
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}
	ledgers := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LedgerAccount))
	expenses := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ExpenseType))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}

	lo.ForEach(lines, func(line mfgmodel.WorkOrderResourceItem, index int) {
		//get rateperhour in summary
		totalCost := line.WorkingHour * line.RatePerHour
		totalCost = lo.Ternary(totalCost > 0, totalCost*-1, totalCost)

		// itemGroup, _ := itemGroups.Get(line.Item.ItemGroupID)
		expenseType, err := expenses.Get(line.ExpenseType)
		if err != nil {
			return
		}

		ledgerAccount, err := ledgers.Get(expenseType.LedgerAccountID)
		if err != nil {
			return
		}

		// if err != nil {
		// 	ledgerAccount, err = ledgers.Get(itemGroup.LedgerAccountIDStock)
		// if err != nil {
		// 	return
		// }
		// }

		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         p.header.Dimension,
			SourceType:        scmmodel.ModuleWorkorder,
			SourceJournalID:   p.header.WorkOrderPlanID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     string(scmmodel.JournalWorkOrder),
			SourceLineNo:      0,
			TrxDate:           p.header.TrxDate,
			Text:              "",
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            totalCost,
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", p.header.WorkOrderPlanID).
				Set("WorkOrderElemenType", "Cost").
				Set("WorkOrderCostType", "Manpower").
				Set("WorkOrderOutputType", "Output"),
		}

		offset, err := ledgers.Get(p.jt.DefaultOffsiteConsumption.AccountID)
		if err != nil {
			err = fmt.Errorf("invalid: offset account on journal type: %s, %s", p.jt.DefaultOffsiteConsumption.AccountID, err.Error())
			return
		}

		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         p.header.Dimension,
			SourceType:        scmmodel.ModuleWorkorder,
			SourceJournalID:   p.header.WorkOrderPlanID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     string(scmmodel.JournalWorkOrder),
			SourceLineNo:      0,
			TrxDate:           p.header.TrxDate,
			Text:              "",
			Account:           *offset,
			Status:            ficomodel.AmountConfirmed,
			Amount:            (-1 * totalCost),
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	})

	if len(ledgerTrxs) > 0 {
		trxs[ledgerTrxs[0].TableName()] = ficologic.ToDataModels(ledgerTrxs)
	}

	return &preview, trxs, 0, err
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *woReportResourcePosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportResource, lines []mfgmodel.WorkOrderResourceItem) []ficomodel.JournalLine {
	return lo.Map(lines, func(line mfgmodel.WorkOrderResourceItem, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ExpenseType) // TODO: fix this!
		jl.OffsetAccount = p.jt.DefaultOffsiteManPower
		return *jl
	})
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *woReportResourcePosting) Post(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlanReportResource, lines []mfgmodel.WorkOrderResourceItem, trxs map[string][]orm.DataModel) (string, error) {
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

	sums := []mfgmodel.WorkOrderSummaryResource{}
	err = db.GetsByFilter(new(mfgmodel.WorkOrderSummaryResource), dbflex.Eq("WorkOrderPlanID", p.header.WorkOrderPlanID), &sums)
	if err != nil {
		return res, err
	}
	sumGroupM := lo.GroupBy(sums, func(d mfgmodel.WorkOrderSummaryResource) string {
		return d.ExpenseType
	})

	for _, line := range lines {
		if sums, exist := sumGroupM[line.ExpenseType]; exist {
			for _, sum := range sums {
				sum.UsedHour = sum.UsedHour + line.WorkingHour
				db.Update(&sum, "UsedHour")
			}
		} else {
			// add new summary resource, tapi sepertinya ini sudah ga akan kejadian krn sudah di lock
			db.Save(&mfgmodel.WorkOrderSummaryResource{
				WorkOrderPlanID: p.header.WorkOrderPlanID,
				ExpenseType:     line.ExpenseType,
				UsedHour:        line.WorkingHour,
				RatePerHour:     0,
			})
		}
	}

	header.Status = ficomodel.JournalStatusPosted
	db.Update(header, "Status")

	res, err = ficologic.PostModelSave(db, header, "WorkOrderVoucherNo", trxs) // validate only
	if err != nil {
		return res, err
	}

	// update status in WorkOrderPlanReport
	p.worep.WorkOrderPlanReportResourceStatus = string(ficomodel.JournalStatusPosted)
	db.UpdateField(p.worep, dbflex.Eq("_id", p.worep.ID), "WorkOrderPlanReportResourceStatus")

	err = UpdateWOPRWhenAllChildReportPosted(db, header.WorkOrderPlanReportID, "Resource")
	if err != nil {
		return res, err
	}

	return res, nil
}

func (p *woReportResourcePosting) Approved() error {
	return nil
}

func (p *woReportResourcePosting) Rejected() error {
	return nil
}

func (p *woReportResourcePosting) GetAccount() string {
	return ""
}

func (p *woReportResourcePosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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

func (p *woReportResourcePosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
