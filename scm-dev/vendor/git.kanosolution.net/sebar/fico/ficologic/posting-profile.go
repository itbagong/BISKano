package ficologic

import (
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

func GetPostingProfile(h *datahub.Hub, id string) *ficomodel.PostingProfile {
	// TODO: read from redis memory using kiva
	profile, err := datahub.GetByID(h, new(ficomodel.PostingProfile), id)
	if err != nil {
		return nil
	}
	return profile
}

type PostingRole string

const (
	PostingPICApprover  string = "APPROVER"
	PostingPICPostinger string = "POSTINGER"
)

type SourceURLRequest struct {
	SourceType string
	SourceID   string
}

type SourceURL struct {
	Menu      string
	URL       string
	JournalID string // used in mfg only
}

type PostingProfilePICs struct {
	Submitters []ficomodel.PostingUsers
	Postingers []ficomodel.PostingUsers
	Approvers  []ficomodel.PostingUsers
}

func GetPostingProfilePICByJournalLine(h *datahub.Hub, postingProfileID string, dim tenantcoremodel.Dimension, lines []ficomodel.JournalLine) PostingProfilePICs {
	pics := GetPostingSetupPICByJournalLine(h, postingProfileID, dim, lines)

	res := PostingProfilePICs{}
	res.Submitters = extractSubmittersFromPics(h, pics)
	res.Approvers = extractApproversFromPics(h, pics)
	res.Postingers = extractPostingersFromPics(h, pics)
	return res
}

func extractApproversFromPics(db *datahub.Hub, pics []*ficomodel.PostingProfilePIC) []ficomodel.PostingUsers {
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

			for _, groupid := range app.GroupIds {
				users := getUsersFromGroup(db, groupid)
				for _, userid := range users {
					current := mapUserStage[userid]
					if stage > current {
						mapUserStage[userid] = stage
					}
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

func GetApproversByJournalLine(h *datahub.Hub, postingProfileID string, dim tenantcoremodel.Dimension, lines []ficomodel.JournalLine) []ficomodel.PostingUsers {
	pics := GetPostingSetupPICByJournalLine(h, postingProfileID, dim, lines)
	return extractApproversFromPics(h, pics)
}

func extractPostingersFromPics(db *datahub.Hub, pics []*ficomodel.PostingProfilePIC) []ficomodel.PostingUsers {
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

			for _, groupid := range app.GroupIds {
				users := getUsersFromGroup(db, groupid)
				for _, userid := range users {
					current := mapUserStage[userid]
					appStage := stage + profilePic.Priority
					if appStage > current {
						mapUserStage[userid] = appStage
					}
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

func GetPostingerByJournalLine(h *datahub.Hub, postingProfileID string, dim tenantcoremodel.Dimension, lines []ficomodel.JournalLine) []ficomodel.PostingUsers {
	pics := GetPostingSetupPICByJournalLine(h, postingProfileID, dim, lines)
	return extractPostingersFromPics(h, pics)
}

func extractSubmittersFromPics(db *datahub.Hub, pics []*ficomodel.PostingProfilePIC) []ficomodel.PostingUsers {
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

			for _, groupid := range app.GroupIds {
				users := getUsersFromGroup(db, groupid)
				for _, userid := range users {
					current := mapUserStage[userid]
					appStage := stage + profilePic.Priority
					if appStage > current {
						mapUserStage[userid] = appStage
					}
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

func GetSubmitterByJournalLine(h *datahub.Hub, postingProfileID string, dim tenantcoremodel.Dimension, lines []ficomodel.JournalLine) []ficomodel.PostingUsers {
	pics := GetPostingSetupPICByJournalLine(h, postingProfileID, dim, lines)

	return extractSubmittersFromPics(h, pics)
}

func GetPostingSetupPICByJournalLine(h *datahub.Hub, postingProfileID string, dim tenantcoremodel.Dimension, lines []ficomodel.JournalLine) []*ficomodel.PostingProfilePIC {
	res := []*ficomodel.PostingProfilePIC{}

	pics, _ := datahub.FindAnyByParm(h, new(ficomodel.PostingProfilePIC), new(ficomodel.PostingProfilePIC).TableName(),
		dbflex.
			NewQueryParam().
			SetWhere(dbflex.Eq("PostingProfileID", postingProfileID)).
			SetSort("Priority", "_id"))

	blockedPriorities := map[int]bool{}
	picsWithDim := []*ficomodel.PostingProfilePIC{}
iteratePicPP:
	for _, pic := range pics {
		if pic.Dimension.Compare(dim) {
			blocked := blockedPriorities[pic.Priority]
			if blocked {
				continue iteratePicPP
			}
			picsWithDim = append(picsWithDim, pic)
			if pic.Exclusive {
				blockedPriorities[pic.Priority] = true
			}
		}
	}

	type accountAmount struct {
		TotalAmount float64
		MaxPrice    float64
	}

	// group line by their account and offset account
	mapLines := lo.MapEntries(
		lo.GroupBy(lines, func(item ficomodel.JournalLine) ficomodel.SubledgerAccount {
			return item.Account
		}), func(k ficomodel.SubledgerAccount, v []ficomodel.JournalLine) (ficomodel.SubledgerAccount, accountAmount) {
			am := accountAmount{}
			am.TotalAmount = lo.SumBy(v, func(i ficomodel.JournalLine) float64 {
				return i.Amount
			})
			prices := lo.Map(v, func(l ficomodel.JournalLine, _ int) float64 {
				if l.Qty == 0 {
					l.Qty = 1
				}
				return l.Amount / l.Qty
			})
			if len(prices) > 0 {
				sort.Float64s(prices)
				am.MaxPrice = prices[len(prices)-1]
			}
			return k, am
		})

	mapOffsetLines := lo.MapEntries(lo.GroupBy(lo.Filter(lines, func(item ficomodel.JournalLine, index int) bool {
		return item.OffsetAccount.AccountID != ""
	}), func(item ficomodel.JournalLine) ficomodel.SubledgerAccount {
		return item.OffsetAccount
	}), func(k ficomodel.SubledgerAccount, v []ficomodel.JournalLine) (ficomodel.SubledgerAccount, accountAmount) {
		am := accountAmount{}
		am.TotalAmount = lo.SumBy(v, func(i ficomodel.JournalLine) float64 {
			return i.Amount
		})
		prices := lo.Map(v, func(l ficomodel.JournalLine, _ int) float64 {
			if l.Qty == 0 {
				l.Qty = 1
			}
			return l.Amount / l.Qty
		})
		if len(prices) == 0 {
			sort.Float64s(prices)
			am.MaxPrice = prices[len(prices)-1]
		}
		return k, am
	})

	for k, m := range mapOffsetLines {
		_, has := mapLines[k]
		if !has {
			mapLines[k] = m
		}
	}

	// find the pic relevant to account
	for k, line := range mapLines {
	iteratePIC:
		for _, pic := range picsWithDim {
			// check account
			if pic.Account.AccountType != "" {
				if !(pic.Account.AccountType == k.AccountType && pic.Account.IsValid(k.AccountID)) {
					continue
				}
			}

			// check range
			if pic.UseRange {
				var baseAmount float64

				switch pic.AmountType {
				case ficomodel.AmountMaxPricePerUnit:
					baseAmount = line.MaxPrice

				default:
					baseAmount = line.TotalAmount
				}

				if !(pic.HiAmount >= baseAmount && pic.LowAmount <= baseAmount) {
					continue iteratePIC
				}
			}

			res = append(res, pic)
		}
	}

	return res
}

func GetOrCreatePostingApproval(h *datahub.Hub, userID, companyID, sourceType, sourceID string,
	postingProfile ficomodel.PostingProfile, dim tenantcoremodel.Dimension, openOnly, createIfNotFound bool, lines []ficomodel.JournalLine,
	accountID, text string, trxDate time.Time, amount float64) (*ficomodel.PostingApproval, bool, error) {
	pa, err := GetPostingApprovalBySource(h, companyID, sourceType, sourceID, openOnly)
	if pa != nil && err == nil {
		return pa, false, nil
	}
	if err != io.EOF {
		return nil, false, err
	} else if !createIfNotFound {
		return nil, false, err
	}
	pics := GetPostingProfilePICByJournalLine(h, postingProfile.ID, dim, lines)
	if postingProfile.LimitSubmission {
		found := false
		for _, c := range pics.Submitters {
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
			Approvers:        pics.Approvers,
			Postingers:       pics.Postingers,
			Dimension:        dim,
			AccountID:        accountID,
			Text:             text,
			TrxDate:          trxDate,
			Amount:           amount,
		}
	} else {
		pa = &ficomodel.PostingApproval{
			CompanyID:        companyID,
			PostingProfileID: postingProfile.ID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			Status:           "APPROVED",
			Postingers:       pics.Postingers,
			Dimension:        dim,
			AccountID:        accountID,
			Text:             text,
			TrxDate:          trxDate,
			Amount:           amount,
		}
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

func GetPostingApprovalBySource(h *datahub.Hub, companyID, sourceType, sourceID string, openOnly bool) (*ficomodel.PostingApproval, error) {
	approvalFilter := dbflex.Eqs("CompanyID", companyID, "SourceType", sourceType, "SourceID", sourceID)
	pa, err := datahub.GetByParm(h, new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(approvalFilter).SetSort("-Created"))

	approvalFilterPending := dbflex.Eqs("CompanyID", companyID, "SourceType", sourceType, "SourceID", sourceID, "Status", "PENDING")
	_, err2 := datahub.GetByParm(h, new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(approvalFilterPending).SetSort("-Created"))
	found := false

	// fmt.Println(companyID, " | ", sourceType, " | ", sourceID, " | ", openOnly, ">", err, "|", err2)

	if err != nil {
		if err != io.EOF {
			return nil, err
		}
	} else if openOnly && pa.Status == "PENDING" {
		found = true
	} else if openOnly && pa.Status == "REJECTED" && err2 != io.EOF {
		found = true
	} else if !openOnly {
		found = true
	}
	if found {
		return pa, nil
	}
	return nil, io.EOF
}

func FindPostingApprovalByUserID(db *datahub.Hub, userID string) ([]*ficomodel.PostingProfile, error) {
	publics, _ := datahub.FindByFilter(db, new(ficomodel.PostingProfile), dbflex.Eq("LimitSubmission", false))
	submitters, _ := datahub.FindByFilter(db, new(ficomodel.PostingProfilePIC), dbflex.Eq("Submitters.UserIDs", userID))

	publicIds := lo.Map(publics, func(s *ficomodel.PostingProfile, index int) interface{} {
		return s.ID
	})

	byUserIDs := lo.Map(submitters, func(s *ficomodel.PostingProfilePIC, index int) interface{} {
		return s.PostingProfileID
	})

	for _, byUserID := range byUserIDs {
		if !lo.Contains(publicIds, byUserID) {
			publicIds = append(publicIds, byUserID)
		}
	}

	res, _ := datahub.Find(db, new(ficomodel.PostingProfile), dbflex.NewQueryParam().SetSort("_id").SetWhere(
		dbflex.In("_id", publicIds...)))
	return res, nil
}

func FindPostingApprovalBySourceUserID(db *datahub.Hub, userID, companyID string, req *GetPostingApprovalBySourceUserRequest, found bool) (PostingApprovalBySourceUserResponse, error) {
	res := PostingApprovalBySourceUserResponse{
		Approvers:  false,
		Submitters: false,
		Postingers: false,
	}

	approvalFilter := dbflex.Eqs("CompanyID", companyID, "SourceType", req.JournalType, "SourceID", req.JournalID)
	pa, err := datahub.GetByParm(db, new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(approvalFilter).SetSort("-Created"))
	openOnly := false

	if err != nil {
		if err != io.EOF {
			return res, err
		}
	} else if openOnly && pa.Status == "PENDING" {
		found = true
	} else if !openOnly {
		found = true
	}

	if found {
		for i, c := range pa.Approvers {
			// current stage start from 1 and index start from (0+1)
			if pa.CurrentStage == (i + 1) {
				if lo.IndexOf(c.UserIDs, userID) >= 0 {
					res.Approvers = true
					break
				}
			}
		}
		for _, c := range pa.Submitters {
			if lo.IndexOf(c.UserIDs, userID) >= 0 {
				res.Submitters = true
				break
			}
		}
		for _, c := range pa.Postingers {
			if lo.IndexOf(c.UserIDs, userID) >= 0 {
				res.Postingers = true
				break
			}
		}
		// check if already approved
		for _, c := range pa.Approvals {
			if c.UserID == userID && c.Status == string(ficomodel.JournalStatusApproved) {
				res.Approvers = false
				break
			}
		}
	}

	return res, nil
}

func getUsersFromGroup(_ *datahub.Hub, _ string) []string {
	res := []string{}

	return res
}

func SyncPostingApprovalBySource(db *datahub.Hub, coID, sourceType, sourceID string, lines []ficomodel.JournalLine) ([]*ficomodel.PostingApproval, error) {
	eqs := dbflex.Eqs("CompanyID", coID, "SourceType", sourceType, "SourceID", sourceID)
	parm := dbflex.NewQueryParam().
		SetWhere(eqs).
		SetGroupBy("CompanyID", "SourceType", "SourceID").
		SetAggr(dbflex.NewAggrItem("Version", dbflex.AggrMax, "Version"))
	groupByVersion := []struct {
		ID struct {
			CompanyID  string
			SourceType string
			SourceID   string
		} `json:"_id" bson:"_id"`
		Version int
	}{}
	db.PopulateByParm(new(ficomodel.PostingApproval).TableName(), parm, &groupByVersion)

	mapPP := sebar.NewMapRecordWithORM(db, new(ficomodel.PostingProfile))
	res := []*ficomodel.PostingApproval{}
	for _, gbv := range groupByVersion {
		eqs = dbflex.Eqs("CompanyID", coID, "SourceType", sourceType, "SourceID", sourceID, "Version", gbv.Version)
		ppa, err := datahub.GetByFilter(db, new(ficomodel.PostingApproval), eqs)
		if err != nil {
			ficoconfig.Config.Log.Errorf("missing: posting approva: %v: %s", ppa, err.Error())
			continue
		}

		pp, err := mapPP.Get(ppa.PostingProfileID)
		if err != nil {
			ficoconfig.Config.Log.Errorf("missing: posting profile: %s: %s", ppa.PostingProfileID, err.Error())
			continue
		}

		err = syncPostingApproval(db, ppa, pp, lines)
		if err == nil {
			res = append(res, ppa)
		}
	}

	return res, nil
}

func syncPostingApproval(db *datahub.Hub, ppa *ficomodel.PostingApproval,
	pp *ficomodel.PostingProfile, lines []ficomodel.JournalLine) error {
	if ppa == nil {
		return errors.New("missing: posting approval")
	}

	if pp == nil {
		return errors.New("missing: posting profile")
	}

	if e := initPostingApprovalFromProfile(db, ppa, pp, "", false, ppa.Dimension, lines, ppa.CompanyID, ppa.SourceType, ppa.SourceID); e != nil {
		return fmt.Errorf("fail sync posting profile: %s", e.Error())
	}

	err := db.Save(ppa)
	if err != nil {
		return fmt.Errorf("save fail: posting profile: %s: %s", ppa.ID, err.Error())
	}
	return nil
}

func initPostingApprovalFromProfile(h *datahub.Hub, pa *ficomodel.PostingApproval, postingProfile *ficomodel.PostingProfile,
	userID string, isCreate bool, dim tenantcoremodel.Dimension, lines []ficomodel.JournalLine,
	companyID, sourceType, sourceID string) error {
	pics := GetPostingProfilePICByJournalLine(h, postingProfile.ID, dim, lines)
	if isCreate && postingProfile.LimitSubmission {
		found := false
		for _, c := range pics.Submitters {
			if lo.IndexOf(c.UserIDs, userID) > 0 {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("submission is not allowed: %s", postingProfile.Name)
		}
	}
	if postingProfile.NeedApproval {
		pa = &ficomodel.PostingApproval{
			CompanyID:        companyID,
			PostingProfileID: postingProfile.ID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			Status:           "PENDING",
			Approvers:        pics.Approvers,
			Postingers:       pics.Postingers,
			Dimension:        dim,
			/*
				AccountID:        accountID,
				Text:             text,
				TrxDate:          trxDate,
				Amount:           amount,
			*/
		}
	} else {
		pa = &ficomodel.PostingApproval{
			CompanyID:        companyID,
			PostingProfileID: postingProfile.ID,
			SourceType:       sourceType,
			SourceID:         sourceID,
			Status:           "APPROVED",
			Postingers:       pics.Postingers,
			Dimension:        dim,
			/*
				AccountID:        accountID,
				Text:             text,
				TrxDate:          trxDate,
				Amount:           amount,
			*/
		}
	}
	return nil
}

func IsLastApprover(ppa *ficomodel.PostingApproval, userID string) bool {
	approvers := ppa.Approvers[len(ppa.Approvers)-1].UserIDs
	return lo.IndexOf(approvers, userID) >= 0
}
