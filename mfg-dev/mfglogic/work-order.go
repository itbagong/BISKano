package mfglogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type WorkOrderEngine struct{}

type ItemRequestRequest struct {
	WOID string
}

func (o *WorkOrderEngine) ItemRequest(ctx *kaos.Context, payload *ItemRequestRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	var e error
	//get wo
	wo, e := datahub.GetByID(h, new(mfgmodel.WorkOrder), payload.WOID)
	if e != nil {
		return nil, fmt.Errorf("missing: wo")
	}

	wo.Status = "WAITING"
	e = h.Save(wo)
	if e != nil {
		return nil, e
	}

	now := time.Now()
	itemRequest := scmmodel.ItemRequest{
		Name:        wo.Name,
		Dimension:   wo.Dimension,
		InventDimTo: wo.InventDim,
		CompanyID:   wo.CompanyID,
		TrxDate:     now,
		WOReff:      wo.ID,
	}

	e = h.Save(&itemRequest)
	if e != nil {
		return nil, e
	}

	lo.ForEach(wo.WorkDescriptions, func(description mfgmodel.WorkDescription, index int) {
		// return
		lo.ForEach(description.ItemUsage, func(item mfgmodel.WorkDescriptionItem, index int) {
			itemDetail := scmmodel.ItemRequestDetail{
				ItemID:        item.ItemID,
				ItemRequestID: itemRequest.ID,
				SKU:           item.SKU,
				QtyRequested:  item.Qty,
				UoM:           item.UnitID,
			}

			e = h.Save(&itemDetail)
			if e != nil {
				return
			}
		})
	})

	if e != nil {
		return nil, e
	}

	return wo, nil
}

type ItemRequestUpdateStatus struct {
	WOID   string
	Status mfgmodel.WOStatus
}

func (o *WorkOrderEngine) UpdateStatus(ctx *kaos.Context, payload *ItemRequestUpdateStatus) (interface{}, error) {
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

	var e error
	//get wo
	wo, e := datahub.GetByID(h, new(mfgmodel.WorkOrder), payload.WOID)
	if e != nil {
		return nil, fmt.Errorf("missing: wo")
	}

	wo.Status = payload.Status
	e = h.Save(wo)

	if wo.SunID != "" {
		payload := struct {
			SunID   string
			AssetID string
		}{
			wo.SunID,
			wo.EquipmentNo,
		}

		Config.EventHub.Publish("/v1/sdp/documentunitchecklist/save-from-wo", payload, nil, nil)
	}

	for wdi, wd := range wo.WorkDescriptions {
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
			Name:         fmt.Sprintf("FROM %s", wo.ID),
			RequestDate:  &now,
			DocumentDate: &now,
			TrxDate:      now,
			WOReff:       wo.ID,
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

		for usgi, usg := range wd.ItemUsage {
			paramusg := struct {
				ItemRequestID string
				ItemID        string
				SKU           string
				QtyRequested  float64
				UoM           string
				QtyAvailable  float64
				Remarks       string
				WarehouseID   string
			}{
				ItemRequestID: irID,
				ItemID:        usg.ItemID,
				SKU:           usg.SKU,
				QtyRequested:  usg.Qty,
				UoM:           usg.UnitID,
				Remarks:       usg.Remarks,
				WarehouseID:   usg.InventDim.WarehouseID,
			}

			var usgResult interface{}
			err := Config.EventHub.Publish(
				"/v1/scm/item/request/detail/insert",
				&paramusg,
				&usgResult,
				&kaos.PublishOpts{Headers: codekit.M{"CompanyID": wo.CompanyID, sebar.CtxJWTReferenceID: userID}},
			)
			fmt.Printf("journal id: %s | /v1/scm/item/request/detail/insert e: %s | res: %s\n", wo.ID, err, codekit.JsonStringIndent(usgResult, "\t"))

			wo.WorkDescriptions[wdi].ItemUsage[usgi].Requested = true
		}
	}

	if e := h.Save(wo); e != nil {
		return nil, e
	}

	return wo, e
}
