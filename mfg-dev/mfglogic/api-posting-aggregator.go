package mfglogic

import (
	"errors"
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

var (
	JournalTypeMap = map[string][]string{
		"fico": []string{
			string(ficomodel.SubledgerVendor),
			string(ficomodel.SubledgerCustomer),
			string(ficomodel.SubledgerAsset),
			string(ficomodel.SubledgerPayroll),
			string(ficomodel.SubledgerExpense),
			string(ficomodel.SubledgerCashBank),
			string(ficomodel.SubledgerTax),
			string(ficomodel.SubledgerNone),
			string(ficomodel.SubledgerAccounting),
		},
		"scm": []string{
			string(ficomodel.Inventory),
			string(ficomodel.GoodReceive),
			string(ficomodel.GoodIssuance),
			string(ficomodel.Transfer),
			string(ficomodel.ItemRequest),
			string(ficomodel.PurchOrder),
			string(ficomodel.PurchRequest),
			string(ficomodel.AssetAcquisition),
		},
		"mfg": []string{
			string(ficomodel.WorkRequest),
			string(ficomodel.WorkOrder),
			string(ficomodel.WorkOrderReportConsumption),
			string(ficomodel.WorkOrderReportResource),
			string(ficomodel.WorkOrderReportOutput),
		},
	}
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
			postParam.JournalType = tenantcoremodel.TrxModule(profile.GetString("SourceType"))
			postParam.JournalID = profile["SourceID"].(string)
			postParams = append(postParams, postParam)
		}
	}

	postParamsGrouped := lo.GroupBy(postParams, func(d ficologic.PostRequest) string {
		for module, journalTypes := range JournalTypeMap {
			if lo.Contains(journalTypes, d.JournalType.String()) {
				return module
			}
		}
		return ""
	})

	// _, err := new(PostingProfileHandler).Post(ctx, postParams)
	// if err != nil {
	// 	return nil, fmt.Errorf("error when posting journals : %s", err.Error())
	// }

	errs := []error{}
	for module, params := range postParamsGrouped {
		if module == "" {
			continue // bypass if JournalType is not categorized yet, please add new one in JournalTypeMap variable above
		}

		switch module {
		case "fico":
			_, err := new(ficologic.PostingProfileHandler).Post(ctx, params)
			if err != nil {
				errs = append(errs, err)
			}
		case "scm":
			_, err := new(scmlogic.PostingProfileHandlerV2).Post(ctx, params)
			if err != nil {
				errs = append(errs, err)
			}
		case "mfg":
			_, err := new(PostingProfileHandler).Post(ctx, params)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		errTexts := lo.Map(errs, func(d error, i int) string {
			return d.Error()
		})
		return nil, fmt.Errorf("error when posting journals : %s", strings.Join(errTexts, " | "))
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
