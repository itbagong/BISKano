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

type overtimePosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.Overtime
	jt      *hcmmodel.JournalType
	trxType string
}

func NewOvertimePosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.Overtime, ficomodel.JournalLine] {
	c := new(overtimePosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.Overtime, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *overtimePosting) Header() (*hcmmodel.Overtime, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.Overtime), p.opt.JournalID)
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

func (p *overtimePosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.Overtime) (string, error) {
	return "", nil
}

func (p *overtimePosting) Approved() error {
	return nil
}

func (p *overtimePosting) Rejected() error {
	return nil
}

func (p *overtimePosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Overtime",
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

func (p *overtimePosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *overtimePosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 1),
	}

	masterIDs := []string{}
	employeeIDs := []string{p.header.RequestorID}
	for _, d := range p.header.Lines {
		masterIDs = append(masterIDs, d.Position)
		employeeIDs = append(employeeIDs, d.EmployeeID)
	}

	employees := []tenantcoremodel.Employee{}
	err := p.opt.Db.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", employeeIDs...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	mapEmployee := lo.Associate(employees, func(m tenantcoremodel.Employee) (string, string) {
		return m.ID, m.Name
	})

	employeeDetails := []bagongmodel.EmployeeDetail{}
	err = p.opt.Db.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("EmployeeID", employeeIDs...),
	), &employeeDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	mapEmployeeDetail := map[string]bagongmodel.EmployeeDetail{}
	for _, d := range employeeDetails {
		mapEmployeeDetail[d.EmployeeID] = d

		masterIDs = append(masterIDs, d.Department)
	}

	masters := []tenantcoremodel.MasterData{}
	err = p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", masterIDs...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	data := make([][]string, 5)
	for i := range data {
		data[i] = make([]string, 2)
	}
	data[0] = []string{"Requestor Name : ", mapEmployee[p.header.RequestorID]}
	data[1] = []string{"Requestor Department : ", mapMaster[mapEmployeeDetail[p.header.RequestorID].Department]}
	data[2] = []string{"Overtime Date : ", p.header.OvertimeDate.Format("02 January 2006")}
	data[3] = []string{"Estimated Start Time : ", p.header.EstimatedStartTime}
	data[4] = []string{"Estimated End Time : ", p.header.EstimatedEndTime}
	preview.Header["Data"] = data

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "",
		Items:       make([][]string, 0),
	}
	section.Items = append(section.Items, []string{"Employee Name", "NIK", "Position", "Department", "Task", "Actual Overtime", "Day Off"})
	for _, l := range p.header.Lines {
		empDetail := mapEmployeeDetail[l.EmployeeID]
		dayOff := "False"
		if l.OffDay {
			dayOff = "True"
		}
		section.Items = append(section.Items, []string{mapEmployee[l.EmployeeID], empDetail.EmployeeNo, mapMaster[empDetail.Position], mapMaster[empDetail.Department], l.Task, codekit.ToString(l.ActualOvertime), dayOff})
	}
	preview.Sections[0] = section

	return preview, nil
}
