package scmlogic

import (
	"fmt"
	"sort"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
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
		ms := []scmmodel.ItemRequest{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		employeeIDs := make([]interface{}, len(ms))
		inventWHIDs := make([]interface{}, len(ms))
		lo.ForEach(ms, func(row scmmodel.ItemRequest, index int) {
			ids[index] = row.ID
			employeeIDs[index] = row.Requestor
			inventWHIDs[index] = row.InventDimTo.WarehouseID
		})

		employees := []tenantcoremodel.Employee{}
		err := h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", employeeIDs...),
		), &employees)
		if err != nil {
			return false, fmt.Errorf("error when get employee: %s", err.Error())
		}

		mapEmployee := lo.Associate(employees, func(v tenantcoremodel.Employee) (string, string) {
			return v.ID, v.Name
		})

		inventWHs := []tenantcoremodel.LocationWarehouse{}
		err = h.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", inventWHIDs...),
		), &inventWHs)
		if err != nil {
			return false, fmt.Errorf("error when get inventory warehouse: %s", err.Error())
		}

		mapInvent := lo.Associate(inventWHs, func(v tenantcoremodel.LocationWarehouse) (string, string) {
			return v.ID, v.Name
		})

		mapApproval, err := findNextApproval(h, ids)
		if err != nil {
			return false, err
		}

		itemDetails := []scmmodel.ItemRequestDetail{}
		err = h.Gets(new(scmmodel.ItemRequestDetail), dbflex.NewQueryParam().SetWhere(
			dbflex.In("ItemRequestID", ids...),
		), &itemDetails)
		if err != nil {
			return false, fmt.Errorf("error when get item request detail: %s", err.Error())
		}

		mapDetail := map[string][]string{}
		for _, i := range itemDetails {
			fullfillments := make([]string, len(i.DetailLines))
			for i, dl := range i.DetailLines {
				fullfillments[i] = string(dl.FulfillmentType)
			}
			mapDetail[i.ItemRequestID] = append(mapDetail[i.ItemRequestID], fullfillments...)
		}

		lo.ForEach(ms, func(row scmmodel.ItemRequest, index int) {
			row.Requestor = mapEmployee[row.Requestor]
			row.InventDimTo.WarehouseID = mapInvent[row.InventDimTo.WarehouseID]
			row.Approvers = mapApproval[row.ID]

			// fill Fulfilled By
			fulfilledBys := lo.Uniq(mapDetail[row.ID])
			sort.Strings(fulfilledBys)
			row.FulfilledBy = strings.Join(fulfilledBys, ", ")

			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func findNextApproval(h *datahub.Hub, ids []interface{}) (map[string][]string, error) {
	mapApproval := map[string][]string{}
	profiles := []ficomodel.PostingApproval{}
	err := h.Gets(new(ficomodel.PostingApproval), dbflex.NewQueryParam().SetWhere(
		dbflex.In("SourceID", ids...),
	), &profiles)
	if err != nil {
		return nil, fmt.Errorf("error when get posting profile: %s", err.Error())
	}

	userIDs := []string{}
	for _, p := range profiles {
		for _, approver := range p.Approvers {
			userIDs = append(userIDs, approver.UserIDs...)
		}
	}

	users := []tenantcoremodel.Employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", userIDs...)), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get user: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(user tenantcoremodel.Employee) (string, string) {
		return user.ID, user.Name
	})

	for _, p := range profiles {
		names := []string{}
		for _, approver := range p.Approvers {
			for _, id := range approver.UserIDs {
				user := mapUser[id]
				isExist := false

				for _, approval := range p.Approvals {
					if id == approval.UserID {
						if approval.Status == "PENDING" {
							isExist = true
							break
						}
					}
				}

				if isExist == true {
					names = append(names, user)
				}
			}
		}
		mapApproval[p.SourceID] = names
	}

	return mapApproval, nil
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

		ids := make([]interface{}, len(ms))
		vendorIDs := make([]interface{}, len(ms))
		companyIDs := make([]interface{}, len(ms))
		inventWHIDs := make([]interface{}, len(ms))
		warehouseIDs := make([]interface{}, len(ms))
		lo.ForEach(ms, func(row codekit.M, index int) {
			ids[index] = row.GetString("_id")
			companyIDs[index] = row.GetString("CompanyID")
			vendorIDs[index] = row.GetString("VendorID")
			inventWHIDs[index] = row.Get("Location", scmmodel.InventDimension{}).(scmmodel.InventDimension).WarehouseID
			warehouseIDs[index] = row.GetString("WarehouseID")
		})

		companies := []tenantcoremodel.Company{}
		err := h.Gets(new(tenantcoremodel.Company), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", companyIDs...),
		), &companies)
		if err != nil {
			return false, fmt.Errorf("error when get company: %s", err.Error())
		}

		mapCompany := lo.Associate(companies, func(v tenantcoremodel.Company) (string, string) {
			return v.ID, v.Name
		})

		vendors := []tenantcoremodel.Vendor{}
		err = h.Gets(new(tenantcoremodel.Vendor), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", vendorIDs...),
		), &vendors)
		if err != nil {
			return false, fmt.Errorf("error when get vendor: %s", err.Error())
		}

		mapVendor := lo.Associate(vendors, func(v tenantcoremodel.Vendor) (string, string) {
			return v.ID, v.Name
		})

		inventWHs := []tenantcoremodel.LocationWarehouse{}
		err = h.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", inventWHIDs...),
		), &inventWHs)
		if err != nil {
			return false, fmt.Errorf("error when get inventory warehouse: %s", err.Error())
		}

		mapInvent := lo.Associate(inventWHs, func(v tenantcoremodel.LocationWarehouse) (string, string) {
			return v.ID, v.Name
		})

		warehouses := []tenantcoremodel.LocationWarehouse{}
		err = h.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", warehouseIDs...),
		), &warehouses)
		if err != nil {
			return false, fmt.Errorf("error when get warehouse: %s", err.Error())
		}

		mapWarehouse := lo.Associate(warehouses, func(v tenantcoremodel.LocationWarehouse) (string, string) {
			return v.ID, v.Name
		})

		mapApproval, err := findNextApproval(h, ids)
		if err != nil {
			return false, err
		}

		lo.ForEach(ms, func(row codekit.M, index int) {
			companyID := row.GetString("CompanyID")
			vendorID := row.GetString("VendorID")
			warehouseID := row.GetString("WarehouseID")
			inventDim := row.Get("Location").(scmmodel.InventDimension)

			inventDim.WarehouseID = mapInvent[inventDim.WarehouseID]
			row.Set("CompanyID", mapCompany[companyID])
			row.Set("VendorID", mapVendor[vendorID])
			row.Set("Location", inventDim)
			row.Set("WarehouseID", mapWarehouse[warehouseID])
			row.Set("Approvers", mapApproval[row.GetString("_id")])
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

		ids := make([]interface{}, len(ms))
		vendorIDs := make([]interface{}, len(ms))
		inventWHIDs := make([]interface{}, len(ms))
		warehouseIDs := make([]interface{}, len(ms))
		for i := range ms {
			vendorID := ms[i].GetString("VendorID")
			inventWHID := ms[i].Get("Location").(scmmodel.InventDimension).WarehouseID
			warehouseID := ms[i].GetString("WarehouseID")

			ids[i] = ms[i].GetString("_id")
			vendorIDs[i] = vendorID
			inventWHIDs[i] = inventWHID
			warehouseIDs[i] = warehouseID
		}

		vendors := []tenantcoremodel.Vendor{}
		err := h.Gets(new(tenantcoremodel.Vendor), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", vendorIDs...),
		), &vendors)
		if err != nil {
			return false, fmt.Errorf("error when get vendor: %s", err.Error())
		}

		mapVendor := lo.Associate(vendors, func(v tenantcoremodel.Vendor) (string, string) {
			return v.ID, v.Name
		})

		inventWHs := []tenantcoremodel.LocationWarehouse{}
		err = h.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", inventWHIDs...),
		), &inventWHs)
		if err != nil {
			return false, fmt.Errorf("error when get inventory warehouse: %s", err.Error())
		}

		mapInvent := lo.Associate(inventWHs, func(v tenantcoremodel.LocationWarehouse) (string, string) {
			return v.ID, v.Name
		})

		warehouses := []tenantcoremodel.LocationWarehouse{}
		err = h.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", warehouseIDs...),
		), &warehouses)
		if err != nil {
			return false, fmt.Errorf("error when get warehouse: %s", err.Error())
		}

		mapWarehouse := lo.Associate(warehouses, func(v tenantcoremodel.LocationWarehouse) (string, string) {
			return v.ID, v.Name
		})

		mapApproval, err := findNextApproval(h, ids)
		if err != nil {
			return false, err
		}

		lo.ForEach(ms, func(row codekit.M, index int) {
			vendorID := row.GetString("VendorID")
			inventDim := row.Get("Location").(scmmodel.InventDimension)
			warehouseID := row.GetString("WarehouseID")

			inventDim.WarehouseID = mapInvent[inventDim.WarehouseID]
			row.Set("VendorID", mapVendor[vendorID])
			row.Set("Location", inventDim)
			row.Set("WarehouseID", mapWarehouse[warehouseID])
			row.Set("Approvers", mapApproval[row.GetString("_id")])
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

			grs := []scmmodel.InventReceiveIssueJournal{}
			h.GetsByFilter(new(scmmodel.InventReceiveIssueJournal), dbflex.And(
				dbflex.Eq("Status", ficomodel.JournalStatusPosted),
				dbflex.Eq("ReffNo", m.ID),
				dbflex.Eq("Lines.References.Value", d.ID),
			), &grs)

			receivedQty := float64(0)

			for _, gr := range grs {
				// get all gr lines that has reference to this po line id
				filterLines := lo.Filter(gr.Lines, func(grLine scmmodel.InventReceiveIssueLine, i int) bool {
					_, _, exist := lo.FindIndexOf(grLine.References, func(grReff tenantcoremodel.ReferenceItem) bool {
						return grReff.Key == string(scmmodel.RefKeyPOLineID) && grReff.Value == d.ID
					})
					return exist
				})

				receivedQty += lo.SumBy(filterLines, func(grLine scmmodel.InventReceiveIssueLine) float64 {
					return grLine.InventJournalLine.Qty
				})
			}

			d.ReceivedQty = receivedQty

			return d
		})

		ctx.Data().Set("FnResult", m)
		return true, nil
	}
}
