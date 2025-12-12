package scmlogic

import (
	"fmt"
	"reflect"
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

type ItemRequestEngine struct {
}

func (o *ItemRequestEngine) GetsV1(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	res := []scmmodel.ItemRequest{}
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

	//get posting profile item request
	jtItemRequests := []scmmodel.ItemRequestJournalType{}
	e := h.Gets(new(scmmodel.ItemRequestJournalType), dbflex.NewQueryParam(), &jtItemRequests)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	ppItemRequests := lo.Map(jtItemRequests, func(jt scmmodel.ItemRequestJournalType, index int) string {
		return jt.PostingProfileID
	})

	sourceIDs, _ := FindSourceIDApprovalByFilter(h, keyword, ppItemRequests[0])
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

	e = h.Gets(new(scmmodel.ItemRequest), parm, &res)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	// get count
	cmd := dbflex.From(new(scmmodel.ItemRequest).TableName())
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

func CombineQueryParam(origin, other *dbflex.QueryParam) *dbflex.QueryParam {
	if origin == nil {
		if other != nil {
			return origin
		}
		return nil
	}

	if other == nil {
		return origin
	}

	if len(other.Aggregates) > 0 {
		origin.Aggregates = append(origin.Aggregates, other.Aggregates...)
	}

	if len(other.GroupBy) > 0 {
		origin.GroupBy = append(origin.GroupBy, other.GroupBy...)
	}

	if len(other.Select) > 0 {
		origin.Select = append(origin.Select, other.Select...)
	}

	if len(other.Sort) > 0 {
		origin.Sort = append(origin.Sort, other.Sort...)
	}

	if other.Skip > 0 {
		origin.Skip = other.Skip
	}

	if other.Take > 0 {
		origin.Take = other.Take
	}

	origin.Where = combineFilter(origin.Where, other.Where)
	return origin
}

func combineFilter(origin, other *dbflex.Filter) *dbflex.Filter {
	origin = filterString2Date(origin)
	other = filterString2Date(other)

	if origin == nil {
		if other != nil {
			return other
		}
		return origin
	}

	if other == nil {
		return origin
	}

	return dbflex.And(origin, other)
}

func filterString2Date(f *dbflex.Filter) *dbflex.Filter {
	if f == nil {
		return f
	}

	if f.Op == dbflex.OpAnd || f.Op == dbflex.OpOr {
		for index, itemF := range f.Items {
			itemF = filterString2Date(itemF)
			f.Items[index] = itemF
		}
		return f
	}

	vf := reflect.ValueOf(f.Value)
	if vf.Kind() == reflect.Ptr {
		vf = vf.Elem()
	}

	if vf.Kind() == reflect.String {
		if dt, err := time.Parse(time.RFC3339, vf.String()); err == nil {
			f.Value = dt
		}
	}

	return f
}

func FindSourceIDApprovalByFilter(h *datahub.Hub, keyword string, pp string) ([]interface{}, error) {
	sourceIDs := []interface{}{}

	if keyword != "" {
		//get posting profile approval by source in and posting profile id
		postingApprovals := []ficomodel.PostingApproval{}
		e := h.GetsByFilter(new(ficomodel.PostingApproval), dbflex.And(
			dbflex.Eq("PostingProfileID", pp),
			dbflex.Eq("Status", "PENDING"),
		), &postingApprovals)
		if e != nil && e != mongo.ErrNoDocuments {
			return sourceIDs, e
		}

		if len(postingApprovals) > 0 {
			approvalFilter, _ := TransformPostingApprovalStageApprove(h, postingApprovals, keyword)
			if len(approvalFilter) > 0 {
				sourceIDs := lo.Map(approvalFilter, func(approval ficomodel.PostingApproval, index int) interface{} {
					return approval.SourceID
				})

				return sourceIDs, nil
				// itemRequestFromApproval := []scmmodel.ItemRequest{}
				// //search IR by source id from approval
				// e = h.GetsByFilter(new(scmmodel.ItemRequest), dbflex.In("_id", sourceIDs...), &itemRequestFromApproval)
				// if e != nil && e != mongo.ErrNoDocuments {
				// 	return resFinal, e
				// }

				// mapItemRequestRes := lo.Associate(res, func(ir scmmodel.ItemRequest) (string, scmmodel.ItemRequest) {
				// 	return ir.ID, ir
				// })

				// for _, ir := range itemRequestFromApproval {
				// 	if _, ok := mapItemRequestRes[ir.ID]; !ok {
				// 		res = append(res, ir)
				// 	}
				// }
			}
		}
	}

	return sourceIDs, nil
}

func TransformPostingApprovalStageApprove(h *datahub.Hub, data []ficomodel.PostingApproval, keyword string) ([]ficomodel.PostingApproval, error) {
	approvals := []ficomodel.PostingApproval{}

	//get employee where name contains keyword
	employees := []tenantcoremodel.Employee{}
	h.GetsByFilter(new(tenantcoremodel.Employee), dbflex.Contains("Name", keyword), &employees)
	if len(employees) == 0 {
		return approvals, nil
	}

	mapEmployee := lo.Associate(employees, func(user tenantcoremodel.Employee) (string, string) {
		return user.ID, user.Name
	})

	newApprovals := []ficomodel.PostingApproval{}
	for _, approval := range data {
		isFound := false
		if approval.Status != "PENDING" {
			continue
		}

		for _, approver := range approval.Approvers {
			for _, user := range approver.UserIDs {
				isExist := false
				for _, approvalUser := range approval.Approvals {
					if user == approvalUser.UserID {
						if userName, ok := mapEmployee[user]; ok {
							if userName != "" && approvalUser.Status == "PENDING" {
								isExist = true
								break
							}
						}
					}
				}

				if isExist == true {
					isFound = true
					break
				}
			}

			if isFound == true {
				break
			}
		}

		if isFound == true {
			newApprovals = append(newApprovals, approval)
		}
	}

	return newApprovals, nil
}
