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

type talentDevelopmentAssesmentPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.TalentDevelopmentAssesment
	jt      *hcmmodel.JournalType
	trxType string
}

func NewTalentDevelopmentAssesmentPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.TalentDevelopmentAssesment, ficomodel.JournalLine] {
	c := new(talentDevelopmentAssesmentPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.TalentDevelopmentAssesment, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *talentDevelopmentAssesmentPosting) Header() (*hcmmodel.TalentDevelopmentAssesment, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.TalentDevelopmentAssesment), p.opt.JournalID)
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

func (p *talentDevelopmentAssesmentPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.TalentDevelopmentAssesment) (string, error) {
	return "", nil
}

func (p *talentDevelopmentAssesmentPosting) Approved() error {
	return nil
}

func (p *talentDevelopmentAssesmentPosting) Rejected() error {
	return nil
}

func (p *talentDevelopmentAssesmentPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				UserTo:                   app.UserID,
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
				IsApproval:               true,
				Menu:                     "Training Development Assessment",
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

func (p *talentDevelopmentAssesmentPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *talentDevelopmentAssesmentPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 2),
	}

	talentDev := new(hcmmodel.TalentDevelopment)
	err := p.opt.Db.GetByID(talentDev, p.header.TalentDevelopmentID)
	if err != nil {
		return nil, fmt.Errorf("error when get talent development: %s", err.Error())
	}

	employee := new(tenantcoremodel.Employee)
	err = p.opt.Db.GetByID(employee, talentDev.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	employeeDetail := new(bagongmodel.EmployeeDetail)
	err = p.opt.Db.GetByFilter(employeeDetail, dbflex.Eq("EmployeeID", talentDev.EmployeeID))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	master := new(tenantcoremodel.MasterData)
	err = p.opt.Db.GetByID(master, p.header.Interview.Conclusion)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error when get master: %s", err.Error())
	}

	data := make([][]string, 5)
	for i := range data {
		data[i] = make([]string, 2)
	}
	data[0] = []string{"Type : ", talentDev.SubmissionType}
	data[1] = []string{"NIK : ", employeeDetail.EmployeeNo}
	data[2] = []string{"Name : ", employee.Name}
	data[3] = []string{"Join Date : ", employee.JoinDate.Format("02 January 2006")}
	data[4] = []string{"Reason : ", talentDev.Reason}
	preview.Header["Data"] = data

	// attendance
	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "",
		Items:       make([][]string, 3),
	}

	for i := range section.Items {
		section.Items[i] = make([]string, 2)
	}
	section.Items[0] = []string{"Title", "Value"}

	assessmentResult := make([]string, 2)
	assessmentResult[0] = "Assessment Result"
	if p.header.Assesment.IsProbationEnd {
		assessmentResult[1] = "Masa kontrak habis"
	} else if p.header.Assesment.IsBecomeEmployee {
		assessmentResult[1] = "Diangkat menjadi karyawan tetap"
	} else if p.header.Assesment.IsPromoted {
		assessmentResult[1] = "Dipromosikan"
	}

	section.Items[1] = assessmentResult
	section.Items[2] = []string{"Interview Result", master.Name}
	preview.Sections[0] = section

	return preview, nil
}
