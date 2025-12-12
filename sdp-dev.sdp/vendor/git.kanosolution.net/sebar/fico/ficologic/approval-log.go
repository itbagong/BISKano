package ficologic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type ApprovalLogHandler struct {
}

type ApprovalLogHandlerGetRequest struct {
	ID string
}

type ApprovalLogHandlerGetResponse struct {
	Status     string
	Text       string
	Reason     string
	Date       *time.Time
	IsApproved bool
}

func (m *ApprovalLogHandler) Get(ctx *kaos.Context, payload *ApprovalLogHandlerGetRequest) ([]ApprovalLogHandlerGetResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	return GetLogApprovalBySourceID(h, payload.ID)
}

func GetLogApprovalBySourceID(h *datahub.Hub, sourceID string) ([]ApprovalLogHandlerGetResponse, error) {
	profile := new(ficomodel.PostingApproval)
	err := h.GetByParm(profile, dbflex.NewQueryParam().SetWhere(dbflex.Eq("SourceID", sourceID)).SetSort("-Created"))
	if err != nil {
		return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
	}

	userIDs := []string{}
	for _, approver := range profile.Approvers {
		userIDs = append(userIDs, approver.UserIDs...)
	}

	users := []tenantcoremodel.Employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", userIDs...)), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get user: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(user tenantcoremodel.Employee) (string, string) {
		return user.ID, user.Name
	})

	result := []ApprovalLogHandlerGetResponse{}
	for _, approver := range profile.Approvers {
		for _, id := range approver.UserIDs {
			user := mapUser[id]
			app := ApprovalLogHandlerGetResponse{}
			isExist := false

			for _, approval := range profile.Approvals {
				if id == approval.UserID {
					app.Text = fmt.Sprintf("Waiting Approval from %s", user)
					app.Status = "PENDING"

					if approval.Status == "APPROVED" && approval.Confirmed != nil {
						app.Text = fmt.Sprintf("Approved By %s", user)
						app.Status = approval.Status
						app.IsApproved = true
					} else if approval.Status == "REJECTED" && approval.Confirmed != nil {
						app.Text = fmt.Sprintf("Rejected By %s", user)
						app.Status = approval.Status
					} else if approval.Status == "PENDING" && approval.Confirmed != nil {
						app.Text = fmt.Sprintf("Waiting Approval from %s", user)
						app.Status = approval.Status
					}

					app.Reason = approval.Text
					app.Date = approval.Confirmed

					isExist = true
					break
				}
			}

			if !isExist {
				app.Status = "PENDING"
				app.Text = fmt.Sprintf("Waiting Approval from %s", user)
			}

			result = append(result, app)
		}
	}

	return result, nil
}

func FindNextApproval(h *datahub.Hub, sourceID string) ([]ApprovalLogHandlerGetResponse, error) {
	profile := new(ficomodel.PostingApproval)
	err := h.GetByParm(profile, dbflex.NewQueryParam().SetWhere(dbflex.Eq("SourceID", sourceID)).SetSort("-Created"))
	if err != nil {
		return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
	}

	userIDs := []string{}
	for _, approver := range profile.Approvers {
		userIDs = append(userIDs, approver.UserIDs...)
	}

	users := []tenantcoremodel.Employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", userIDs...)), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get user: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(user tenantcoremodel.Employee) (string, string) {
		return user.ID, user.Name
	})

	result := []ApprovalLogHandlerGetResponse{}
	for _, approver := range profile.Approvers {
		for _, id := range approver.UserIDs {
			user := mapUser[id]
			app := ApprovalLogHandlerGetResponse{}
			isExist := false

			for _, approval := range profile.Approvals {
				if id == approval.UserID {
					if approval.Status == "PENDING" {
						isExist = true
						break
					}
					// app.Text = ""
					// app.Status = "PENDING"

					// if approval.Status == "APPROVED" && approval.Confirmed != nil {
					// 	app.Text = fmt.Sprintf("Approved By %s", user)
					// 	app.Status = approval.Status
					// 	app.IsApproved = true
					// } else if approval.Status == "REJECTED" && approval.Confirmed != nil {
					// 	app.Text = fmt.Sprintf("Rejected By %s", user)
					// 	app.Status = approval.Status
					// } else
					// if approval.Status == "PENDING" && approval.Confirmed != nil {
					// 	app.Text = fmt.Sprintf("%s", user)
					// 	app.Status = approval.Status
					// }

					// app.Reason = approval.Text
					// app.Date = approval.Confirmed
				}
			}

			if isExist == true {
				app.Status = "PENDING"
				app.Text = fmt.Sprintf("%s", user)
				result = append(result, app)
			}
		}
	}

	return result, nil
}
