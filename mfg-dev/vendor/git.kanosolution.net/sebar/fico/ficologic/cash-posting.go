package ficologic

import (
	"fmt"
	"io"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type cashPosting struct {
	header *ficomodel.CashJournal
	pp     *ficomodel.PostingProfile
	jt     *ficomodel.CashJournalType

	opt *PostingHubCreateOpt
	cb  *tenantcoremodel.CashBank

	coas           *sebar.MapRecord[*tenantcoremodel.LedgerAccount]
	expenses       *sebar.MapRecord[*tenantcoremodel.ExpenseType]
	banks          *sebar.MapRecord[*tenantcoremodel.CashBank]
	bankGroups     *sebar.MapRecord[*tenantcoremodel.CashBankGroup]
	vendors        *sebar.MapRecord[*tenantcoremodel.Vendor]
	customers      *sebar.MapRecord[*tenantcoremodel.Customer]
	customerGroups *sebar.MapRecord[*tenantcoremodel.CustomerGroup]
	vendorGroups   *sebar.MapRecord[*tenantcoremodel.VendorGroup]

	trxs map[string][]orm.DataModel

	calcHeader bool
}

func NewCashPosting(opt PostingHubCreateOpt) *PostingHub[*ficomodel.CashJournal, ficomodel.JournalLine] {
	c := new(cashPosting)
	c.opt = &opt
	c.coas = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.LedgerAccount))
	c.expenses = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.ExpenseType))
	c.banks = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.CashBank))
	c.vendors = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.Vendor))
	c.customers = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.Customer))
	c.customerGroups = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.CustomerGroup))
	c.vendorGroups = sebar.NewMapRecordWithORM(c.opt.Db, new(tenantcoremodel.VendorGroup))
	c.trxs = make(map[string][]orm.DataModel)
	pvd := PostingProvider[*ficomodel.CashJournal, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *cashPosting) Header() (*ficomodel.CashJournal, *ficomodel.PostingProfile, error) {
	var (
		err error
	)

	if p.calcHeader {
		return p.header, p.pp, nil
	}

	p.header, err = datahub.GetByID(p.opt.Db, new(ficomodel.CashJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract header: %s: %s", p.opt.JournalID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(ficomodel.CashJournalType), p.header.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract journal type: %s: %s", p.header.JournalTypeID, err.Error())
	}
	p.jt = jt

	for idx, _ := range p.header.Lines {
		p.header.Lines[idx].TrxType = jt.TransactionType
	}
	p.header.PostingProfileID = jt.PostingProfileID

	p.pp, err = datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), p.header.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract posting profile: %s: %s", p.header.PostingProfileID, err.Error())
	}

	p.cb, err = datahub.GetByID(p.opt.Db, new(tenantcoremodel.CashBank), p.header.CashBookID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract cash bank: %s: %s", p.header.CashBookID, err.Error())
	}

	cbg, err := datahub.GetByID(p.opt.Db, new(tenantcoremodel.CashBankGroup), p.cb.CashBankGroupID)
	if err != nil {
		return nil, nil, fmt.Errorf("extract cash bank group: %s: %s", p.cb.CashBankGroupID, err.Error())
	}

	p.header.Dimension = tenantcorelogic.TernaryDimension(p.header.Dimension, jt.Dimension, p.cb.Dimension, cbg.Dimension)

	p.cb.MainBalanceAccount = tenantcorelogic.TernaryString(p.cb.MainBalanceAccount, cbg.MainBalanceAccount)
	_, err = p.coas.Get(p.cb.MainBalanceAccount)
	if err != nil {
		return nil, nil, fmt.Errorf("cash main balance account: %s: %s", p.cb.MainBalanceAccount, err.Error())
	}

	p.calcHeader = true
	return p.header, p.pp, err
}

func (p *cashPosting) Lines() ([]ficomodel.JournalLine, error) {
	if len(p.header.Lines) == 0 {
		return nil, fmt.Errorf("no lines")
	}

	lines := []ficomodel.JournalLine{}
	for idx, line := range p.header.Lines {
		if line.Ignore {
			continue
		}
		if line.LineNo == 0 {
			line.LineNo = idx + 1
		}
		if line.OffsetAccount.AccountID == "" {
			line.OffsetAccount = ficomodel.NewSubAccount(ficomodel.SubledgerCashBank, p.header.CashBookID)
		}
		line.Dimension = tenantcorelogic.TernaryDimension(line.Dimension, p.header.Dimension)

		if p.header.CashJournalType == strings.ToUpper(string(ficomodel.CashOut)) ||
			p.header.CashJournalType == strings.ToUpper("Submission Cash Out") {
			line.Amount = -line.Amount
		}

		lines = append(lines, line)
	}

	return lines, nil
}

func (p *cashPosting) ToJournalLines(opt PostingHubExecOpt, header *ficomodel.CashJournal, lines []ficomodel.JournalLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line ficomodel.JournalLine, index int) ficomodel.JournalLine {
		return line
	})
}

func (p *cashPosting) Calculate(opt PostingHubExecOpt, header *ficomodel.CashJournal, lines []ficomodel.JournalLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}

	offsetCoa, err := p.coas.Get(p.cb.MainBalanceAccount)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("ledger account: %s: %s", p.cb.MainBalanceAccount, err.Error())
	}

	allowedModule := []tenantcoremodel.TrxModule{
		ficomodel.SubledgerCashBank,
		ficomodel.SubledgerCustomer, ficomodel.SubledgerVendor,
		ficomodel.SubledgerAccounting, ficomodel.SubledgerExpense}

	for _, line := range lines {
		needCashSchedule := !(line.Account.AccountType == ficomodel.SubledgerCashBank && line.Account.AccountID != "")
		if line.Amount == 0 {
			continue
		}
		if line.Ignore {
			continue
		}
		if !codekit.HasMember(allowedModule, line.Account.AccountType) {
			return nil, nil, 0, fmt.Errorf("line %d: allowed account type: %v", line.LineNo, allowedModule)
		}
		ct := &ficomodel.CashTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCashBank,
			SourceJournalID:   p.header.ID,
			SourceLineNo:      line.LineNo,
			SourceJournalType: p.header.JournalTypeID,
			Dimension:         line.Dimension,
			CashBank:          *p.cb,
			TrxDate:           p.header.TrxDate,
			TrxType:           line.TrxType,
			Amount:            line.Amount,
			Status:            ficomodel.AmountConfirmed,
			Text:              line.Text,
		}

		var cs *ficomodel.CashSchedule
		if needCashSchedule {
			cs = &ficomodel.CashSchedule{
				CompanyID:       p.header.CompanyID,
				SourceType:      ficomodel.SubledgerCashBank,
				SourceJournalID: p.header.ID,
				SourceLineNo:    line.LineNo,
				Expected:        p.header.TrxDate,
				Text:            line.Text,
				Account:         ficomodel.NewSubAccount(line.Account.AccountType, line.Account.AccountID),
				Amount:          line.Amount,
				Status:          ficomodel.CashScheduled,
			}
		}

		var (
			expense   *tenantcoremodel.ExpenseType
			mainCoa   *tenantcoremodel.LedgerAccount
			customer  *tenantcoremodel.Customer
			vendor    *tenantcoremodel.Vendor
			bankLawan *tenantcoremodel.CashBank
			cashSched *ficomodel.CashSchedule
			err       error
		)

		switch line.Account.AccountType {
		case ficomodel.SubledgerExpense:
			expense, mainCoa, err = GetSubledgerFromMapRecord(line.Account.AccountID, "LedgerAccountID", p.expenses, p.coas)

			if err == nil {
				cashSched = &ficomodel.CashSchedule{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.CashScheduled,
					Account:         line.Account,
					Amount:          -line.Amount,
					Expected:        p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting,
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}
			}

		case ficomodel.SubledgerAccounting:
			mainCoa, err = p.coas.Get(line.Account.AccountID)

		case ficomodel.SubledgerCustomer:
			customer, mainCoa, err = GetSubledgerFromMapRecordAndGroup(line.Account.AccountID,
				"Setting.MainBalanceAccount", "GroupID",
				p.customers, p.customerGroups, p.coas)
			if err == nil {
				cashSched = &ficomodel.CashSchedule{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.CashScheduled,
					Account:         line.Account,
					Amount:          -line.Amount,
					Expected:        p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting,
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}

				custTrx := ficomodel.CustomerTransaction{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.AmountConfirmed,
					Customer:        *customer,
					Amount:          -line.Amount,
					TrxDate:         p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting.String(),
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}
				trxs[custTrx.TableName()] = append(trxs[custTrx.TableName()], &custTrx)
			}

		case ficomodel.SubledgerVendor:
			vendor, mainCoa, err = GetSubledgerFromMapRecordAndGroup(line.Account.AccountID, "MainBalanceAccount", "GroupID", p.vendors, p.vendorGroups, p.coas)
			if err == nil {
				cashSched = &ficomodel.CashSchedule{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.CashScheduled,
					Account:         line.Account,
					Amount:          line.Amount,
					Expected:        p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting,
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}

				vendTrx := ficomodel.VendorTransaction{
					CompanyID:       p.header.CompanyID,
					Status:          ficomodel.AmountConfirmed,
					Vendor:          *vendor,
					Amount:          line.Amount,
					TrxDate:         p.header.TrxDate,
					Text:            line.Text,
					SourceType:      ficomodel.SubledgerAccounting.String(),
					SourceJournalID: p.header.ID,
					SourceLineNo:    line.LineNo,
				}
				trxs[vendTrx.TableName()] = append(trxs[vendTrx.TableName()], &vendTrx)
			}

		case ficomodel.SubledgerCashBank:
			bankLawan, mainCoa, err = GetSubledgerFromMapRecordAndGroup(line.Account.AccountID,
				"MainBalanceAccount", "CashBankGroupID",
				p.banks, p.bankGroups, p.coas)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: line account bank: %s, %s", line.Account.AccountID, err.Error())
			}
			ctLawan := &ficomodel.CashTransaction{
				CompanyID:         p.header.CompanyID,
				SourceType:        ficomodel.SubledgerCashBank,
				SourceJournalID:   p.header.ID,
				SourceLineNo:      line.LineNo,
				SourceJournalType: p.header.JournalTypeID,
				Dimension:         line.Dimension,
				CashBank:          *bankLawan,
				TrxDate:           p.header.TrxDate,
				TrxType:           line.TrxType,
				Amount:            -line.Amount,
				Status:            ficomodel.AmountConfirmed,
			}
			trxs[ctLawan.TableName()] = append(trxs[ctLawan.TableName()], ctLawan)

		default:
			return nil, nil, 0, fmt.Errorf("invalid: line account type: %d, %s", line.LineNo, line.Account.AccountType)
		}

		if err != nil {
			return nil, nil, 0, fmt.Errorf("invalid: account: %s, %s: %s", line.Account.AccountType, line.Account.AccountID, err.Error())
		}

		if cashSched != nil {
			trxs[cashSched.TableName()] = append(trxs[cashSched.TableName()], cashSched)
		}

		ledgerTrxOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCashBank,
			SourceTrxType:     line.TrxType,
			SourceJournalID:   p.header.ID,
			SourceLineNo:      line.LineNo,
			SourceJournalType: p.header.JournalTypeID,
			Dimension:         line.Dimension,
			Account:           *offsetCoa,
			TrxDate:           p.header.TrxDate,
			Amount:            line.Amount,
			Status:            ficomodel.AmountConfirmed,
		}

		ledgerTrxMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			SourceType:        ficomodel.SubledgerCashBank,
			SourceTrxType:     line.TrxType,
			SourceJournalID:   p.header.ID,
			SourceLineNo:      line.LineNo,
			SourceJournalType: p.header.JournalTypeID,
			Dimension:         line.Dimension,
			Account:           *mainCoa,
			Expense:           expense,
			TrxDate:           p.header.TrxDate,
			Amount:            -line.Amount,
			Status:            ficomodel.AmountConfirmed,
		}

		trxs[ct.TableName()] = append(trxs[ct.TableName()], ct)
		if cs != nil {
			trxs[cs.TableName()] = append(trxs[cs.TableName()], cs)
		}
		trxs[ledgerTrxMain.TableName()] = append(trxs[ledgerTrxMain.TableName()], ledgerTrxMain, ledgerTrxOffset)
	}

	p.trxs = trxs
	return &preview, trxs, header.Amount, nil
}

func (p *cashPosting) Post(opt PostingHubExecOpt, header *ficomodel.CashJournal, lines []ficomodel.JournalLine, trxs map[string][]orm.DataModel) (string, error) {
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

	// check balance
	cts := FromDataModels(p.trxs[new(ficomodel.CashTransaction).TableName()], new(ficomodel.CashTransaction))
	if len(cts) == 0 {
		return "", fmt.Errorf("missing: cash transaction")
	}
	for _, ct := range cts {
		ct.Status = ficomodel.AmountConfirmed
	}
	ctTotals := lo.MapEntries(lo.GroupBy(cts, func(ct *ficomodel.CashTransaction) string {
		return ct.CashBank.ID
	}), func(bankID string, values []*ficomodel.CashTransaction) (string, float64) {
		return bankID, lo.SumBy(values, func(c *ficomodel.CashTransaction) float64 {
			return c.Amount
		})
	})
	for bankID, amount := range ctTotals {
		if amount > 0 {
			continue
		}

		bals, _ := NewCashBalanceHub(db).Get(nil, CashBalanceOpt{
			CompanyID:  p.header.CompanyID,
			AccountIDs: []string{bankID},
			Dimension:  p.header.Dimension,
		})

		total := lo.SumBy(bals, func(bal *ficomodel.CashBalance) float64 {
			return bal.Balance
		})

		if total < amount {
			bank, _ := datahub.GetByID(db, new(tenantcoremodel.CashBank), bankID)
			if !bank.AllowNegative {
				return "", fmt.Errorf("balance not available: %s, %s", bankID, bank.Name)
			}
		}
	}

	// delete reserved or planned cash transaction
	db.DeleteByFilter(new(ficomodel.CashTransaction),
		dbflex.Eqs("CompanyID", p.header.CompanyID,
			"SourceType", cts[0].SourceType, "SourceJournalID", cts[0].SourceJournalID))

	// save all trxs
	res, err = PostModelSave(db, p.header, "CustomerVoucherNo", trxs)
	return res, err
}

func (p *cashPosting) Approved() error {
	cts := FromDataModels(p.trxs[new(ficomodel.CashTransaction).TableName()], new(ficomodel.CashTransaction))
	db := p.opt.Db
	for _, ct := range cts {
		ct.Status = lo.Ternary(ct.Amount > 0, ficomodel.AmountPlanned, ficomodel.AmountReserved)
		db.Save(ct)

		_, e := NewCustomerBalanceHub(db).Sync(nil, CustomerBalanceOpt{
			CompanyID:  ct.CompanyID,
			AccountIDs: []string{ct.CashBank.ID},
		})
		if e != nil {
			return fmt.Errorf("fail: approved calc balance: %s", e.Error())
		}
	}
	return nil
}

func (p *cashPosting) Rejected() error {
	return nil
}

func (p *cashPosting) GetAccount() string {
	return p.cb.ID
}

func (p *cashPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				TrxType:                  p.header.CashJournalType,
				Text:                     p.header.Text,
				Menu:                     "Cash Journal",
				Amount:                   p.header.Amount,
				UserTo:                   app.UserID,
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

func (p *cashPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error {
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
