package scmlogic

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/mongo"
)

type InventJournalEngine struct{}

func (o *InventJournalEngine) GetsV1(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	res := []scmmodel.InventJournal{}
	resFinal := codekit.M{
		"count": len(res),
		"data":  res,
	}

	transactionType := ""
	filterQueryParam := []*dbflex.Filter{}
	if hr, ok := ctx.Data().Get("http_request", nil).(*http.Request); ok {
		queryValues := hr.URL.Query()
		for k, vs := range queryValues {
			if len(vs) > 0 {
				if strings.ToLower(k) == "trxtype" {
					if vs[0] == "Transfer" {
						transactionType = "Item Transfer"
					} else if vs[0] == "Movement In" {
						transactionType = "Movement In"
					} else if vs[0] == "Movement Out" {
						transactionType = "Movement Out"
					}
				}

				filterQueryParam = append(filterQueryParam, dbflex.Eq(k, vs[0]))
			}
		}
		// if len(fs) == 1 {
		// 	parm = combineQueryParam(parm, dbflex.NewQueryParam().SetWhere(fs[0]))
		// } else if len(fs) > 1 {
		// 	parm = combineQueryParam(parm, dbflex.NewQueryParam().SetWhere(dbflex.And(fs...)))
		// }
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

	//get posting profile item request
	jtInventJournals := []scmmodel.InventJournalType{}
	e := h.Gets(new(scmmodel.InventJournalType), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("TransactionType", transactionType),
	), &jtInventJournals)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	ppInventJournals := lo.Map(jtInventJournals, func(jt scmmodel.InventJournalType, index int) string {
		return jt.PostingProfileID
	})

	sourceIDs, _ := FindSourceIDApprovalByFilter(h, keyword, ppInventJournals[0])
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

	req.SetWhere(combineFilter(req.Where, dbflex.And(filterQueryParam...))) // fix: IVJ202409114947 -> Transfer ga boleh masuk, kalo TrxType == Movement In

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

	e = h.Gets(new(scmmodel.InventJournal), parm, &res)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	// get count
	cmd := dbflex.From(new(scmmodel.InventJournal).TableName())
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
