package scmlogic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/mongo"
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

	filters := []*dbflex.Filter{}
	filters = append(filters, dbflex.Eq("IsUsedVendor", false))
	filters = append(filters, dbflex.Eq("Status", ficomodel.JournalStatusPosted))

	if payload != nil {
		if payload.Where != nil {
			filters2 := []*dbflex.Filter{}
			if len(payload.Where.Items) > 0 {
				vItems := payload.Where.Items
				if len(vItems) > 0 {
					for _, val := range vItems {
						fieldVal := val.Field
						opVal := val.Op
						if opVal == dbflex.OpContains {
							aInterface := val.Value.([]interface{})
							aString := make([]string, len(aInterface))
							for i, v := range aInterface {
								aString[i] = v.(string)
							}
							if len(aString) > 0 {
								if aString[0] != "" {
									filters2 = append(filters2, dbflex.Contains(fieldVal, aString[0]))
								}
							}
						}
					}
				}
			} else {
				fieldVal := payload.Where.Field
				opVal := payload.Where.Op
				if opVal == dbflex.OpEq {
					aInterface := payload.Where.Value.(string)
					filters2 = append(filters2, dbflex.Eq(fieldVal, aInterface))
				}
			}
			if len(filters2) > 0 {
				filters = append(filters, dbflex.Or(filters2...))
			}
		}
	}

	query = query.SetWhere(dbflex.And(filters...))

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
	countData := 0
	countTotal := 0
	if len(listGR) > 0 {
		for _, val := range listGR {
			if len(val.ReffNo) > 0 {
				for _, valno := range val.ReffNo {
					if lo.Contains(listPONo, valno) {
						resultTmp := FindByVendorResponse{}
						resultTmp.InventReceiveIssueJournal = val
						resultTmp.MainBalanceAccount = vendor.MainBalanceAccount

						countTotal++
						if payload.Skip != 0 {
							if countTotal > payload.Skip {
								if countData < payload.Take {
									result = append(result, resultTmp)
									countData++
								}
							}
						} else {
							if countData < payload.Take {
								result = append(result, resultTmp)
								countData++
							}
						}

						break
					}
				}
			}
		}
	}

	resultFin := codekit.M{}.Set("data", result).Set("count", countTotal)
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

func (o *InventReceiveJournalEngine) GetsV1(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	res := []scmmodel.InventReceiveIssueJournal{}
	resFinal := codekit.M{
		"count": len(res),
		"data":  res,
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return resFinal, fmt.Errorf("missing: connection")
	}

	if req == nil {
		return resFinal, fmt.Errorf("missing: payload")
	}

	//clear param
	keyword := ""
	if req.Where != nil {
		filter := req.Where
		if filter.Op == dbflex.OpAnd {
			clearItems := []*dbflex.Filter{}
			for _, filterItem := range filter.Items {
				if filterItem.Field == "Keyword" {
					keyword = filterItem.Value.(string)
				} else {
					clearItems = append(clearItems, filterItem)
				}
			}

			filter.Items = clearItems
		}

		if len(filter.Items) > 0 {
			req.Where = filter
		} else {
			req.Where = nil
		}
	}

	//get posting profile invent journal type
	jtInventJournalTypes := []scmmodel.InventJournalType{}
	e := h.Gets(new(scmmodel.InventJournalType), dbflex.NewQueryParam().SetWhere(dbflex.Eq("_id", "JT-InventoryReceive")), &jtInventJournalTypes)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	ppReceiveJournals := lo.Map(jtInventJournalTypes, func(jt scmmodel.InventJournalType, index int) string {
		return jt.PostingProfileID
	})

	sourceIDs, _ := FindSourceIDApprovalByFilter(h, keyword, ppReceiveJournals[0])
	if len(sourceIDs) > 0 {
		if req.Where != nil {
			filter := req.Where
			if filter.Op == dbflex.OpAnd {
				clearItems := []*dbflex.Filter{}
				for _, filterItem := range filter.Items {
					if filterItem.Field == "Keyword" {
						keyword = filterItem.Value.(string)
					} else {
						if filterItem.Op == dbflex.OpOr {
							filterItem.Items = append(filterItem.Items, dbflex.In("_id", sourceIDs...))
						}

						clearItems = append(clearItems, filterItem)
					}
				}

				filter.Items = clearItems
			}

			if len(filter.Items) > 0 {
				req.Where = filter
			} else {
				req.Where = nil
			}
		}
	}

	fs := ctx.Data().Get("DBModFilter", []*dbflex.Filter{}).([]*dbflex.Filter)
	parm := req
	if len(fs) == 1 {
		parm = CombineQueryParam(req, dbflex.NewQueryParam().SetWhere(fs[0]))
	} else if len(fs) > 1 {
		parm = CombineQueryParam(req, dbflex.NewQueryParam().SetWhere(dbflex.And(fs...)))
	}

	if parm.Where != nil {
		if parm.Where.Op == dbflex.OpAnd {
			if len(parm.Where.Items) == 1 {
				if parm.Where.Items[0].Op == dbflex.OpOr {
					parm.Where = parm.Where.Items[0]
				}
			}
		}
	}

	// Parse "TrxDate" in query's "Where" clause
	// Convert string to RFC3339 time format if present
	// Ensures proper formatting for database query
	if parm.Where != nil && parm.Where.Items != nil {
		for i, item := range parm.Where.Items {
			if item.Field == "TrxDate" {
				if strValue, ok := item.Value.(string); ok {
					if parsedTime, err := time.Parse(time.RFC3339, strValue); err == nil {
						parm.Where.Items[i].Value = parsedTime
					}
				}
			}
		}
	}

	e = h.Gets(new(scmmodel.InventReceiveIssueJournal), parm, &res)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	// get count
	cmd := dbflex.From(new(scmmodel.InventReceiveIssueJournal).TableName())
	if parm != nil && parm.Where != nil {
		cmd.Where(parm.Where)
	}

	connIdx, conn, err := h.GetConnection()
	if err == nil {
		defer h.CloseConnection(connIdx, conn)
		resFinal.Set("count", conn.Cursor(cmd, nil).Count())
	}

	resFinal.Set("data", res)
	return resFinal, nil
}
