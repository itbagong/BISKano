package hcmlogic

import (
	"errors"
	"fmt"
	"io"
	"sort"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
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
	HandlePostRequest(ctx *kaos.Context, request *ficologic.PostRequest) (*tenantcoremodel.PreviewReport, error)
}

type PostingProvider[H orm.DataModel, L any] interface {
	Header() (H, *ficomodel.PostingProfile, error)
	Approved() error
	Rejected() error
	Post(opt ficologic.PostingHubExecOpt, header H) (string, error)
	SubmitNotification(pa *ficomodel.PostingApproval) error
	ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error
	Preview() (*tenantcoremodel.PreviewReport, error)
}

type PostingHub[H orm.DataModel, L any] struct {
	opt       ficologic.PostingHubExecOpt
	provider  PostingProvider[H, L]
	header    H
	dimension *tenantcoremodel.Dimension
	preview   *tenantcoremodel.PreviewReport
}

func NewPostingHub[O orm.DataModel, L any](pvd PostingProvider[O, L], createOpt ficologic.PostingHubCreateOpt) *PostingHub[O, L] {
	opt := ficologic.PostingHubExecOpt{
		PostingHubCreateOpt: createOpt,
	}

	ph := new(PostingHub[O, L])
	ph.opt = opt
	ph.provider = pvd
	return ph
}

func (p *PostingHub[H, L]) Header() (H, error) {
	var err error
	p.header, p.opt.PostingProfile, err = p.provider.Header()
	if err != nil {
		return p.header, fmt.Errorf("extract header: %s", err.Error())
	}

	var headerDimension *tenantcoremodel.Dimension
	reflector.From(p.header).GetTo("Dimension", &headerDimension)
	p.dimension = headerDimension
	return p.header, nil
}

func (p *PostingHub[H, L]) Status() string {
	status := ""
	reflector.From(p.header).GetTo("Status", &status)
	return status
}

func (p *PostingHub[H, L]) Preview() (*tenantcoremodel.PreviewReport, error) {
	var err error

	p.preview, err = p.provider.Preview()
	if err != nil {
		return nil, err
	}

	pv, _ := tenantcorelogic.GetPreviewBySource(p.opt.Db, p.opt.ModuleID, p.opt.JournalID,
		"", "Default")
	pv.PreviewReport = p.preview
	p.opt.Db.Save(pv)

	return p.preview, nil
}

func (p *PostingHub[H, L]) Submit() (*ficomodel.PostingApproval, string, error) {
	pa, isNew, err := GetOrCreatePostingApproval(p.opt.Db, p.opt.UserID, p.opt.CompanyID,
		p.opt.ModuleID, p.opt.JournalID,
		*p.opt.PostingProfile, p.dimension,
		true, true, "", 0)

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
	pa, err := ficologic.GetPostingApprovalBySource(p.opt.Db, p.opt.CompanyID, p.opt.ModuleID, p.opt.JournalID, true)
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
	p.provider.ApproveRejectNotification(pa, ficologic.PostOp(op))

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
		pa, err = ficologic.GetPostingApprovalBySource(p.opt.Db, p.opt.CompanyID, p.opt.ModuleID, p.opt.JournalID, false)

		if err != nil {
			return "", fmt.Errorf("posting approval: %s", err.Error())
		}
	} else {
		pa, _, err = GetOrCreatePostingApproval(p.opt.Db, p.opt.UserID, p.opt.CompanyID,
			p.opt.ModuleID, p.opt.JournalID,
			*p.opt.PostingProfile, p.dimension,
			true, true, "", 0)

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
	voucherNo, err := p.provider.Post(p.opt, p.header)
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

func (obj *PostingHub[H, L]) HandlePostRequest(ctx *kaos.Context, request *ficologic.PostRequest) (*tenantcoremodel.PreviewReport, error) {
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
		if (request.Op == ficologic.PostOpSubmit || request.Op == ficologic.PostOpApprove || request.Op == ficologic.PostOpPost) && perr == nil {
			new(tenantcoremodel.Log).Add(tenantcoremodel.LogParam{
				Hub:           obj.opt.Db,
				Menu:          request.JournalType.String(),
				Action:        string(request.Op),
				TransactionID: request.JournalID,
				UserLogin:     userID,
			})
		}
	}()

	_, err := obj.Header()
	if err != nil {
		return res, err
	}

	preview, err := obj.Preview()
	if err != nil {
		return res, err
	}

	if preview.Header == nil {
		preview.Header = codekit.M{}
	}

	op := string(request.Op)
	approveTxt := request.Text
	switch op {
	case string(ficologic.PostOpPreview):
		return preview, nil

	case string(ficologic.PostOpSubmit):
		_, _, err := obj.Submit()
		perr = err
		return preview, err

	case string(ficologic.PostOpApprove), string(ficologic.PostOpReject):
		_, _, err := obj.Approve(op, approveTxt)
		perr = err
		return preview, err

	case string(ficologic.PostOpPost):
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

func GetOrCreatePostingApproval(h *datahub.Hub, userID, companyID, sourceType, sourceID string,
	postingProfile ficomodel.PostingProfile, dim *tenantcoremodel.Dimension, openOnly, createIfNotFound bool,
	text string, amount float64) (*ficomodel.PostingApproval, bool, error) {
	pa, err := ficologic.GetPostingApprovalBySource(h, companyID, sourceType, sourceID, openOnly)
	if pa != nil && err == nil {
		return pa, false, nil
	}
	if err != io.EOF {
		return nil, false, err
	} else if !createIfNotFound {
		return nil, false, err
	}

	// get posting profile
	ppPics, _ := datahub.FindAnyByParm(h, new(ficomodel.PostingProfilePIC), new(ficomodel.PostingProfilePIC).TableName(),
		dbflex.
			NewQueryParam().
			SetWhere(dbflex.Eq("PostingProfileID", postingProfile.ID)).
			SetSort("Priority", "_id"))

	blockedPriorities := map[int]bool{}
	pics := []*ficomodel.PostingProfilePIC{}
	if dim != nil {
	iteratePicPP:
		for _, pic := range ppPics {
			if pic.Dimension.Compare(*dim) {
				blocked := blockedPriorities[pic.Priority]
				if blocked {
					continue iteratePicPP
				}
				pics = append(pics, pic)
				if pic.Exclusive {
					blockedPriorities[pic.Priority] = true
				}
			}
		}
	} else {
		pics = ppPics
	}

	if postingProfile.LimitSubmission {
		found := false
		for _, c := range extractSubmittersFromPics(pics) {
			if lo.IndexOf(c.UserIDs, userID) > 0 {
				found = true
				break
			}
		}
		if !found {
			return nil, false, fmt.Errorf("submission is not allowed: %s", postingProfile.Name)
		}
	}
	if postingProfile.NeedApproval {
		pa = &ficomodel.PostingApproval{
			CompanyID:        companyID,
			PostingProfileID: postingProfile.ID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			Status:           "PENDING",
			Approvers:        extractApproversFromPics(pics),
			Postingers:       extractPostingersFromPics(pics),
			Text:             text,
			Amount:           amount,
		}
	} else {
		pa = &ficomodel.PostingApproval{
			CompanyID:        companyID,
			PostingProfileID: postingProfile.ID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			Status:           "APPROVED",
			Postingers:       extractPostingersFromPics(pics),
			Text:             text,
			Amount:           amount,
		}
	}

	if dim != nil {
		pa.Dimension = *dim
	}

	pa.UpdateStage(h)

	if postingProfile.NeedApproval && len(pa.Approvals) == 0 {
		return nil, false, fmt.Errorf("no approval: %s", postingProfile.ID)
	}

	if err = h.Insert(pa); err != nil {
		return nil, false, fmt.Errorf("create approval error: %s, %s: %s", sourceType, sourceID, err.Error())
	}

	return pa, true, nil
}

func extractApproversFromPics(pics []*ficomodel.PostingProfilePIC) []ficomodel.PostingUsers {
	// remove redundant
	mapUserStage := map[string]int{}
	mapMinimalApproveCount := map[int]int{}
	for _, profilePic := range pics {
		for stage, app := range profilePic.Approvers {
			stage := stage + profilePic.Priority
			mapMinimalApproveCount[stage] = app.MinimalApproverCount

			for _, userid := range app.UserIDs {
				current := mapUserStage[userid]
				if stage > current {
					mapUserStage[userid] = stage
				}
			}
		}
	}

	// return approvers sorted by stage
	mapStageApprovers := map[int][]string{}
	for userid, stage := range mapUserStage {
		mapStageApprovers[stage] = append(mapStageApprovers[stage], userid)
	}
	sortedStage := lo.Keys(mapStageApprovers)
	sort.Ints(sortedStage)

	res := lo.Map(sortedStage, func(stage, index int) ficomodel.PostingUsers {
		return ficomodel.PostingUsers{
			UserIDs:              mapStageApprovers[stage],
			MinimalApproverCount: mapMinimalApproveCount[stage],
		}
	})

	return res
}

func extractPostingersFromPics(pics []*ficomodel.PostingProfilePIC) []ficomodel.PostingUsers {
	// remove redundant
	mapUserStage := map[string]int{}
	for _, profilePic := range pics {
		for stage, app := range profilePic.Postingers {
			for _, userid := range app.UserIDs {
				current := mapUserStage[userid]
				appStage := stage + profilePic.Priority
				if appStage > current {
					mapUserStage[userid] = appStage
				}
			}
		}
	}

	// return postingers sorted by stage
	mapStagePostingers := map[int][]string{}
	for userid, stage := range mapUserStage {
		mapStagePostingers[stage] = append(mapStagePostingers[stage], userid)
	}
	sortedStage := lo.Keys(mapStagePostingers)
	sort.Ints(sortedStage)

	res := lo.Map(sortedStage, func(stage, index int) ficomodel.PostingUsers {
		return ficomodel.PostingUsers{
			UserIDs: mapStagePostingers[stage],
		}
	})

	return res
}

func extractSubmittersFromPics(pics []*ficomodel.PostingProfilePIC) []ficomodel.PostingUsers {
	// remove redundant
	mapUserStage := map[string]int{}
	for _, profilePic := range pics {
		for stage, app := range profilePic.Submitters {
			for _, userid := range app.UserIDs {
				current := mapUserStage[userid]
				appStage := stage + profilePic.Priority
				if appStage > current {
					mapUserStage[userid] = appStage
				}
			}
		}
	}

	// return approvers sorted by stage
	mapStageSubmitters := map[int][]string{}
	for userid, stage := range mapUserStage {
		mapStageSubmitters[stage] = append(mapStageSubmitters[stage], userid)
	}
	sortedStage := lo.Keys(mapStageSubmitters)
	sort.Ints(sortedStage)

	res := lo.Map(sortedStage, func(stage, index int) ficomodel.PostingUsers {
		return ficomodel.PostingUsers{
			UserIDs: mapStageSubmitters[stage],
		}
	})

	return res
}
