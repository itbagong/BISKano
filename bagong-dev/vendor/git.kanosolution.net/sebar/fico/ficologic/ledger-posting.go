package ficologic

import (
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

type ledgerPosting struct {
	opt     *PostingHubCreateOpt
	header  *ficomodel.LedgerJournal
	trxType string

	coas     *sebar.MapRecord[*tenantcoremodel.LedgerAccount]
	expenses *sebar.MapRecord[*tenantcoremodel.ExpenseType]
}

func NewLedgerPosting(opt PostingHubCreateOpt) *PostingHub[*ficomodel.LedgerJournal, ficomodel.JournalLine] {
	c := new(ledgerPosting)
	c.opt = &opt
	c.coas = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.LedgerAccount))
	c.expenses = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.ExpenseType))
	pvd := PostingProvider[*ficomodel.LedgerJournal, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *ledgerPosting) ToJournalLines(opt PostingHubExecOpt, header *ficomodel.LedgerJournal, lines []ficomodel.JournalLine) []ficomodel.JournalLine {
	return lines
}

func (p *ledgerPosting) Header() (*ficomodel.LedgerJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(ficomodel.LedgerJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(ficomodel.LedgerJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}
	p.trxType = jt.TrxType
	if p.trxType == "" {
		p.trxType = "LedgerJournal"
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
	if p.header.DefaultOffset.AccountType == "" || p.header.DefaultOffset.AccountID == "" {
		p.header.DefaultOffset = jt.DefaultOffset
	}

	return j, pp, nil
}

func (p *ledgerPosting) Lines() ([]ficomodel.JournalLine, error) {
	lines := make([]ficomodel.JournalLine, len(p.header.Lines))
	for idx, line := range p.header.Lines {
		line.LineNo = idx + 1
		if p.header.DefaultOffset.AccountID != "" && line.OffsetAccount.AccountID == "" {
			line.OffsetAccount = p.header.DefaultOffset
		}
		line.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, p.header.Dimension)
		line.TrxType = p.trxType
		lines[idx] = line
	}
	return lines, nil
}

func (p *ledgerPosting) Calculate(opt PostingHubExecOpt, header *ficomodel.LedgerJournal, lines []ficomodel.JournalLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		err error
	)

	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}
	totalAmount := float64(0)

	// create ledger transaction
	ledgTrxs := []orm.DataModel{}
	for _, line := range lines {
		var (
			expense       *tenantcoremodel.ExpenseType
			ledgerAccount *tenantcoremodel.LedgerAccount
		)
		if !(line.Account.AccountType == ficomodel.SubledgerAccounting || line.Account.AccountType == ficomodel.SubledgerExpense) {
			return nil, nil, 0, fmt.Errorf("line %d: allowed account type only Ledger and Expense", line.LineNo)
		}
		if line.OffsetAccount.AccountType != ficomodel.SubledgerAccounting {
			return nil, nil, 0, fmt.Errorf("line %d: allowed offset account type only %s", line.LineNo, ficomodel.SubledgerAccounting)
		}

		if line.Account.AccountType == ficomodel.SubledgerExpense {
			expense, ledgerAccount, err = GetSubledgerFromMapRecord(line.Account.AccountID, "LedgerAccountID", p.expenses, p.coas)
		} else {
			ledgerAccount, _, err = GetSubledgerFromMapRecord(line.Account.AccountID, "", p.coas, p.coas)
		}
		if err != nil {
			return nil, nil, 0, fmt.Errorf("missing: ledger account: %s, line %d", line.Account.AccountID, line.LineNo)
		}

		offsetCoaID := line.OffsetAccount.AccountID
		ledgerOffset, err := p.coas.Get(offsetCoaID)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("missing: ledger account: %s, line %d", offsetCoaID, line.LineNo)
		}

		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerAccounting,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Expense:           expense,
			Amount:            line.Amount,
		}

		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        ficomodel.SubledgerAccounting,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     line.TrxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Account:           *ledgerOffset,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -line.Amount,
		}

		ledgTrxs = append(ledgTrxs, ltMain, ltOffset)

		// create cash schedule if expense != nil
		if expense != nil {
			schedule := &ficomodel.CashSchedule{
				Account:         ficomodel.NewSubAccount(line.Account.AccountType, line.Account.AccountID),
				CompanyID:       ltMain.CompanyID,
				SourceType:      ltMain.SourceType,
				SourceJournalID: ltMain.SourceJournalID,
				SourceLineNo:    ltMain.SourceLineNo,
				Status:          ficomodel.CashScheduled,
				Text:            ltMain.Text,
				Expected:        p.header.TrxDate,
				Amount:          -ltMain.Amount,
			}
			trxs[schedule.TableName()] = append(trxs[schedule.TableName()], schedule)
		}

		totalAmount += line.Amount
	}
	trxs[ledgTrxs[0].TableName()] = ledgTrxs

	return &preview, trxs, totalAmount, nil
}

func (p *ledgerPosting) Post(opt PostingHubExecOpt, header *ficomodel.LedgerJournal, lines []ficomodel.JournalLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		err error
		res string
	)

	sebar.Tx(p.opt.Db, false, func(tx *datahub.Hub) error {
		res, err = PostModelSave(p.opt.Db, p.header, "LedgerVoucherNo", trxs)
		return err
	})

	return res, err
}

func (p *ledgerPosting) Approved() error {
	return nil
}

func (p *ledgerPosting) Rejected() error {
	return nil
}

func (p *ledgerPosting) GetAccount() string {
	return p.header.DefaultOffset.AccountID
}

func (p *ledgerPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Ledger Journal",
				UserTo:                   app.UserID,
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
				IsApproval:               true,
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

func (p *ledgerPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error {
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
	latestNotif.IsApproval = false

	if op == PostOpApprove {
		latestNotif.Message = fmt.Sprintf("Your Submission ID. %s has been approved", p.header.ID)
	} else {
		latestNotif.Message = fmt.Sprintf("Your Submission ID. %s has been rejected", p.header.ID)
	}

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
				latestNotif.IsApproval = true

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
