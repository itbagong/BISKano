package mfglogic

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWRoutine() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		// force sort by ExecutionDate DESC when sort is -_id
		if payload != nil {
			if req, ok := payload.(*dbflex.QueryParam); ok {
				if len(req.Sort) == 1 && req.Sort[0] == "-_id" {
					req.Sort[0] = "-ExecutionDate"
					serde.Serde(req, payload)
				}
			}
		}

		return true, nil
	}
}

func MWPostRoutine() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		siteIDs := lo.Map(ms, func(m codekit.M, index int) interface{} {
			return m.GetString("SiteID")
		})

		dims := []tenantcoremodel.DimensionMaster{}
		if e := h.GetsByFilter(new(tenantcoremodel.DimensionMaster), dbflex.In("_id", siteIDs...), &dims); e != nil {
			ctx.Log().Errorf("Failed populate data: %s", e.Error())
		}

		dimM := lo.Associate(dims, func(elem tenantcoremodel.DimensionMaster) (string, tenantcoremodel.DimensionMaster) {
			return elem.ID, elem
		})

		for _, m := range ms {
			siteID := m.GetString("SiteID")
			if dim, ok := dimM[siteID]; ok {
				m.Set("SiteName", dim.Label)
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostRoutineDetail() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		mapMasterData := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))

		serde.Serde(res["data"], &ms)

		assetIDs := lo.Map(ms, func(m codekit.M, index int) interface{} {
			return m.GetString("AssetID")
		})

		assets := []tenantcoremodel.Asset{}
		if e := h.GetsByFilter(new(tenantcoremodel.Asset), dbflex.In("_id", assetIDs...), &assets); e != nil {
			ctx.Log().Errorf("Failed populate data: %s", e.Error())
		}

		assetM := lo.Associate(assets, func(elem tenantcoremodel.Asset) (string, tenantcoremodel.Asset) {
			return elem.ID, elem
		})

		for _, m := range ms {
			assetID := m.GetString("AssetID")
			if as, ok := assetM[assetID]; ok {
				masterData, _ := mapMasterData.Get(as.AssetType)
				m.Set("AssetName", as.Name)
				m.Set("AssetTypeName", masterData.Name)
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPreWorkOrderSave() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		if m.GetString("WOType") == "PRODUKSI" {
			// validate WO with WOType = PRODUKSI
			if m.GetString("WorkRequestID") == "" {
				return false, fmt.Errorf("Work Request ID is mandatory when WOType is PRODUKSI")
			}

			h := sebar.GetTenantDBFromContext(ctx)
			if h == nil {
				return false, errors.New("missing: db connection")
			}

			wr := new(mfgmodel.WorkRequest)
			if e := h.GetByID(wr, m.GetString("WorkRequestID")); e != nil {
				return false, fmt.Errorf("work request data not found: %s, e: %s", m.GetString("WorkRequestID"), e)
			}

			if wr.SourceType == mfgmodel.WRSourceTypeRoutineCheck {
				return false, fmt.Errorf("Work Request - Source Type is Routine Check, could not proceed WO with PRODUKSI")
			}
		}

		return true, nil
	}
}

func MWPostWorkOrder() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		coID, userID, err := GetCompanyAndUserIDFromContext(ctx)
		if err != nil {
			return false, err
		}

		if _, ok := ctx.Data().Data()["FnResult"]; !ok {
			return true, nil
		}

		res := ctx.Data().Data()["FnResult"]
		m := new(mfgmodel.WorkOrder)
		if e := serde.Serde(res, &m); e != nil {
			fmt.Println("MWPostWorkOrder() ERROR SERDE:", e)
			return true, nil // bypass anything if error occurs
		}

		// inject AvailableStock
		for wdi, wd := range m.WorkDescriptions {
			for usgi, usg := range wd.ItemUsage {
				avail := "Not Available"

				// get balance
				parm := []scmlogic.ItemBalanceGetQtyRequest{{
					ItemID:    usg.ItemID,
					CompanyID: m.CompanyID,
					SKU:       usg.SKU,
					InventDim: usg.InventDim,
				}}
				bals := []scmlogic.ItemBalanceGetQtyResponse{}

				err := Config.EventHub.Publish(
					"/v1/scm/item/balance/get-qty",
					&parm,
					&bals,
					&kaos.PublishOpts{Headers: codekit.M{"CompanyID": coID, sebar.CtxJWTReferenceID: userID}},
				)
				if err == nil && len(bals) > 0 {
					if bals[0].Qty == 0 {
						avail = "Not Available"
					} else if bals[0].Qty >= usg.Qty {
						avail = "Available"
					} else if bals[0].Qty < usg.Qty {
						avail = "Insufficient"
					}
				}

				m.WorkDescriptions[wdi].ItemUsage[usgi].AvailableStock = avail
			}
		}

		ctx.Data().Set("FnResult", m)
		return true, nil
	}
}

func MWPostWorkOrderDailyReport() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		coID, userID, err := GetCompanyAndUserIDFromContext(ctx)
		if err != nil {
			return false, err
		}

		h := sebar.GetTenantDBFromContext(ctx)
		res := ctx.Data().Data()["FnResult"]

		m := new(mfgmodel.WorkOrderDailyReport)
		if e := serde.Serde(res, &m); e != nil {
			fmt.Println("MWPostWorkOrderDailyReport() ERROR SERDE:", e)
			return true, nil // bypass anything if error occurs
		}

		wo := new(mfgmodel.WorkOrder)
		h.GetByID(wo, m.WorkOrderJournalID)

		// inject AvailableStock
		for usgi, usg := range m.ItemUsage {
			avail := "Not Available"

			// get balance
			parm := []scmlogic.ItemBalanceGetQtyRequest{{
				ItemID:    usg.ItemID,
				CompanyID: wo.CompanyID,
				SKU:       usg.SKU,
				InventDim: usg.InventDim,
			}}
			bals := []scmlogic.ItemBalanceGetQtyResponse{}

			err := Config.EventHub.Publish(
				"/v1/scm/item/balance/get-qty",
				&parm,
				&bals,
				&kaos.PublishOpts{Headers: codekit.M{"CompanyID": coID, sebar.CtxJWTReferenceID: userID}},
			)
			if err == nil && len(bals) > 0 {
				if bals[0].Qty == 0 {
					avail = "Not Available"
				} else if bals[0].Qty >= usg.Qty {
					avail = "Available"
				} else if bals[0].Qty < usg.Qty {
					avail = "Insufficient"
				}
			}

			m.ItemUsage[usgi].AvailableStock = avail
		}

		ctx.Data().Set("FnResult", m)
		return true, nil
	}
}

func MWPostWorkOrderGrid() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		mapWorkRequests := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkRequest))
		mapJournalType := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkOrderJournalType))
		mapPostingProfile := sebar.NewMapRecordWithORM(h, new(ficomodel.PostingProfile))
		mapEmployee := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Employee))
		mapMasterData := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))
		mapAssets := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Asset))

		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		lo.ForEach(ms, func(row codekit.M, index int) {
			workRequestID := row.GetString("WorkRequestID")
			woJournalTypeID := row.GetString("JournalTypeID")
			postingProfileID := row.GetString("PostingProfileID")
			employeeID := row.GetString("Name")
			departmentID := row.GetString("RequestorDepartment")
			assetID := row.GetString("EquipmentNo")
			assignedPICID := row.GetString("AssignedPIC")

			workRequest, _ := mapWorkRequests.Get(workRequestID)
			journalType, _ := mapJournalType.Get(woJournalTypeID)
			postingProfile, _ := mapPostingProfile.Get(postingProfileID)
			employee, _ := mapEmployee.Get(employeeID)
			department, _ := mapMasterData.Get(departmentID)
			asset, _ := mapAssets.Get(assetID)
			pic, _ := mapEmployee.Get(assignedPICID)

			row.Set("WorkRequestID", workRequest.Name)
			row.Set("JournalTypeID", journalType.Name)
			row.Set("PostingProfileID", postingProfile.Name)
			row.Set("Name", employee.Name)
			row.Set("RequestorDepartment", department.Name)
			row.Set("EquipmentNo", asset.Name)
			row.Set("AssignedPIC", pic.Name)

			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil

	}
}

func MWPostWorkRequestGrid() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		mapJournalType := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkRequestorJournalType))
		mapPostingProfile := sebar.NewMapRecordWithORM(h, new(ficomodel.PostingProfile))
		mapEmployee := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Employee))
		mapMasterData := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))
		mapAssets := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Asset))
		mapDimensions := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.DimensionMaster))

		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		lo.ForEach(ms, func(row codekit.M, index int) {
			wrJournalTypeID := row.GetString("JournalTypeID")
			postingProfileID := row.GetString("PostingProfileID")
			employeeID := row.GetString("Name")
			departmentID := row.GetString("Department")
			assetID := row.GetString("EquipmentNo")
			dimension := row.Get("Dimension", tenantcoremodel.Dimension{}).(tenantcoremodel.Dimension)

			siteID := dimension.Get("Site")
			journalType, _ := mapJournalType.Get(wrJournalTypeID)
			postingProfile, _ := mapPostingProfile.Get(postingProfileID)
			employee, _ := mapEmployee.Get(employeeID)
			department, _ := mapMasterData.Get(departmentID)
			asset, _ := mapAssets.Get(assetID)
			siteDimension, _ := mapDimensions.Get(siteID)

			row.Set("JournalTypeID", journalType.Name)
			row.Set("PostingProfileID", postingProfile.Name)
			row.Set("Name", employee.Name)
			row.Set("Department", department.Name)
			row.Set("EquipmentNo", asset.Name)
			row.Set("Site", siteDimension.Label)

			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostWorkOrderPlanGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		workRequestIDs := []interface{}{}
		jtIDs := []interface{}{}
		empRequestorIDs := []interface{}{}
		assetIDs := []interface{}{}
		warehouseIDs := []interface{}{}
		masterDataIDs := []interface{}{}
		siteMasterIDs := []interface{}{}

		for _, m := range ms {
			wr := m.GetString("WorkRequestID")
			if wr != "" {
				workRequestIDs = append(workRequestIDs, wr)
			}

			jt := m.GetString("JournalTypeID")
			if jt != "" {
				jtIDs = append(jtIDs, jt)
			}

			empRequestor := m.GetString("RequestorWOName")
			if empRequestor != "" {
				empRequestorIDs = append(empRequestorIDs, empRequestor)
			}

			department := m.GetString("RequestorDepartment")
			if department != "" {
				masterDataIDs = append(masterDataIDs, department)
			}

			asset := m.GetString("Asset")
			if asset != "" {
				assetIDs = append(assetIDs, asset)
			}

			inventDim, ok := m.Get("InventDim").(scmmodel.InventDimension)
			if ok {
				warehouse := inventDim.WarehouseID
				if warehouse != "" {
					warehouseIDs = append(warehouseIDs, warehouse)
				}
			}

			dimension, ok := m.Get("Dimension").(tenantcoremodel.Dimension)
			if ok {
				siteID := dimension.Get("Site")
				if siteID != "" {
					siteMasterIDs = append(siteMasterIDs, siteID)
				}
			}

			if m.GetString("JournalTypeID") == "WO_PrevMaintenance" {
				woName := m.GetString("WOName")
				masterDataIDs = append(masterDataIDs, woName)
			}
		}

		workRequestKV := map[string]*mfgmodel.WorkRequest{}
		if len(workRequestIDs) > 0 {
			workrequests, _ := datahub.Find(h, new(mfgmodel.WorkRequest), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", workRequestIDs...)))
			workRequestKV = lo.Associate(workrequests, func(wr *mfgmodel.WorkRequest) (string, *mfgmodel.WorkRequest) {
				return wr.ID, wr
			})
		}

		jtKV := map[string]*mfgmodel.WorkOrderJournalType{}
		if len(jtIDs) > 0 {
			jts, _ := datahub.Find(h, new(mfgmodel.WorkOrderJournalType), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", jtIDs...)))
			jtKV = lo.Associate(jts, func(jt *mfgmodel.WorkOrderJournalType) (string, *mfgmodel.WorkOrderJournalType) {
				return jt.ID, jt
			})
		}

		empRequestorKV := map[string]*tenantcoremodel.Employee{}
		if len(empRequestorIDs) > 0 {
			emps, _ := datahub.Find(h, new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", empRequestorIDs...)))
			empRequestorKV = lo.Associate(emps, func(emp *tenantcoremodel.Employee) (string, *tenantcoremodel.Employee) {
				return emp.ID, emp
			})
		}

		assetKV := map[string]*tenantcoremodel.Asset{}
		if len(assetIDs) > 0 {
			assets, _ := datahub.Find(h, new(tenantcoremodel.Asset), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", assetIDs...)))
			assetKV = lo.Associate(assets, func(asset *tenantcoremodel.Asset) (string, *tenantcoremodel.Asset) {
				return asset.ID, asset
			})
		}

		warehouseKV := map[string]*tenantcoremodel.LocationWarehouse{}
		if len(warehouseIDs) > 0 {
			warehouses, _ := datahub.Find(h, new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", warehouseIDs...)))
			warehouseKV = lo.Associate(warehouses, func(warehouse *tenantcoremodel.LocationWarehouse) (string, *tenantcoremodel.LocationWarehouse) {
				return warehouse.ID, warehouse
			})
		}

		masterDataKV := map[string]*tenantcoremodel.MasterData{}
		if len(masterDataIDs) > 0 {
			masterdatas, _ := datahub.Find(h, new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", masterDataIDs...)))
			masterDataKV = lo.Associate(masterdatas, func(masterData *tenantcoremodel.MasterData) (string, *tenantcoremodel.MasterData) {
				return masterData.ID, masterData
			})
		}

		dimensionMasterKV := map[string]*tenantcoremodel.DimensionMaster{}
		if len(siteMasterIDs) > 0 {
			dimensionMasters, _ := datahub.Find(h, new(tenantcoremodel.DimensionMaster), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", siteMasterIDs...)))
			dimensionMasterKV = lo.Associate(dimensionMasters, func(dimensionMaster *tenantcoremodel.DimensionMaster) (string, *tenantcoremodel.DimensionMaster) {
				return dimensionMaster.ID, dimensionMaster
			})
		}

		for _, m := range ms {
			wr, ok := workRequestKV[m.GetString("WorkRequestID")]
			if !ok {
				wr = new(mfgmodel.WorkRequest)
			}
			m.Set("WorkRequestName", wr.Name)

			jt, ok := jtKV[m.GetString("JournalTypeID")]
			if !ok {
				jt = new(mfgmodel.WorkOrderJournalType)
			}
			m.Set("JournalTypeName", jt.Name)

			empRequestor, ok := empRequestorKV[m.GetString("RequestorWOName")]
			if !ok {
				empRequestor = new(tenantcoremodel.Employee)
			}
			m.Set("RequestorNameFix", empRequestor.Name)

			department, ok := masterDataKV[m.GetString("RequestorDepartment")]
			if !ok {
				department = new(tenantcoremodel.MasterData)
			}
			m.Set("DepartmentName", department.Name)

			asset, ok := assetKV[m.GetString("Asset")]
			if !ok {
				asset = new(tenantcoremodel.Asset)
			}
			m.Set("AssetName", asset.Name)

			inventDim, ok := m.Get("InventDim").(scmmodel.InventDimension)
			if ok {
				warehouse, ok := warehouseKV[inventDim.WarehouseID]
				if !ok {
					warehouse = new(tenantcoremodel.LocationWarehouse)
				}
				m.Set("WarehouseName", warehouse.Name)
			} else {
				m.Set("WarehouseName", "")
			}

			dimension, ok := m.Get("Dimension").(tenantcoremodel.Dimension)
			if ok {
				siteID := dimension.Get("Site")
				siteMaster, ok := dimensionMasterKV[siteID]
				if !ok {
					siteMaster = new(tenantcoremodel.DimensionMaster)
				}
				m.Set("SiteName", siteMaster.Label)
			} else {
				m.Set("SiteName", "-")
			}

			if m.GetString("JournalTypeID") == "WO_PrevMaintenance" {
				mdWOName, ok := masterDataKV[m.GetString("WOName")]
				if !ok {
					mdWOName = new(tenantcoremodel.MasterData)
				}

				if mdWOName.Name != "" {
					m.Set("WOName", mdWOName.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostWorkOrderPlanGet() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res := ctx.Data().Data()["FnResult"]
		if _, ok := res.(codekit.M); ok {
			return true, nil // bypass if api is gets
		}

		m := new(mfgmodel.WorkOrderPlan)
		if e := serde.Serde(res, &m); e != nil {
			return true, nil
		}

		bgAsset, _ := BagongAssetGet(m.Asset)
		if bgAsset != nil {
			if len(bgAsset.CurrentUserInfo) > 0 {
				m.NoHullCustomer = bgAsset.CurrentUserInfo[0].NoHullCustomer
			}
		}

		ctx.Data().Set("FnResult", m)
		return true, nil
	}
}

func MWPostWorkOrderPlanSave() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		// h := sebar.GetTenantDBFromContext(ctx)
		// res := ctx.Data().Data()["FnResult"]

		// m := new(mfgmodel.WorkOrderPlan)
		// if e := serde.Serde(res, &m); e != nil {
		// 	fmt.Println("MWPostWorkOrderPlanSave() ERROR SERDE:", e)
		// 	return true, nil
		// }

		// zeroTime, _ := time.Parse("2006-01-02", "1900-01-01")
		// if m.TrxDate.IsZero() || m.TrxDate.Before(zeroTime) {
		// 	m.TrxDate = time.Now()
		// }

		// Material is controlled by UI already
		// if m.Status == ficomodel.JournalStatusDraft && m.BOM != "" {
		// 	bomMaterials := []mfgmodel.BoMMaterial{}
		// 	h.GetsByFilter(new(mfgmodel.BoMMaterial), dbflex.Eq("BoMID", m.BOM), &bomMaterials)
		// 	h.DeleteByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", m.ID))

		// 	for _, bm := range bomMaterials {
		// 		h.Save(&mfgmodel.WorkOrderSummaryMaterial{
		// 			WorkOrderPlanID: m.ID,
		// 			ItemID:          bm.ItemID,
		// 			SKU:             bm.SKU,
		// 			UnitID:          bm.UoM,
		// 			Required:        float64(bm.Qty),
		// 		})
		// 	}
		// }

		return true, nil
	}
}

func MWPostWorkOrderSummaryMaterialGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		items := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
		specs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
		units := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.UoM))

		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		for _, m := range ms {
			it, _ := items.Get(m.GetString("ItemID"))
			m.Set("ItemName", it.Name)

			sk, _ := specs.Get(m.GetString("SKU"))
			m.Set("SKUName", sk.SKU) // TODO: change to SKU full description similar to /gets-info

			u, _ := units.Get(m.GetString("UnitID"))
			m.Set("UnitName", u.Name)
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostWorkOrderSummaryResourceGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		exps := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ExpenseType))

		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		for _, m := range ms {
			exp, _ := exps.Get(m.GetString("ExpenseType"))
			m.Set("ExpenseTypeName", exp.Name)
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostWorkOrderSummaryOutputGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		items := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
		ledgers := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LedgerAccount))
		specs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
		groups := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemGroup))
		units := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.UoM))

		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		for _, m := range ms {
			switch mfgmodel.WorkOrderOutputType(m.GetString("Type")) {
			case mfgmodel.WorkOrderOutputTypeWOOutput, mfgmodel.WorkOrderOutputTypeWasteItem:
				it, _ := items.Get(m.GetString("InventoryLedgerAccID"))
				m.Set("InventoryLedgerAccName", it.Name)

				gr, _ := groups.Get(m.GetString("Group"))
				m.Set("GroupName", gr.Name)
			case mfgmodel.WorkOrderOutputTypeWasteLedger:
				lg, _ := ledgers.Get(m.GetString("InventoryLedgerAccID"))
				m.Set("InventoryLedgerAccName", lg.Name)
			}

			sk, _ := specs.Get(m.GetString("SKU"))
			m.Set("SKUName", sk.SKU) // TODO: change to SKU full description similar to /gets-info

			u, _ := units.Get(m.GetString("UnitID"))
			m.Set("UnitName", u.Name)
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

type WorkOrderPlanReportSaveReqNorm struct {
	mfgmodel.WorkOrderPlanReport
	WorkOrderPlanReportConsumptionLines []WorkOrderMaterialItemNorm
	WorkOrderPlanReportResourceLines    []WorkOrderResourceItemNorm
	WorkOrderPlanReportOutputLines      []WorkOrderOutputItemNorm
}

type WorkOrderMaterialItemNorm struct {
	mfgmodel.WorkOrderMaterialItem
	ItemName string
	SKUName  string
	UnitName string
}

type WorkOrderResourceItemNorm struct {
	mfgmodel.WorkOrderResourceItem
	ExpenseTypeName string
}

type WorkOrderOutputItemNorm struct {
	mfgmodel.WorkOrderOutputItem
	InventoryLedgerAccName string
	SKUName                string
	GroupName              string
	UnitName               string
}

func MWPostWorkOrderPlanReportGet() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		// h := sebar.GetTenantDBFromContext(ctx)
		res := ctx.Data().Data()["FnResult"]

		m := new(WorkOrderPlanReportSaveReq)
		if e := serde.Serde(res, &m); e != nil {
			fmt.Println("MWPostWorkOrderPlanSave() ERROR SERDE:", e)
			return true, nil
		}

		// TODO: di komeng, masih nyebabin data inconsistency
		// items := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
		// ledgers := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LedgerAccount))
		// specs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
		// groups := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemGroup))
		// units := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.UoM))
		// exps := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ExpenseType))

		// linesCons := lo.Map(m.WorkOrderPlanReportConsumptionLines, func(d mfgmodel.WorkOrderMaterialItem, i int) WorkOrderMaterialItemNorm {
		// 	item, _ := items.Get(d.ItemID)
		// 	spec, _ := specs.Get(d.SKU)
		// 	unit, _ := units.Get(d.UnitID)

		// 	return WorkOrderMaterialItemNorm{
		// 		WorkOrderMaterialItem: d,
		// 		ItemName:              item.Name,
		// 		SKUName:               spec.SKU,
		// 		UnitName:              unit.Name,
		// 	}
		// })

		// linesRes := lo.Map(m.WorkOrderPlanReportResourceLines, func(d mfgmodel.WorkOrderResourceItem, i int) WorkOrderResourceItemNorm {
		// 	exp, _ := exps.Get(d.ExpenseType)

		// 	return WorkOrderResourceItemNorm{
		// 		ExpenseTypeName: exp.Name,
		// 	}
		// })

		// linesOut := lo.Map(m.WorkOrderPlanReportOutputLines, func(d mfgmodel.WorkOrderOutputItem, i int) WorkOrderOutputItemNorm {
		// 	inventoryLedgerAccName := ""

		// 	switch d.Type {
		// 	case mfgmodel.WorkOrderOutputTypeWOOutput, mfgmodel.WorkOrderOutputTypeWasteItem:
		// 		it, _ := items.Get(d.InventoryLedgerAccID)
		// 		inventoryLedgerAccName = it.Name
		// 	case mfgmodel.WorkOrderOutputTypeWasteLedger:
		// 		lg, _ := ledgers.Get(d.InventoryLedgerAccID)
		// 		inventoryLedgerAccName = lg.Name
		// 	}

		// 	spec, _ := specs.Get(d.SKU)
		// 	gr, _ := groups.Get(d.GroupID)
		// 	unit, _ := units.Get(d.UnitID)

		// 	return WorkOrderOutputItemNorm{
		// 		InventoryLedgerAccName: inventoryLedgerAccName,
		// 		SKUName:                spec.SKU,
		// 		GroupName:              gr.Name,
		// 		UnitName:               unit.Name,
		// 	}
		// })

		// newres := WorkOrderPlanReportSaveReqNorm{
		// 	WorkOrderPlanReport:                 m.WorkOrderPlanReport,
		// 	WorkOrderPlanReportConsumptionLines: linesCons,
		// 	WorkOrderPlanReportResourceLines:    linesRes,
		// 	WorkOrderPlanReportOutputLines:      linesOut,
		// }

		// ctx.Data().Set("FnResult", newres)
		ctx.Data().Set("FnResult", m)
		return true, nil
	}
}

// MWPreAssignCompanyID middleware untuk menggunakan CompanyID
func MWPreAssignCompanyID() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		ctx.Data().Set(tenantcoremodel.DBModFilter, []*dbflex.Filter{dbflex.Eq("CompanyID", tenantcorelogic.GetCompanyIDFromContext(ctx))})

		field := "CompanyID"

		if payload == nil {
			return true, nil
		}

		t := reflect.TypeOf(payload)
		if t.Kind() == reflect.Pointer {
			t = reflect.Indirect(reflect.ValueOf(payload)).Type()
		}

		haveCompanyID := false
		if t.Kind() == reflect.Struct {
			for index := 0; index < t.NumField(); index++ {
				if strings.ToLower(t.Field(index).Name) == strings.ToLower(field) {
					haveCompanyID = true
				}
			}
		}

		if !haveCompanyID {
			return true, nil
		}

		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		coID, _ := m[field].(string)
		if coID != "" {
			return true, nil
		}

		coID = tenantcorelogic.GetCompanyIDFromContext(ctx)

		m.Set(field, coID)
		serde.Serde(m, payload)

		return true, nil
	}
}

func MWPostRoutineTemplateGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []mfgmodel.RoutineTemplate{}
		serde.Serde(res["data"], &ms)

		masterDataORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))

		for mI, m := range ms {
			ctg, _ := masterDataORM.Get(m.CategoryID)
			ms[mI].CategoryID = ctg.Name

			mda, _ := masterDataORM.Get(m.AssetType)
			ms[mI].AssetType = mda.Name

			mdd, _ := masterDataORM.Get(m.DriveType)
			ms[mI].DriveType = mdd.Name
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostBomGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []mfgmodel.BoM{}
		serde.Serde(res["data"], &ms)

		LedgerAccount := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LedgerAccount))

		for mI, m := range ms {
			ms[mI].ItemID = tenantcorelogic.ItemVariantName(h, m.ItemID, m.SKU)
			ctg, _ := LedgerAccount.Get(m.LedgerID)
			ms[mI].LedgerID = ctg.Name
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
