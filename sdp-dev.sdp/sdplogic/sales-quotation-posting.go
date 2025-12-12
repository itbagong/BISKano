package sdplogic

import (
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type salesQuotationPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *sdpmodel.SalesQuotation
	trxType string

	// lines      []ficomodel.JournalLine
	// inventTrxs *ficomodel.CustomerJournal
	items *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewSalesQuotationPosting(opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*sdpmodel.SalesQuotation, sdpmodel.SalesQuotationLine] {
	p := new(salesQuotationPosting)
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*sdpmodel.SalesQuotation, sdpmodel.SalesQuotationLine](p) //-------
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *salesQuotationPosting) Header() (*sdpmodel.SalesQuotation, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(sdpmodel.SalesQuotation), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(sdpmodel.SalesOrderJournalType), j.JournalType)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalType, err.Error())
	}

	p.trxType = string(jt.TrxType)
	if p.trxType == "" {
	}

	j.PostingProfileID = tenantcorelogic.TernaryString(j.PostingProfileID, jt.PostingProfileID)
	if j.PostingProfileID == "" {
		return nil, nil, fmt.Errorf("missing: posting profile")
	}

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	p.header = j
	return j, pp, nil
}

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *salesQuotationPosting) Lines() ([]sdpmodel.SalesQuotationLine, error) {
	// rils := []ficomodel.JournalLine{}
	Lines := []sdpmodel.SalesQuotationLine{}
	for _, line := range p.header.Lines {
		Lines = append(Lines, line)
	}

	return Lines, nil
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *salesQuotationPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *sdpmodel.SalesQuotation, lines []sdpmodel.SalesQuotationLine) []ficomodel.JournalLine {
	// return receiveIssuelineToficoLines(p.lines)
	return lo.Map(lines, func(line sdpmodel.SalesQuotationLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		return *jl
	})
}

func (p *salesQuotationPosting) Calculate(opt ficologic.PostingHubExecOpt, header *sdpmodel.SalesQuotation, lines []sdpmodel.SalesQuotationLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}

	// TODO: set preview (belum nemu contoh)
	// TODO: bener ga kalo disini kita pake inventTrxs instead of ledgTrxs (fixo LedgerTransaction)

	return &preview, trxs, 0.0, nil
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *salesQuotationPosting) Post(opt ficologic.PostingHubExecOpt, header *sdpmodel.SalesQuotation, lines []sdpmodel.SalesQuotationLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	db, _ = p.opt.Db.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	header.Status = ficomodel.JournalStatusPosted
	err = db.Save(header)

	return res, err
}

func (p *salesQuotationPosting) Approved() error {
	return nil
}

func (p *salesQuotationPosting) Rejected() error {
	return nil
}

func (p *salesQuotationPosting) GetAccount() string {
	return ""
}

func (p *salesQuotationPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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
				JournalType:              p.header.JournalType,
				PostingProfileApprovalID: pa.ID,
				TrxDate:                  p.header.TrxDate,
				Text:                     p.header.Text,
				UserTo:                   app.UserID,
				Menu:                     "Sales Quotation",
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
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

func (p *salesQuotationPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
