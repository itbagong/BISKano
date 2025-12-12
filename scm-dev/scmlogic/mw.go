package scmlogic

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostMovementInItemDetail() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}

		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		itemDetailIDs := lo.Map(ms, func(m codekit.M, index int) interface{} {
			return m.GetString("ItemID")
		})

		items := []tenantcoremodel.Item{}
		e := h.GetsByFilter(new(tenantcoremodel.Item), dbflex.In("_id", itemDetailIDs...), &items)
		if e != nil {
			ctx.Log().Errorf("Failed populate data asset groups: %s", e.Error())
		}

		mapItems := lo.Associate(items, func(item tenantcoremodel.Item) (string, tenantcoremodel.Item) {
			return item.ID, item
		})

		for _, m := range ms {
			if _, ok := mapItems[m.GetString("ItemID")]; ok {
				m.Set("Item", mapItems[m.GetString("ItemID")])
				//get movement in and inject item balance
				movementInID := m.GetString("MovementInID")
				itemID := m.GetString("ItemID")
				SKU := m.GetString("SKU")

				movementIn := new(scmmodel.MovementIn)
				e = h.GetByID(movementIn, movementInID)
				if e != nil {
					ctx.Log().Errorf("Movement in not found with id: %s %s", movementInID, e.Error())
				} else {
					//get item balance
					itemBalance := new(scmmodel.ItemBalance)
					h.GetByFilter(itemBalance, new(scmmodel.ItemBalance).UniqueFilter(scmmodel.ItemBalanceUniqueFilterParam{
						ItemID:             itemID,
						SKU:                SKU,
						InventoryDimension: movementIn.InventoryDimension,
					}))

					// m.Set("Qty", itemBalance.QtyPlan)
					m.Set("ReceivedQty", itemBalance.QtyPlanned)
					m.Set("ItemBalance", itemBalance)
				}

			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostMovementOutItemDetail() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}

		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		itemDetailIDs := lo.Map(ms, func(m codekit.M, index int) interface{} {
			return m.GetString("ItemID")
		})

		items := []tenantcoremodel.Item{}
		e := h.GetsByFilter(new(tenantcoremodel.Item), dbflex.In("_id", itemDetailIDs...), &items)
		if e != nil {
			ctx.Log().Errorf("Failed populate data asset groups: %s", e.Error())
		}

		mapItems := lo.Associate(items, func(item tenantcoremodel.Item) (string, tenantcoremodel.Item) {
			return item.ID, item
		})

		for _, m := range ms {
			if _, ok := mapItems[m.GetString("ItemID")]; ok {
				m.Set("Item", mapItems[m.GetString("ItemID")])
				//get movement in and inject item balance
				movementOutID := m.GetString("MovementOutID")
				itemID := m.GetString("ItemID")
				SKU := m.GetString("SKU")

				movementOut := new(scmmodel.MovementOut)
				e = h.GetByID(movementOut, movementOutID)
				if e != nil {
					ctx.Log().Errorf("Movement out not found with id: %s %s", movementOutID, e.Error())
				} else {
					//get item balance
					itemBalance := new(scmmodel.ItemBalance)
					h.GetByFilter(itemBalance, new(scmmodel.ItemBalance).UniqueFilter(scmmodel.ItemBalanceUniqueFilterParam{
						ItemID:             itemID,
						SKU:                SKU,
						InventoryDimension: movementOut.InventoryDimension,
					}))

					// m.Set("Qty", itemBalance.QtyPlan)
					m.Set("ReceivedQty", itemBalance.QtyPlanned)
					m.Set("ItemBalance", itemBalance)
				}

			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostItem(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"ItemID"}
		}

		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}

		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		//collect ids from response
		itemIDs := []interface{}{}
		for _, field := range fields {
			itemIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			itemIDs = append(itemIDs, itemIDFields...)
		}

		//get list of item by list of ids
		items := []tenantcoremodel.Item{}
		e := h.GetsByFilter(new(tenantcoremodel.Item), dbflex.In("_id", itemIDs...), &items)
		if e != nil {
			ctx.Log().Errorf("Failed populate data items: %s", e.Error())
		}

		//convert list asset group to map[string]AssetGroup
		mapItems := lo.Associate(items, func(item tenantcoremodel.Item) (string, tenantcoremodel.Item) {
			return item.ID, item
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapItems[m.GetString(field)]; ok {
					m.Set("Item", v)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostInventoryJournal() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		//convert response to map[string]interface{}
		ms := []scmmodel.InventJournal{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		warehouseIDs := make([]interface{}, len(ms)*2)
		i := 0
		lo.ForEach(ms, func(row scmmodel.InventJournal, index int) {
			ids[index] = row.ID
			warehouseIDs[i] = row.InventDim.WarehouseID
			i++
			warehouseIDs[i] = row.InventDimTo.WarehouseID
			i++
		})

		warehouses := []tenantcoremodel.LocationWarehouse{}
		err := h.Gets(new(tenantcoremodel.LocationWarehouse), dbflex.NewQueryParam().SetWhere(
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

		lo.ForEach(ms, func(row scmmodel.InventJournal, index int) {
			inventDim := row.InventDim
			inventDimTo := row.InventDimTo
			inventDim.WarehouseID = mapWarehouse[row.InventDim.WarehouseID]
			inventDimTo.WarehouseID = mapWarehouse[row.InventDimTo.WarehouseID]

			row.InventDim.WarehouseID = mapWarehouse[row.InventDim.WarehouseID]
			row.InventDimTo.WarehouseID = mapWarehouse[row.InventDimTo.WarehouseID]
			row.Approvers = mapApproval[row.ID]

			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostInventoryReceive() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		vendorIDs := make([]interface{}, 0)
		warehouseIDs := make([]interface{}, 0)
		sectionIDs := make([]interface{}, len(ms))
		itemIDs := make([]string, 0)
		skuIDs := make([]string, 0)
		lo.ForEach(ms, func(row codekit.M, index int) {
			ids[index] = row.GetString("_id")
			warehouseIDs = append(warehouseIDs, row.GetString("WarehouseID"))
			sectionIDs[index] = row.GetString("SectionID")

			lines, ok := row.Get("Lines").([]scmmodel.InventReceiveIssueLine)
			if ok {
				for _, l := range lines {
					vendorIDs = append(vendorIDs, l.VendorID)
					warehouseIDs = append(warehouseIDs, l.InventDim.WarehouseID)
					itemIDs = append(itemIDs, l.ItemID)
					skuIDs = append(skuIDs, l.SKU)
				}
			}
		})

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

		sections := []tenantcoremodel.LocationSection{}
		err = h.Gets(new(tenantcoremodel.LocationSection), dbflex.NewQueryParam().SetWhere(
			dbflex.In("_id", sectionIDs...),
		), &sections)
		if err != nil {
			return false, fmt.Errorf("error when get section: %s", err.Error())
		}

		mapSection := lo.Associate(sections, func(v tenantcoremodel.LocationSection) (string, string) {
			return v.ID, v.Name
		})

		mapApproval, err := findNextApproval(h, ids)
		if err != nil {
			return false, err
		}

		mapAssignItem, err := AssignItem(h, itemIDs, skuIDs)
		if err != nil {
			return false, err
		}

		lo.ForEach(ms, func(row codekit.M, index int) {
			row.Set("WarehouseID", mapWarehouse[row.GetString("WarehouseID")])
			row.Set("SectionID", mapSection[row.GetString("SectionID")])
			lines, ok := row.Get("Lines").([]scmmodel.InventReceiveIssueLine)
			if ok {
				lines = lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
					itemName := mapAssignItem[line.ItemID+line.SKU]
					line.Item.Name = itemName
					line.ItemName = itemName
					line.VendorName = mapVendor[line.VendorID]
					line.OriginDestinationName = mapWarehouse[line.InventDim.WarehouseID]

					return line
				})
			}
			row.Set("Lines", lines)
			row.Set("Approvers", mapApproval[row.GetString("_id")])
			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPreInventoryReceiveSave() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		m := new(scmmodel.InventReceiveIssueJournal)
		if e := serde.Serde(payload, &m); e != nil {
			fmt.Println("MWPostWorkOrderPlanSave() ERROR SERDE:", e)
			return true, nil
		}

		if len(m.Lines) >= 2 {
			// check semua WH dan Vendor harus sama m.Lines
			whID := m.Lines[0].InventDim.WarehouseID
			vendorID := m.Lines[0].VendorID

			for _, line := range m.Lines {
				if line.InventDim.WarehouseID != whID || line.VendorID != vendorID {
					return false, fmt.Errorf("Warehouse and Vendor must be same")
				}
			}
		}

		return true, nil
	}
}

func MWPreItemBalance() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		fs := ctx.Data().Get("DBModFilter", []*dbflex.Filter{}).([]*dbflex.Filter)
		if len(fs) == 0 {
			h := sebar.GetTenantDBFromContext(ctx)
			//set default filter warehouse
			//get dimension from jwt
			jwtData := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
			if jwtData == nil {
				return true, nil // bypass
			}

			dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})
			if len(dimIface) == 0 {
				return true, nil // bypass
			}

			dim := tenantcoremodel.Dimension{}
			if err := serde.Serde(dimIface, &dim); err != nil {
				return true, nil // bypass
			}

			dimFilter := dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", dim.Get("Site")))
			warehouses := []tenantcoremodel.LocationWarehouse{}
			err := h.GetsByFilter(new(tenantcoremodel.LocationWarehouse), dimFilter, &warehouses)
			if err != nil {
				return true, nil
			}

			whIDs := lo.Map(warehouses, func(wh tenantcoremodel.LocationWarehouse, _ int) interface{} {
				return wh.ID
			})

			filters := []*dbflex.Filter{
				dbflex.In("InventDim.WarehouseID", whIDs...),
			}

			ctx.Data().Set("DBModFilter", filters)
		}

		return true, nil
	}
}

func MWPostItemBalance() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		mapWarehouses := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))
		mapSections := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationSection))
		mapAisel := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationAisle))
		mapBox := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationBox))
		mapItems := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
		mapItemSpecs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
		// mapItemSpecVariants := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.SpecVariant))
		mapUoM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.UoM))

		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		// generateItemName := func(item *tenantcoremodel.Item, spec *tenantcoremodel.ItemSpec, variant *tenantcoremodel.SpecVariant) string {
		// 	texts := []string{}
		// 	separator := "-"
		// 	if item != nil {
		// 		if item.Name != "" {
		// 			texts = append(texts, item.Name)
		// 			if item.OtherName != "" {
		// 				texts = append(texts, item.OtherName)
		// 			}
		// 		}

		// 		if spec != nil {
		// 			if spec.SKU != "" {
		// 				texts = append(texts, spec.SKU)
		// 			}
		// 		}

		// 		if variant != nil {
		// 			if variant.Name != "" {
		// 				texts = append(texts, variant.Name)
		// 			}
		// 		}
		// 	}

		// 	if len(texts) > 0 {
		// 		return strings.Join(texts, separator)
		// 	}

		// 	return ""
		// }

		lo.ForEach(ms, func(row codekit.M, index int) {
			itemID := row.GetString("ItemID")
			item, _ := mapItems.Get(itemID)
			inventDim := row.Get("InventDim", scmmodel.InventDimension{}).(scmmodel.InventDimension)
			specID := row.GetString("SKU")

			row.Set("WarehouseID", "")
			row.Set("SectionID", "")
			row.Set("AisleID", "")
			row.Set("BoxID", "")

			if inventDim.WarehouseID != "" {
				warehouse, _ := mapWarehouses.Get(inventDim.WarehouseID)
				row.Set("WarehouseName", warehouse.Name)
				inventDim.WarehouseID = warehouse.Name
			} else {
				row.Set("WarehouseName", "")
			}

			if inventDim.SectionID != "" {
				warehouse, _ := mapSections.Get(inventDim.SectionID)
				inventDim.SectionID = warehouse.Name
				row.Set("SectionName", warehouse.Name)
			} else {
				row.Set("SectionName", "")
			}

			if inventDim.AisleID != "" {
				warehouse, _ := mapAisel.Get(inventDim.AisleID)
				inventDim.AisleID = warehouse.Name
				row.Set("AisleName", warehouse.Name)
			} else {
				row.Set("AisleName", "")
			}

			if inventDim.BoxID != "" {
				warehouse, _ := mapBox.Get(inventDim.BoxID)
				inventDim.BoxID = warehouse.Name
				row.Set("BoxName", warehouse.Name)
			} else {
				row.Set("BoxName", "")
			}

			spec, _ := mapItemSpecs.Get(specID)
			// variant := new(tenantcoremodel.SpecVariant)
			// if spec != nil {
			// 	variant, _ = mapItemSpecVariants.Get(spec.SpecVariantID)
			// }

			// itemName := generateItemName(item, spec, variant)
			itemName := tenantcorelogic.ItemVariantName(h, item.ID, spec.ID)
			row.Set("ItemName", itemName)
			row.Set("InventDim", inventDim)
			row.Set("ItemID", item.Name)
			row.Set("SKU", spec.SKU)

			unit, _ := mapUoM.Get(item.DefaultUnitID)
			row.Set("DefaultUnitID", item.DefaultUnitID)
			row.Set("UnitName", unit.Name)

			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostInventTrx() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		h := sebar.GetTenantDBFromContext(ctx)
		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		ms := []codekit.M{}
		//convert response to map[string]interface{}
		serde.Serde(res["data"], &ms)

		lo.ForEach(ms, func(row codekit.M, index int) {
			itemName := ""
			if item, ok := row.Get("Item").(tenantcoremodel.Item); ok {
				sku := row.GetString("SKU")
				itemName = tenantcorelogic.ItemVariantName(h, item.ID, sku)
			}

			row.Set("ItemName", itemName)
			ms[index] = row
		})

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

type NumSeqClaimRespond struct {
	Number string
}

type NumSeqClaimPayload struct {
	NumberSequenceID string
	Date             time.Time
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
		if coID == "" {
			coID = tenantcorelogic.GetCompanyIDFromContext(ctx)
		}

		// coID = tenantcorelogic.GetCompanyIDFromContext(ctx)
		userID := sebar.GetUserIDFromCtx(ctx)
		m.Set(field, coID)

		//validation ID
		id := m.GetString("_id")
		if id == "" {
			h := sebar.GetTenantDBFromContext(ctx)
			if h != nil {
				ev, _ := ctx.DefaultEvent()
				if ev != nil {
					setup := tenantcorelogic.GetSequenceSetup(h, "ItemRequest", coID)
					if setup != nil {
						resp := new(NumSeqClaimRespond)
						e := ev.Publish("/v1/tenant/numseq/claim", &NumSeqClaimPayload{NumberSequenceID: setup.NumSeqID, Date: time.Now()}, resp, &kaos.PublishOpts{Headers: codekit.M{
							"CompanyID": coID, sebar.CtxJWTReferenceID: userID}})
						if e != nil {
							log.Println(e.Error())
						} else {
							if resp.Number != "" {
								m.Set("_id", resp.Number)
							}
						}
					}
				}
			}

		}
		serde.Serde(m, payload)
		return true, nil
	}
}

// MWPreInjectFindByKeywordApproval untuk mencari journal posting berdasarkan username yang melakukan approve
func MWPreInjectFindByKeywordApproval() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		typeOfPayload := reflect.TypeOf(payload)
		if payload == nil {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		if typeOfPayload.String() != "*dbflex.QueryParam" {
			return true, nil
		}

		param := dbflex.NewQueryParam()
		e := serde.Serde(payload, param)
		if e != nil {
			return true, nil
		}

		if param.Where != nil {
			where := param.Where

			//berarti cuma search keyword
			if where.Op == dbflex.OpOr {

			} else {
				items := where.Items
				newFilter := []*dbflex.Filter{}
				lo.ForEach(items, func(filterItem *dbflex.Filter, _ int) {
					if filterItem.Op != dbflex.OpOr {
						newFilter = append(newFilter, filterItem)
					} else {
						keyword := ""
						if len(filterItem.Items) > 0 {
							//dapatin text yang di search
							if val, ok := filterItem.Items[0].Value.(string); ok {
								keyword = val
							}
						}

						if keyword != "" {
							employees := []tenantcoremodel.Employee{}
							//filter Or ketemu, disini mulai inject posting approvals
							h.GetsByFilter(new(tenantcoremodel.Employee), dbflex.Contains("Name", keyword), employees)
							if len(employees) > 0 {

							}
						}
					}
				})
			}
		}
		return true, nil
	}
}

// MWPrePRCalcTax middleware untuk menghitung ulang tax pada line
func MWPrePRCalcTax() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//cek type dari payload apakah tipe dari *scmmodel.PurchaseRequestJournal
		typeOfPayload := reflect.TypeOf(payload)
		if payload == nil {
			return true, nil
		}

		if typeOfPayload.String() != "*scmmodel.PurchaseRequestJournal" {
			return true, nil
		}

		prJournal := new(scmmodel.PurchaseRequestJournal)
		if e := serde.Serde(payload, prJournal); e != nil {
			return true, nil
		}

		prJournal.TotalAmount = 0         //sub total
		prJournal.TotalDiscountAmount = 0 //discount line
		prJournal.PPN = 0                 //ppn
		prJournal.PPH = 0

		lo.ForEach(prJournal.Lines, func(line scmmodel.PurchaseJournalLine, index int) {
			prJournal.TotalAmount += line.SubTotal
			prJournal.TotalDiscountAmount += line.DiscountAmount
			prJournal.PPH += line.PPH
			prJournal.PPN += line.PPN
		})

		serde.Serde(prJournal, payload)
		return true, nil
	}
}

// MWPreAssignSite middleware untuk menggunakan CompanyID
func MWPreAssignSite(inventDimField ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		idimField := "InventDim"
		if len(inventDimField) > 0 && inventDimField[0] != "" {
			idimField = inventDimField[0]
		}

		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		warehouseORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))

		if inventDim, ok := m.Get(idimField).(scmmodel.InventDimension); ok {
			wh, _ := warehouseORM.Get(inventDim.WarehouseID)
			site := wh.Dimension.Get("Site")

			if dim, dimOK := m.Get("Dimension").(tenantcoremodel.Dimension); dimOK {
				dim = dim.Set("Site", site)
				m.Set("Dimension", dim)
			} else {
				dim := tenantcoremodel.Dimension{}
				dim = dim.Set("Site", site)
				m.Set("Dimension", dim)
			}
		}

		serde.Serde(m, payload)
		return true, nil
	}
}

func MWPreSiteHONoFilter() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		HOSiteID := "SITE020"

		jwtData := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
		if jwtData == nil {
			return true, nil // bypass
		}

		dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})
		if len(dimIface) == 0 {
			return true, nil // bypass
		}

		dim := tenantcoremodel.Dimension{}
		if err := serde.Serde(dimIface, &dim); err != nil {
			return true, nil // bypass
		}

		userSite := dim.Get("Site")
		if userSite != HOSiteID {
			return true, nil // bypass
		}

		// set filter to see all data
		fs := ctx.Data().Get("DBModFilter", []*dbflex.Filter{}).([]*dbflex.Filter)

		for f1Idx, f1 := range fs {
			newItems := []*dbflex.Filter{}

			for _, f2 := range f1.Items {
				// cek f2.Items nya apakah dimension atau bukan, kalo iya, ga usah dimasukkan
				_, _, isDim := lo.FindIndexOf(f2.Items, func(d *dbflex.Filter) bool {
					// return d.Field == "Key" && (d.Value == "PC" || d.Value == "Site" || d.Value == "CC")
					return d.Field == "Dimension"
				})
				if isDim == false {
					newItems = append(newItems, f2)
				}
			}

			f1.Items = newItems
			fs[f1Idx] = f1
		}

		// filter hanya fs yang Items nya ada isinya biar ga error
		fs = lo.FilterMap(fs, func(d *dbflex.Filter, i int) (*dbflex.Filter, bool) {
			return d, len(d.Items) > 0
		})
		ctx.Data().Set("DBModFilter", fs)

		if len(fs) == 0 {
			ctx.Data().Remove("DBModFilter") // remove kalo kosong biar ga error
		}

		return true, nil
	}
}

func MWPreAssignSequenceNoForNATSApi(sequenceKind string, field string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if field == "" {
			field = "_id"
		}

		test := ctx.Data().Get("jwt_data", codekit.M{})
		_ = test
		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, sequenceKind)
		m.Set(field, id)
		serde.Serde(m, payload)

		return true, nil
	}
}

func MWPostGetGoodReceiveIssuance() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)

		res := ctx.Data().Data()["FnResult"]
		receiveIssuance := scmmodel.InventReceiveIssueJournal{}
		if e := serde.Serde(res, &receiveIssuance); e != nil {
			fmt.Println("MWPostGetGoodReceive() ERROR SERDE:", e)
			return true, nil
		}

		receiveIssuance.Lines = lo.Map(receiveIssuance.Lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
			line.Item.Name = tenantcorelogic.ItemVariantName(h, line.ItemID, line.SKU)

			return line
		})

		ctx.Data().Set("FnResult", receiveIssuance)
		return true, nil
	}
}

func MWPrePurchaseOrderSave() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)
		vendorORM := sebar.NewMapRecordWithORM(h, new(bagongmodel.BGVendor))

		m := new(scmmodel.PurchaseOrderJournal)
		if e := serde.Serde(payload, &m); e != nil {
			fmt.Println("MWPostWorkOrderPlanSave() ERROR SERDE:", e)
			return true, nil
		}

		vnd, _ := vendorORM.Get(m.VendorID)
		if vnd.Terms.Taxes1 != "" && !lo.Contains(m.TaxCodes, vnd.Terms.Taxes1) {
			return false, fmt.Errorf("Tax %s not found, please re-select Vendor to update Tax Codes", vnd.Terms.Taxes1)
		}

		if vnd.Terms.Taxes2 != "" && !lo.Contains(m.TaxCodes, vnd.Terms.Taxes2) {
			return false, fmt.Errorf("Tax %s not found, please re-select Vendor to update Tax Codes", vnd.Terms.Taxes2)
		}

		return true, nil
	}
}

func MWPreItemMinMaxSave() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)
		itemSpecORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))

		m := new(scmmodel.ItemMinMax)
		if e := serde.Serde(payload, &m); e != nil {
			fmt.Println("MWPreItemMinMaxSave() ERROR SERDE:", e)
			return true, nil
		}

		if m.SKU != "" {
			itemSpec, _ := itemSpecORM.Get(m.SKU)
			m.InventoryDimension.Size = itemSpec.SpecSizeID
			m.InventoryDimension.Grade = itemSpec.SpecGradeID
			m.InventoryDimension.VariantID = itemSpec.SpecVariantID
		}
		m.InventoryDimension = *m.InventoryDimension.Calc()

		if m.ID == "" {
			// validation if new
			exm := new(scmmodel.ItemMinMax)
			h.GetByFilter(exm, dbflex.And(
				dbflex.Eq("ItemID", m.ItemID),
				dbflex.Eq("SKU", m.SKU),
				dbflex.Eq("InventoryDimension.InventDimID", m.InventoryDimension.InventDimID),
			))
			if exm.ID != "" {
				return false, fmt.Errorf("Item Min Max setting already exist")
			}
		} else {
			// validation if update
			exms := []scmmodel.ItemMinMax{}
			h.GetsByFilter(new(scmmodel.ItemMinMax), dbflex.And(
				dbflex.Eq("ItemID", m.ItemID),
				dbflex.Eq("SKU", m.SKU),
				dbflex.Eq("InventoryDimension.InventDimID", m.InventoryDimension.InventDimID),
				dbflex.Ne("_id", m.ID), // yg bukan id yg di update
			), &exms)
			if len(exms) > 0 {
				return false, fmt.Errorf("Item Min Max setting already exist")
			}
		}

		serde.Serde(m, payload)
		return true, nil
	}
}

func MWPostItemMinMaxGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)
		whORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))
		sectionORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationSection))
		aisleORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationAisle))
		boxORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationBox))

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		for _, m := range ms {
			itemID := m.GetString("ItemID")
			sku := m.GetString("SKU")
			itemName := tenantcorelogic.ItemVariantName(h, itemID, sku)

			m.Set("ItemName", itemName)

			if invdim, ok := m.Get("InventoryDimension").(scmmodel.InventDimension); ok {
				wh, _ := whORM.Get(invdim.WarehouseID)
				m.Set("WarehouseName", wh.Name)
				section, _ := sectionORM.Get(invdim.SectionID)
				m.Set("SectionName", section.Name)
				aisle, _ := aisleORM.Get(invdim.AisleID)
				m.Set("AisleName", aisle.Name)
				box, _ := boxORM.Get(invdim.BoxID)
				m.Set("BoxName", box.Name)
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostInventoryAdjustmentGets() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)
		whORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationWarehouse))
		sectionORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationSection))
		aisleORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationAisle))
		boxORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LocationBox))

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		for _, m := range ms {
			if invdim, ok := m.Get("InventDim").(scmmodel.InventDimension); ok {
				wh, _ := whORM.Get(invdim.WarehouseID)
				m.Set("WarehouseName", wh.Name)
				section, _ := sectionORM.Get(invdim.SectionID)
				m.Set("SectionName", section.Name)
				aisle, _ := aisleORM.Get(invdim.AisleID)
				m.Set("AisleName", aisle.Name)
				box, _ := boxORM.Get(invdim.BoxID)
				m.Set("BoxName", box.Name)
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
