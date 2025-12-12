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

type talentDevelopmentPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.TalentDevelopment
	jt      *hcmmodel.JournalType
	trxType string
}

func NewTalentDevelopmentPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.TalentDevelopment, ficomodel.JournalLine] {
	c := new(talentDevelopmentPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.TalentDevelopment, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *talentDevelopmentPosting) Header() (*hcmmodel.TalentDevelopment, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.TalentDevelopment), p.opt.JournalID)
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

func (p *talentDevelopmentPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.TalentDevelopment) (string, error) {
	return "", nil
}

func (p *talentDevelopmentPosting) Approved() error {
	return nil
}

func (p *talentDevelopmentPosting) Rejected() error {
	return nil
}

func (p *talentDevelopmentPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Training Development",
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

func (p *talentDevelopmentPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *talentDevelopmentPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
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
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	data := make([][]string, 6)
	for i := range data {
		data[i] = make([]string, 2)
	}
	data[0] = []string{"Type : ", p.header.SubmissionType}
	data[1] = []string{"NIK : ", employeeDetail.EmployeeNo}
	data[2] = []string{"Name : ", employee.Name}
	data[3] = []string{"Point of Hire : ", employeeDetail.POH}
	data[4] = []string{"Join Date : ", employee.JoinDate.Format("02 January 2006")}
	data[5] = []string{"Reason : ", p.header.Reason}
	preview.Header["Data"] = data

	// attendance
	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Benefit Detail",
		Items:       make([][]string, 10),
	}

	for i := range section.Items {
		section.Items[i] = make([]string, 3)
	}

	detail := new(hcmmodel.TalentDevelopmentDetail)
	err = p.opt.Db.GetByID(detail, p.header.ID)
	if err != nil {
		return nil, fmt.Errorf("error when get talent development detail: %s", err.Error())
	}

	ids := []string{employeeDetail.Department, employeeDetail.Position, employeeDetail.Grade, employeeDetail.Group, employeeDetail.SubGroup, employeeDetail.POH,
		detail.Department, detail.Position, detail.Grade, detail.Group, detail.SubGroup, detail.PointOfHire}
	masters := []tenantcoremodel.MasterData{}
	err = p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", ids...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	sites := []tenantcoremodel.DimensionMaster{}
	err = p.opt.Db.Gets(new(tenantcoremodel.DimensionMaster), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{employee.Dimension.Get("Site"), detail.Site}...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapSite := lo.Associate(sites, func(m tenantcoremodel.DimensionMaster) (string, string) {
		return m.ID, m.Label
	})

	section.Items[0] = []string{"Description", "Existing", "New Purpose"}
	section.Items[1] = []string{"Department", mapMaster[employeeDetail.Department], mapMaster[detail.Department]}
	section.Items[2] = []string{"Position", mapMaster[employeeDetail.Position], mapMaster[detail.Position]}
	section.Items[3] = []string{"Grade", mapMaster[employeeDetail.Grade], mapMaster[detail.Grade]}
	section.Items[4] = []string{"Group", mapMaster[employeeDetail.Group], mapMaster[detail.Group]}
	section.Items[5] = []string{"SubGroup", mapMaster[employeeDetail.SubGroup], mapMaster[detail.SubGroup]}
	section.Items[6] = []string{"Site", mapSite[employee.Dimension.Get("Site")], mapSite[detail.Site]}
	section.Items[7] = []string{"PointOfHire", mapMaster[employeeDetail.POH], mapMaster[detail.PointOfHire]}
	section.Items[8] = []string{"BasicSalary", codekit.ToString(employeeDetail.BasicSalary), codekit.ToString(detail.BasicSalary)}
	section.Items[9] = []string{"Allowance", "", detail.Allowance}
	preview.Sections[0] = section

	return preview, nil
}
