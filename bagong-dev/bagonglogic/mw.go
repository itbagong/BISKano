package bagonglogic

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

func MWPostGetTariff() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		fields := []string{"SiteID"}

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
		siteIDs := []interface{}{}
		for _, field := range fields {
			siteIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			siteIDs = append(siteIDs, siteIDFields...)
		}

		//get list of ledger account by list of ids
		sites := []bagongmodel.Site{}
		e := h.GetsByFilter(new(bagongmodel.Site), dbflex.In("_id", siteIDs...), &sites)
		if e != nil {
			ctx.Log().Errorf("Failed populate data site: %s", e.Error())
		}

		//convert list ledger account to map[string]LedgerAccount
		mapSites := lo.Associate(sites, func(detail bagongmodel.Site) (string, bagongmodel.Site) {
			return detail.ID, detail
		})

		var revenue, expense, income float64
		for _, m := range ms {
			m.Set("RowType", "row")
			for _, field := range fields {
				if v, ok := mapSites[m.GetString(field)]; ok {
					m.Set(field, v.Name)
				}
			}
			revenue += m.GetFloat64("Revenue")
			expense += m.GetFloat64("Expense")
			income += m.GetFloat64("Income")
		}

		total := codekit.M{
			"RowType": "total",
			"Revenue": revenue,
			"Expense": expense,
			"Income":  income,
		}
		var mt []codekit.M
		mt = append(mt, total)
		res.Set("data", ms)
		res.Set("summary", mt)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPreExtenseFind(ctx *kaos.Context, payload interface{}) (bool, error) {
	parm, ok := payload.(*dbflex.QueryParam)
	if !ok {
		return false, fmt.Errorf("invalid: Payload, got %t", payload)
	}
	parm.MergeWhere(false, dbflex.Eq("TerminalExpense", true))

	return true, nil
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
					m.Set(field, v.Name)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostSiteEntry(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"SiteID"}
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
		siteIDs := []interface{}{}
		for _, field := range fields {
			siteIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			siteIDs = append(siteIDs, siteIDFields...)
		}

		//get list of ledger account by list of ids
		sites := []bagongmodel.Site{}
		e := h.GetsByFilter(new(bagongmodel.Site), dbflex.In("_id", siteIDs...), &sites)
		if e != nil {
			ctx.Log().Errorf("Failed populate data site: %s", e.Error())
		}

		//convert list ledger account to map[string]LedgerAccount
		mapSites := lo.Associate(sites, func(detail bagongmodel.Site) (string, bagongmodel.Site) {
			return detail.ID, detail
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapSites[m.GetString(field)]; ok {
					m.Set("SiteName", v.Name)
				}
			}
			m.Set("SiteID", m.GetString("SiteID"))
		}
		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostDimension(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"DimensionID"}
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
		dimensionIDs := []interface{}{}
		for _, field := range fields {
			dimensionIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})

			dimensionIDs = append(dimensionIDs, dimensionIDFields...)
		}

		//get list of dimension by list of ids
		dimensions := []tenantcoremodel.DimensionMaster{}
		e := h.GetsByFilter(new(tenantcoremodel.DimensionMaster), dbflex.In("_id", dimensionIDs...), &dimensions)
		if e != nil {
			ctx.Log().Errorf("Failed populate data master dimensions: %s", e.Error())
		}

		//convert list dimension to map[string]MasterDimension
		mapDimensions := lo.Associate(dimensions, func(group tenantcoremodel.DimensionMaster) (string, tenantcoremodel.DimensionMaster) {
			return group.ID, group
		})

		for _, m := range ms {
			for _, field := range fields {
				if v, ok := mapDimensions[m.GetString(field)]; ok {
					m.Set(field, v.Label)
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostTerminal(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"Terminals"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}

		serde.Serde(res["data"], &ms)

		terminalIDs := []interface{}{}
		for _, field := range fields {
			for _, m := range ms {
				if val, ok := m.Get(field).([]string); ok {
					for _, v := range val {
						terminalIDs = append(terminalIDs, v)
					}
				}
			}
		}

		terminals := []bagongmodel.Terminal{}
		h.GetsByFilter(new(bagongmodel.Terminal), dbflex.In("_id", terminalIDs...), &terminals)
		mapTerminals := lo.Associate(terminals, func(terminal bagongmodel.Terminal) (string, bagongmodel.Terminal) {
			return terminal.ID, terminal
		})

		for _, m := range ms {
			for _, field := range fields {
				terminalIDs, ok := m.Get(field).([]string)
				if ok {
					for keyIDOfCharge, id := range terminalIDs {
						if v, ok := mapTerminals[id]; ok {
							terminalIDs[keyIDOfCharge] = v.Name
						}
					}

					if len(terminalIDs) > 0 {
						m.Set(field, terminalIDs)
					}
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostSiteEntryAsset() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		fields := []string{"AssetID"}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)

		ms := []codekit.M{}

		serde.Serde(res["data"], &ms)

		// get header site entry
		getSiteEntryID := func() string {
			for _, c := range ms {
				return c.GetString("SiteEntryID")
			}
			return ""
		}
		siteEntryID := getSiteEntryID()
		siteEntry := new(bagongmodel.SiteEntry)
		if e := h.GetByID(siteEntry, siteEntryID); e != nil {
			return false, fmt.Errorf("site entry not found: %s", siteEntryID)
		}

		// get site entry asset id
		siteEntryAssetIDs := lo.Map(ms, func(m codekit.M, index int) interface{} {
			return m.GetString("_id")
		})

		// get site entry project id
		siteEntryAssetProjectIDs := lo.Map(ms, func(m codekit.M, index int) interface{} {
			return m.GetString("ProjectID")
		})

		// get master site
		site := new(bagongmodel.Site)
		if e := h.GetByID(site, siteEntry.SiteID); e != nil {
			return false, fmt.Errorf("site not found: %s", siteEntry.SiteID)
		}

		// get master asset
		assetIDs := []interface{}{}
		for _, field := range fields {
			assetIDFields := lo.Map(ms, func(m codekit.M, index int) interface{} {
				return m.GetString(field)
			})
			assetIDs = append(assetIDs, assetIDFields...)
		}

		assets := []bagongmodel.Asset{}
		h.GetsByFilter(new(bagongmodel.Asset), dbflex.In("_id", assetIDs...), &assets)
		mapAssets := lo.Associate(assets, func(asset bagongmodel.Asset) (string, bagongmodel.Asset) {
			return asset.ID, asset
		})

		measureProjects := []sdpmodel.MeasuringProject{}
		h.GetsByFilter(new(sdpmodel.MeasuringProject), dbflex.In("_id", siteEntryAssetProjectIDs...), &measureProjects)
		mapProjects := lo.Associate(measureProjects, func(measureProjects sdpmodel.MeasuringProject) (string, string) {
			return measureProjects.ID, measureProjects.ProjectName
		})

		// get master asset in tenant
		tenantAssets := []tenantcoremodel.Asset{}
		h.GetsByFilter(new(tenantcoremodel.Asset), dbflex.In("_id", assetIDs...), &tenantAssets)
		mapTenantAssets := lo.Associate(tenantAssets, func(asset tenantcoremodel.Asset) (string, tenantcoremodel.Asset) {
			return asset.ID, asset
		})

		// get key assetType
		assetTypeIDs := []interface{}{}
		for _, field := range tenantAssets {
			assetTypeIDs = append(assetTypeIDs, field.AssetType)
		}

		// get master data asset
		masterdata := []tenantcoremodel.MasterData{}
		h.GetsByFilter(new(tenantcoremodel.MasterData), dbflex.In("_id", assetTypeIDs...), &masterdata)
		mapTenantMasterData := lo.Associate(masterdata, func(master tenantcoremodel.MasterData) (string, tenantcoremodel.MasterData) {
			return master.ID, master
		})

		trayekIDs := lo.Map(assets, func(m bagongmodel.Asset, index int) interface{} {
			return m.DetailUnit.TrayekID
		})

		// get driver name by purpose
		mapDrivers := make(map[string]string)
		mapStatuses := make(map[string]string)
		mapTrayeks := make(map[string]string)
		driverIds := []interface{}{}
		if site.Purpose == "BTS" {
			details := []bagongmodel.SiteEntryBTSDetail{}
			h.GetsByFilter(new(bagongmodel.SiteEntryBTSDetail), dbflex.In("_id", siteEntryAssetIDs...), &details)
			mapDrivers = lo.Associate(details, func(detail bagongmodel.SiteEntryBTSDetail) (string, string) {
				// todo mapping driver
				return detail.ID, ""
			})
			mapStatuses = lo.Associate(details, func(detail bagongmodel.SiteEntryBTSDetail) (string, string) {
				return detail.ID, detail.Status
			})
		} else if site.Purpose == "Mining" {
			details := []bagongmodel.SiteEntryMiningDetail{}
			h.GetsByFilter(new(bagongmodel.SiteEntryMiningDetail), dbflex.In("_id", siteEntryAssetIDs...), &details)
			mapDrivers = lo.Associate(details, func(detail bagongmodel.SiteEntryMiningDetail) (string, string) {
				return detail.ID, detail.DriverID
			})
			mapStatuses = lo.Associate(details, func(detail bagongmodel.SiteEntryMiningDetail) (string, string) {
				return detail.ID, detail.Status
			})
		} else if site.Purpose == "Trayek" {
			details := []bagongmodel.SiteEntryTrayekDetail{}
			h.GetsByFilter(new(bagongmodel.SiteEntryTrayekDetail), dbflex.In("_id", siteEntryAssetIDs...), &details)
			mapDrivers = lo.Associate(details, func(detail bagongmodel.SiteEntryTrayekDetail) (string, string) {
				return detail.ID, detail.DriverID
			})
			mapStatuses = lo.Associate(details, func(detail bagongmodel.SiteEntryTrayekDetail) (string, string) {
				return detail.ID, detail.Status
			})

			trayeks := []bagongmodel.Trayek{}
			h.GetsByFilter(new(bagongmodel.Trayek), dbflex.In("_id", trayekIDs...), &trayeks)
			mapTrayeks = lo.Associate(trayeks, func(detail bagongmodel.Trayek) (string, string) {
				return detail.ID, detail.Name
			})
		} else if site.Purpose == "Tourism" {
			details := []bagongmodel.SiteEntryTourismDetail{}
			h.GetsByFilter(new(bagongmodel.SiteEntryTourismDetail), dbflex.In("_id", siteEntryAssetIDs...), &details)
			mapDrivers = lo.Associate(details, func(detail bagongmodel.SiteEntryTourismDetail) (string, string) {
				return detail.ID, detail.DriverID
			})
			mapStatuses = lo.Associate(details, func(detail bagongmodel.SiteEntryTourismDetail) (string, string) {
				return detail.ID, detail.Status
			})
		}

		for _, c := range mapDrivers {
			driverIds = append(driverIds, c)
		}
		employees := []tenantcoremodel.Employee{}
		h.GetsByFilter(new(tenantcoremodel.Employee), dbflex.In("_id", driverIds...), &employees)
		mapEmployees := lo.Associate(employees, func(emp tenantcoremodel.Employee) (string, string) {
			return emp.ID, emp.Name
		})

		// get master customer
		customerIDs := []interface{}{}
		for _, c := range assets {
			for _, d := range c.UserInfo {
				customerIDs = append(customerIDs, d.CustomerID)
			}
		}

		customers := []tenantcoremodel.Customer{}
		h.GetsByFilter(new(tenantcoremodel.Customer), dbflex.In("_id", customerIDs...), &customers)
		mapCustomers := lo.Associate(customers, func(customer tenantcoremodel.Customer) (string, tenantcoremodel.Customer) {
			return customer.ID, customer
		})
		// mapping res
		for _, m := range ms {
			for _, field := range fields {
				m.Set("RowType", "row")
				if v, ok := mapAssets[m.GetString(field)]; ok {
					m.Set("PoliceNum", v.DetailUnit.PoliceNum)
					m.Set("UnitType", v.DetailUnit.UnitType)

					if len(v.UserInfo) > 0 {
						userInfo := v.UserInfo[0]

						for i := 1; i < len(v.UserInfo); i++ {
							dateFrom1 := userInfo.AssetDateFrom

							dateFrom2 := v.UserInfo[i].AssetDateFrom

							if dateFrom2.After(dateFrom1) {
								userInfo = v.UserInfo[i]
							}
						}

						if v2, ok2 := mapCustomers[userInfo.CustomerID]; ok2 {
							m.Set("CustomerName", v2.Name)
						}
					}

					if t, ok := mapTrayeks[v.DetailUnit.TrayekID]; ok {
						m.Set("TrayekName", t)
					} else {
						m.Set("TrayekName", "")
					}
				}

				if v3, ok3 := mapDrivers[m.GetString("_id")]; ok3 {
					m.Set("DriverName", mapEmployees[v3])
				}
				if v4, ok4 := mapStatuses[m.GetString("_id")]; ok4 {
					m.Set("Status", v4)
				}
				if v5, ok5 := mapProjects[m.GetString("ProjectID")]; ok5 {
					m.Set("ProjectID", v5)
				}
			}

			if v, ok := mapTenantAssets[m.GetString("AssetID")]; ok {
				m.Set("AssetType", v.AssetType)
			}

			if v, ok := mapTenantMasterData[m.GetString("AssetType")]; ok {
				m.Set("AssetTypeName", v.Name)
			}
		}

		res.Set("data", ms)

		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostAccidentFund(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"EmployeeID"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		empIDs := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					empIDs[i] = val
				}
			}
		}

		emps := []tenantcoremodel.Employee{}
		h.GetsByFilter(new(tenantcoremodel.Employee), dbflex.In("_id", empIDs...), &emps)
		mapEmployee := lo.Associate(emps, func(emp tenantcoremodel.Employee) (string, tenantcoremodel.Employee) {
			return emp.ID, emp
		})

		for _, m := range ms {
			for _, field := range fields {
				empIDs, ok := m.Get(field).(string)
				if ok {
					if v, ok := mapEmployee[empIDs]; ok {
						m[field] = v.Name
					}
				}
			}
		}

		res.Set("data", ms)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostAsset(fields ...string) kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		if len(fields) == 0 {
			fields = []string{"_id"}
		}

		res, ok := ctx.Data().Data()["FnResult"].(codekit.M)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ms := []codekit.M{}
		serde.Serde(res["data"], &ms)

		ids := make([]interface{}, len(ms))
		for _, field := range fields {
			for i, m := range ms {
				if val, ok := m.Get(field).(string); ok {
					ids[i] = val
				}
			}
		}

		assets := []tenantcoremodel.Asset{}
		h.GetsByFilter(new(tenantcoremodel.Asset), dbflex.In("_id", ids...), &assets)
		mapAsset := lo.Associate(assets, func(asset tenantcoremodel.Asset) (string, tenantcoremodel.Asset) {
			return asset.ID, asset
		})

		now := time.Now()
		data := make([]codekit.M, 0)
		for _, m := range ms {
			m["AcquisitionDate"] = nil
			m["AcquisitionAmount"] = 0.0
			m["LifeTime"] = nil
			m["DepreciationAmount"] = 0.0
			m["NBV"] = 0.0
			Lc, ok := m["LatestCustomer"].(string)
			if ok {
				m["LatestCustomer"] = Lc
			} else {
				m["LatestCustomer"] = ""
			}
			DimAsset, ok := m["Dimension"].(tenantcoremodel.Dimension)
			if ok {
				m["Dimension"] = DimAsset
			} else {
				m["Dimension"] = tenantcoremodel.Dimension{}
			}

			if v, ok := mapAsset[m["_id"].(string)]; ok {
				m["Name"] = v.Name
				m["GroupID"] = v.GroupID
				m["AssetType"] = v.AssetType

				acquisitionAmount := 0.0
				// set AcquisitionDate & AcquisitionAmount
				for _, r := range m["References"].([]codekit.M) {
					switch strings.ToLower(r["Key"].(string)) {
					case "acquisition date":
						if !(r["Value"] == nil) {
							date, err := time.Parse(time.RFC3339Nano, r["Value"].(string))
							if err != nil {
								return true, fmt.Errorf(fmt.Sprintf("error when parse acquisition date: %s", err.Error()))
							}
							m["AcquisitionDate"] = date
						} else {
							m["AcquisitionDate"] = ""
						}
					case "acquisition cost":
						var err error
						t := reflect.TypeOf(r["Value"]).Kind()
						if t == reflect.Int32 {
							acquisitionAmount = float64(r["Value"].(int32))
						} else if t == reflect.Int64 {
							acquisitionAmount = float64(r["Value"].(int64))
						}

						if err != nil {
							return true, fmt.Errorf(fmt.Sprintf("error when parse acquisition cost: %s", err.Error()))
						}
						m["AcquisitionAmount"] = acquisitionAmount
					}
				}

				// set LifeTime
				dep, ok := m["Depreciation"].(bagongmodel.Depreciation)
				if ok {
					// set DepreciationAmount
					depreciationAmount := 0.0
					if float64(dep.AssetDuration) > 0 {
						depreciationAmount = (acquisitionAmount - dep.ResidualAmount) / float64(dep.AssetDuration)
					}
					depreciationAmount = lo.Ternary(math.IsNaN(depreciationAmount), 0, depreciationAmount)
					m["DepreciationAmount"] = depreciationAmount
					if dep.DepreciationDate != nil {
						if dep.DepreciationPeriod == "Monthly" {
							m["LifeTime"] = dep.DepreciationDate.AddDate(0, dep.AssetDuration, 0)
						} else if dep.DepreciationPeriod == "Yearly" {
							m["LifeTime"] = dep.DepreciationDate.AddDate(dep.AssetDuration, 0, 0)
						}

						// calculate nbv
						start := time.Date(dep.DepreciationDate.Year(), dep.DepreciationDate.Month(), dep.DepreciationDate.Day(), 0, 0, 0, 0, dep.DepreciationDate.Location())
						end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, dep.DepreciationDate.Location())
						count := 0.0
						for {
							var st time.Time
							if dep.DepreciationPeriod == "Monthly" {
								st = start.AddDate(0, 1, 0)
							} else if dep.DepreciationPeriod == "Yearly" {
								st = start.AddDate(1, 0, 0)
							}

							if st.After(end) || count > float64(dep.AssetDuration) {
								break
							}

							count++
							start = st
						}

						// set NBV
						m["NBV"] = acquisitionAmount - (depreciationAmount * count)
					}
				}

				data = append(data, m)
			} else {
				continue
			}
		}

		res.Set("data", data)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}

func MWPostGetAsset() kaos.MWFunc {
	return func(ctx *kaos.Context, payload interface{}) (bool, error) {
		res, ok := ctx.Data().Data()["FnResult"].(*bagongmodel.Asset)
		if !ok {
			return true, nil
		}

		h := sebar.GetTenantDBFromContext(ctx)
		ids := lo.Map(res.DepreciationActivity, func(m bagongmodel.DepreciationActivity, index int) string {
			return m.ID
		})

		journals := []ficomodel.LedgerJournal{}
		e := h.GetsByFilter(new(ficomodel.LedgerJournal), dbflex.ElemMatch("Lines.References", dbflex.Eq("Key", "DepreciationID"), dbflex.In("Value", ids...)), &journals)
		if e != nil {
			ctx.Log().Errorf("Failed populate data site: %s", e.Error())
		}

		mapJournal := map[string]string{}
		for _, id := range ids {
			for _, j := range journals {
				isExist := false
				for _, l := range j.Lines {
					val := l.References.Get("DepreciationID", "").(string)
					if val == id {
						mapJournal[id] = j.ID
						isExist = true
						break
					}
				}

				if isExist {
					break
				}
			}
		}

		ms := codekit.M{}
		serde.Serde(res, &ms)

		deps := make([]codekit.M, len(res.DepreciationActivity))
		for i, d := range res.DepreciationActivity {
			dep := codekit.M{}
			serde.Serde(d, &dep)

			if v, ok := mapJournal[d.ID]; ok {
				dep["JournalID"] = v
			} else {
				dep["JournalID"] = ""
			}

			deps[i] = dep
		}
		ms["DepreciationActivity"] = deps

		serde.Serde(ms, res)
		ctx.Data().Set("FnResult", res)
		return true, nil
	}
}
