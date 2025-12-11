package scmlogic

import (
	"fmt"
	"strings"
	"time"

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
	Error                  string
}

func (p *PurchaseOrderEngine) RollbackRemainingQty(ctx *kaos.Context, req *RollbackRemainingQtyRequest) (string, error) {
	if req.Error != "" {
		errorPart := strings.Split(req.Error, "|")
		if len(errorPart) > 0 && strings.Contains(strings.ToLower(strings.TrimSpace(errorPart[0])), "[validate]") {
			return "OK", nil
		}
	}
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

type PurchaseOrderCancelRequest struct {
	ID string
}

func (o *PurchaseOrderEngine) IsCancelable(ctx *kaos.Context, req *PurchaseOrderCancelRequest) (bool, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return false, fmt.Errorf("missing: connection")
	}

	return o.isCancelable(h, req.ID)
}

func (o *PurchaseOrderEngine) Cancel(ctx *kaos.Context, req *PurchaseOrderCancelRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: connection")
	}

	err := sebar.Tx(h, true, func(tx *datahub.Hub) error {
		po := new(scmmodel.PurchaseOrderJournal)
		if e := h.GetByID(po, req.ID); e != nil {
			return e
		}

		isCancelable, e := o.isCancelable(h, req.ID)
		if e != nil {
			return e
		}
		if !isCancelable {
			return fmt.Errorf("this PO cannot be cancelled")
		}

		// proses Invent Trx di minus kan
		trxs := []scmmodel.InventTrx{}
		h.GetsByFilter(new(scmmodel.InventTrx), dbflex.And(
			dbflex.Eq("SourceJournalID", po.ID),
			dbflex.Eq("Status", string(scmmodel.ItemPlanned)),
		), &trxs)

		for _, trx := range trxs {
			trx.ID = ""
			trx.Created = time.Time{}
			trx.LastUpdate = time.Time{}
			trx.Qty = trx.Qty * -1
			trx.TrxQty = trx.TrxQty * -1
			trx.AmountPhysical = trx.AmountPhysical * -1
			trx.AmountFinancial = trx.AmountFinancial * -1
			trx.AmountAdjustment = trx.AmountAdjustment * -1
			tx.Save(&trx)
		}

		// Balance Sync
		balanceOpt := ItemBalanceOpt{}
		balanceOpt.CompanyID = po.CompanyID
		balanceOpt.DisableGrouping = true
		balanceOpt.ConsiderSKU = true
		balanceOpt.ItemIDs = lo.Map(trxs, func(t scmmodel.InventTrx, index int) string {
			return t.Item.ID
		})

		if _, e := NewItemBalanceHub(tx).Sync(nil, balanceOpt); e != nil {
			return e
		}

		// return remaining Qty nya tiap PR
		if e := PRLinesUpdateRemainingQty(tx, po, Increase); e != nil {
			return e
		}

		// update isCanceled
		po.IsCanceled = true
		po.LastUpdate = time.Now()
		tx.Update(po, "IsCanceled", "LastUpdate")

		return nil
	})
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func (p *PurchaseOrderEngine) isCancelable(h *datahub.Hub, poID string) (bool, error) {
	po := new(scmmodel.PurchaseOrderJournal)
	if e := h.GetByID(po, poID); e != nil {
		return false, e
	}

	if po.Status != ficomodel.JournalStatusPosted {
		return false, nil
	}

	// ngambil semua Invent Trx untuk PO ini dan cek apakah ada Status yang confirmed (ada: tidak bisa di close)
	trxs := []scmmodel.InventTrx{}
	h.GetsByFilter(new(scmmodel.InventTrx), dbflex.And(
		dbflex.Eq("SourceJournalID", po.ID),
		dbflex.Eq("Status", string(scmmodel.ItemConfirmed)),
	), &trxs)

	if len(trxs) > 0 {
		return false, nil
	}

	return true, nil
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

type PurchaseOrderSyncJournalStatusRequest struct {
	PurchaseOrderID string
}

func (p *PurchaseOrderEngine) SyncJournalStatus(ctx *kaos.Context, req *PurchaseOrderSyncJournalStatusRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "NOK", fmt.Errorf("missing: connection")
	}

	if req == nil {
		return "NOK", fmt.Errorf("missing: payload")
	}

	//get purchase order journal
	po, err := datahub.GetByID(h, new(scmmodel.PurchaseOrderJournal), req.PurchaseOrderID)
	if err != nil && err != mongo.ErrNoDocuments {
		return "NOK", fmt.Errorf("Error when get purchase order journal with error: %s", err.Error())
	}

	//get posting profile approval
	param := dbflex.NewQueryParam()
	param.SetSort("-LastUpdate")
	param.SetWhere(dbflex.Eq("SourceID", req.PurchaseOrderID))

	postingApproval, err := datahub.GetByParm(h, new(ficomodel.PostingApproval), param)
	if err != nil {
		return "NOK", fmt.Errorf("Error when get posting approval with error: %s", err.Error())
	}

	journalStatus := ValidatePostingApproval(postingApproval.Status)

	if po.Status != ficomodel.JournalStatus(journalStatus) && journalStatus != "" {
		po.Status = ficomodel.JournalStatus(journalStatus)
		err = datahub.Update(h, po)
		if err != nil {
			return "NOK", fmt.Errorf("Error when update journal status with error: %s", err.Error())
		}
	}

	return "OK", nil
}
