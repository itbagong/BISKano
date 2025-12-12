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

type plottingPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.OLPlotting
	jt      *hcmmodel.JournalType
	trxType string
}

func NewPlottingPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.OLPlotting, ficomodel.JournalLine] {
	c := new(plottingPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.OLPlotting, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *plottingPosting) Header() (*hcmmodel.OLPlotting, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.OLPlotting), p.opt.JournalID)
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

func (p *plottingPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.OLPlotting) (string, error) {
	return "", nil
}

func (p *plottingPosting) Approved() error {
	return nil
}

func (p *plottingPosting) Rejected() error {
	return nil
}

func (p *plottingPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "OL & Plotting",
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

func (p *plottingPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *plottingPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 1),
	}

	employee := new(tenantcoremodel.Employee)
	err := p.opt.Db.GetByID(employee, p.header.CandidateID)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	employeeDetail := new(bagongmodel.EmployeeDetail)
	err = p.opt.Db.GetByID(employeeDetail, p.header.CandidateID)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail : %s", err.Error())
	}

	manpower := new(hcmmodel.ManpowerRequest)
	err = p.opt.Db.GetByID(manpower, p.header.JobVacancyID)
	if err != nil {
		return nil, fmt.Errorf("error when get manpower : %s", err.Error())
	}

	site := new(tenantcoremodel.DimensionMaster)
	err = p.opt.Db.GetByID(site, p.header.Plotting)
	if err != nil {
		return nil, fmt.Errorf("error when get site : %s", err.Error())
	}

	masters := []tenantcoremodel.MasterData{}
	err = p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{employeeDetail.Position, employeeDetail.Level, employeeDetail.SubGroup, employeeDetail.Department, employeeDetail.POH, employeeDetail.EmployeeStatus}...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	data := make([][]string, 4)
	for i := range data {
		data[i] = make([]string, 2)
	}
	data[0] = []string{"Candidate ID : ", p.header.CandidateID}
	data[1] = []string{"Candidate Name : ", employee.Name}
	data[2] = []string{"Job Vacancy : ", manpower.Name}
	data[3] = []string{"Plotting : ", site.Label}
	preview.Header["Data"] = data

	// attendance
	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Attendance",
		Items:       make([][]string, 16),
	}

	for i := range section.Items {
		section.Items[i] = make([]string, 2)
	}

	section.Items[0] = []string{"Title", "Value"}
	section.Items[1] = []string{"Position", mapMaster[employeeDetail.Position]}
	section.Items[2] = []string{"Level", mapMaster[employeeDetail.Level]}
	section.Items[3] = []string{"Sub Group", mapMaster[employeeDetail.SubGroup]}
	section.Items[4] = []string{"Department", mapMaster[employeeDetail.Department]}
	section.Items[5] = []string{"POH", mapMaster[employeeDetail.POH]}
	section.Items[6] = []string{"Work Location", site.Label}
	section.Items[7] = []string{"Employee Status", mapMaster[employeeDetail.EmployeeStatus]}
	section.Items[8] = []string{"Contract Period", codekit.ToString(p.header.ContractPeriod)}
	section.Items[9] = []string{"Salary", codekit.ToString(p.header.Salary)}
	section.Items[10] = []string{"Benefit", p.header.Benefit}
	section.Items[11] = []string{"Working Hour", codekit.ToString(p.header.WorkingHour)}
	section.Items[12] = []string{"THR", codekit.ToString(p.header.THR)}
	section.Items[13] = []string{"BPJS TK", codekit.ToString(p.header.BPJSTK)}
	section.Items[14] = []string{"BPJS Kesehatan", codekit.ToString(p.header.BPJSHealth)}
	section.Items[15] = []string{"Facility", p.header.Facility}
	preview.Sections[0] = section

	return preview, nil
}
