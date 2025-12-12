package ficologic

import (
	"errors"
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type assetPosting struct {
	header *ficomodel.LedgerJournal
	pp     *ficomodel.PostingProfile

	opt *PostingHubCreateOpt

	coas     *sebar.MapRecord[*tenantcoremodel.LedgerAccount]
	assets   *sebar.MapRecord[*tenantcoremodel.Asset]
	expenses *sebar.MapRecord[*tenantcoremodel.ExpenseType]

	calcHeader bool
}

func NewAssetPosting(opt PostingHubCreateOpt) *PostingHub[*ficomodel.LedgerJournal, ficomodel.JournalLine] {
	c := new(assetPosting)
	c.opt = &opt
	c.coas = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.LedgerAccount))
	c.expenses = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.ExpenseType))
	pvd := PostingProvider[*ficomodel.LedgerJournal, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *assetPosting) Header() (*ficomodel.LedgerJournal, *ficomodel.PostingProfile, error) {
	var (
		err error
	)

	if p.calcHeader {
		return p.header, p.pp, nil
	}

	p.header, err = datahub.GetByID(p.opt.Db, new(ficomodel.LedgerJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract header: %s: %s", p.opt.JournalID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(ficomodel.CashJournalType), p.header.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract journal type: %s: %s", p.header.JournalTypeID, err.Error())
	}

	p.header.PostingProfileID = tenantcorelogic.TernaryString(p.header.PostingProfileID, jt.PostingProfileID)
	p.pp, err = datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), p.header.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract posting profile: %s: %s", p.header.PostingProfileID, err.Error())
	}

	p.header.Dimension = tenantcorelogic.TernaryDimension(p.header.Dimension, jt.Dimension)
	p.calcHeader = true
	return p.header, p.pp, err
}

func (p *assetPosting) Lines() ([]ficomodel.JournalLine, error) {
	if len(p.header.Lines) == 0 {
		return nil, errors.New("lines is empty")
	}

	lines := lo.Map(p.header.Lines, func(line ficomodel.JournalLine, index int) ficomodel.JournalLine {
		line.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, p.header.Dimension)

		return line
	})

	return lines, nil
}

func (p *assetPosting) ToJournalLines(opt PostingHubExecOpt, header *ficomodel.LedgerJournal, lines []ficomodel.JournalLine) []ficomodel.JournalLine {
	return lines
}

func (p *assetPosting) Calculate(opt PostingHubExecOpt, header *ficomodel.LedgerJournal, lines []ficomodel.JournalLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		preview *tenantcoremodel.PreviewReport
		trxs    = map[string][]orm.DataModel{}
	)

	totalAmount := 0.0
	for idx, line := range lines {
		if line.Account.AccountType != ficomodel.SubledgerAsset {
			return nil, nil, 0, fmt.Errorf("line %d: account type should be asset", idx)
		}

		if !(line.OffsetAccount.AccountType == ficomodel.SubledgerAccounting || line.OffsetAccount.AccountType != ficomodel.SubledgerExpense) {
			return nil, nil, 0, fmt.Errorf("line %d: offset account type accepted only Ledger or Expense", idx)
		}

		asset, assetLedgerAccount, err := GetSubledgerFromMapRecord(line.Account.AccountID, "LedgerAccountID", p.assets, p.coas)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("line %d: asset and ledger: %s: %s", idx, line.Account.AccountID, err.Error())
		}

		var (
			offsetLedgerAccount *tenantcoremodel.LedgerAccount
			expense             *tenantcoremodel.ExpenseType
		)
		if line.OffsetAccount.AccountType == ficomodel.SubledgerExpense {
			expense, offsetLedgerAccount, err = GetSubledgerFromMapRecord(line.OffsetAccount.AccountID, "LedgerAccountID", p.expenses, p.coas)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("extract expense: %s: %s", line.OffsetAccount.AccountID, err.Error())
			}
		}

		mainTrx := &ficomodel.LedgerTransaction{
			CompanyID:       p.header.CompanyID,
			SourceType:      ficomodel.SubledgerAsset,
			SourceJournalID: p.header.ID,
			SourceLineNo:    line.LineNo,
			SourceTrxType:   line.TrxType,
			Dimension:       line.Dimension,
			Status:          ficomodel.AmountConfirmed,
			Account:         *assetLedgerAccount,
			Amount:          line.Amount,
			Text:            line.Text,
		}
		mainTrx.References.Set("AssetID", asset.ID)

		offsetTrx := &ficomodel.LedgerTransaction{
			CompanyID:       p.header.CompanyID,
			SourceType:      ficomodel.SubledgerAsset,
			SourceJournalID: p.header.ID,
			SourceLineNo:    line.LineNo,
			Dimension:       line.Dimension,
			Status:          ficomodel.AmountConfirmed,
			Account:         *offsetLedgerAccount,
			Amount:          -line.Amount,
			Text:            line.Text,
		}
		if expense != nil {
			offsetTrx.Expense = expense
			cashSched := &ficomodel.CashSchedule{
				CompanyID:       p.header.CompanyID,
				SourceType:      ficomodel.SubledgerAsset,
				SourceJournalID: p.header.ID,
				SourceLineNo:    line.LineNo,
				Account:         ficomodel.NewSubAccount(ficomodel.SubledgerExpense, expense.ID),
				Status:          ficomodel.CashScheduled,
				Expected:        p.header.TrxDate,
				Text:            line.Text,
				Amount:          -line.Amount,
			}
			trxs[cashSched.TableName()] = append(trxs[cashSched.TableName()], cashSched)
		}

		totalAmount += line.Amount

		trxs[mainTrx.TableName()] = append(trxs[mainTrx.TableName()], mainTrx, offsetTrx)
	}

	return preview, trxs, totalAmount, nil
}

func (p *assetPosting) Post(opt PostingHubExecOpt, header *ficomodel.LedgerJournal, lines []ficomodel.JournalLine, models map[string][]orm.DataModel) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (p *assetPosting) Approved() error {
	return nil
}

func (p *assetPosting) Rejected() error {
	return nil
}

func (p *assetPosting) GetAccount() string {
	return p.header.DefaultOffset.AccountID
}

func (p *assetPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Asset Journal",
				UserTo:                   app.UserID,
				Status:                   app.Status,
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

func (p *assetPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error {
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

	if op == PostOpApprove {
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
	if len(approvals) == pa.Approvers[userApprovals[0].Line-1].MinimalApproverCount &&
		userApprovals[0].Line == pa.CurrentStage {
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
