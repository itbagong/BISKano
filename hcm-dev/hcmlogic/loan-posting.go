package hcmlogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type loanPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.Loan
	jt      *hcmmodel.JournalType
	trxType string
}

func NewLoanPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.Loan, ficomodel.JournalLine] {
	c := new(loanPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.Loan, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *loanPosting) Header() (*hcmmodel.Loan, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.Loan), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}
	p.header = j
	jt, err := datahub.GetByID(p.opt.Db, new(hcmmodel.JournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}
	p.jt = jt
	p.trxType = p.jt.TransactionType

	j.PostingProfileID = p.jt.PostingProfileID
	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	return j, pp, nil
}

func (p *loanPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.Loan) (string, error) {
	return "", nil
}

func (p *loanPosting) Approved() error {
	return nil
}

func (p *loanPosting) Rejected() error {
	return nil
}

func (p *loanPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				TrxType:                  p.jt.TransactionType,
				Menu:                     "Loan",
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

func (p *loanPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
	latestNotif.IsApproval = false

	if op == ficologic.PostOpApprove {
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

func (p *loanPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 1),
	}

	employee := new(tenantcoremodel.Employee)
	err := p.opt.Db.GetByID(employee, p.header.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	employeeDetail := new(bagongmodel.EmployeeDetail)
	err = p.opt.Db.GetByFilter(employeeDetail, dbflex.Eq("EmployeeID", p.header.EmployeeID))
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	masters := []tenantcoremodel.MasterData{}
	err = p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{employeeDetail.Position, employeeDetail.Department}...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	data := make([][]string, 10)
	for i := range data {
		data[i] = make([]string, 2)
	}
	data[0] = []string{"Request Date : ", p.header.RequestDate.Format("02 January 2006")}
	data[1] = []string{"NIK : ", employeeDetail.EmployeeNo}
	data[2] = []string{"Employee Name : ", employee.Name}
	data[3] = []string{"Position : ", mapMaster[employeeDetail.Position]}
	data[4] = []string{"Department : ", mapMaster[employeeDetail.Department]}
	data[5] = []string{"Work Location : ", ""}
	data[6] = []string{"Employee Status : ", employeeDetail.EmployeeStatus}
	data[7] = []string{"Mobile Phone No : ", employeeDetail.Phone}
	data[8] = []string{"Period of Employment : ", employeeDetail.WorkingPeriod}
	data[9] = []string{"Salary : ", codekit.ToString(employeeDetail.BasicSalary)}
	preview.Header["Data"] = data

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "",
		Items:       make([][]string, 9),
	}

	for i := range section.Items {
		section.Items[i] = make([]string, 2)
	}

	section.Items[0] = []string{"Title", "Value"}
	section.Items[1] = []string{"Loan Purpose", p.header.LoanPurpose}
	section.Items[2] = []string{"Loan Application Amount", codekit.ToString(p.header.LoanApplication)}
	section.Items[3] = []string{"Loan Period", codekit.ToString(p.header.LoanPeriod)}
	section.Items[4] = []string{"Installment", codekit.ToString(p.header.Installment)}
	section.Items[5] = []string{"Approved Loan Amount", codekit.ToString(p.header.ApprovedLoan)}
	section.Items[6] = []string{"Approved Loan Period", codekit.ToString(p.header.ApprovedLoanPeriod)}
	section.Items[7] = []string{"Approved Installment", codekit.ToString(p.header.ApprovedInstallment)}
	section.Items[8] = []string{"Notes", codekit.ToString(p.header.Notes)}
	preview.Sections[0] = section

	return preview, nil
}
