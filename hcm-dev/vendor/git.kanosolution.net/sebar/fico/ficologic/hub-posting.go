package ficologic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type PostRequestHandler interface {
	HandlePostRequest(ctx *kaos.Context, request *PostRequest) (*tenantcoremodel.PreviewReport, error)
	Recalc(db *datahub.Hub, companyID string, request *PostRequest) (string, error)
}

type PostingProvider[H orm.DataModel, L any] interface {
	Header() (H, *ficomodel.PostingProfile, error)
	Lines() ([]L, error)
	ToJournalLines(opt PostingHubExecOpt, header H, lines []L) []ficomodel.JournalLine
	Calculate(opt PostingHubExecOpt, header H, lines []L) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error)
	Approved() error
	Rejected() error
	Post(opt PostingHubExecOpt, header H, lines []L, models map[string][]orm.DataModel) (string, error)
	GetAccount() string
	SubmitNotification(pa *ficomodel.PostingApproval) error
	ApproveRejectNotification(pa *ficomodel.PostingApproval, op PostOp) error
}

type PostingHubCreateOpt struct {
	Db        *datahub.Hub
	UserID    string
	CompanyID string
	JournalID string
	ModuleID  string
	Op        PostOp
}

type PostingHubExecOpt struct {
	PostingHubCreateOpt
	PostingProfile *ficomodel.PostingProfile
}

type PostingProfileApprovalDetail struct {
	AccountID string
	Text      string
	TrxDate   time.Time
	Amount    float64
}

type PostingHub[H orm.DataModel, L any] struct {
	opt      PostingHubExecOpt
	provider PostingProvider[H, L]

	headerCalc bool
	lineCalc   bool
	header     H
	lines      []L

	dimension tenantcoremodel.Dimension
	preview   *tenantcoremodel.PreviewReport
	trxs      map[string][]orm.DataModel

	approvalDetail *PostingProfileApprovalDetail
}

func NewPostingHub[O orm.DataModel, L any](pvd PostingProvider[O, L], createOpt PostingHubCreateOpt) *PostingHub[O, L] {
	opt := PostingHubExecOpt{
		PostingHubCreateOpt: createOpt,
	}

	ph := new(PostingHub[O, L])
	ph.opt = opt
	ph.provider = pvd
	return ph
}

func (p *PostingHub[H, L]) Header() (H, error) {
	var err error
	if !p.headerCalc {
		p.header, p.opt.PostingProfile, err = p.provider.Header()
		if err != nil {
			return p.header, fmt.Errorf("extract header: %s", err.Error())
		}
		p.headerCalc = true

		var headerDimension tenantcoremodel.Dimension
		err = reflector.From(p.header).GetTo("Dimension", &headerDimension)
		if err != nil {
			return p.header, errors.New("fail extract dimension from " + p.header.TableName())
		}
		p.dimension = headerDimension

		p.approvalDetail = &PostingProfileApprovalDetail{}
		var trxDate time.Time
		err = reflector.From(p.header).GetTo("TrxDate", &trxDate)
		if err != nil {
			return p.header, errors.New("fail extract transaction date from " + p.header.TableName())
		}
		p.approvalDetail.TrxDate = trxDate

		var text string
		err = reflector.From(p.header).GetTo("Text", &text)
		if err != nil {
			return p.header, errors.New("fail extract text from " + p.header.TableName())
		}
		p.approvalDetail.Text = text

		p.approvalDetail.AccountID = p.provider.GetAccount()
	}
	return p.header, nil
}

func (p *PostingHub[H, L]) Lines() ([]L, error) {
	var err error
	if !p.lineCalc {
		p.lines, err = p.provider.Lines()
		if err != nil {
			return p.lines, fmt.Errorf("extract header: %s", err.Error())
		}
		p.lineCalc = true
	}

	errorTexts := []string{}
	for index, line := range p.lines {
		var dim tenantcoremodel.Dimension
		rl := reflector.From(&line)
		err := rl.GetTo("Dimension", &dim)
		if err != nil {
			continue
		}
		for _, d := range dim {
			if d.Value == "" && d.Key != "Asset" {
				errorTexts = append(errorTexts, fmt.Sprintf("missing: line %d dimension %s", index+1, d.Key))
			}
		}
	}
	if len(errorTexts) > 0 {
		return p.lines, fmt.Errorf(strings.Join(errorTexts, " | "))
	}
	return p.lines, nil
}

func (p *PostingHub[H, L]) Status() string {
	status := ""
	reflector.From(p.header).GetTo("Status", &status)
	return status
}

func (p *PostingHub[H, L]) Calculate() error {
	var err error

	if len(p.lines) == 0 {
		return errors.New("no lines")
	}

	p.preview, p.trxs, p.approvalDetail.Amount, err = p.provider.Calculate(p.opt, p.header, p.lines)
	if err != nil {
		return err
	}

	ledgerTrxs := FromDataModels(p.trxs[new(ficomodel.LedgerTransaction).TableName()], new(ficomodel.LedgerTransaction))
	if len(ledgerTrxs) > 0 {
		totalAmt := lo.SumBy(ledgerTrxs, func(l *ficomodel.LedgerTransaction) float64 {
			return l.Amount
		})
		if totalAmt != 0 {
			return fmt.Errorf("total ledger transaction amount is %.2f (not balance)", totalAmt)
		}
	}

	pv, _ := tenantcorelogic.GetPreviewBySource(p.opt.Db, p.opt.ModuleID, p.opt.JournalID,
		"", "Default")
	pv.PreviewReport = p.preview
	p.opt.Db.Save(pv)

	return nil
}

func (p *PostingHub[H, L]) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *PostingHub[H, L]) Transactions(name string) ([]orm.DataModel, error) {
	trxs, ok := p.trxs[name]
	if !ok {
		return nil, fmt.Errorf("missing: %s", name)
	}
	return trxs, nil
}

func (p *PostingHub[H, L]) Submit() (*ficomodel.PostingApproval, string, error) {
	pa, isNew, err := GetOrCreatePostingApproval(p.opt.Db, p.opt.UserID, p.opt.CompanyID,
		p.opt.ModuleID, p.opt.JournalID,
		*p.opt.PostingProfile, p.dimension,
		true, true, p.provider.ToJournalLines(p.opt, p.header, p.lines),
		p.approvalDetail.AccountID, p.approvalDetail.Text, p.approvalDetail.TrxDate, p.approvalDetail.Amount)

	if err != nil {
		return nil, "", err
	}

	if pa == nil {
		return nil, "", errors.New("no posting approval")
	}

	if p.opt.PostingProfile.NeedApproval {
		if err != nil {
			return nil, "", fmt.Errorf("create approval: %s", err.Error())
		}
		if !isNew {
			return nil, "", fmt.Errorf("duplicate: approval: %s, %s", ficomodel.SubledgerAccounting, p.opt.JournalID)
		}
		reflector.From(p.header).Set("Status", ficomodel.JournalStatusSubmitted).Flush()
		p.opt.Db.Save(p.header)
		p.opt.Db.Update(p.header, "Status")

		// create notification
		p.provider.SubmitNotification(pa)
	} else {
		if vch, err := p.markAsReady(); err != nil {
			return nil, "", err
		} else {
			return pa, vch, nil
		}
	}
	return pa, "", nil
}

func (p *PostingHub[H, L]) Approve(op, txt string) (string, string, error) {
	pa, err := GetPostingApprovalBySource(p.opt.Db, p.opt.CompanyID, p.opt.ModuleID, p.opt.JournalID, true)
	if err != nil {
		return "", "", fmt.Errorf("posting approval: %s", err.Error())
	}

	if err = pa.UpdateApproval(p.opt.Db, p.opt.UserID, op, txt); err != nil {
		return pa.Status, "", fmt.Errorf("posting approval: %s", err.Error())
	}
	if err = p.opt.Db.Save(pa); err != nil {
		return pa.Status, "", fmt.Errorf("posting approval save: %s", err.Error())
	}

	// create notification
	p.provider.ApproveRejectNotification(pa, PostOp(op))

	switch ficomodel.JournalStatus(pa.Status) {
	case ficomodel.JournalStatusRejected:
		reflector.From(p.header).Set("Status", ficomodel.JournalStatusRejected).Flush()
		p.opt.Db.Save(p.header)
		p.opt.Db.Update(p.header, "Status")

	case ficomodel.JournalStatusApproved:
		if vch, err := p.markAsReady(); err != nil {
			return pa.Status, vch, err
		}
	}

	return pa.Status, "", nil
}

func (p *PostingHub[H, L]) Post() (string, error) {
	var (
		pa  *ficomodel.PostingApproval
		err error
	)
	if p.opt.PostingProfile.NeedApproval {
		pa, err = GetPostingApprovalBySource(p.opt.Db, p.opt.CompanyID, p.opt.ModuleID, p.opt.JournalID, false)

		if err != nil {
			return "", fmt.Errorf("posting approval: %s", err.Error())
		}
	} else {
		pa, _, err = GetOrCreatePostingApproval(p.opt.Db, p.opt.UserID, p.opt.CompanyID,
			p.opt.ModuleID, p.opt.JournalID,
			*p.opt.PostingProfile, p.dimension,
			true, true, p.provider.ToJournalLines(p.opt, p.header, p.lines),
			p.approvalDetail.AccountID, p.approvalDetail.Text, p.approvalDetail.TrxDate, p.approvalDetail.Amount)

		if err != nil {
			return "", err
		}
	}
	if pa.Status != "APPROVED" {
		return "", fmt.Errorf("invalid: posting approval status: %s: %s", pa.ID, pa.Status)
	}

	if !p.opt.PostingProfile.DirectPosting {
		usersCanPost := []string{}
		for _, postinger := range pa.Postingers {
			usersCanPost = append(usersCanPost, postinger.UserIDs...)
		}
		if !codekit.HasMember(usersCanPost, p.opt.UserID) {
			return "", fmt.Errorf("no access: %s: posting", p.opt.UserID)
		}
	}

	return p.PostJournal()
}

func (p *PostingHub[H, L]) PostJournal() (string, error) {
	voucherNo, err := p.provider.Post(p.opt, p.header, p.lines, p.trxs)
	if err != nil {
		return "", err
	}

	reflector.From(p.header).Set("Status", ficomodel.JournalStatusPosted)
	p.opt.Db.Save(p.header)
	p.opt.Db.Update(p.header, "Status")

	pv, _ := tenantcorelogic.GetPreviewBySource(p.opt.Db, p.opt.ModuleID, p.opt.JournalID,
		"", "Default")
	if p.preview == nil {
		p.preview = new(tenantcoremodel.PreviewReport)
	}
	if p.preview.Header == nil {
		p.preview.Header = codekit.M{}
	}
	p.preview.Header.Set("VoucherNo", voucherNo)
	pv.VoucherNo = voucherNo
	pv.PreviewReport = p.preview
	p.opt.Db.Save(pv)

	return voucherNo, nil
}

func (p *PostingHub[H, L]) markAsReady() (string, error) {
	if err := reflector.From(p.header).
		Set("Status", ficomodel.JournalStatusReady).
		Flush(); err != nil {
		return "", err
	}
	if err := p.opt.Db.Save(p.header); err != nil {
		return "", err
	}
	p.opt.Db.Update(p.header, "Status")

	if err := p.provider.Approved(); err != nil {
		return "", err
	}

	// insert approval log
	p.opt.Db.Insert(&ficomodel.PostingProfileApprovalLog{
		PostingProfile: p.opt.PostingProfile,
		Journal:        p.header,
		Action:         string(p.opt.PostingHubCreateOpt.Op),
	})

	if p.opt.PostingProfile.DirectPosting {
		if vch, postError := p.Post(); postError != nil {
			return "", postError
		} else {
			return vch, nil
		}
	}

	return "", nil
}

func (obj *PostingHub[H, L]) HandlePostRequest(ctx *kaos.Context, request *PostRequest) (*tenantcoremodel.PreviewReport, error) {
	var res *tenantcoremodel.PreviewReport
	var perr error

	if ctx == nil {
		return nil, errors.New("ctx is nil")
	}

	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	defer func() {
		interfaceText, _ := reflector.From(obj.header).Get("Text")
		text := ""
		if interfaceText != nil {
			text = interfaceText.(string)
		}
		if (request.Op == PostOpSubmit || request.Op == PostOpApprove || request.Op == PostOpPost) && perr == nil {
			new(tenantcoremodel.Log).Add(tenantcoremodel.LogParam{
				Hub:           obj.opt.Db,
				Menu:          request.JournalType.String(),
				Action:        string(request.Op),
				TransactionID: request.JournalID,
				Name:          text, // TODO: ganti dengan journal name tp masih bingung gimana dapetinnya, sudah diganti perlu dicheck apakah sudah benar
				UserLogin:     userID,
			})
		}
	}()

	_, err := obj.Header()
	if err != nil {
		return res, err
	}

	_, err = obj.Lines()
	if err != nil {
		return res, err
	}

	err = obj.Calculate()
	if err != nil {
		return res, fmt.Errorf("calculate: %s", err.Error())
	}

	preview := obj.Preview()
	if preview.Header == nil {
		preview.Header = codekit.M{}
	}

	op := string(request.Op)
	approveTxt := request.Text
	switch op {
	case string(PostOpPreview):
		return preview, nil

	case string(PostOpSubmit):
		_, vch, err := obj.Submit()
		preview.Header.Set("VoucherNo", vch)
		perr = err
		return preview, err

	case string(PostOpApprove), string(PostOpReject):
		_, vch, err := obj.Approve(op, approveTxt)
		preview.Header.Set("VoucherNo", vch)
		perr = err
		return preview, err

	case string(PostOpPost):
		vch, err := obj.Post()
		preview.Header.Set("VoucherNo", vch)
		perr = err
		return preview, err

	default:
		if op == "" {
			op = "BLANK"
		}
		perr = err
		return res, fmt.Errorf("invalid: op %s", op)
	}
}
