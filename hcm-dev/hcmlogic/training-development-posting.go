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

type trainingDevelopmentPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.TrainingDevelopment
	jt      *hcmmodel.TDCJournalType
	trxType string
}

func NewTrainingDevelopmentPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.TrainingDevelopment, ficomodel.JournalLine] {
	c := new(trainingDevelopmentPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.TrainingDevelopment, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *trainingDevelopmentPosting) Header() (*hcmmodel.TrainingDevelopment, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.TrainingDevelopment), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}
	p.header = j
	jt, err := datahub.GetByID(p.opt.Db, new(hcmmodel.TDCJournalType), j.JournalTypeID)
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

func (p *trainingDevelopmentPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.TrainingDevelopment) (string, error) {
	return "", nil
}

func (p *trainingDevelopmentPosting) Approved() error {
	return nil
}

func (p *trainingDevelopmentPosting) Rejected() error {
	return nil
}

func (p *trainingDevelopmentPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Training Development Center",
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

func (p *trainingDevelopmentPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *trainingDevelopmentPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 1),
	}

	participants := []hcmmodel.TrainingDevelopmentParticipant{}
	err := p.opt.Db.Gets(new(hcmmodel.TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("TrainingCenterID", p.header.ID),
	), &participants)
	if err != nil {
		return nil, fmt.Errorf("error when get participant: %s", err.Error())
	}

	ids := make([]string, len(participants)+1)
	ids[0] = p.header.TrainingRequestor
	for i, p := range participants {
		ids[i+1] = p.EmployeeID
	}
	fmt.Println(ids)
	employees := []tenantcoremodel.Employee{}
	err = p.opt.Db.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", ids...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	dimensionMasterIDs := make([]string, len(employees)*2)
	i := 0
	mapEmployee := map[string]tenantcoremodel.Employee{}
	for _, p := range employees {
		dimensionMasterIDs[i] = p.Dimension.Get("Site")
		i++
		dimensionMasterIDs[i] = p.Dimension.Get("CC")
		i++

		mapEmployee[p.ID] = p
	}
	fmt.Println(mapEmployee)
	masters := []tenantcoremodel.DimensionMaster{}
	err = p.opt.Db.Gets(new(tenantcoremodel.DimensionMaster), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", dimensionMasterIDs...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get dimension master: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.DimensionMaster) (string, string) {
		return m.ID, m.Label
	})

	data := make([][]string, 7)

	data[0] = []string{"ID : ", p.header.ID}
	data[1] = []string{"Training Type : ", p.header.TrainingType}
	data[2] = []string{"Training Title : ", p.header.TrainingTitle}
	data[3] = []string{"Requestor Date : ", p.header.RequestDate.Format("02 January 2006")}
	data[4] = []string{"Requestor Name : ", mapEmployee[p.header.TrainingRequestor].Name}
	data[5] = []string{"Request Training Date From : ", p.header.RequestTrainingDateFrom.Format("02 January 2006")}
	data[6] = []string{"Request Training Date To : ", p.header.RequestTrainingDateTo.Format("02 January 2006")}

	preview.Header["Data"] = data

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "Participant",
		Items:       make([][]string, len(participants)+1),
	}

	section.Items[0] = []string{"No.", "Employee", "Site", "Department"}
	for i, p := range participants {
		fmt.Println(p.EmployeeID)
		emp := mapEmployee[p.EmployeeID]
		section.Items[i+1] = []string{codekit.ToString(i + 1), emp.Name, mapMaster[emp.Dimension.Get("Site")], mapMaster[emp.Dimension.Get("CC")]}
	}
	preview.Sections[0] = section

	return preview, nil
}
