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

type workTerminationPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.WorkTermination
	jt      *hcmmodel.JournalType
	trxType string
}

func NewWorkTerminationPosting(opt ficologic.PostingHubCreateOpt) *PostingHub[*hcmmodel.WorkTermination, ficomodel.JournalLine] {
	c := new(workTerminationPosting)
	c.opt = &opt
	pvd := PostingProvider[*hcmmodel.WorkTermination, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *workTerminationPosting) Header() (*hcmmodel.WorkTermination, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.WorkTermination), p.opt.JournalID)
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

func (p *workTerminationPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.WorkTermination) (string, error) {
	return "", nil
}

func (p *workTerminationPosting) Approved() error {
	return nil
}

func (p *workTerminationPosting) Rejected() error {
	return nil
}

func (p *workTerminationPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				Menu:                     "Work Termination",
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

func (p *workTerminationPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *workTerminationPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header:   make(codekit.M),
		Sections: make([]tenantcoremodel.PreviewSection, 4),
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

	site := new(bagongmodel.Site)
	err = p.opt.Db.GetByID(site, employee.Dimension.Get("Site"))
	if err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	masters := []tenantcoremodel.MasterData{}
	err = p.opt.Db.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{employeeDetail.POH, employeeDetail.Grade, employeeDetail.Position, employeeDetail.Department}...),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get master data: %s", err.Error())
	}

	mapMaster := lo.Associate(masters, func(m tenantcoremodel.MasterData) (string, string) {
		return m.ID, m.Name
	})

	row := 0
	col := 0
	if p.header.Type == "Resign" {
		row = 8
		col = 2
	} else {
		row = 14
		col = 2
	}

	data := make([][]string, row)
	for i := range data {
		data[i] = make([]string, col)
	}

	data[0] = []string{"ID : ", p.header.ID}
	data[1] = []string{"Name : ", employee.Name}
	data[2] = []string{"NIK : ", employeeDetail.EmployeeNo}
	data[3] = []string{"Type : ", p.header.Type}
	data[4] = []string{"Resign Date : ", p.header.ResignDate.Format("02 January 2006")}
	data[5] = []string{"Site : ", site.Name}
	data[6] = []string{"Point of Hire : ", mapMaster[employeeDetail.POH]}
	data[7] = []string{"Joined Date : ", employee.JoinDate.Format("02 January 2006")}
	data[7] = []string{"Reason : ", p.header.Reason}

	// add another field if not resign
	if p.header.Type != "Resign" {
		data[8] = []string{"Grade : ", mapMaster[employeeDetail.Grade]}
		data[9] = []string{"Position : ", mapMaster[employeeDetail.Position]}
		data[10] = []string{"Department : ", mapMaster[employeeDetail.Department]}
		data[11] = []string{"Education : ", employeeDetail.LastEducation}
		data[12] = []string{"Working Period : ", employeeDetail.WorkingPeriod}
		data[13] = []string{"UMK Site : ", codekit.ToString(site.Configuration.UMK)}
	}

	preview.Header["Data"] = data

	if p.header.Type != "Resign" {
		section := tenantcoremodel.PreviewSection{
			SectionType: tenantcoremodel.PreviewAsGrid,
			Title:       "Hak Pekerja",
			Items:       make([][]string, 0),
		}

		section.Items = append(section.Items, []string{"No.", "Komponen Penghasilan Kena Pajak Tidak Final", "Perhitungan", "Nomor PP", "Jumlah (Rp)"})
		for _, d := range p.header.NonTaxableIncome {
			section.Items = append(section.Items, []string{d.Number, d.Name, codekit.ToString(d.Calculation), d.PPNo, codekit.ToString(d.Amount)})
		}
		preview.Sections[0] = section

		section.Title = ""
		section.Items = make([][]string, 0)
		section.Items = append(section.Items, []string{"No.", "Komponen Penghasilan Kena Pajak Final", "Perhitungan", "Jumlah (Rp)"})
		for _, d := range p.header.TaxableIncome {
			section.Items = append(section.Items, []string{d.Number, d.Name, codekit.ToString(d.Calculation), codekit.ToString(d.Amount)})
		}
		preview.Sections[1] = section

		section.Title = "Kewajiban Pekerja"
		section.Items = make([][]string, 0)
		section.Items = append(section.Items, []string{"No.", "Komponen", "Perhitungan", "Nomor RV", "Jumlah (Rp)"})
		for _, d := range p.header.MandatoryWork {
			section.Items = append(section.Items, []string{d.Number, d.Name, codekit.ToString(d.Calculation), d.RVNo, codekit.ToString(d.Amount)})
		}
		preview.Sections[2] = section

		section.Title = ""
		section.Items = make([][]string, 0)
		section.Items = append(section.Items, []string{"Sebab Terjadinya PHK"})
		if p.header.ResignOnCompanyInitiative != "" {
			section.Items = append(section.Items, []string{"Keluar Atas Inisiatif Perusahaan"})
			section.Items = append(section.Items, []string{p.header.ResignOnCompanyInitiative})
			if p.header.ResignOnCompanyInitiativeEtc != "" {
				section.Items = append(section.Items, []string{p.header.ResignOnCompanyInitiativeEtc})
			}
		}
		if p.header.ResignOnEmployeeInitiative != "" {
			section.Items = append(section.Items, []string{"Keluar Atas Inisiatif Pekerja"})
			section.Items = append(section.Items, []string{p.header.ResignOnEmployeeInitiative})
			if p.header.ResignOnEmployeeInitiativeEtc != "" {
				section.Items = append(section.Items, []string{p.header.ResignOnEmployeeInitiativeEtc})
			}
		}
		preview.Sections[3] = section
	}
	return preview, nil
}
