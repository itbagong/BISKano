package sdplogic

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/sebarcode/codekit"
)

func MWPreMeasuringProject() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)

		//collect ids
		// ACT01 - Revenue
		// ACT02 - Expense
		itemIDs := []string{"ACT02", "ACT01"}

		//get list of item by list of ids
		items := []tenantcoremodel.LedgerAccount{}
		e := h.GetsByFilter(new(tenantcoremodel.LedgerAccount), dbflex.In("AccountType", itemIDs...), &items)
		if e != nil {
			ctx.Log().Errorf("Failed populate data items: %s", e.Error())
		}

		o, ok := payload.(*sdpmodel.MeasuringProject)
		if !ok {
			ctx.Log().Errorf("Invalid: Payload, got %s", e.Error())
		}

		if o.ID == "" {
			// days := o.EndPeriodMonth.Sub(o.StartPeriodMonth).Hours() / 24
			// o.ProjectPeriod = int(days)

			for _, l := range items {
				StartPeriodMonth := o.StartPeriodMonth
				EndPeriodMonth := o.EndPeriodMonth
				for i := o.StartPeriodMonth.Year(); i <= o.EndPeriodMonth.Year(); i++ {
					Line := sdpmodel.LinesMeasuringProject{}
					Line.Budget = string(l.AccountType)
					Line.Year = i
					Line.LedgerAccount = l.ID
					// Line.LedgerAccountType = string(l.AccountType)

					var arrMonths map[string]float64
					arrMonths = map[string]float64{}
					for EndPeriodMonth.After(StartPeriodMonth) {
						sMonth := time.Month(StartPeriodMonth.Month()).String()
						arrMonths[sMonth] = 0.0

						StartPeriodMonth = StartPeriodMonth.AddDate(0, 1, 0)
						if int(StartPeriodMonth.Month()) == 1 {
							break
						}
					}

					Line.Month = arrMonths
					o.Lines = append(o.Lines, Line)
				}
			}
		} else {
			RevenueEstimation := 0.0
			ExpenseEstimation := 0.0

			for _, l := range o.Lines {
				if l.Budget == "ACT01" {
					for _, m := range l.Month {
						RevenueEstimation += m
					}
				} else {
					for _, m := range l.Month {
						ExpenseEstimation += m
					}
				}
			}
			o.RevenueEstimation = RevenueEstimation
			o.ExpenseEstimation = ExpenseEstimation

		}

		return true, nil
	}
}

func MWPreSalesOrder() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)
		// Sales Order       : SO/COUNTER/BDM-HO-CUST/MM/YY > SO/0001/BDM-HO-CK/03/24

		//get list of item by list of ids
		items := []sdpmodel.SalesOrder{}
		e := h.GetsByFilter(new(sdpmodel.SalesOrder), dbflex.Eq("Year", time.Now().Year()), &items)
		if e != nil {
			ctx.Log().Errorf("Failed populate data order: %s", e.Error())
		}

		o, ok := payload.(*sdpmodel.SalesOrder)
		if !ok {
			ctx.Log().Errorf("Invalid: Payload, got %s", e.Error())
		}

		CU := new(tenantcoremodel.Customer)
		e = h.GetByID(CU, o.CustomerID)
		if e != nil {
			ctx.Log().Errorf("Invalid: Payload, got %s", e.Error())
		}

		if o.ID == "" {
			t := time.Now()
			sTime := t.Format("01/2006")
			o.Year = t.Year()
			o.SalesOrderNo = "SO/" + strconv.Itoa(len(items)) + "/BDM-HO-" + CU.CustomerAlias + "/" + sTime
		}

		return true, nil
	}
}

func MWPreSalesOpportunity() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		h := sebar.GetTenantDBFromContext(ctx)
		// Sales Opportunity : OP/COUNTER/BDM-HO/MM/YY > OP/0001/BDM-HO/03/24

		//get list of item by list of ids
		items := []sdpmodel.SalesOpportunity{}
		e := h.GetsByFilter(new(sdpmodel.SalesOpportunity), dbflex.Eq("Year", time.Now().Year()), &items)
		if e != nil {
			ctx.Log().Errorf("Failed populate data order: %s", e.Error())
		}

		o, ok := payload.(*sdpmodel.SalesOpportunity)
		if !ok {
			ctx.Log().Errorf("Invalid: Payload, got %s", e.Error())
		}

		// CU := new(tenantcoremodel.Customer)
		// e = h.GetByID(CU, o.CustomerID)
		// if e != nil {
		// 	ctx.Log().Errorf("Invalid: Payload, got %s", e.Error())
		// }

		if o.ID == "" {
			t := time.Now()
			sTime := t.Format("01/2006")
			// o.Year = t.Year()
			o.OpportunityNo = "OP/" + strconv.Itoa(len(items)) + "/BDM-HO/" + sTime
		}

		return true, nil
	}
}

// MWPreAssignCompanyID middleware untuk menggunakan CompanyID
func MWPreAssignCompanyID() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		ctx.Data().Set(tenantcoremodel.DBModFilter, []*dbflex.Filter{dbflex.Eq("CompanyID", tenantcorelogic.GetCompanyIDFromContext(ctx))})

		field := "CompanyID"
		companyID := tenantcorelogic.GetCompanyIDFromContext(ctx)

		if payload == nil {
			return true, nil
		}

		t := reflect.TypeOf(payload)
		if t.Kind() == reflect.Pointer {
			t = reflect.Indirect(reflect.ValueOf(payload)).Type()
		}

		m := codekit.M{}
		if e := serde.Serde(payload, &m); e != nil {
			return true, nil
		}

		rv := reflect.Indirect(reflect.ValueOf(payload))

		if t.Kind() == reflect.Struct {
			for index := 0; index < t.NumField(); index++ {
				if t.Field(index).Type.Kind() == reflect.Struct {
					for i := 0; i < t.Field(index).Type.NumField(); i++ {
						if strings.ToLower(t.Field(index).Type.Field(i).Name) == strings.ToLower(field) {
							rv.Field(index).Field(i).SetString(companyID)
						}
					}
				}
				if strings.ToLower(t.Field(index).Name) == strings.ToLower(field) {
					rv.Field(index).SetString(companyID)
				}
			}
		}

		return true, nil
	}
}
