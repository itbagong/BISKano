package scmlogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/mongo"
)

type PurchaseRequestEngine struct{}

type PurchaseRequestGetLinesRequest struct {
	VendorID    string
	Site        string
	DeliveryTo  string   // WarehouseID
	Statuses    []string // empty -> POSTED only
	ReffNo      string
	PRRequestor string
}

type PurchaseRequestGetLinesResponse struct {
	scmmodel.PurchaseJournalLine
	Header       PRHeader
	PRID         string
	PRDate       *time.Time
	WarehouseID  string
	RemainQty    float64 // from item balance - qty planned
	VendorID     string
	VendorName   string
	SourceLineNo int
}

type PRHeader struct {
	scmmodel.PurchaseRequestJournal
	DueDate time.Time
}

func (o *PurchaseRequestEngine) GetLines(ctx *kaos.Context, payload *PurchaseRequestGetLinesRequest) ([]PurchaseRequestGetLinesResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	filters := []*dbflex.Filter{}

	if payload.VendorID != "" {
		filters = append(filters, dbflex.Eq("VendorID", payload.VendorID))
	}

	if payload.DeliveryTo != "" {
		filters = append(filters, dbflex.Eq("Location.WarehouseID", payload.DeliveryTo))
	}

	if payload.Site != "" {
		filters = append(filters, dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Site)))
	}

	if len(payload.Statuses) > 0 {
		filters = append(filters, dbflex.In("Status", payload.Statuses...))
	} else {
		filters = append(filters, dbflex.Eq("Status", ficomodel.JournalStatusPosted))
	}

	if payload.ReffNo != "" {
		filters = append(filters, dbflex.Contains("_id", payload.ReffNo))
	}

	if payload.PRRequestor != "" {
		filters = append(filters, dbflex.Eq("Requestor", payload.PRRequestor))
	}

	var f *dbflex.Filter
	if len(filters) > 0 {
		f = dbflex.And(filters...)
	}

	prs := []scmmodel.PurchaseRequestJournal{}
	if e := h.GetsByFilter(new(scmmodel.PurchaseRequestJournal), f, &prs); e != nil {
		return nil, e
	}

	bal := NewInventBalanceCalc(h)
	specORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
	itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
	specVariantORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant))
	specSizeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecSize))
	specGradeORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecGrade))

	res := []PurchaseRequestGetLinesResponse{}
	for _, pr := range prs {
		lines := lo.Map(pr.Lines, func(item scmmodel.PurchaseJournalLine, index int) PurchaseRequestGetLinesResponse {
			prHeader := pr
			prHeader.ReffNo = []string{pr.ID}
			prHeader.Lines = []scmmodel.PurchaseJournalLine{}
			return PurchaseRequestGetLinesResponse{
				PurchaseJournalLine: item,
				Header:              PRHeader{PurchaseRequestJournal: prHeader, DueDate: *pr.PRDate},
				PRID:                pr.ID,
				PRDate:              pr.PRDate,
				WarehouseID:         pr.WarehouseID,
				VendorID:            pr.VendorID,
				VendorName:          pr.VendorName,
				SourceLineNo:        item.LineNo,
			}
		})

		// tambah query find remaining qty > 0
		prg := []PurchaseRequestGetLinesResponse{}
		for li, l := range lines {
			spec, _ := specORM.Get(lines[li].SKU)
			specName := spec.GenerateName(itemORM, specVariantORM, specSizeORM, specGradeORM)
			item, _ := itemORM.Get(l.ItemID)

			lines[li].Text = lo.Ternary(specName != "", specName, item.Name)

			b, _ := bal.Get(&InventBalanceCalcOpts{
				CompanyID: pr.CompanyID,
				ItemID:    []string{l.ItemID},
				InventDim: l.InventDim,
				// BalanceDate: ???,
			})

			if len(b) > 0 {
				lines[li].RemainQty = b[0].QtyPlanned
			}
			if l.InventJournalLine.RemainingQty > 0 {
				prg = append(prg, lines[li])
			}
		}

		res = append(res, prg...)
	}

	return res, nil
}

func (o *PurchaseRequestEngine) GetsV1(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	res := []scmmodel.PurchaseRequestJournal{}
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

	//get posting profile purchase request
	jtPurchaseRequests := []scmmodel.PurchaseRequestJournalType{}
	e := h.Gets(new(scmmodel.PurchaseRequestJournalType), dbflex.NewQueryParam(), &jtPurchaseRequests)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	ppPurchaseRequests := lo.Map(jtPurchaseRequests, func(jt scmmodel.PurchaseRequestJournalType, index int) string {
		return jt.PostingProfileID
	})

	sourceIDs, _ := FindSourceIDApprovalByFilter(h, keyword, ppPurchaseRequests[0])
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

	e = h.Gets(new(scmmodel.PurchaseRequestJournal), parm, &res)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	// get count
	cmd := dbflex.From(new(scmmodel.PurchaseRequestJournal).TableName())
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
