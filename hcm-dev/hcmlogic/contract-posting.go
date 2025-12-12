package hcmlogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type contractPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.Contract
	jt      *hcmmodel.JournalType
	trxType string
}

func NewContractPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.Contract, ficomodel.JournalLine] {
	c := new(contractPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.Contract, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *contractPosting) Header() (*hcmmodel.Contract, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.Contract), p.opt.JournalID)
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

func (p *contractPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.Contract) (string, error) {
	if p.header.IsContractExtended {
		newContract := new(hcmmodel.Contract)
		newContract.CompanyID = p.header.CompanyID
		newContract.JournalTypeID = p.header.JournalTypeID
		newContract.PostingProfileID = p.header.PostingProfileID
		newContract.EmployeeID = p.header.EmployeeID
		newContract.JobID = p.header.JobID
		newContract.JobTitle = p.header.JobTitle
		newContract.Reviewer = p.header.Reviewer
		newContract.Status = ficomodel.JournalStatusDraft
		newContract.JoinedDate = p.header.ExpiredContractDate.AddDate(0, 0, 1)
		newContract.ExpiredContractDate = *p.header.ExtendedExpiredContractDate

		err := p.opt.Db.Insert(newContract)
		if err != nil {
			return "", fmt.Errorf("error when save new contract: %s", err.Error())
		}
	}

	return "", nil
}

func (p *contractPosting) Approved() error {
	return nil
}

func (p *contractPosting) Rejected() error {
	return nil
}

func (p *contractPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Contract",
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

func (p *contractPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *contractPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 2),
	}

	master := new(tenantcoremodel.MasterData)
	err := p.opt.Db.GetByID(master, p.header.JobTitle)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	employees := []tenantcoremodel.Employee{}
	err = p.opt.Db.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{p.header.EmployeeID, p.header.Reviewer}...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(m tenantcoremodel.Employee) (string, string) {
		return m.ID, m.Name
	})

	data := make([][]string, 7)
	for i := range data {
		data[i] = make([]string, 2)
	}
	data[0] = []string{"ID : ", p.header.ID}
	data[1] = []string{"Employee Name : ", mapEmployee[p.header.EmployeeID]}
	data[2] = []string{"Job Title : ", master.Name}
	data[3] = []string{"Join Date : ", p.header.JoinedDate.Format("02 January 2006")}
	data[4] = []string{"Expired Contract Date : ", p.header.ExpiredContractDate.Format("02 January 2006")}
	data[5] = []string{"Reviewer : ", mapEmployee[p.header.Reviewer]}

	data[6][0] = "Result : "
	if p.header.IsBecomeEmployee {
		data[6][1] = "Diangkat menjadi karyawan tetap"
	} else if p.header.IsProbationEnd {
		data[6][1] = "Habis Kontrak"
	} else if p.header.IsContractExtended {
		data[6][1] = "Peranjang Masa Kontrak"
	}
	preview.Header["Data"] = data

	// attendance
	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Attendance",
		Items:       make([][]string, 6),
	}

	for i := range section.Items {
		section.Items[i] = make([]string, 8)
	}

	section.Items[0] = []string{"Absent Name", "Month 1", "Month 2", "Month 3", "Month 4", "Month 5", "Month 6", "Total"}
	for i, d := range p.header.Attendace.Presence {
		section.Items[1][i] = codekit.ToString(d.Score)
	}
	for i, d := range p.header.Attendace.Absent {
		section.Items[2][i] = codekit.ToString(d.Score)
	}
	for i, d := range p.header.Attendace.Sick {
		section.Items[3][i] = codekit.ToString(d.Score)
	}
	for i, d := range p.header.Attendace.Leave {
		section.Items[4][i] = codekit.ToString(d.Score)
	}
	for i, d := range p.header.Attendace.Late {
		section.Items[5][i] = codekit.ToString(d.Score)
	}
	preview.Sections[0] = section

	// assessment
	section.Title = "Assessment"
	section.Items = make([][]string, 0)
	section.Items = append(section.Items, []string{"Aspect", "Max Score", "Achieved Score"})
	for _, d := range p.header.ItemDetails {
		section.Items = append(section.Items, []string{d.Aspect, codekit.ToString(d.MaxScore), codekit.ToString(d.AchievedScore)})
	}
	section.Items = append(section.Items, []string{"Total Score", codekit.ToString(p.header.MaxScoreTotal), codekit.ToString(p.header.AchievedScoreTotal)})
	section.Items = append(section.Items, []string{"Final Score", "", codekit.ToString(p.header.FinalScore)})
	preview.Sections[1] = section

	return preview, nil
}
