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

type businessTripPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.BusinessTrip
	jt      *hcmmodel.JournalType
	trxType string
}

func NewBusinessTripPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.BusinessTrip, ficomodel.JournalLine] {
	c := new(businessTripPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.BusinessTrip, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *businessTripPosting) Header() (*hcmmodel.BusinessTrip, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.BusinessTrip), p.opt.JournalID)
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

func (p *businessTripPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.BusinessTrip) (string, error) {
	return "", nil
}

func (p *businessTripPosting) Approved() error {
	return nil
}

func (p *businessTripPosting) Rejected() error {
	return nil
}

func (p *businessTripPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Business Trip",
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

func (p *businessTripPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *businessTripPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 1),
	}

	empIds := make([]string, len(p.header.Lines)+1)
	empIds[0] = p.header.RequestorID
	for i, d := range p.header.Lines {
		empIds[i+1] = d.EmployeeID
	}

	employees := []tenantcoremodel.Employee{}
	err := p.opt.Db.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", empIds...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	siteIds := make([]string, len(employees)+1)
	mapEmployee := map[string]tenantcoremodel.Employee{}
	for i, d := range employees {
		mapEmployee[d.ID] = d
		siteIds[i] = d.Dimension.Get("Site")
	}

	employeeDetails := []bagongmodel.EmployeeDetail{}
	err = p.opt.Db.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("EmployeeID", empIds...),
	), &employeeDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	masterIds := make([]string, len(employeeDetails)*3)
	mapEmployeeDetail := map[string]bagongmodel.EmployeeDetail{}
	i := 0
	for _, d := range employeeDetails {
		mapEmployeeDetail[d.EmployeeID] = d
		masterIds[i] = d.Department
		i++
		masterIds[i] = d.PostCode
		i++
		masterIds[i] = d.Level
		i++
	}

	masters := []tenantcoremodel.MasterData{}
	err = p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", masterIds...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	sites := []bagongmodel.Site{}
	err = p.opt.Db.Gets(new(bagongmodel.Site), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", siteIds...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	mapSite := lo.Associate(sites, func(m bagongmodel.Site) (string, string) {
		return m.ID, m.Name
	})

	data := make([][]string, 8)
	for i := range data {
		data[i] = make([]string, 2)
	}
	emp := mapEmployee[p.header.RequestorID]
	detail := mapEmployeeDetail[p.header.RequestorID]
	data[0] = []string{"Requestor Name : ", emp.Name}
	data[1] = []string{"Requestor NIK : ", detail.EmployeeNo}
	data[2] = []string{"Requestor Position : ", mapMaster[detail.Position]}
	data[3] = []string{"Requestor Email : ", emp.Email}
	data[4] = []string{"Requestor Site : ", mapSite[emp.Dimension.Get("Site")]}
	data[5] = []string{"Request Date : ", p.header.RequestDate.Format("02 January 2006")}
	data[6] = []string{"Date From : ", p.header.DateFrom.Format("02 January 2006")}
	data[7] = []string{"Date To : ", p.header.DateTo.Format("02 January 2006")}
	preview.Header["Data"] = data

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Detail",
		Items:       make([][]string, 0),
	}

	section.Items = append(section.Items, []string{"NIK", "Name", "Position", "Level", "Department", "Site", "Location", "Description", "Cost", "Total Cost"})
	for _, l := range p.header.Lines {
		emp := mapEmployee[l.EmployeeID]
		detail := mapEmployeeDetail[l.EmployeeID]

		total := 0.0
		items := make([][]string, 0)
		for i, d := range l.Details {
			total += d.Cost
			if i == 0 {
				items = append(items, []string{detail.EmployeeNo, emp.Name, mapMaster[detail.Position], mapMaster[detail.Level], mapMaster[detail.Department], mapSite[emp.Dimension.Get("Site")], l.Location, d.Description, codekit.ToString(d.Cost), ""})
			} else {
				items = append(items, []string{"", "", "", "", "", "", "", d.Description, codekit.ToString(d.Cost), ""})
			}
		}
		items[0][9] = codekit.ToString(total)
		section.Items = append(section.Items, items...)
	}
	preview.Sections[0] = section

	return preview, nil
}
