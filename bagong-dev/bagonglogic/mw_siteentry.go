package bagonglogic

import (
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostSiteEntryTrayekRitase() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {

		routePath := ctx.Data().Data()["RoutePath"].(string)
		arrPath := strings.Split(string(routePath), "/")
		lastPath := arrPath[len(arrPath)-1]

		if lastPath == "get" {

			h := sebar.GetTenantDBFromContext(ctx)
			ms := codekit.M{}

			res, ok := ctx.Data().Data()["FnResult"].(*bagongmodel.SiteEntryTrayekRitase)
			if !ok {
				return true, nil
			}

			//convert response to map[string]interface{}
			serde.Serde(res, &ms)
			//collect ids from response
			vendorJournalIDs := []interface{}{}
			customerJournalIDs := []interface{}{}

			for _, d := range res.FixExpense {
				vendorJournalIDs = append(vendorJournalIDs, d.JournalID)
			}
			for _, d := range res.OtherExpense {
				vendorJournalIDs = append(vendorJournalIDs, d.JournalID)
			}
			for _, d := range res.OtherIncome {
				customerJournalIDs = append(customerJournalIDs, d.JournalID)
			}
			customerJournalIDs = append(customerJournalIDs, res.PassengerIncome.JournalID)
			customerJournalIDs = append(customerJournalIDs, res.DepositIncome.JournalID)

			//get list of vendor account by list of ids
			vendorJournal := []ficomodel.VendorJournal{}
			e := h.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
			if e != nil {
				ctx.Log().Errorf("Failed populate data VendorJournal: %s", e.Error())
			}
			//get list of customer account by list of ids
			customerJournal := []ficomodel.CustomerJournal{}
			e = h.GetsByFilter(new(ficomodel.CustomerJournal), dbflex.In("_id", customerJournalIDs...), &customerJournal)
			if e != nil {
				ctx.Log().Errorf("Failed populate data CustomerJournal: %s", e.Error())
			}
			cashSchedule := []ficomodel.CashSchedule{}
			e = h.GetsByFilter(new(ficomodel.CashSchedule), dbflex.Or(dbflex.In("SourceJournalID", customerJournalIDs...), dbflex.In("SourceJournalID", vendorJournalIDs...)), &cashSchedule)
			if e != nil {
				ctx.Log().Errorf("Failed populate data CustomerJournal: %s", e.Error())
			}
			//convert list vendor journal to map[string]VendorJournal
			mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
				return detail.ID, detail
			})
			//convert list customer journal to map[string]CustomerJournal
			mapCustomerJournal := lo.Associate(customerJournal, func(detail ficomodel.CustomerJournal) (string, ficomodel.CustomerJournal) {
				return detail.ID, detail
			})
			//convert list cashschedule to map[string]CashSchedule
			mapCashSchedule := lo.Associate(cashSchedule, func(detail ficomodel.CashSchedule) (string, ficomodel.CashSchedule) {
				return string(detail.SourceLineNo) + "~" + string(detail.SourceJournalID), detail
			})

			fixExpense := []bagongmodel.SiteExpense{}
			for _, n := range ms.Get("FixExpense").([]bagongmodel.SiteExpense) {
				tmp := n
				if v, ok := mapCashSchedule[string(n.LineNo)+"~"+n.JournalID]; ok {
					tmp.VoucherID = v.VoucherNo
				}
				if v, ok := mapVendorJournal[n.JournalID]; ok {
					tmp.ApprovalStatus = v.Status
				}
				fixExpense = append(fixExpense, tmp)
			}
			otherExpense := []bagongmodel.SiteExpense{}
			for _, n := range ms.Get("OtherExpense").([]bagongmodel.SiteExpense) {
				tmp := n
				if v, ok := mapCashSchedule[string(n.LineNo)+"~"+n.JournalID]; ok {
					tmp.VoucherID = v.VoucherNo
				}
				if v, ok := mapVendorJournal[n.JournalID]; ok {
					tmp.ApprovalStatus = v.Status
				}
				otherExpense = append(otherExpense, tmp)
			}
			otherIncome := []bagongmodel.SiteIncome{}
			for _, n := range ms.Get("OtherIncome").([]bagongmodel.SiteIncome) {
				tmp := n
				if v, ok := mapCashSchedule[string(n.LineNo)+"~"+n.JournalID]; ok {
					tmp.VoucherID = v.VoucherNo
				}
				if v2, ok := mapCustomerJournal[n.JournalID]; ok {
					tmp.ApprovalStatus = string(v2.Status)
				}
				otherIncome = append(otherIncome, tmp)
			}

			passengerIncome := res.PassengerIncome
			depositIncome := res.DepositIncome

			if v, ok := mapCashSchedule[string(passengerIncome.LineNo)+"~"+passengerIncome.JournalID]; ok {
				passengerIncome.VoucherID = v.VoucherNo
			}
			if v, ok := mapCustomerJournal[passengerIncome.JournalID]; ok {
				passengerIncome.ApprovalStatus = string(v.Status)
			}

			if v, ok := mapCashSchedule[string(depositIncome.LineNo)+"~"+depositIncome.JournalID]; ok {
				depositIncome.VoucherID = v.VoucherNo
			}
			if v, ok := mapCustomerJournal[depositIncome.JournalID]; ok {
				depositIncome.ApprovalStatus = string(v.Status)
			}

			res.FixExpense = fixExpense
			res.OtherExpense = otherExpense
			res.OtherIncome = otherIncome
			res.PassengerIncome = passengerIncome
			res.DepositIncome = depositIncome

			ctx.Data().Set("FnResult", res)
		}
		return true, nil
	}
}

func MWPostSiteEntryTourism() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		routePath := ctx.Data().Data()["RoutePath"].(string)
		arrPath := strings.Split(string(routePath), "/")
		lastPath := arrPath[len(arrPath)-1]

		if lastPath == "get" {
			h := sebar.GetTenantDBFromContext(ctx)

			//get data from response
			res, ok := ctx.Data().Data()["FnResult"].(*bagongmodel.SiteEntryTourismDetail)
			if !ok {
				return true, nil
			}

			//collect ids from response
			vendorJournalIDs := []interface{}{}
			for _, c := range res.OperationalExpense {
				vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
			}
			for _, c := range res.OtherExpense {
				vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
			}

			//get list of vendor account by list of ids
			vendorJournal := []ficomodel.VendorJournal{}
			e := h.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
			if e != nil {
				ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
			}

			//convert list vendor journal to map[string]VendorJournal
			mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
				return detail.ID, detail
			})

			// get list of vendor trx by list of ids
			cashSchedule := []ficomodel.CashSchedule{}
			e = h.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
			if e != nil {
				ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
			}

			// join with vendor journal to get status and voucher no
			opExpenses := []bagongmodel.SiteExpense{}
			otExpenses := []bagongmodel.SiteExpense{}

			for _, d := range res.OperationalExpense {
				tmp := d
				if v, ok := mapVendorJournal[d.JournalID]; ok {
					for _, lTrx := range cashSchedule {
						if lTrx.SourceJournalID == v.ID {
							tmp.VoucherID = lTrx.VoucherNo
							break
						}
					}
					tmp.ApprovalStatus = v.Status
				}
				opExpenses = append(opExpenses, tmp)
			}
			for _, d := range res.OtherExpense {
				tmp := d
				if v, ok := mapVendorJournal[d.JournalID]; ok {
					for _, lTrx := range cashSchedule {
						if lTrx.SourceJournalID == v.ID {
							tmp.VoucherID = lTrx.VoucherNo
							break
						}
					}
					tmp.ApprovalStatus = v.Status
				}
				otExpenses = append(otExpenses, tmp)
			}

			res.OperationalExpense = opExpenses
			res.OtherExpense = otExpenses

			// res.Set("data", ms)
			ctx.Data().Set("FnResult", res)
		}
		return true, nil
	}
}

func MWPostSiteEntryBTS() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		routePath := ctx.Data().Data()["RoutePath"].(string)
		arrPath := strings.Split(string(routePath), "/")
		lastPath := arrPath[len(arrPath)-1]

		if lastPath == "get" {
			h := sebar.GetTenantDBFromContext(ctx)

			//get data from response
			res, ok := ctx.Data().Data()["FnResult"].(*bagongmodel.SiteEntryBTSDetail)
			if !ok {
				return true, nil
			}

			//collect ids from response
			vendorJournalIDs := []interface{}{}
			for _, c := range res.Expense {
				vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
			}

			//get list of vendor account by list of ids
			vendorJournal := []ficomodel.VendorJournal{}
			e := h.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
			if e != nil {
				ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
			}

			//convert list vendor journal to map[string]VendorJournal
			mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
				return detail.ID, detail
			})

			// get list of vendor trx by list of ids
			cashSchedule := []ficomodel.CashSchedule{}
			e = h.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
			if e != nil {
				ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
			}

			// join with vendor journal to get status and voucher no
			siteExpenses := []bagongmodel.SiteExpense{}

			for _, d := range res.Expense {
				tmp := d
				if v, ok := mapVendorJournal[d.JournalID]; ok {
					for _, lTrx := range cashSchedule {
						if lTrx.SourceJournalID == v.ID {
							tmp.VoucherID = lTrx.VoucherNo
							break
						}
					}
					tmp.ApprovalStatus = v.Status
				}
				siteExpenses = append(siteExpenses, tmp)
			}

			res.Expense = siteExpenses

			// res.Set("data", ms)
			ctx.Data().Set("FnResult", res)
		}
		return true, nil
	}
}

func MWPostSiteEntryMining() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		routePath := ctx.Data().Data()["RoutePath"].(string)
		arrPath := strings.Split(string(routePath), "/")
		lastPath := arrPath[len(arrPath)-1]

		if lastPath == "get" {
			h := sebar.GetTenantDBFromContext(ctx)

			//get data from response
			res, ok := ctx.Data().Data()["FnResult"].(*bagongmodel.SiteEntryMiningDetail)
			if !ok {
				return true, nil
			}

			//collect ids from response
			vendorJournalIDs := []interface{}{}
			for _, c := range res.Expense {
				vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
			}

			//get list of vendor account by list of ids
			vendorJournal := []ficomodel.VendorJournal{}
			e := h.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
			if e != nil {
				ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
			}

			//convert list vendor journal to map[string]VendorJournal
			mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
				return detail.ID, detail
			})

			// get list of vendor trx by list of ids
			cashSchedule := []ficomodel.CashSchedule{}
			e = h.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
			if e != nil {
				ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
			}

			// join with vendor journal to get status and voucher no
			siteExpenses := []bagongmodel.SiteExpense{}

			for _, d := range res.Expense {
				tmp := d
				if v, ok := mapVendorJournal[d.JournalID]; ok {
					for _, lTrx := range cashSchedule {
						if lTrx.SourceJournalID == v.ID {
							tmp.VoucherID = lTrx.VoucherNo
							break
						}
					}
					tmp.ApprovalStatus = v.Status
				}
				siteExpenses = append(siteExpenses, tmp)
			}

			res.Expense = siteExpenses

			// res.Set("data", ms)
			ctx.Data().Set("FnResult", res)
		}
		return true, nil
	}
}

func MWPostNonAsset(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		routePath := ctx.Data().Data()["RoutePath"].(string)
		arrPath := strings.Split(string(routePath), "/")
		lastPath := arrPath[len(arrPath)-1]

		h := sebar.GetTenantDBFromContext(ctx)

		if lastPath == "get" {

			if len(fields) == 0 {
				fields = []string{"ExpenseDetail"}
			}

			//get data from response
			res, ok := ctx.Data().Data()["FnResult"].(*bagongmodel.SiteEntryNonAsset)
			if !ok {
				return true, nil
			}

			//collect ids from response
			vendorJournalIDs := []interface{}{}
			for _, c := range res.ExpenseDetail {
				vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
			}

			//get list of vendor account by list of ids
			vendorJournal := []ficomodel.VendorJournal{}
			e := h.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
			if e != nil {
				ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
			}

			//convert list vendor journal to map[string]VendorJournal
			mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
				return detail.ID, detail
			})

			// get list of vendor trx by list of ids
			cashSchedule := []ficomodel.CashSchedule{}
			e = h.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
			if e != nil {
				ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
			}

			// join with vendor journal to get status and voucher no
			siteExpense := []bagongmodel.SiteExpense{}

			for _, d := range res.ExpenseDetail {
				tmp := d
				if v, ok := mapVendorJournal[d.JournalID]; ok {
					for _, lTrx := range cashSchedule {
						if lTrx.SourceJournalID == v.ID {
							tmp.VoucherID = lTrx.VoucherNo
							break
						}
					}
					tmp.ApprovalStatus = v.Status
				}
				siteExpense = append(siteExpense, tmp)
			}

			res.ExpenseDetail = siteExpense

			// res.Set("data", ms)
			ctx.Data().Set("FnResult", res)
		}
		return true, nil
	}
}
