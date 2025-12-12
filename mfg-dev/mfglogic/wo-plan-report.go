package mfglogic

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type WorkOrderPlanReportEngine struct{}

func (o *WorkOrderPlanReportEngine) Get(ctx *kaos.Context, req []string) (*WorkOrderPlanReportSaveReq, error) {
	if len(req) == 0 && req[0] == "" {
		return nil, fmt.Errorf("id required")
	}
	id := req[0]

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	wopr := new(mfgmodel.WorkOrderPlanReport)
	if e := h.GetByID(wopr, id); e != nil {
		return nil, e
	}

	mat := new(mfgmodel.WorkOrderPlanReportConsumption)
	if wopr.WorkOrderPlanReportConsumptionID != "" {
		if e := h.GetByID(mat, wopr.WorkOrderPlanReportConsumptionID); e != nil {
			return nil, e
		}
		wopr.WorkOrderPlanReportConsumptionStatus = string(mat.Status)

		wop := new(mfgmodel.WorkOrderPlan)
		if e := h.GetByID(wop, wopr.WorkOrderPlanID); e != nil {
			return nil, e
		}

		sums := []mfgmodel.WorkOrderSummaryMaterial{}
		h.GetsByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", wop.ID), &sums)
		sumM := lo.SliceToMap(sums, func(d mfgmodel.WorkOrderSummaryMaterial) (string, mfgmodel.WorkOrderSummaryMaterial) {
			return fmt.Sprintf("%s||%s", d.ItemID, d.SKU), d
		})

		for lineI, line := range mat.Lines {
			reservedButNotUsed := float64(0)
			if sum, exist := sumM[fmt.Sprintf("%s||%s", line.ItemID, line.SKU)]; exist {
				reservedButNotUsed = sum.Reserved - sum.Used // bila item dan sku ada di Summary, dapetin sisa yang sudah di reserved
			}

			bals := GetAvailableStocks(h, GetAvailableStocksParam{
				CompanyID: wop.CompanyID,
				InventDim: line.InventDim,
				Items:     []GetAvailableStocksParamItem{{ItemID: line.ItemID, SKU: line.SKU}},
			})
			if len(bals) > 0 {
				mat.Lines[lineI].QtyAvailable = bals[0].Qty + reservedButNotUsed
			} else {
				mat.Lines[lineI].QtyAvailable = reservedButNotUsed
			}
		}
	}

	rsc := new(mfgmodel.WorkOrderPlanReportResource)
	if wopr.WorkOrderPlanReportResourceID != "" {
		if e := h.GetByID(rsc, wopr.WorkOrderPlanReportResourceID); e != nil {
			return nil, e
		}
		wopr.WorkOrderPlanReportResourceStatus = string(rsc.Status)
	}

	out := new(mfgmodel.WorkOrderPlanReportOutput)
	if wopr.WorkOrderPlanReportOutputID != "" {
		if e := h.GetByID(out, wopr.WorkOrderPlanReportOutputID); e != nil {
			return nil, e
		}
		wopr.WorkOrderPlanReportOutputStatus = string(out.Status)
	}

	return &WorkOrderPlanReportSaveReq{
		WorkOrderPlanReport:                           *wopr,
		WorkOrderPlanReportConsumptionLines:           mat.Lines,
		WorkOrderPlanReportConsumptionAdditionalLines: mat.AdditionalLines,
		WorkOrderPlanReportResourceLines:              rsc.Lines,
		WorkOrderPlanReportOutputLines:                out.Lines,
	}, nil
}

type WorkOrderPlanReportSaveReq struct {
	mfgmodel.WorkOrderPlanReport
	WorkOrderPlanReportConsumptionLines           []mfgmodel.WorkOrderMaterialItem
	WorkOrderPlanReportConsumptionAdditionalLines []mfgmodel.WorkOrderMaterialItem
	WorkOrderPlanReportResourceLines              []mfgmodel.WorkOrderResourceItem
	WorkOrderPlanReportOutputLines                []mfgmodel.WorkOrderOutputItem
}

func (o *WorkOrderPlanReportEngine) Save(ctx *kaos.Context, req *WorkOrderPlanReportSaveReq) (*WorkOrderPlanReportSaveReq, error) {
	var (
		h         *datahub.Hub
		err       error
		committed bool
	)

	if req.WorkOrderPlanID == "" {
		return nil, fmt.Errorf("missing WorkOrderPlanID in payload")
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: connection")
	}

	h, _ = db.BeginTx()
	defer func() {
		if h.IsTx() {
			if err == nil && committed == false {
				h.Commit()
			} else {
				h.Rollback()
			}
		}
	}()

	if req.WorkOrderPlanReport.ID == "" {
		// format Report ID
		existingReports := []mfgmodel.WorkOrderPlanReport{}
		h.GetsByFilter(new(mfgmodel.WorkOrderPlanReport), dbflex.Eq("WorkOrderPlanID", req.WorkOrderPlanID), &existingReports)
		numb := len(existingReports) + 1
		req.WorkOrderPlanReport.ID = fmt.Sprintf("%s-%03d", req.WorkOrderPlanID, numb)

		// req.WorkOrderPlanReport.ID = primitive.NewObjectID().Hex()
	}

	if req.WorkOrderPlanReport.Status == "" {
		req.WorkOrderPlanReport.Status = mfgmodel.WorkOrderPlanStatusOverallDraft
	}

	err = h.Save(&req.WorkOrderPlanReport)
	if err != nil {
		return nil, err
	}

	wop := new(mfgmodel.WorkOrderPlan)
	err = h.GetByID(wop, req.WorkOrderPlanID)
	if err != nil {
		return nil, err
	}

	jt := new(mfgmodel.WorkOrderJournalType)
	err = h.GetByID(jt, wop.JournalTypeID)
	if err != nil {
		return nil, err
	}

	r1 := new(mfgmodel.WorkOrderPlanReportConsumption)
	if len(req.WorkOrderPlanReportConsumptionLines) > 0 || len(req.WorkOrderPlanReportConsumptionAdditionalLines) > 0 {
		h.GetByID(r1, req.WorkOrderPlanReport.WorkOrderPlanReportConsumptionID)

		req.WorkOrderPlanReportConsumptionLines = lo.Map(req.WorkOrderPlanReportConsumptionLines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
			d.Date = req.WorkOrderPlanReport.WorkDate
			return d
		})

		req.WorkOrderPlanReportConsumptionAdditionalLines = lo.Map(req.WorkOrderPlanReportConsumptionAdditionalLines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
			d.Date = req.WorkOrderPlanReport.WorkDate
			return d
		})

		//handle flow reopen
		if req.WorkOrderPlanReport.WorkOrderPlanReportConsumptionStatus == string(ficomodel.JournalStatusDraft) {
			if req.WorkOrderPlanReport.WorkOrderPlanReportConsumptionStatus != string(r1.Status) && r1.Status == ficomodel.JournalStatusRejected {
				r1.Status = ficomodel.JournalStatusDraft

				// //remove posting approval
				// err := h.DeleteByFilter(new(ficomodel.PostingApproval), dbflex.Eq("_id", r1.ID))
				// if err != nil {
				// 	return nil, err
				// }
			}
		}

		r1.ID = req.WorkOrderPlanReport.WorkOrderPlanReportConsumptionID // will be "" if new
		r1.WorkOrderPlanReportID = req.WorkOrderPlanReport.ID
		r1.WorkOrderPlanID = req.WorkOrderPlanID
		r1.CompanyID = wop.CompanyID
		r1.JournalTypeID = wop.JournalTypeID
		r1.TrxType = mfgmodel.JournalWorkOrderReportConsumption
		r1.Lines = req.WorkOrderPlanReportConsumptionLines
		r1.AdditionalLines = req.WorkOrderPlanReportConsumptionAdditionalLines
		r1.Dimension = wop.Dimension

		if r1.TrxDate.IsZero() {
			r1.TrxDate = time.Now()
		}

		if r1.Status == "" {
			r1.Status = ficomodel.JournalStatusDraft
		}

		err = h.Save(r1)
		if err != nil {
			return nil, err
		}
	}

	r2 := new(mfgmodel.WorkOrderPlanReportResource)
	if len(req.WorkOrderPlanReportResourceLines) > 0 {
		h.GetByID(r2, req.WorkOrderPlanReport.WorkOrderPlanReportResourceID)

		// sumRess := []mfgmodel.WorkOrderSummaryResource{}
		// h.GetsByFilter(new(mfgmodel.WorkOrderSummaryResource), dbflex.Eq("WorkOrderPlanID", wop.ID), &sumRess)
		// sumResM := lo.SliceToMap(sumRess, func(d mfgmodel.WorkOrderSummaryResource) (string, mfgmodel.WorkOrderSummaryResource) {
		// 	return d.ExpenseType, d
		// })

		req.WorkOrderPlanReportResourceLines = lo.Map(req.WorkOrderPlanReportResourceLines, func(d mfgmodel.WorkOrderResourceItem, i int) mfgmodel.WorkOrderResourceItem {
			d.Date = req.WorkOrderPlanReport.WorkDate

			// rate := sumResM[d.ExpenseType].RatePerHour
			// d.Total = d.WorkingHour * rate
			d.Total = d.WorkingHour * d.RatePerHour
			return d
		})

		r2.ID = req.WorkOrderPlanReport.WorkOrderPlanReportResourceID // will be "" if new
		r2.WorkOrderPlanReportID = req.WorkOrderPlanReport.ID
		r2.WorkOrderPlanID = req.WorkOrderPlanID
		r2.CompanyID = wop.CompanyID
		r2.JournalTypeID = wop.JournalTypeID
		r2.TrxType = mfgmodel.JournalWorkOrderReportConsumption
		r2.Lines = req.WorkOrderPlanReportResourceLines
		r2.Dimension = wop.Dimension

		//handle flow reopen
		if req.WorkOrderPlanReport.WorkOrderPlanReportResourceStatus == string(ficomodel.JournalStatusDraft) {
			if req.WorkOrderPlanReport.WorkOrderPlanReportResourceStatus != string(r2.Status) && r2.Status == ficomodel.JournalStatusRejected {
				r2.Status = ficomodel.JournalStatusDraft

				// //remove posting approval
				// err := h.DeleteByFilter(new(ficomodel.PostingApproval), dbflex.Eq("_id", r2.ID))
				// if err != nil {
				// 	return nil, err
				// }
			}
		}

		if r2.TrxDate.IsZero() {
			r2.TrxDate = time.Now()
		}

		if r2.Status == "" {
			r2.Status = ficomodel.JournalStatusDraft
		}

		err = h.Save(r2)
		if err != nil {
			return nil, err
		}
	}

	r3 := new(mfgmodel.WorkOrderPlanReportOutput)
	if len(req.WorkOrderPlanReportOutputLines) > 0 {
		h.GetByID(r3, req.WorkOrderPlanReport.WorkOrderPlanReportOutputID)

		req.WorkOrderPlanReportOutputLines = lo.Map(req.WorkOrderPlanReportOutputLines, func(d mfgmodel.WorkOrderOutputItem, i int) mfgmodel.WorkOrderOutputItem {
			d.Date = req.WorkOrderPlanReport.WorkDate
			return d
		})

		r3.ID = req.WorkOrderPlanReport.WorkOrderPlanReportOutputID // will be "" if new
		r3.WorkOrderPlanReportID = req.WorkOrderPlanReport.ID
		r3.WorkOrderPlanID = req.WorkOrderPlanID
		r3.CompanyID = wop.CompanyID
		r3.JournalTypeID = wop.JournalTypeID
		r3.TrxType = mfgmodel.JournalWorkOrderReportConsumption
		r3.Lines = req.WorkOrderPlanReportOutputLines
		r3.Dimension = wop.Dimension

		//handle flow reopen
		if req.WorkOrderPlanReport.WorkOrderPlanReportOutputStatus == string(ficomodel.JournalStatusDraft) {
			if req.WorkOrderPlanReport.WorkOrderPlanReportOutputStatus != string(r3.Status) && (r3.Status == ficomodel.JournalStatusPosted || r3.Status == ficomodel.JournalStatusRejected) {
				r3.Status = ficomodel.JournalStatusDraft

				// //remove posting approval
				// err := h.DeleteByFilter(new(ficomodel.PostingApproval), dbflex.Eq("_id", r3.ID))
				// if err != nil {
				// 	return nil, err
				// }
			}
		}

		if r3.TrxDate.IsZero() {
			r3.TrxDate = time.Now()
		}

		if r3.Status == "" {
			r3.Status = ficomodel.JournalStatusDraft
		}

		err = h.Save(r3)
		if err != nil {
			return nil, err
		}
	}

	req.WorkOrderPlanReport.WorkOrderPlanReportConsumptionID = r1.ID
	req.WorkOrderPlanReport.WorkOrderPlanReportResourceID = r2.ID
	req.WorkOrderPlanReport.WorkOrderPlanReportOutputID = r3.ID
	req.WorkOrderPlanReport.WorkOrderPlanReportConsumptionPPID = jt.PostingProfileIDConsumption
	req.WorkOrderPlanReport.WorkOrderPlanReportResourcePPID = jt.PostingProfileIDResource
	req.WorkOrderPlanReport.WorkOrderPlanReportOutputPPID = jt.PostingProfileIDOutput

	h.Update(&req.WorkOrderPlanReport,
		"WorkOrderPlanReportConsumptionID",
		"WorkOrderPlanReportResourceID",
		"WorkOrderPlanReportOutputID",
		"WorkOrderPlanReportConsumptionPPID",
		"WorkOrderPlanReportResourcePPID",
		"WorkOrderPlanReportOutputPPID",
	)

	return req, nil
}

type WorkOrderPlanReportSubmitReq struct {
	ID string // WorkOrderPlanReportID
}

func (o *WorkOrderPlanReportEngine) Submit(ctx *kaos.Context, req *WorkOrderPlanReportSubmitReq) (interface{}, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	wopr := new(mfgmodel.WorkOrderPlanReport)
	if e := h.GetByID(wopr, req.ID); e != nil {
		return nil, e
	}

	if wopr.WorkOrderPlanReportConsumptionID != "" {
		mat := new(mfgmodel.WorkOrderPlanReportConsumption)
		if e := h.GetByID(mat, wopr.WorkOrderPlanReportConsumptionID); e != nil {
			return nil, e
		}

		wop := new(mfgmodel.WorkOrderPlan)
		if e := h.GetByID(wop, wopr.WorkOrderPlanID); e != nil {
			return nil, e
		}

		// TODO: kalo ada validasi ini, waktu Daily Report jadi ga bisa Submit kalo di ItemBalance + (Reserved - Used) tidak mencukupi, padahal kan bisa IR -> make sure this!
		// validate each NEW line (haven't exist in Summary Material) Qty against current available stock
		sums := []mfgmodel.WorkOrderSummaryMaterial{}
		h.GetsByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", wop.ID), &sums)
		sumM := lo.SliceToMap(sums, func(d mfgmodel.WorkOrderSummaryMaterial) (string, mfgmodel.WorkOrderSummaryMaterial) {
			return fmt.Sprintf("%s||%s", d.ItemID, d.SKU), d
		})

		itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
		unavailableItemNames := []string{}
		for _, line := range mat.Lines {
			qtyCheck := line.Qty

			if sum, exist := sumM[fmt.Sprintf("%s||%s", line.ItemID, line.SKU)]; exist {
				// item dan sku ada di summary material
				if sum.Reserved >= sum.Used+line.Qty {
					continue // bypass validation, item dan sku ini ada & cukup in Summary Material
				}

				qtyCheck = (line.Qty + sum.Used) - sum.Reserved // item dan sku ini lebih dari yang telah di reserved, cek sisanya
			}

			bals := GetAvailableStocks(h, GetAvailableStocksParam{
				CompanyID: wop.CompanyID,
				InventDim: line.InventDim,
				Items:     []GetAvailableStocksParamItem{{ItemID: line.ItemID, SKU: line.SKU}},
			})
			if len(bals) > 0 && bals[0].Qty >= qtyCheck {
				continue // balance qty is enough
			}

			// balance qty not enough, Mohon maaf bapak barangnya tidak cukup, jadi bapak tidak dapat bekerja, silahkan coba lagi besok :)
			item, _ := itemORM.Get(line.ItemID)
			unavailableItemNames = append(unavailableItemNames, item.Name)
		}

		if len(unavailableItemNames) > 0 {
			return nil, fmt.Errorf("Item not available: %s", strings.Join(unavailableItemNames, ", "))
		}

		if e := submitPostingProfile(ctx, mat.CompanyID, userID, mat.ID, mfgmodel.JournalWorkOrderReportConsumption); e != nil {
			return nil, e
		}
	}

	if wopr.WorkOrderPlanReportResourceID != "" {
		rsc := new(mfgmodel.WorkOrderPlanReportResource)
		if e := h.GetByID(rsc, wopr.WorkOrderPlanReportResourceID); e != nil {
			return nil, e
		}

		if e := submitPostingProfile(ctx, rsc.CompanyID, userID, rsc.ID, mfgmodel.JournalWorkOrderReportResource); e != nil {
			return nil, e
		}
	}

	if wopr.WorkOrderPlanReportOutputID != "" {
		out := new(mfgmodel.WorkOrderPlanReportOutput)
		if e := h.GetByID(out, wopr.WorkOrderPlanReportOutputID); e != nil {
			return nil, e
		}

		if e := submitPostingProfile(ctx, out.CompanyID, userID, out.ID, mfgmodel.JournalWorkOrderReportOutput); e != nil {
			return nil, e
		}
	}

	wopr.Status = mfgmodel.WorkOrderPlanStatusOverallInProgress
	h.Update(wopr, "Status")

	return "ok", nil
}

type WOPlanCreateItemRequestParam struct {
	WorkOrderPlanReportID string
}

func (o *WorkOrderPlanReportEngine) CreateAdditionalItemRequest(ctx *kaos.Context, req *WOPlanCreateItemRequestParam) (interface{}, error) {
	coID, userID, e := GetCompanyAndUserIDFromContext(ctx)
	if e != nil {
		return nil, e
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	rc := new(mfgmodel.WorkOrderPlanReportConsumption)
	if e := h.GetByAttr(rc, "WorkOrderPlanReportID", req.WorkOrderPlanReportID); e != nil {
		return nil, e
	}

	wop := new(mfgmodel.WorkOrderPlan)
	if e := h.GetByID(wop, rc.WorkOrderPlanID); e != nil {
		return nil, e
	}

	irLines := lo.Map(rc.AdditionalLines, func(line mfgmodel.WorkOrderMaterialItem, i int) ItemRequestLineParam {
		inventDim := scmlogic.NewInventDimHelper(scmlogic.InventDimHelperOpt{DB: h, SKU: line.SKU}).TernaryInventDimension(&line.InventDim)

		return ItemRequestLineParam{
			ItemID:       line.ItemID,
			SKU:          line.SKU,
			QtyRequested: line.Qty,
			UoM:          line.UnitID,
			WarehouseID:  inventDim.WarehouseID,
			InventDimTo:  line.InventDim,
			Dimension:    wop.Dimension,
		}
	})

	if len(irLines) == 0 {
		return nil, fmt.Errorf("")
	}

	_, err := ItemRequestInsert(ctx, &ItemRequestInsertParam{
		Name:        fmt.Sprintf("FROM WO: %s", wop.ID),
		WOReff:      wop.ID,
		Requestor:   wop.RequestorName,
		Department:  wop.RequestorDepartment,
		InventDimTo: irLines[0].InventDimTo,
		Dimension:   wop.Dimension,
		Lines:       irLines,
	}, coID, userID)
	if err != nil {
		return nil, err
	}

	// err = Posting([]ficologic.PostRequest{{
	// 	JournalType: tenantcoremodel.TrxModule(scmmodel.ItemRequestType),
	// 	JournalID:   irID,
	// 	Op:          ficologic.PostOpSubmit,
	// }}, coID, userID)
	// if err != nil {
	// 	return nil, err
	// }

	h.Update(&mfgmodel.WorkOrderPlanReport{
		ID:           req.WorkOrderPlanReportID,
		HasRequested: true,
	}, "HasRequested")

	return "OK", nil
}

func submitPostingProfile(ctx *kaos.Context, companyID, userID string, journalID string, trxType mfgmodel.InventTrxType) error {
	_, e := new(PostingProfileHandler).Post(ctx, []ficologic.PostRequest{{
		JournalType: tenantcoremodel.TrxModule(trxType),
		JournalID:   journalID,
		Op:          ficologic.PostOpSubmit,
	}})

	// preview := []*tenantcoremodel.PreviewReport{}
	// e := Config.EventHub.Publish("/v1/mfg/postingprofile/post",
	// 	[]ficologic.PostRequest{{
	// 		JournalType: tenantcoremodel.TrxModule(trxType),
	// 		JournalID:   journalID,
	// 		Op:          ficologic.PostOpSubmit,
	// 	}},
	// 	&preview,
	// 	&kaos.PublishOpts{Headers: codekit.M{"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}},
	// )
	// fmt.Printf("%s | journal id: %s | /mfg/postingprofile/post e: %s\n", trxType, journalID, e)

	return e
}

func UpdateWOPRWhenAllChildReportPosted(h *datahub.Hub, workOrderPlanReportID, runFrom string) error {
	wopr := new(mfgmodel.WorkOrderPlanReport)
	if e := h.GetByID(wopr, workOrderPlanReportID); e != nil {
		return e
	}

	if runFrom != "Consumption" {
		cons := new(mfgmodel.WorkOrderPlanReportConsumption)
		h.GetByID(cons, wopr.WorkOrderPlanReportConsumptionID)
		if cons.ID != "" && cons.Status != ficomodel.JournalStatusPosted {
			return nil
		}
	}

	if runFrom != "Resource" {
		res := new(mfgmodel.WorkOrderPlanReportResource)
		h.GetByID(res, wopr.WorkOrderPlanReportResourceID)
		if res.ID != "" && res.Status != ficomodel.JournalStatusPosted {
			return nil
		}
	}

	if runFrom != "Output" {
		out := new(mfgmodel.WorkOrderPlanReportOutput)
		h.GetByID(out, wopr.WorkOrderPlanReportOutputID)
		if out.ID != "" && out.Status != ficomodel.JournalStatusPosted {
			return nil
		}
	}

	// if goes here, then all child reports are posted then update wopr
	wopr.Status = mfgmodel.WorkOrderPlanStatusOverallEnd
	if e := h.Update(wopr, "Status"); e != nil {
		return e
	}

	return nil
}
