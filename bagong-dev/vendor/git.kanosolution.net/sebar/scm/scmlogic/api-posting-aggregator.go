package scmlogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/sebarcode/codekit"
)

type ApprovalAggregatorHandler struct {
}

func (m *ApprovalAggregatorHandler) PostByGroup(ctx *kaos.Context, payloads []ficologic.ApprovalAggregatorPostByGroupRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	postParams := []ficologic.PostRequest{}
	for _, payload := range payloads {
		postParam := ficologic.PostRequest{
			Op:   payload.Op,
			Text: payload.Text,
		}

		param := ficologic.ApprovalAggregatorGetJournal{
			GroupBy: payload.GroupBy,
			Status:  "PENDING",
			Type:    payload.Type,
		}

		filters := m.getFilterPostingProfile(ctx, &param)
		profiles := codekit.Ms{}
		err := h.Gets(new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(
			dbflex.And(filters...),
		).SetSelect("SourceType", "SourceID"), &profiles)
		if err != nil {
			return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
		}

		for _, profile := range profiles {
			// switch profile["SourceType"].(string) {
			// case string(ficomodel.SubledgerCashBank):
			// 	postParam.JournalType = ficomodel.SubledgerCashBank
			// case string(ficomodel.SubledgerVendor):
			// 	postParam.JournalType = ficomodel.SubledgerVendor
			// case string(ficomodel.SubledgerCustomer):
			// 	postParam.JournalType = ficomodel.SubledgerCustomer
			// case string(ficomodel.SubledgerAccounting):
			// 	postParam.JournalType = ficomodel.SubledgerAccounting
			// default:
			// 	continue
			// }

			postParam.JournalType = tenantcoremodel.TrxModule(profile.GetString("SourceType"))
			postParam.JournalID = profile["SourceID"].(string)
			postParams = append(postParams, postParam)
		}
	}

	_, err := new(PostingProfileHandlerV2).Post(ctx, postParams)
	if err != nil {
		return nil, fmt.Errorf("error when posting journals : %s", err.Error())
	}

	return "Success", nil
}

func (m *ApprovalAggregatorHandler) getFilterPostingProfile(ctx *kaos.Context, payload *ficologic.ApprovalAggregatorGetJournal) []*dbflex.Filter {
	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	if ctx.Data().Get("CompanyID", "").(string) != "" {
		coID = ctx.Data().Get("CompanyID", "").(string)
	}
	user := sebar.GetUserIDFromCtx(ctx)

	filter := []*dbflex.Filter{
		dbflex.Eq("CompanyID", coID),
		dbflex.ElemMatch("Approvals",
			dbflex.And(
				dbflex.Eq("Status", payload.Status),
				dbflex.Eq("UserID", user),
			),
		),
	}

	if payload.GroupBy == ficologic.ApprovalAggregatorGroupBySite {
		filter = append(filter, dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", payload.Type)))
	} else {
		filter = append(filter, dbflex.Eq("SourceType", payload.Type))
	}

	return filter
}
