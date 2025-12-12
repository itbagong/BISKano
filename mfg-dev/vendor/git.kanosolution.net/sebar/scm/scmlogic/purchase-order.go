package scmlogic

import (
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/mongo"
)

type PurchaseOrderEngine struct{}

type RollbackRemainingQtyRequest struct {
	PurchaseOrderJournalID string
}

func (p *PurchaseOrderEngine) RollbackRemainingQty(ctx *kaos.Context, req *RollbackRemainingQtyRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: connection")
	}

	po, _ := datahub.GetByID(h, new(scmmodel.PurchaseOrderJournal), req.PurchaseOrderJournalID)
	if po == nil {
		po = new(scmmodel.PurchaseOrderJournal)
	}

	err := p.calcRemainingQtyInReff(h, po)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func (o *PurchaseOrderEngine) GetsV1(ctx *kaos.Context, req *dbflex.QueryParam) (codekit.M, error) {
	res := []scmmodel.PurchaseOrderJournal{}
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

	//get posting profile purchase order
	jtPurchaseOrders := []scmmodel.PurchaseOrderJournalType{}
	e := h.Gets(new(scmmodel.PurchaseOrderJournalType), dbflex.NewQueryParam(), &jtPurchaseOrders)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	ppPurchaseOrders := lo.Map(jtPurchaseOrders, func(jt scmmodel.PurchaseOrderJournalType, index int) string {
		return jt.PostingProfileID
	})

	sourceIDs, _ := FindSourceIDApprovalByFilter(h, keyword, ppPurchaseOrders[0])
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

	e = h.Gets(new(scmmodel.PurchaseOrderJournal), parm, &res)
	if e != nil && e != mongo.ErrNoDocuments {
		return resFinal, e
	}

	// get count
	cmd := dbflex.From(new(scmmodel.PurchaseOrderJournal).TableName())
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

func (p *PurchaseOrderEngine) calcRemainingQtyInReff(db *datahub.Hub, header *scmmodel.PurchaseOrderJournal) error {
	// proses perhitungan remaining qty bila memiliki ReffNo
	err := sebar.Tx(db, true, func(tx *datahub.Hub) error {
		linePerPR := lo.GroupBy(header.Lines, func(d scmmodel.PurchaseJournalLine) string {
			return d.PRID
		})

		for prID, poLines := range linePerPR {
			//cari apakah ada PO yang sudah menggunakan PR yang sama
			poExists := []scmmodel.PurchaseOrderJournal{}
			tx.GetsByFilter(new(scmmodel.PurchaseOrderJournal), dbflex.And(
				dbflex.Ne("_id", header.ID),
				dbflex.Eq("ReffNo", prID),
				dbflex.Ne("Status", ficomodel.JournalStatusDraft),
				dbflex.Ne("Status", ficomodel.JournalStatusRejected),
			), &poExists)

			existingLines := map[string]scmmodel.PurchaseJournalLine{}
			if len(poExists) > 0 {
				lo.ForEach(poExists, func(po scmmodel.PurchaseOrderJournal, index int) {
					lo.ForEach(po.Lines, func(line scmmodel.PurchaseJournalLine, index int) {
						key := fmt.Sprintf("%s|%s|%d", line.ItemID, line.SKU, line.SourceLineNo)
						if _, ok := existingLines[key]; !ok {
							existingLines[key] = scmmodel.PurchaseJournalLine{}
						}

						existLine := existingLines[key]
						existLine.Qty += line.Qty
						existingLines[key] = existLine
					})
				})
			}

			pr, e := datahub.GetByID(tx, new(scmmodel.PurchaseRequestJournal), prID)
			if e != nil {
				fmt.Printf("WARNING PO Line doesn't have PRID but has ReffNo | PO ID: %s | ReffNo: %s\n", header.ID, strings.Join(header.ReffNo, ", "))
				continue // bypass if PR not found
			}

			// update each PR Lines
			for prLineIdx, prLine := range pr.Lines {
				//calculate existing remaining
				key := fmt.Sprintf("%s|%s|%s|%d", prLine.ItemID, prLine.SKU, prLine.LineNo)
				qtyConvertExisting := 0.0
				if _, ok := existingLines[key]; ok {
					qtyConvertExisting, _ = ConvertUnit(tx, existingLines[key].Qty, existingLines[key].UnitID, prLine.UnitID)

				}

				// get PO Lines with the same ItemID, SKU and line no
				fpolines := lo.Filter(poLines, func(d scmmodel.PurchaseJournalLine, i int) bool {
					return d.ItemID == prLine.ItemID && d.SKU == prLine.SKU && d.SourceLineNo == prLine.LineNo
				})
				if len(fpolines) == 0 {
					continue // bypass if not exist
				}

				// calculate total converted PO Line Qty
				totalConvertedQty := float64(0)
				for _, fpl := range fpolines {
					convQty, e := ConvertUnit(tx, fpl.Qty, fpl.UnitID, prLine.UnitID)
					if e != nil {
						return e
					}

					totalConvertedQty = totalConvertedQty + convQty
				}

				availRemainingQty := prLine.Qty - (prLine.RemainingQty + qtyConvertExisting)
				if totalConvertedQty > availRemainingQty {
					return nil
				} else {
					pr.Lines[prLineIdx].RemainingQty = (prLine.RemainingQty + qtyConvertExisting) + totalConvertedQty // increase
				}
			}

			if e := tx.Save(pr); e != nil {
				return e
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
