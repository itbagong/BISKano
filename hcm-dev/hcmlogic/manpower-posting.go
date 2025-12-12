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

type manpowerPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.ManpowerRequest
	jt      *hcmmodel.JournalType
	trxType string
}

func NewManpowerPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.ManpowerRequest, ficomodel.JournalLine] {
	c := new(manpowerPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.ManpowerRequest, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *manpowerPosting) Header() (*hcmmodel.ManpowerRequest, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.ManpowerRequest), p.opt.JournalID)
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

func (p *manpowerPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.ManpowerRequest) (string, error) {
	return "", nil
}

func (p *manpowerPosting) Approved() error {
	return nil
}

func (p *manpowerPosting) Rejected() error {
	return nil
}

func (p *manpowerPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Manpower Request",
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

func (p *manpowerPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *manpowerPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 1),
	}

	masters := []tenantcoremodel.MasterData{}
	err := p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("MasterDataTypeID", []string{"EmploymentType", "EmployeeSource", "PTE"}...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	employees := []tenantcoremodel.Employee{}
	err = p.opt.Db.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{p.header.RequestorID, p.header.ReplacedEmployeeName}...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(m tenantcoremodel.Employee) (string, string) {
		return m.ID, m.Name
	})

	var data [][]string
	data = append(data, []string{"No : ", p.header.ID})
	data = append(data, []string{"Requestor ID : ", mapEmployee[p.header.RequestorID]})
	data = append(data, []string{"Name : ", p.header.Name})
	data = append(data, []string{"Job Vacancy Title : ", mapMaster[p.header.JobVacancyTitle]})
	data = append(data, []string{"Employment Type : ", mapMaster[p.header.EmployementType]})
	data = append(data, []string{"Onsite Required Date : ", p.header.OnsiteRequiredDate.Format("02 January 2006")})
	data = append(data, []string{"Employee Source : ", mapMaster[p.header.EmployeeSource]})
	data = append(data, []string{"Notes : ", p.header.Notes})
	preview.Header["Data"] = data

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "",
	}

	if p.header.AdditionalNumber != 0 {
		section.Items = make([][]string, 2)
		section.Items[0] = []string{"Additional Number", "Existing Employee Number", "Estimate Cost Per Month", "Employee Number Total"}
		section.Items[1] = []string{codekit.ToString(p.header.AdditionalNumber), codekit.ToString(p.header.ExistingEmployeeNumber), codekit.ToString(p.header.EstimateCostPerMonth), codekit.ToString(p.header.EmployeeNumberTotal)}
	} else if p.header.ReplacedEmployeeName != "" {
		section.Items = make([][]string, 2)
		section.Items[0] = []string{"Position", "Class", "Replaced Employee Name", "Replacement Reason"}
		section.Items[1] = []string{mapMaster[p.header.Position], p.header.Class, mapEmployee[p.header.ReplacedEmployeeName], p.header.ReplacementSeason}
	}
	preview.Sections[0] = section

	return preview, nil
}
