package bagonglogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/eaciit/toolkit"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteEntryEngine struct{}

type SiteEntryAssetReq struct {
	SiteEntryID string
}

type SiteEntryReq struct {
	SiteID    string
	CompanyID string
	TrxDate   time.Time
}

type EmployeeReq struct {
	Search string
	Prefix string
}

type TariffReq struct {
	TrayekID string
	From     string
	To       string
}

type JournalPostReq struct {
	SiteID string
	Type   string
}

type SiteEntryLedgerJournalPostReq struct {
	SiteID string
}

type SumTotal struct {
	RowType string
	Revenue float64
	Expense float64
	Income  float64
}

type PayloadSiteEntryLedgerJournal struct {
	Purpose         string
	SiteEntry       codekit.M
	LedgerJournal   ficomodel.LedgerJournal
	CustomerJournal ficomodel.CustomerJournal
}

func (engine *SiteEntryEngine) FindByExpenseTypesTrue(ctx *kaos.Context, payload *dbflex.QueryParam) ([]tenantcoremodel.ExpenseType, error) {
	expenseType := []tenantcoremodel.ExpenseType{}
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return expenseType, errors.New("missing: connection")
	}

	r := ctx.Data().Get("http_request", nil).(*http.Request)

	name := strings.TrimSpace(r.URL.Query().Get("Name"))

	filters := []*dbflex.Filter{}

	if name != "" {
		filters = append(filters, dbflex.Contains("Name", name))
	}
	filters = append(filters, dbflex.Eq("Cashable", true))

	if payload != nil {
		if payload.Where != nil {
			fieldVal := payload.Where.Field
			if fieldVal == "_id" {
				aInterface := payload.Where.Value.(interface{})
				vID := aInterface.(string)
				if vID != "" {
					filters = append(filters, dbflex.Eq("_id", vID))
				}
			} else if fieldVal == "Name" {
				aInterface := payload.Where.Value.([]interface{})
				aString := make([]string, len(aInterface))
				for i, v := range aInterface {
					aString[i] = v.(string)
				}
				if len(aString) > 0 {
					filters = append(filters, dbflex.Contains("Name", aString[0]))
				}
			}
		}
	}

	if len(filters) > 0 {
		e := h.Gets(new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(dbflex.And(filters...)).SetSort("Name"), &expenseType)
		if e != nil {
			return expenseType, errors.New("Failed populate data driver: " + e.Error())
		}
	} else {
		e := h.Gets(new(tenantcoremodel.Employee), nil, &expenseType)
		if e != nil {
			return expenseType, errors.New("Failed populate data driver: " + e.Error())
		}
	}

	return expenseType, nil
}

func (engine *SiteEntryEngine) FindByExpenseTypesFalse(ctx *kaos.Context, payload *dbflex.QueryParam) ([]tenantcoremodel.ExpenseType, error) {
	expenseType := []tenantcoremodel.ExpenseType{}
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return expenseType, errors.New("missing: connection")
	}

	r := ctx.Data().Get("http_request", nil).(*http.Request)

	name := strings.TrimSpace(r.URL.Query().Get("Name"))

	filters := []*dbflex.Filter{}

	if name != "" {
		filters = append(filters, dbflex.Contains("Name", name))
	}
	filters = append(filters, dbflex.Eq("Cashable", false))

	if payload != nil {
		if payload.Where != nil {
			fieldVal := payload.Where.Field
			if fieldVal == "_id" {
				aInterface := payload.Where.Value.(interface{})
				vID := aInterface.(string)
				if vID != "" {
					filters = append(filters, dbflex.Eq("_id", vID))
				}
			} else if fieldVal == "Name" {
				aInterface := payload.Where.Value.([]interface{})
				aString := make([]string, len(aInterface))
				for i, v := range aInterface {
					aString[i] = v.(string)
				}
				if len(aString) > 0 {
					filters = append(filters, dbflex.Contains("Name", aString[0]))
				}
			}
		}
	}

	if len(filters) > 0 {
		e := h.Gets(new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(dbflex.And(filters...)).SetSort("Name"), &expenseType)
		if e != nil {
			return expenseType, errors.New("Failed populate data driver: " + e.Error())
		}
	} else {
		e := h.Gets(new(tenantcoremodel.Employee), nil, &expenseType)
		if e != nil {
			return expenseType, errors.New("Failed populate data driver: " + e.Error())
		}
	}

	return expenseType, nil
}

func (engine *SiteEntryEngine) GetExpenseTypes(ctx *kaos.Context, payload tenantcoremodel.ExpenseType) ([]tenantcoremodel.ExpenseType, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	et := []tenantcoremodel.ExpenseType{}
	e := h.Gets(new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(dbflex.Eq("Cashable", payload.Cashable)), &et)
	if e != nil {
		return et, errors.New("Failed populate data ExpenseType: " + e.Error())
	}

	return et, nil
}

func (engine *SiteEntryEngine) GetExpenseTypesBySite(ctx *kaos.Context, payload SiteEntryLedgerJournalPostReq) ([]tenantcoremodel.ExpenseType, error) {
	h := sebar.GetTenantDBFromContext(ctx)

	site := new(bagongmodel.Site)
	e := h.GetByID(site, payload.SiteID)
	if e != nil {
		return nil, errors.New("Failed populate data ExpenseType: " + e.Error())
	}

	return site.Expense, nil
}

func (engine *SiteEntryEngine) GetTotalSiteEntry(ctx *kaos.Context, payload []bagongmodel.SiteEntry) ([]SumTotal, error) {
	totals := []SumTotal{}
	total := SumTotal{
		RowType: "total",
	}

	for _, m := range payload {
		total.Expense += m.Expense
		total.Revenue += m.Revenue
		total.Income += m.Income
	}

	totals = append(totals, total)

	return totals, nil
}

func (engine *SiteEntryEngine) GetTotalSiteEntryAsset(ctx *kaos.Context, payload []bagongmodel.SiteEntryAsset) ([]SumTotal, error) {
	h := sebar.GetTenantDBFromContext(ctx)

	totals := []SumTotal{}

	subTotal := SumTotal{
		RowType: "subTotal",
	}

	var siteEntryID string
	for _, m := range payload {
		siteEntryID = m.SiteEntryID
		subTotal.Expense += m.Expense
		subTotal.Revenue += m.Revenue
		subTotal.Income += m.Income
	}

	siteEntryNonAsset := new(bagongmodel.SiteEntryNonAsset)
	if e := h.GetByFilter(siteEntryNonAsset, dbflex.Eq("_id", siteEntryID)); e != nil {
		return nil, fmt.Errorf("site entry non asset not found: %s", siteEntryID)
	}

	totalNonAsset := SumTotal{
		RowType: "totalNonAsset",
		Expense: siteEntryNonAsset.Expense,
		Revenue: siteEntryNonAsset.Revenue,
		Income:  siteEntryNonAsset.Income,
	}

	total := SumTotal{
		RowType: "total",
		Expense: siteEntryNonAsset.Expense + subTotal.Expense,
		Revenue: siteEntryNonAsset.Revenue + subTotal.Revenue,
		Income:  siteEntryNonAsset.Income + subTotal.Income,
	}

	totals = append(totals, subTotal)
	totals = append(totals, totalNonAsset)
	totals = append(totals, total)

	return totals, nil
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func (engine *SiteEntryEngine) GetSiteEntryAsset(ctx *kaos.Context, payload SiteEntryAssetReq) (string, error) {
	if payload.SiteEntryID == "" {
		return "", errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	siteEntry := new(bagongmodel.SiteEntry)
	if e := hub.GetByID(siteEntry, payload.SiteEntryID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("SiteEntry not found: %s", payload.SiteEntryID))
	}

	// Generate site entry asset from master
	siteEntryAssetCompare := []bagongmodel.SiteEntryAsset{}
	e := hub.GetsByFilter(new(bagongmodel.SiteEntryAsset), dbflex.Eq("SiteEntryID", payload.SiteEntryID), &siteEntryAssetCompare)
	if e != nil {
		return "", e
	}

	// Generate site entry asset from master
	tenantAssets := []tenantcoremodel.Asset{}
	e = hub.GetsByFilter(new(tenantcoremodel.Asset), dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", siteEntry.SiteID)), &tenantAssets)
	if e != nil {
		return "", e
	}

	ids := []interface{}{}
	for _, c := range tenantAssets {
		ids = append(ids, c.ID)
	}

	bagongAssets := []bagongmodel.Asset{}
	e = hub.GetsByFilter(new(bagongmodel.Asset), dbflex.In("_id", ids...), &bagongAssets)
	if e != nil {
		return "", e
	}

	siteEntryAsset := []bagongmodel.SiteEntryAsset{}
	for _, c := range bagongAssets {
		isInPeriod := false
		userInfo := bagongmodel.UserInfo{}
		// check period
		if len(c.UserInfo) > 0 {
			for _, d := range c.UserInfo {
				fmt.Println(d.AssetDateFrom.Format(time.DateTime), d.AssetDateTo.Format(time.DateTime), siteEntry.TrxDate.Format(time.DateTime))
				if inTimeSpan(d.AssetDateFrom, d.AssetDateTo, siteEntry.TrxDate) {
					isInPeriod = true
					userInfo = d
				}
			}
		}

		// check if unit
		tenantAsset := new(tenantcoremodel.Asset)
		e = hub.GetByID(tenantAsset, c.ID)
		if e != nil {
			return "", e
		}

		if tenantAsset.GroupID == "ELC" || tenantAsset.GroupID == "PRT" {
			continue
		}

		if !isInPeriod {
			continue
		}

		se := bagongmodel.SiteEntryAsset{
			SiteEntryID: payload.SiteEntryID,
			AssetID:     c.ID,
			PoliceNo:    c.DetailUnit.PoliceNum,
			UnitType:    c.DetailUnit.UnitType,
			ProjectID:   userInfo.ProjectID,
			HullNo:      userInfo.NoHullCustomer,
		}

		// if asset exist do update else insert
		for _, d := range siteEntryAssetCompare {
			if c.ID == d.AssetID {
				se = d
				se.HullNo = userInfo.NoHullCustomer
				break
			}
		}

		siteEntryAsset = append(siteEntryAsset, se)
	}

	bagongSite := new(bagongmodel.Site)
	e = hub.GetByID(bagongSite, siteEntry.SiteID)
	if e != nil {
		return "", e
	}

	hub.BeginTx()
	for _, c := range siteEntryAsset {
		e = hub.Save(&c)
		if e != nil {
			hub.Rollback()
			return "", e
		}
	}
	hub.Commit()

	return "success", nil
}

func (engine *SiteEntryEngine) GetSiteEntry(ctx *kaos.Context, payload SiteEntryReq) (string, error) {

	if payload.SiteID == "" && toolkit.IsNilOrEmpty(payload.TrxDate) {
		return "", errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	tzJakarta, e := time.LoadLocation("Asia/Jakarta")
	if e != nil {
		return "", errors.New("missing: timezone not found")
	}

	trxDate := payload.TrxDate.Format("2006-01-02")
	res, _ := time.Parse("2006-01-02", trxDate)

	// Check if siteentry exist
	isExistSiteEntry := new(bagongmodel.SiteEntry)
	e = hub.GetByFilter(isExistSiteEntry, dbflex.And(dbflex.Eq("SiteID", payload.SiteID), dbflex.Eq("TrxDate", res.In(tzJakarta).UTC())))
	if e == nil {
		return "", errors.New("missing: invalid request, site entry already registered")
	}

	// Generate site entry asset from master
	tenantAssets := []tenantcoremodel.Asset{}
	e = hub.GetsByFilter(new(tenantcoremodel.Asset), dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", payload.SiteID)), &tenantAssets)
	if e != nil {
		return "", e
	}

	ids := []interface{}{}
	for _, c := range tenantAssets {
		ids = append(ids, c.ID)
	}

	bagongAssets := []bagongmodel.Asset{}
	e = hub.GetsByFilter(new(bagongmodel.Asset), dbflex.In("_id", ids...), &bagongAssets)
	if e != nil {
		return "", e
	}

	siteEntryID := primitive.NewObjectID().Hex()
	siteEntryAsset := []bagongmodel.SiteEntryAsset{}

	for _, c := range bagongAssets {
		userInfo := bagongmodel.UserInfo{}
		isInPeriod := false

		// check period
		if len(c.UserInfo) > 0 {
			for _, d := range c.UserInfo {
				if inTimeSpan(d.AssetDateFrom, d.AssetDateTo, payload.TrxDate) {
					isInPeriod = true
					userInfo = d
				}
			}
		}

		// check if unit
		tenantAsset := new(tenantcoremodel.Asset)
		e = hub.GetByID(tenantAsset, c.ID)
		if e != nil {
			return "", e
		}

		if tenantAsset.GroupID == "ELC" || tenantAsset.GroupID == "PRT" {
			continue
		}

		if !isInPeriod {
			continue
		}

		se := bagongmodel.SiteEntryAsset{
			SiteEntryID: siteEntryID,
			AssetID:     c.ID,
			PoliceNo:    c.DetailUnit.PoliceNum,
			UnitType:    c.DetailUnit.UnitType,
			ProjectID:   userInfo.ProjectID,
			HullNo:      userInfo.NoHullCustomer,
		}

		for _, ui := range c.UserInfo {
			if payload.TrxDate.After(ui.AssetDateFrom) && payload.TrxDate.Before(ui.AssetDateTo) {
				se.CustomerID = ui.CustomerID
				break
			}
		}

		siteEntryAsset = append(siteEntryAsset, se)
	}

	bagongSite := new(bagongmodel.Site)
	e = hub.GetByID(bagongSite, payload.SiteID)
	if e != nil {
		return "", e
	}

	hub.BeginTx()

	e = hub.Save(&bagongmodel.SiteEntry{
		ID:      siteEntryID,
		TrxDate: res.In(tzJakarta).UTC(),
		SiteID:  bagongSite.ID,
		Purpose: bagongSite.Purpose,
	})
	if e != nil {
		return "", e
	}
	for _, c := range siteEntryAsset {
		e = hub.Save(&c)
		if e != nil {
			hub.Rollback()
			return "", e
		}
	}
	e = hub.Save(&bagongmodel.SiteEntryNonAsset{
		Created: res.In(tzJakarta).UTC(),
		ID:      siteEntryID,
	})
	if e != nil {
		hub.Rollback()
		return "", e
	}
	hub.Commit()

	return "success", nil
}

func (engine *SiteEntryEngine) SaveAssetDetail(ctx *kaos.Context, payload codekit.M) (interface{}, error) {
	if payload.GetString("_id") == "" {
		return "", errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get SiteEntryAsset
	siteEntryAsset := new(bagongmodel.SiteEntryAsset)
	if e := hub.GetByID(siteEntryAsset, payload.GetString("_id")); e != nil {
		return "", errors.New(fmt.Sprintf("SiteEntryAsset not found: %s", payload.GetString("_id")))
	}

	// get SiteEntry
	siteEntry := new(bagongmodel.SiteEntry)
	if e := hub.GetByID(siteEntry, siteEntryAsset.SiteEntryID); e != nil {
		return "", errors.New(fmt.Sprintf("SiteEntry not found: %s", siteEntryAsset.SiteEntryID))
	}

	if siteEntry.Purpose == "BTS" {
		siteEntryDetail := new(bagongmodel.SiteEntryBTSDetail)
		body, _ := json.Marshal(payload)
		json.Unmarshal(body, &siteEntryDetail)

		if e := hub.GetByID(new(bagongmodel.SiteEntryBTSDetail), siteEntryDetail.ID); e != nil {
			return "", errors.New(fmt.Sprintf("SiteEntryBTSDetail not found: %s", siteEntryDetail.ID))
		}

		if e := hub.Save(siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("save SiteEntryBTSDetail error: %s.", e.Error()))
		}

		if e := SumBTS(hub, siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("calculate summary BTS error: %s.", e.Error()))
		}

		return "success", nil
	}

	if siteEntry.Purpose == "Mining" {
		siteEntryDetail := new(bagongmodel.SiteEntryMiningDetail)
		body, _ := json.Marshal(payload)
		json.Unmarshal(body, &siteEntryDetail)

		if e := hub.GetByID(new(bagongmodel.SiteEntryMiningDetail), siteEntryDetail.ID); e != nil {
			return "", errors.New(fmt.Sprintf("SiteEntryMiningDetail not found: %s", siteEntryDetail.ID))
		}

		if e := hub.Save(siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("save SiteEntryMiningDetail error: %s.", e.Error()))
		}

		if e := SumMining(hub, siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("calculate summary Mining error: %s.", e.Error()))
		}

		return "success", nil
	}

	if siteEntry.Purpose == "Trayek" {
		siteEntryDetail := new(bagongmodel.SiteEntryTrayekDetail)
		body, _ := json.Marshal(payload)
		json.Unmarshal(body, &siteEntryDetail)

		if e := hub.GetByID(new(bagongmodel.SiteEntryTrayekDetail), siteEntryDetail.ID); e != nil {
			return "", errors.New(fmt.Sprintf("SiteEntryTrayekDetail not found: %s", siteEntryDetail.ID))
		}

		if e := hub.Save(siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("save SiteEntryTrayekDetail error: %s.", e.Error()))
		}

		if e := SumTrayek(hub, siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("calculate summary Trayek error: %s.", e.Error()))
		}

		return "success", nil
	}

	if siteEntry.Purpose == "Tourism" {
		siteEntryDetail := new(bagongmodel.SiteEntryTourismDetail)
		body, _ := json.Marshal(payload)
		json.Unmarshal(body, &siteEntryDetail)

		if e := hub.GetByID(new(bagongmodel.SiteEntryTourismDetail), siteEntryDetail.ID); e != nil {
			return "", errors.New(fmt.Sprintf("SiteEntryTourismDetail not found: %s", siteEntryDetail.ID))
		}

		if e := hub.Save(siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("save SiteEntryTourismDetail error: %s.", e.Error()))
		}

		if e := SumTourism(hub, siteEntryDetail); e != nil {
			return "", errors.New(fmt.Sprintf("calculate summary Tourism error: %s.", e.Error()))
		}

		return "success", nil
	}

	return "", errors.New(fmt.Sprintf("Purpose not found: %s", siteEntry.Purpose))
}

func (engine *SiteEntryEngine) GetEmployeeByPosition(ctx *kaos.Context, payload EmployeeReq) (float64, error) {
	if payload.Search == "" || payload.Prefix == "" {
		return 0, errors.New("missing: invalid request, please check your payload")
	}

	// hub := sebar.GetTenantDBFromContext(ctx)
	// if hub == nil {
	// 	return 0, errors.New("missing: connection")
	// }

	// trayek := new(bagongmodel.Trayek)
	// e := hub.GetByID(trayek, payload.TrayekID)
	// if e != nil {
	// 	return 0, e
	// }

	// rate, e := trayek.GetTariff(payload.From, payload.To)
	// if e != nil {
	// 	return 0, e
	// }

	return 0, nil
}

func (engine *SiteEntryEngine) GetTariff(ctx *kaos.Context, payload TariffReq) (float64, error) {
	if payload.From == "" || payload.To == "" {
		return 0, errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return 0, errors.New("missing: connection")
	}

	trayek := new(bagongmodel.Trayek)
	e := hub.GetByID(trayek, payload.TrayekID)
	if e != nil {
		return 0, e
	}

	rate, e := trayek.GetTariff(payload.From, payload.To)
	if e != nil {
		return 0, e
	}

	return rate, nil
}

func (engine *SiteEntryEngine) GetDetailLedgerJournalPost(ctx *kaos.Context, payload JournalPostReq) (interface{}, error) {
	if payload.SiteID == "" {
		return 0, errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return 0, errors.New("missing: connection")
	}

	mapSiteJournal := new(tenantcoremodel.SiteEntryJournalType)
	e := hub.GetByFilter(mapSiteJournal, dbflex.And(dbflex.Eq("SiteID", payload.SiteID), dbflex.Eq("Type", payload.Type)))
	if e != nil {
		return 0, e
	}

	if payload.Type == "Revenue" {
		siteJournal := new(ficomodel.CustomerJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	} else if payload.Type == "Expense" {
		siteJournal := new(ficomodel.LedgerJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	} else if payload.Type == "Employee Expense" {
		siteJournal := new(ficomodel.VendorJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	} else {
		siteJournal := new(ficomodel.LedgerJournalType)
		e = hub.GetByID(siteJournal, mapSiteJournal.JournalTypeID)
		if e != nil {
			return 0, e
		}
		return siteJournal, nil
	}
}

func GenerateSummary(ctx *kaos.Context, siteEntryAssetID string) error {
	hub := sebar.GetTenantDBFromContext(ctx)

	siteEntryAsset := new(bagongmodel.SiteEntryAsset)
	e := hub.GetByID(siteEntryAsset, siteEntryAssetID)
	if e != nil {
		return errors.New(fmt.Sprintf("siteEntry not found: %s", siteEntryAssetID))
	}

	siteEntry := new(bagongmodel.SiteEntry)
	e = hub.GetByID(siteEntry, siteEntryAsset.SiteEntryID)
	if e != nil {
		return errors.New(fmt.Sprintf("siteEntry not found: %s", siteEntryAssetID))
	}

	if siteEntry.Purpose == "Trayek" {
		trayekRitase := []bagongmodel.SiteEntryTrayekRitase{}
		if e = hub.GetsByFilter(new(bagongmodel.SiteEntryTrayekRitase), dbflex.Eq("SiteEntryAssetID", siteEntryAssetID), &trayekRitase); e != nil {
			return errors.New(fmt.Sprintf("SiteEntryTrayekRitase not found: %s", siteEntryAssetID))
		}

		var income, expense, revenue, bonus float64

		for _, ritase := range trayekRitase {
			income += ritase.Income
			revenue += ritase.Revenue
			expense += ritase.Expense
			bonus += ritase.RitaseSummary.TotalBonus
		}

		if e := SaveSumAsset(hub, siteEntryAssetID, income, expense, revenue, bonus, "", ""); e != nil {
			return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
		}
	} else if siteEntry.Purpose == "Tourism" {
		siteEntryDetail := new(bagongmodel.SiteEntryTourismDetail)
		e = hub.GetByID(siteEntryDetail, siteEntryAsset.ID)
		if e != nil {
			return errors.New(fmt.Sprintf("siteEntry not found: %s", siteEntryAssetID))
		}

		income := siteEntryDetail.Rate
		expense := 0.0

		if len(siteEntryDetail.OperationalExpense) > 0 {
			for _, val := range siteEntryDetail.OperationalExpense {
				expense += val.TotalAmount
			}
		}

		if len(siteEntryDetail.OtherExpense) > 0 {
			for _, val := range siteEntryDetail.OtherExpense {
				expense += val.TotalAmount
			}
		}

		if e := SaveSumAsset(hub, siteEntryDetail.ID, income, expense, 0, 0, siteEntryDetail.Status, siteEntryDetail.SpareAsset); e != nil {
			return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
		}
	} else if siteEntry.Purpose == "BTS" {
		siteEntryDetail := new(bagongmodel.SiteEntryBTSDetail)
		e = hub.GetByID(siteEntryDetail, siteEntryAsset.ID)
		if e != nil {
			return errors.New(fmt.Sprintf("siteEntry not found: %s", siteEntryAssetID))
		}

		income := 0.0
		expense := 0.0

		for _, val := range siteEntryDetail.Expense {
			expense += val.TotalAmount
		}

		if e := SaveSumAsset(hub, siteEntryDetail.ID, income, expense, 0, 0, siteEntryDetail.Status, siteEntryDetail.SpareAsset); e != nil {
			return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
		}
	} else if siteEntry.Purpose == "Mining" {
		siteEntryDetail := new(bagongmodel.SiteEntryMiningDetail)
		e = hub.GetByID(siteEntryDetail, siteEntryAsset.ID)
		if e != nil {
			return errors.New(fmt.Sprintf("siteEntry not found: %s", siteEntryAssetID))
		}

		income := siteEntryDetail.RateRental + siteEntryDetail.RateBreakdown + siteEntryDetail.RateStandby + siteEntryDetail.RateOvertime
		expense := 0.0

		for _, val := range siteEntryDetail.Expense {
			expense += val.TotalAmount
		}

		if e := SaveSumAsset(hub, siteEntryDetail.ID, income, expense, 0, 0, siteEntryDetail.Status, siteEntryDetail.SpareAsset); e != nil {
			return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
		}

		return nil
	}

	return nil
}

func SumBTS(hub *datahub.Hub, parm *bagongmodel.SiteEntryBTSDetail) error {
	income := 0.0
	expense := 0.0

	for _, val := range parm.Expense {
		expense += val.TotalAmount
	}

	if e := SaveSumAsset(hub, parm.ID, income, expense, 0, 0, parm.Status, parm.SpareAsset); e != nil {
		return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
	}

	return nil
}

func SumMining(hub *datahub.Hub, parm *bagongmodel.SiteEntryMiningDetail) error {
	income := parm.RateRental + parm.RateBreakdown + parm.RateStandby + parm.RateOvertime
	expense := 0.0

	for _, val := range parm.Expense {
		expense += val.TotalAmount
	}

	if e := SaveSumAsset(hub, parm.ID, income, expense, 0, 0, parm.Status, parm.SpareAsset); e != nil {
		return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
	}

	return nil
}

func SumTrayek(hub *datahub.Hub, parm *bagongmodel.SiteEntryTrayekDetail) error {

	trayekRitase := []bagongmodel.SiteEntryTrayekRitase{}
	if e := hub.GetsByFilter(new(bagongmodel.SiteEntryTrayekRitase), dbflex.Eq("SiteEntryAssetID", parm.ID), &trayekRitase); e != nil {
		return errors.New(fmt.Sprintf("SiteEntryTrayekRitase not found: %s", parm.ID))
	}

	var income, expense, revenue, bonus float64

	for _, ritase := range trayekRitase {
		income += ritase.Income
		revenue += ritase.Revenue
		expense += ritase.Expense
		bonus += ritase.RitaseSummary.TotalBonus
	}

	if e := SaveSumAsset(hub, parm.ID, income, expense, revenue, bonus, parm.Status, parm.SpareAsset); e != nil {
		return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
	}

	return nil
}

func SumTourism(hub *datahub.Hub, parm *bagongmodel.SiteEntryTourismDetail) error {
	income := parm.Rate
	expense := 0.0

	for _, val := range parm.OperationalExpense {
		expense += val.TotalAmount
	}

	for _, val := range parm.OtherExpense {
		expense += val.TotalAmount
	}

	if e := SaveSumAsset(hub, parm.ID, income, expense, 0, 0, parm.Status, parm.SpareAsset); e != nil {
		return errors.New(fmt.Sprintf("save summary asset error: %s.", e.Error()))
	}

	return nil
}

func SaveSumAsset(hub *datahub.Hub, id string, income, expense, revenue, bonus float64, status, spareAsset string) error {
	// get SiteEntryAsset
	siteEntryAsset := new(bagongmodel.SiteEntryAsset)
	if e := hub.GetByID(siteEntryAsset, id); e != nil {
		return errors.New(fmt.Sprintf("SiteEntryAsset not found: %s", id))
	}

	siteEntry := new(bagongmodel.SiteEntry)
	if e := hub.GetByID(siteEntry, siteEntryAsset.SiteEntryID); e != nil {
		return errors.New(fmt.Sprintf("SiteEntry not found: %s", siteEntryAsset.SiteEntryID))
	}

	// calculate running, standby, and breakdown
	vRunning, vStandby, vBreakDown := 0, 0, 0
	switch status {
	case "Running":
		vRunning++
	case "Standby":
		vStandby++
	case "Standby Ready":
		vRunning++
	case "Partial":
		if spareAsset != "" {
			vRunning++
		} else {
			vBreakDown++
		}
	case "Breakdown":
		if spareAsset != "" {
			vRunning++
		} else {
			vBreakDown++
		}
	case "Breakdown No Driver":
		vBreakDown++
	}

	siteEntryAsset.Running = vRunning
	siteEntryAsset.Standby = vStandby
	siteEntryAsset.Breakdown = vBreakDown
	siteEntryAsset.Income = income
	siteEntryAsset.Expense = expense

	if siteEntry.Purpose == "Trayek" {
		siteEntryAsset.Revenue = revenue
		siteEntryAsset.Bonus = bonus
	} else {
		siteEntryAsset.Revenue = 0
		siteEntryAsset.Bonus = 0
	}

	if e := hub.Save(siteEntryAsset); e != nil {
		return errors.New(fmt.Sprintf("save SiteEntryBTSDetail error: %s.", e.Error()))
	}

	// gets SiteEntryAsset
	siteEntryAssets := []bagongmodel.SiteEntryAsset{}
	if e := hub.GetsByFilter(new(bagongmodel.SiteEntryAsset), dbflex.Eq("SiteEntryID", siteEntryAsset.SiteEntryID), &siteEntryAssets); e != nil {
		return errors.New(fmt.Sprintf("SiteEntryAsset not found: %s", siteEntryAsset.SiteEntryID))
	}

	siteEntryNonAssets := new(bagongmodel.SiteEntryNonAsset)
	e := hub.GetByID(siteEntryNonAssets, siteEntry.ID)
	if e != nil {
		return errors.New(fmt.Sprintf("SiteEntryNonAsset not found: %s", siteEntryAsset.SiteEntryID))
	}

	siteIncome := 0.0
	siteExpense := 0.0
	siteBonus := 0.0
	siteRunning, siteStandby, siteBreakdown := 0, 0, 0
	for _, val := range siteEntryAssets {
		siteIncome += val.Income
		siteExpense += val.Expense
		siteRunning += val.Running
		siteStandby += val.Standby
		siteBreakdown += val.Breakdown
		siteBonus += val.Bonus
	}

	// get SiteEntry
	siteEntry.Running = siteRunning
	siteEntry.Standby = siteStandby
	siteEntry.Breakdown = siteBreakdown
	siteEntry.Income = siteIncome + siteEntryNonAssets.Income
	siteEntry.Expense = siteExpense + siteEntryNonAssets.Expense
	if siteEntry.Purpose == "Trayek" {
		siteEntry.Revenue = (siteIncome - siteExpense - siteBonus)
		siteEntry.Expense = siteEntry.Expense + siteBonus
	} else {
		siteEntry.Revenue = 0
	}

	if e := hub.Save(siteEntry); e != nil {
		return errors.New(fmt.Sprintf("save SiteEntry error: %s.", e.Error()))
	}

	return nil
}
