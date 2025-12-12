package tenantcorelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostVendorName(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"VendorID"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return false, errors.New("invalid result")
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}

		serde.Serde(res["data"], &ms)

		vendorIDs := []interface{}{}
		for _, field := range fields {
			vendorIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})
			vendorIDs = append(vendorIDs, vendorIDFields...)
		}

		vendors := []tenantcoremodel.Vendor{}
		h.GetsByFilter(new(tenantcoremodel.Vendor), dbflex.In("_id", vendorIDs...), &vendors)
		mapVendor := lo.Associate(vendors, func(vendor tenantcoremodel.Vendor) (string, tenantcoremodel.Vendor) {
			return vendor.ID, vendor
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapVendor[m.GetString(field)]; ok {
					m.Set(field, fmt.Sprintf("%s %s", v.ID, v.Name))
				}
			}
		}
		res.Set("data", ms)

		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostCashbankGroup(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"CashBankGroupID", "CurrencyID"}
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

		//collect ids cashbank groups from response
		IDOfCashbankGroups := []interface{}{}
		for _, field := range fields {
			cashbankIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})
			IDOfCashbankGroups = append(IDOfCashbankGroups, cashbankIDFields...)
		}

		//transform KV data model
		mapCashbankGroups := map[string]tenantcoremodel.CashBankGroup{}

		cashbankGroup := func() {
			//get list of cashbank group by list of ids
			cashbankGroups := []tenantcoremodel.CashBankGroup{}
			e := h.GetsByFilter(new(tenantcoremodel.CashBankGroup), dbflex.In("_id", IDOfCashbankGroups...), &cashbankGroups)
			if e != nil {
				ctx.Log().Errorf("Failed populate data cashbank group: %s", e.Error())
			}

			//convert list cashbank group to map[string]CashBankGroup
			mapCashbankGroups = lo.Associate(cashbankGroups, func(cashbank tenantcoremodel.CashBankGroup) (string, tenantcoremodel.CashBankGroup) {
				return cashbank.ID, cashbank
			})
		}

		cashbankGroup()

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapCashbankGroups[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostCurrency(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"CurrencyID"}
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

		//collect ids currency from response
		IDOfCurrencies := []interface{}{}
		for _, field := range fields {
			currencyIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})
			IDOfCurrencies = append(IDOfCurrencies, currencyIDFields...)
		}

		mapCurrencies := map[string]tenantcoremodel.Currency{}
		currency := func() {
			currencies := []tenantcoremodel.Currency{}
			e := h.GetsByFilter(new(tenantcoremodel.Currency), dbflex.In("_id", IDOfCurrencies...), &currencies)
			if e != nil {
				ctx.Log().Errorf("Failed populate data vendor: %s", e.Error())
			}

			//convert list vendor to map[string]Vendor
			mapCurrencies = lo.Associate(currencies, func(currency tenantcoremodel.Currency) (string, tenantcoremodel.Currency) {
				return currency.ID, currency
			})
		}

		currency()

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapCurrencies[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostCustomerGroup(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"GroupID"}
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
		customerGroupIDs := []interface{}{}
		for _, field := range fields {
			customerGroupIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			customerGroupIDs = append(customerGroupIDs, customerGroupIDFields...)
		}

		//get list of customer group by list of ids
		customerGroups := []tenantcoremodel.CustomerGroup{}
		e := h.GetsByFilter(new(tenantcoremodel.CustomerGroup), dbflex.In("_id", customerGroupIDs...), &customerGroups)
		if e != nil {
			ctx.Log().Errorf("Failed populate data customer groups: %s", e.Error())
		}

		//convert list customer group to map[string]CustomerGroup
		mapCustomerGroups := lo.Associate(customerGroups, func(group tenantcoremodel.CustomerGroup) (string, tenantcoremodel.CustomerGroup) {
			return group.ID, group
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapCustomerGroups[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostEmployeeGroup(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"EmployeeGroupID"}
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
		employeeGroupIDs := []interface{}{}
		for _, field := range fields {
			employeeGroupIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			employeeGroupIDs = append(employeeGroupIDs, employeeGroupIDFields...)
		}

		//get list of employee group by list of ids
		employeeGroups := []tenantcoremodel.EmployeeGroup{}
		e := h.GetsByFilter(new(tenantcoremodel.EmployeeGroup), dbflex.In("_id", employeeGroupIDs...), &employeeGroups)
		if e != nil {
			ctx.Log().Errorf("Failed populate data employee groups: %s", e.Error())
		}

		//convert list employee group to map[string]EmployeeGroup
		mapEmployeeGroups := lo.Associate(employeeGroups, func(group tenantcoremodel.EmployeeGroup) (string, tenantcoremodel.EmployeeGroup) {
			return group.ID, group
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapEmployeeGroups[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostVendorGroup(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"GroupID"}
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
		vendorGroupIDs := []interface{}{}
		for _, field := range fields {
			vendorGroupIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			vendorGroupIDs = append(vendorGroupIDs, vendorGroupIDFields...)
		}

		//get list of vendor group by list of ids
		vendorGroups := []tenantcoremodel.VendorGroup{}
		e := h.GetsByFilter(new(tenantcoremodel.VendorGroup), dbflex.In("_id", vendorGroupIDs...), &vendorGroups)
		if e != nil {
			ctx.Log().Errorf("Failed populate data vendor groups: %s", e.Error())
		}

		//convert list vendor group to map[string]VendorGroup
		mapVendorGroups := lo.Associate(vendorGroups, func(group tenantcoremodel.VendorGroup) (string, tenantcoremodel.VendorGroup) {
			return group.ID, group
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapVendorGroups[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostLedgerAccount(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"MainBalanceAccount"}
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
		ledgerAccountIDs := []interface{}{}
		for _, field := range fields {
			ledgerAccountIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			ledgerAccountIDs = append(ledgerAccountIDs, ledgerAccountIDFields...)
		}

		//get list of ledger account by list of ids
		ledgerAccounts := []tenantcoremodel.LedgerAccount{}
		e := h.GetsByFilter(new(tenantcoremodel.LedgerAccount), dbflex.In("_id", ledgerAccountIDs...), &ledgerAccounts)
		if e != nil {
			ctx.Log().Errorf("Failed populate data ledger accounts: %s", e.Error())
		}

		//convert list ledger account to map[string]LedgerAccount
		mapLedgerAccounts := lo.Associate(ledgerAccounts, func(group tenantcoremodel.LedgerAccount) (string, tenantcoremodel.LedgerAccount) {
			return group.ID, group
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapLedgerAccounts[m.GetString(field)]; ok {
					m.Set(field, fmt.Sprintf("%s | %s", v.ID, v.Name))
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostAssetGroup(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"GroupID"}
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
		assetGroupIDs := []interface{}{}
		for _, field := range fields {
			assetGroupIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			assetGroupIDs = append(assetGroupIDs, assetGroupIDFields...)
		}

		//get list of asset groups by list of ids
		assetGroups := []tenantcoremodel.AssetGroup{}
		e := h.GetsByFilter(new(tenantcoremodel.AssetGroup), dbflex.In("_id", assetGroupIDs...), &assetGroups)
		if e != nil {
			ctx.Log().Errorf("Failed populate data asset groups: %s", e.Error())
		}

		//convert list asset group to map[string]AssetGroup
		mapAssetGroups := lo.Associate(assetGroups, func(group tenantcoremodel.AssetGroup) (string, tenantcoremodel.AssetGroup) {
			return group.ID, group
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapAssetGroups[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostItemModel() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		//get data from response
		res, ok := ctx.Data().Get("FnResult", &[]tenantcoremodel.Item{}).(*[]tenantcoremodel.Item)
		if !ok {
			return true, nil
		}

		// h := sebar.GetTenantDBFromContext(ctx)

		for k, item := range *res {
			// fmt.Println(item.Name, item.OtherName)

			// itemRow, _ := datahub.GetByID(h, new(tenantcoremodel.Item), item.ID)
			// if itemRow.OtherName != "" {
			// 	item.Name = fmt.Sprintf("%s ( %s )", itemRow.Name, itemRow.OtherName)
			// }

			(*res)[k] = item
		}

		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPreItem() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)

		m := new(tenantcoremodel.Item)
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		if m.ID != "" || m.ItemGroupID == "" {
			return true, nil // bypass when item already has ID or it doesn't belong to any Item Group
		}

		itemGroup := new(tenantcoremodel.ItemGroup)
		h.GetByID(itemGroup, m.ItemGroupID)
		if itemGroup.ID == "" {
			return true, nil
		}

		itemID := itemGroup.FormatItemID()
		h.Save(itemGroup)

		m.ID = itemID
		serde.Serde(m, payload)

		return true, nil
	}
}

func MWPreEmployeeFilter(fieldEmp, fieldDest string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if fieldEmp == "" || fieldDest == "" {
			return true, nil
		}

		fs := ctx.Data().Get("DBModFilter", []*dbflex.Filter{}).([]*dbflex.Filter)

		hub := sebar.GetTenantDBFromContext(ctx)
		if hub == nil {
			return false, errors.New("missing: connection")
		}

		userID := sebar.GetUserIDFromCtx(ctx)

		// var idParam []interface{}
		// idParam = append(idParam, "EMP-BDM-00002")
		param := struct {
			ID string
		}{
			ID: userID,
			// ID: "EMP-BDM-00002",
		}

		// get data employee detail
		ev, _ := ctx.DefaultEvent()
		if ev == nil {
			return false, errors.New("nil: EventHub")
		}

		res := tenantcoremodel.EmployeeTenantBagong{}
		e := ev.Publish("/v1/bagong/employee/get-employee-by-id", &param, &res, nil)
		fmt.Printf("employee id: %s | employee err: %s\n", userID, e)
		// fmt.Println("res", codekit.JsonString(res))

		if res.ID == "" {
			return true, nil
		}

		m := codekit.M{}
		if e := serde.Serde(res, &m); e != nil {
			return true, nil
		}

		filterVal, _ := m[fieldEmp].(string)
		if filterVal == "" {
			return true, nil
		}

		fs = append(fs, dbflex.Eq(fieldDest, filterVal))
		ctx.Data().Set("DBModFilter", fs)

		return true, nil
	}
}
