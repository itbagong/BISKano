package mfglogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type WODailyReportEngine struct{}

type ManPowerViewRequest struct {
	WorkOrderJournalID string
}

func (o WODailyReportEngine) Resume(ctx *kaos.Context, payload *ManPowerViewRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	//get WODailyReport
	woDailyReports := []mfgmodel.WorkOrderDailyReport{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrderDailyReport), dbflex.Eq("WorkOrderJournalID", payload.WorkOrderJournalID), &woDailyReports)
	if e != nil {
		return nil, fmt.Errorf("Work Order report not found")
	}

	resumes := lo.Map(woDailyReports, func(item mfgmodel.WorkOrderDailyReport, index int) mfgmodel.WorkOrderDailyResume {
		resume := mfgmodel.WorkOrderDailyResume{}
		resume.ID = item.ID
		resume.WorkDescription = item.WorkDescription
		resume.WorkDate = item.WorkDate
		resume.Consumption = len(item.ItemUsage)
		resume.Manpower = len(item.ManpowerUsage)
		resume.Output = len(item.Output)
		resume.Status = item.Status
		return resume
	})
	return resumes, nil
}

func (o *WODailyReportEngine) ManPowerView(ctx *kaos.Context, payload *ManPowerViewRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	//get WODailyReport
	woDailyReports := []mfgmodel.WorkOrderDailyReport{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrderDailyReport), dbflex.Eq("WorkOrderJournalID", payload.WorkOrderJournalID), &woDailyReports)
	if e != nil {
		return nil, fmt.Errorf("Work Order report not found")
	}

	manPowers := lo.MapEntries(lo.GroupBy(woDailyReports, func(item mfgmodel.WorkOrderDailyReport) string {
		return codekit.ToString(item.WorkDescriptionNo) + "|" + item.WorkDescription
	}), func(key string, items []mfgmodel.WorkOrderDailyReport) (string, []mfgmodel.ManPowerUsage) {
		manPowers := []mfgmodel.ManPowerUsage{}
		lo.ForEach(items, func(dailyreport mfgmodel.WorkOrderDailyReport, index int) {
			manPowers = append(manPowers, dailyreport.ManpowerUsage...)
		})

		return key, manPowers
	})

	return manPowers, nil
}

func (o *WODailyReportEngine) OutputView(ctx *kaos.Context, payload *ManPowerViewRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	//get WODailyReport
	woDailyReports := []mfgmodel.WorkOrderDailyReport{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrderDailyReport), dbflex.Eq("WorkOrderJournalID", payload.WorkOrderJournalID), &woDailyReports)
	if e != nil {
		return nil, fmt.Errorf("Work Order report not found")
	}

	outputs := lo.MapEntries(lo.GroupBy(woDailyReports, func(item mfgmodel.WorkOrderDailyReport) string {
		return codekit.ToString(item.WorkDescriptionNo) + "|" + item.WorkDescription
	}), func(key string, items []mfgmodel.WorkOrderDailyReport) (string, []mfgmodel.Output) {
		outputs := []mfgmodel.Output{}
		lo.ForEach(items, func(dailyreport mfgmodel.WorkOrderDailyReport, index int) {
			outputs = append(outputs, dailyreport.Output...)
		})

		return key, outputs
	})

	return outputs, nil
}

func (o *WODailyReportEngine) ItemUsageView(ctx *kaos.Context, payload *ManPowerViewRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	//get WODailyReport
	woDailyReports := []mfgmodel.WorkOrderDailyReport{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrderDailyReport), dbflex.Eq("WorkOrderJournalID", payload.WorkOrderJournalID), &woDailyReports)
	if e != nil {
		return nil, fmt.Errorf("Work Order report not found")
	}

	workDescriptions := lo.GroupBy(woDailyReports, func(item mfgmodel.WorkOrderDailyReport) string {
		return codekit.ToString(item.WorkDescriptionNo) + "|" + item.WorkDescription
	})

	workDescriptionItem := map[string][]mfgmodel.ItemUsage{}
	for key, dailyReports := range workDescriptions {
		itemGroup := map[string]mfgmodel.ItemUsage{}
		lo.ForEach(dailyReports, func(report mfgmodel.WorkOrderDailyReport, index int) {
			lo.ForEach(report.ItemUsage, func(item mfgmodel.ItemUsage, index int) {
				if _, ok := itemGroup[item.ItemID]; !ok {
					itemGroup[item.ItemID] = item
				} else {
					itemExisting := itemGroup[item.ItemID]
					itemExisting.Qty += item.Qty
					itemGroup[item.ItemID] = itemExisting
				}
			})
		})

		itemFinalGroup := []mfgmodel.ItemUsage{}
		for _, item := range itemGroup {
			itemFinalGroup = append(itemFinalGroup, item)
		}

		workDescriptionItem[key] = itemFinalGroup
	}

	return workDescriptionItem, nil
}

type WODRSubmitRequest struct {
	WOID      string
	WODRID    string
	WOStatus  string // btn Submit: "IN PROGRESS", btn Mark As Completed: "COMPLETED"
	Component string
}

func (o *WODailyReportEngine) Submit(ctx *kaos.Context, payload *WODRSubmitRequest) (interface{}, error) {
	coID, userID, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
	// userID := sebar.GetUserIDFromCtx(ctx)
	// if userID == "" {
	// 	userID = "SYSTEM"
	// }

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	wo, e := datahub.GetByID(h, new(mfgmodel.WorkOrder), payload.WOID)
	if e != nil {
		return nil, fmt.Errorf("work order not found")
	}

	wodr, e := datahub.GetByID(h, new(mfgmodel.WorkOrderDailyReport), payload.WODRID)
	if e != nil {
		return nil, fmt.Errorf("work order daily report not found")
	}

	lineNumber := 1
	itemUsageGI := lo.Map(wodr.ItemUsage, func(item mfgmodel.ItemUsage, index int) *scmmodel.InventTrx {
		inventTrx := new(scmmodel.InventTrx)
		if item.InventReceiveIssueLine.Qty > 0 {
			inventTrx.Qty *= -1
		}

		itemTenant, _ := datahub.GetByID(h, new(tenantcoremodel.Item), item.ItemID)
		inventTrx.Item = *itemTenant
		inventTrx.SKU = item.SKU
		inventTrx.InventDim = wo.InventDim
		inventTrx.Dimension = wo.Dimension
		inventTrx.TrxDate = time.Now()
		inventTrx.Status = scmmodel.ItemConfirmed
		inventTrx.TrxQty = inventTrx.Qty
		inventTrx.TrxUnitID = item.UnitID
		inventTrx.SourceType = tenantcoremodel.TrxModule(scmmodel.JournalWorkOrder)
		inventTrx.SourceTrxType = string(scmmodel.JournalWorkOrder)
		inventTrx.SourceJournalID = wo.ID
		inventTrx.SourceLineNo = lineNumber
		inventTrx.SourceLineNo = lineNumber
		lineNumber++

		return inventTrx
	})

	lineNumber = 1
	lineGR := []*scmmodel.InventTrx{}
	lo.ForEach(wodr.Output, func(item mfgmodel.Output, index int) {
		if item.Output == "item" {
			inventTrx := new(scmmodel.InventTrx)
			if item.Qty < 0 {
				inventTrx.Qty = float64(item.Qty * -1)
			}

			itemTenant, _ := datahub.GetByID(h, new(tenantcoremodel.Item), item.ItemID)
			inventTrx.Item = *itemTenant
			inventTrx.SKU = item.SKU
			inventTrx.InventDim = wo.InventDim
			inventTrx.Dimension = wo.Dimension
			inventTrx.TrxDate = time.Now()
			inventTrx.Status = scmmodel.ItemConfirmed
			inventTrx.TrxQty = inventTrx.Qty
			inventTrx.TrxUnitID = item.UOM
			inventTrx.SourceType = tenantcoremodel.TrxModule(scmmodel.JournalWorkOrder)
			inventTrx.SourceTrxType = string(scmmodel.JournalWorkOrder)
			inventTrx.SourceJournalID = wo.ID
			inventTrx.SourceLineNo = lineNumber
			inventTrx.SourceLineNo = lineNumber
			lineNumber++

			lineGR = append(lineGR, inventTrx)
		}
	})

	if len(itemUsageGI) > 0 {
		var e error
		lo.ForEach(itemUsageGI, func(item *scmmodel.InventTrx, index int) {
			e = h.Save(item)
			if e != nil {
				return
			}
		})

		if e != nil {
			return nil, e
		}

		if _, err := scmlogic.NewInventBalanceCalc(h).Sync(itemUsageGI); err != nil {
			return nil, fmt.Errorf("update balance: %s", err.Error())
		}
	}

	if len(lineGR) > 0 {
		var e error
		lo.ForEach(lineGR, func(item *scmmodel.InventTrx, index int) {
			e = h.Save(item)
			if e != nil {
				return
			}
		})

		if e != nil {
			return nil, e
		}

		if _, err := scmlogic.NewInventBalanceCalc(h).Sync(lineGR); err != nil {
			return nil, fmt.Errorf("update balance: %s", err.Error())
		}
	}

	// auto create item request from wo daily report
	unRequestedAddItems := lo.FilterMap(wodr.AdditionalItem, func(item mfgmodel.ItemUsage, i int) (mfgmodel.ItemUsage, bool) {
		return item, item.Requested == false // check any unrequested additional item
	})
	if len(unRequestedAddItems) > 0 {
		// Create Item Request for Additional Items
		now := time.Now()

		param := struct {
			CompanyID    string
			Name         string // FROM [WO_ID]
			RequestDate  *time.Time
			DocumentDate *time.Time
			TrxDate      time.Time
			WOReff       string
			Requestor    string
			Department   string
			TrxType      string // "Item Request"
			Status       string // "Draft"
			InventDimTo  struct {
				WarehouseID string
				AisleID     string
				SectionID   string
				BoxID       string
			}
		}{
			CompanyID:    coID,
			Name:         fmt.Sprintf("FROM %s", wodr.ID),
			RequestDate:  &now,
			DocumentDate: &now,
			TrxDate:      now,
			WOReff:       wodr.ID,
			Requestor:    wo.Name,
			Department:   wo.RequestorDepartment,
			TrxType:      "Item Request",
			Status:       string(ficomodel.JournalStatusDraft),
			InventDimTo: struct {
				WarehouseID string
				AisleID     string
				SectionID   string
				BoxID       string
			}{
				WarehouseID: wo.InventDim.WarehouseID,
				AisleID:     wo.InventDim.AisleID,
				SectionID:   wo.InventDim.SectionID,
				BoxID:       wo.InventDim.BoxID,
			},
		}

		var irResult interface{}
		e = Config.EventHub.Publish(
			"/v1/scm/item/request/insert",
			&param,
			&irResult,
			&kaos.PublishOpts{Headers: codekit.M{"CompanyID": wo.CompanyID, sebar.CtxJWTReferenceID: userID}},
		)
		fmt.Printf("journal id: %s | /v1/scm/item/request/insert e: %s | res: %s\n", wo.ID, e, codekit.JsonStringIndent(irResult, "\t"))

		irM, _ := codekit.ToM(irResult)
		irID := irM.GetString("_id")

		for itemi, item := range wodr.AdditionalItem {
			if item.Requested {
				continue // only create unrequested one
			}

			desc, _ := GetItemSpecDescription(h, item.SKU)

			paramusg := struct {
				ItemRequestID string
				ItemID        string
				SKU           string
				Description   string
				QtyRequested  float64
				UoM           string
				Remarks       string
				WarehouseID   string
			}{
				ItemRequestID: irID,
				ItemID:        item.ItemID,
				SKU:           item.SKU,
				Description:   desc,
				QtyRequested:  item.Qty,
				UoM:           item.UnitID,
				Remarks:       item.Remarks,
				WarehouseID:   item.InventDim.WarehouseID,
			}

			var usgResult interface{}
			err := Config.EventHub.Publish(
				"/v1/scm/item/request/detail/insert",
				&paramusg,
				&usgResult,
				&kaos.PublishOpts{Headers: codekit.M{"CompanyID": wo.CompanyID, sebar.CtxJWTReferenceID: userID}},
			)
			fmt.Printf("journal id: %s | /v1/scm/item/request/detail/insert e: %s | res: %s\n", wodr.ID, err, codekit.JsonStringIndent(usgResult, "\t"))

			wodr.AdditionalItem[itemi].Requested = true
		}

		h.Update(wodr, "AdditionalItem")
	}

	wodr.Status = mfgmodel.WOStatus(payload.WOStatus)
	h.Update(wodr, "Status")

	// check other Daily Report under the same WO
	allWODRS := []mfgmodel.WorkOrderDailyReport{}
	if e := h.GetsByFilter(new(mfgmodel.WorkOrderDailyReport), dbflex.Eq("WorkOrderJournalID", wo.ID), &allWODRS); e != nil {
		return nil, e
	}

	_, idx, exist := lo.FindIndexOf(allWODRS, func(d mfgmodel.WorkOrderDailyReport) bool {
		return d.Status != mfgmodel.WOStatusCompleted
	})
	if !exist && idx == -1 {
		// all Daily Reports have been COMPLETED, then update WO
		wo.Status = mfgmodel.WOStatusCompleted
		h.Update(wo, "Status")
	}

	return wodr, nil
}
