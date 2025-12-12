package scmlogic

import (
	"errors"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type InventReceiveJournalEngine struct{}

type InvReceiveUpdateLinePayload struct {
	InventReceive scmmodel.InventReceiveIssueJournal
}

func (o *InventReceiveJournalEngine) UpdateLines(ctx *kaos.Context, payload scmmodel.InventReceiveIssueJournal) (interface{}, error) {
	//check id empty or not
	if payload.ID == "" {
		return nil, errors.New("Invent Receive ID Empty")
	}

	//get connection
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	inventReceiveJournal, e := datahub.GetByID(h, new(scmmodel.InventReceiveIssueJournal), payload.ID)
	if e != nil {
		return nil, errors.New("Invent Receive Journal not found")
	}

	maxLineNo := 0
	lo.ForEach(inventReceiveJournal.Lines, func(line scmmodel.InventReceiveIssueLine, index int) {
		if maxLineNo < line.LineNo {
			maxLineNo = line.LineNo
		}
	})

	finalLines := lo.Map(payload.Lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
		line.SourceLineNo = maxLineNo + 1
		maxLineNo += 1
		return line
	})

	inventReceiveJournal.Lines = append(inventReceiveJournal.Lines, finalLines...)
	e = h.Save(inventReceiveJournal)
	if e != nil {
		return nil, errors.New("Error append invent receive line")
	}

	return inventReceiveJournal, nil
}

type FindByVendorResponse struct {
	scmmodel.InventReceiveIssueJournal
	MainBalanceAccount string
}

func (o *InventReceiveJournalEngine) FindByVendor(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	//get connection
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	query := dbflex.NewQueryParam()
	// if payload.Skip != 0 {
	// 	query = query.SetSkip(payload.Skip)
	// }

	// if payload.Take != 0 {
	// 	query = query.SetTake(payload.Take)
	// }

	if len(payload.Sort) > 0 {
		query = query.SetSort(payload.Sort...)
	}

	query = query.SetWhere(dbflex.And(dbflex.Eq("IsUsedVendor", false), dbflex.Eq("Status", ficomodel.JournalStatusPosted)))

	filterPO := []*dbflex.Filter{}

	r := ctx.Data().Get("http_request", nil).(*http.Request)
	fieldVal := strings.TrimSpace(r.URL.Query().Get("VendorID"))

	if fieldVal != "" {
		filterPO = append(filterPO, dbflex.Eq("VendorID", fieldVal))
	}

	listGR := []scmmodel.InventReceiveIssueJournal{}
	e := h.Gets(new(scmmodel.InventReceiveIssueJournal), query, &listGR)
	if e != nil {
		return nil, errors.New("Failed populate data GR: " + e.Error())
	}

	refNo := []string{}
	if len(listGR) > 0 {
		for _, val := range listGR {
			refNo = append(refNo, val.ReffNo...)
		}
		filterPO = append(filterPO, dbflex.In("_id", refNo...))
	}

	listPO := []scmmodel.PurchaseOrderJournal{}
	e = h.Gets(new(scmmodel.PurchaseOrderJournal), dbflex.NewQueryParam().SetWhere(dbflex.And(filterPO...)), &listPO)
	if e != nil {
		return nil, errors.New("failed populate data GR: " + e.Error())
	}

	listPONo := lo.Map(listPO, func(po scmmodel.PurchaseOrderJournal, index int) string {
		return po.ID
	})

	vendor := new(tenantcoremodel.Vendor)
	e = h.GetByID(vendor, fieldVal)
	if e != nil {
		return nil, errors.New("failed populate data vendor: " + e.Error())
	}

	result := []FindByVendorResponse{}
	if len(listGR) > 0 {
		for _, val := range listGR {
			if len(val.ReffNo) > 0 {
				for _, valno := range val.ReffNo {
					if lo.Contains(listPONo, valno) {
						resultTmp := FindByVendorResponse{}
						resultTmp.InventReceiveIssueJournal = val
						resultTmp.MainBalanceAccount = vendor.MainBalanceAccount
						result = append(result, resultTmp)
						break
					}
				}
			}
		}
	}

	resultFin := codekit.M{}.Set("data", result).Set("count", len(result))
	return resultFin, nil
}

type GRUsedPayload struct {
	GRNumber []string
}

func (o *InventReceiveJournalEngine) CheckGrUsed(ctx *kaos.Context, payload *GRUsedPayload) (string, error) {
	//get connection
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", errors.New("missing: connection")
	}

	// get data GR
	listGR := []scmmodel.InventReceiveIssueJournal{}
	e := h.Gets(new(scmmodel.InventReceiveIssueJournal), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", payload.GRNumber...),
			dbflex.Eq("IsUsedVendor", true),
		)), &listGR)
	if e != nil {
		return "", errors.New("Failed populate data GR: " + e.Error())
	}

	if len(listGR) > 0 {
		return "", errors.New("GR number already used")
	}

	return "Ok", nil
}

func (o *InventReceiveJournalEngine) UpdateGrUsed(ctx *kaos.Context, payload *GRUsedPayload) (string, error) {
	//get connection
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", errors.New("missing: connection")
	}

	if e := h.UpdateField(&scmmodel.InventReceiveIssueJournal{IsUsedVendor: true}, dbflex.In("_id", payload.GRNumber...), "IsUsedVendor"); e != nil {
		return "", e
	}

	return "Ok", nil
}
