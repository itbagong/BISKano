package scmlogic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostItemRequest() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		mapEmployees := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Employee))
		// mapMasterDatas := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.MasterData))
		mapWarehouses := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))

		ms := []scmmodel.ItemRequest{}
		serde.Serde(res["data"], &ms)

		lo.ForEach(ms, func(row scmmodel.ItemRequest, index int) {
			// departmentID := row.GetString("Department")
			requestor := row.Requestor
			// priorityID := row.GetString("Priority")
			inventDim := row.InventDimTo

			employee, _ := mapEmployees.Get(requestor)
			// department, _ := mapMasterDatas.Get(departmentID)
			warehouse, _ := mapWarehouses.Get(inventDim.WarehouseID)
			// priority, _ := mapMasterDatas.Get(priorityID)

			// employee, _ := datahub.GetByID(h, new(tenantcoremodel.Employee), requestor)
			// department, _ := datahub.GetByID(h, new(tenantcoremodel.MasterData), departmentID)
			// warehouse, _ := datahub.GetByID(h, new(tenantcoremodel.LocationWarehouse), inventDim.WarehouseID)

			inventDim.WarehouseID = warehouse.Name
			// row.Set("Department", department.Name)
			// row.Set("Requestor", employee.Name)
			// row.Set("InventDimTo", inventDim)
			// row.Set("Priority", priority.Name)
			row.Requestor = employee.Name
			row.InventDimTo = inventDim

			approvers := []string{}
			nextApprover, _ := ficologic.FindNextApproval(h, row.ID)
			if len(nextApprover) > 0 {
				approvers = lo.Map(nextApprover, func(approver ficologic.ApprovalLogHandlerGetResponse, index int) string {
					return approver.Text
				})
			}

			row.Approvers = approvers
			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

// sementara taruh sini nanti tak pindah
func MWPostPurchaseRequestGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lo.ForEach(ms, func(row codekit.M, index int) {
			companyID := row.GetString("CompanyID")
			vendorID := row.GetString("VendorID")
			inventDim := row.Get("Location", scmmodel.InventDimension{}).(scmmodel.InventDimension)
			warehouseID := row.GetString("WarehouseID")

			company, _ := datahub.GetByID(h, new(tenantcoremodel.Company), companyID)
			vendor, _ := datahub.GetByID(h, new(tenantcoremodel.Vendor), vendorID)
			warehouse, _ := datahub.GetByID(h, new(tenantcoremodel.LocationWarehouse), inventDim.WarehouseID)
			warehouse1, _ := datahub.GetByID(h, new(tenantcoremodel.LocationWarehouse), warehouseID)

			inventDim.WarehouseID = warehouse.Name
			row.Set("CompanyID", company.Name)
			row.Set("VendorID", vendor.Name)
			row.Set("Location", inventDim)
			row.Set("WarehouseID", warehouse1.Name)

			approvers := []string{}
			nextApprover, _ := ficologic.FindNextApproval(h, row.GetString("_id"))
			if len(nextApprover) > 0 {
				approvers = lo.Map(nextApprover, func(approver ficologic.ApprovalLogHandlerGetResponse, index int) string {
					return approver.Text
				})
			}

			row.Set("Approvers", approvers)
			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostPurchaseRequestGet() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res := ctx.Data().Data()["FnResult"]
		if _, ok := res.(codekit.M); ok {
			return true, nil // bypass if api is gets
		}

		m := new(scmmodel.PurchaseRequestJournal)
		if e := serde.Serde(res, &m); e != nil {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))

		m.Lines = lo.Map(m.Lines, func(d scmmodel.PurchaseJournalLine, i int) scmmodel.PurchaseJournalLine {
			item, _ := itemORM.Get(d.ItemID)
			d.Item = item
			return d
		})

		return true, nil
	}
}

func MWPostPurchaseOrderGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		lo.ForEach(ms, func(row codekit.M, index int) {
			vendorID := row.GetString("VendorID")
			inventDim := row.Get("Location").(scmmodel.InventDimension)
			warehouseID := row.GetString("WarehouseID")

			vendor, _ := datahub.GetByID(h, new(tenantcoremodel.Vendor), vendorID)
			warehouse, _ := datahub.GetByID(h, new(tenantcoremodel.LocationWarehouse), inventDim.WarehouseID)
			warehouse1, _ := datahub.GetByID(h, new(tenantcoremodel.LocationWarehouse), warehouseID)

			inventDim.WarehouseID = warehouse.Name
			row.Set("VendorID", vendor.Name)
			row.Set("Location", inventDim)
			row.Set("WarehouseID", warehouse1.Name)

			approvers := []string{}
			nextApprover, _ := ficologic.FindNextApproval(h, row.GetString("_id"))
			if len(nextApprover) > 0 {
				approvers = lo.Map(nextApprover, func(approver ficologic.ApprovalLogHandlerGetResponse, index int) string {
					return approver.Text
				})
			}

			row.Set("Approvers", approvers)
			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostPurchaseOrderGet() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res := ctx.Data().Data()["FnResult"]
		if _, ok := res.(codekit.M); ok {
			return true, nil // bypass if api is gets
		}

		m := new(scmmodel.PurchaseOrderJournal)
		if e := serde.Serde(res, &m); e != nil {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))

		m.Lines = lo.Map(m.Lines, func(d scmmodel.PurchaseJournalLine, i int) scmmodel.PurchaseJournalLine {
			item, _ := itemORM.Get(d.ItemID)
			d.Item = item
			return d
		})

		ctx.Data().Set("FnResult", m)
		return true, nil
	}
}
