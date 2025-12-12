package hcmlogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
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

type talentDevelopmentSKPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *hcmmodel.TalentDevelopmentSK
	jt      *hcmmodel.JournalType
	trxType string
	ctx     *kaos.Context
}

func NewTalentDevelopmentSKPosting(opt ficologic.PostingHubCreateOpt, ctx *kaos.Context) *PostingHub[*hcmmodel.TalentDevelopmentSK, ficomodel.JournalLine] {
	c := new(talentDevelopmentSKPosting)
	c.opt = &opt
	c.ctx = ctx
	pvd := PostingProvider[*hcmmodel.TalentDevelopmentSK, ficomodel.JournalLine](c)
	return NewPostingHub(pvd, opt)
}

func (p *talentDevelopmentSKPosting) Header() (*hcmmodel.TalentDevelopmentSK, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(hcmmodel.TalentDevelopmentSK), p.opt.JournalID)
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

func (p *talentDevelopmentSKPosting) Post(opt ficologic.PostingHubExecOpt, header *hcmmodel.TalentDevelopmentSK) (string, error) {
	return "", nil
}

func (p *talentDevelopmentSKPosting) Approved() error {
	return nil
}

func (p *talentDevelopmentSKPosting) Rejected() error {
	return nil
}

func (p *talentDevelopmentSKPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
			}

			if p.header.Type == hcmmodel.TalentDevelopmentSKTypeActing {
				notification.Menu = "Training Development SK Acting"
			} else {
				notification.Menu = "Training Development SK Tetap"
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

func (p *talentDevelopmentSKPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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

func (p *talentDevelopmentSKPosting) Preview() (*tenantcoremodel.PreviewReport, error) {
	preview := &tenantcoremodel.PreviewReport{
		Header: make(codekit.M),
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

	if p.header.Type == hcmmodel.TalentDevelopmentSKTypeActing {
		director := new(tenantcoremodel.Employee)
		err = p.opt.Db.GetByID(director, p.header.DirectorID)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error when get director: %s", err.Error())
		}

		directorDetail := new(bagongmodel.EmployeeDetail)
		err = p.opt.Db.GetByFilter(directorDetail, dbflex.Eq("EmployeeID", director.ID))
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("error when get director detail: %s", err.Error())
		}

		ids := []string{employeeDetail.Department,
			employeeDetail.Position,
			employeeDetail.Grade,
			employeeDetail.Group,
			employeeDetail.POH,
			employeeDetail.Level,
			employeeDetail.Gender,
			directorDetail.Position,
		}
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

		data := make([][]string, 17)
		for i := range data {
			data[i] = make([]string, 2)
		}
		data[0] = []string{"ID : ", p.header.ID}
		data[1] = []string{"Name : ", employee.Name}
		data[2] = []string{"Position : ", mapMaster[employeeDetail.Position]}
		data[3] = []string{"Department : ", mapMaster[employeeDetail.Department]}
		data[4] = []string{"Grade : ", mapMaster[employeeDetail.Grade]}
		data[5] = []string{"Level : ", mapMaster[employeeDetail.Level]}
		data[6] = []string{"Group : ", mapMaster[employeeDetail.Group]}
		data[7] = []string{"Employee No. : ", employeeDetail.EmployeeNo}
		data[8] = []string{"Identity Card No. : ", employeeDetail.IdentityCardNo}
		data[9] = []string{"Place of Birth, Date of Birth : ", fmt.Sprintf("%s, %s", employeeDetail.PlaceOfBirth, employeeDetail.DateOfBirth.Format("02 January 2006"))}
		data[10] = []string{"Gender : ", mapMaster[employeeDetail.Gender]}
		data[11] = []string{"Religion : ", employeeDetail.Religion}
		data[12] = []string{"Phone No. : ", employeeDetail.Phone}
		data[13] = []string{"Address : ", employeeDetail.Address}
		data[14] = []string{"Director Name : ", director.Name}
		data[15] = []string{"Director Position : ", mapMaster[directorDetail.Position]}
		data[16] = []string{"Director Address : ", directorDetail.Address}
		preview.Header["Data"] = data

		preview.Sections = make([]tenantcoremodel.PreviewSection, 2)

		section := tenantcoremodel.PreviewSection{
			SectionType: tenantcoremodel.PreviewAsGrid,
			Title:       "",
			Items:       make([][]string, len(p.header.Notices)+1),
		}

		section.Items[0] = []string{"Memperhatikan"}
		for i, d := range p.header.Notices {
			section.Items[i+1] = []string{d}
		}
		preview.Sections[0] = section

		section = tenantcoremodel.PreviewSection{
			SectionType: tenantcoremodel.PreviewAsGrid,
			Items:       make([][]string, len(p.header.Decides)+1),
		}

		section.Items[0] = []string{"Memutuskan"}
		for i, d := range p.header.Decides {
			section.Items[i+1] = []string{d}
		}
		preview.Sections[1] = section
	} else {
		data := make([][]string, 6)
		for i := range data {
			data[i] = make([]string, 2)
		}
		data[0] = []string{"ID : ", p.header.ID}
		data[1] = []string{"Name : ", employee.Name}
		data[2] = []string{"Employee No. : ", employeeDetail.EmployeeNo}
		data[3] = []string{"POH : ", employeeDetail.POH}
		data[3] = []string{"Joined Date : ", employee.JoinDate.Format("02 January 2006")}
		data[5] = []string{"Effective Promoted Date : ", p.header.EffectivePromotedDate.Format("02 January 2006")}
		preview.Header["Data"] = data
		detail, err := new(TalentDevelopmentHandler).GetDetail(p.ctx, &TalentDevelopmentGetDetailRequest{
			ID:         p.header.TalentDevelopmentID,
			EmployeeID: talentDev.EmployeeID,
		})
		if err != nil {
			return nil, fmt.Errorf("error when get talent development detail: %s", err.Error())
		}

		ids := []string{
			detail.Existing.Department,
			detail.Existing.Position,
			detail.Existing.Grade,
			detail.Existing.Group,
			detail.Existing.PointOfHire,
			detail.Existing.SubGroup,
			detail.NewPropose.Department,
			detail.NewPropose.Position,
			detail.NewPropose.Grade,
			detail.NewPropose.Group,
			detail.NewPropose.PointOfHire,
			detail.NewPropose.SubGroup,
		}

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

		sites := []bagongmodel.Site{}
		err = p.opt.Db.Gets(new(bagongmodel.Site), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", []string{detail.NewPropose.Site, detail.Existing.Site}...),
		), &sites)
		if err != nil {
			return nil, fmt.Errorf("error when get site: %s", err.Error())
		}

		mapSite := lo.Associate(sites, func(m bagongmodel.Site) (string, string) {
			return m.ID, m.Name
		})

		section := tenantcoremodel.PreviewSection{
			SectionType: tenantcoremodel.PreviewAsGrid,
			Title:       "Benefit Detail",
			Items:       make([][]string, 10),
		}

		section.Items[0] = []string{"Description", "Existing", "New Purpose"}
		section.Items[1] = []string{"Department", mapMaster[detail.Existing.Department], mapMaster[detail.NewPropose.Department]}
		section.Items[2] = []string{"Position", mapMaster[detail.Existing.Position], mapMaster[detail.NewPropose.Position]}
		section.Items[3] = []string{"Grade", mapMaster[detail.Existing.Grade], mapMaster[detail.NewPropose.Grade]}
		section.Items[4] = []string{"Group", mapMaster[detail.Existing.Group], mapMaster[detail.NewPropose.Group]}
		section.Items[5] = []string{"SubGroup", mapMaster[detail.Existing.SubGroup], mapMaster[detail.NewPropose.SubGroup]}
		section.Items[6] = []string{"Site", mapSite[detail.Existing.Site], mapSite[detail.NewPropose.Site]}
		section.Items[7] = []string{"PointOfHire", detail.Existing.PointOfHire, detail.NewPropose.PointOfHire}
		section.Items[8] = []string{"BasicSalary", codekit.ToString(detail.Existing.BasicSalary), codekit.ToString(detail.NewPropose.BasicSalary)}
		section.Items[9] = []string{"Allowance", detail.Existing.Allowance, detail.NewPropose.Allowance}
		preview.Sections = append(preview.Sections, section)
	}

	return preview, nil
}
