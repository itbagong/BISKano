package bagonglogic

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetEngine struct{}

// GetsFilter gets tenantcoremodel.Asset data by filtering from bagong Asset fields
func (o *AssetEngine) GetsFilter(ctx *kaos.Context, _ *interface{}) (*[]tenantcoremodel.Asset, error) {
	p := GetURLQueryParams(ctx)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// filtering by tenantcore Asset fields
	filters := []*dbflex.Filter{}
	if p["GroupID"] != "" {
		filters = append(filters, dbflex.Eq("GroupID", p["GroupID"]))
	}

	filter := new(dbflex.Filter)
	if len(filters) > 0 {
		filter = dbflex.And(filters...)
	}

	res := []tenantcoremodel.Asset{}
	if e := h.GetsByFilter(new(tenantcoremodel.Asset), filter, &res); e != nil {
		return nil, e
	}

	// filtering by bagong Asset field if any
	bgaFilters := []*dbflex.Filter{}

	if p["HasAcquisitionDate"] != "" {
		if p["HasAcquisitionDate"] == "true" {
			bgaFilters = append(bgaFilters, dbflex.Gt("Depreciation.AcquisitionDate", time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)))
		} else {
			bgaFilters = append(bgaFilters, dbflex.Eq("Depreciation.AcquisitionDate", nil))
		}
	}

	// another bagong Asset field filtering process

	if len(bgaFilters) > 0 && len(res) > 0 {
		ids := lo.Map(res, func(item tenantcoremodel.Asset, index int) interface{} {
			return item.ID
		})
		bgaFilters = append(bgaFilters, dbflex.In("_id", ids...))

		bgass := []bagongmodel.Asset{}
		if e := h.GetsByFilter(new(bagongmodel.Asset), dbflex.And(bgaFilters...), &bgass); e != nil {
			return nil, e
		}

		bgassIDs := lo.Map(bgass, func(item bagongmodel.Asset, index int) string {
			return item.ID
		})

		res = lo.Filter(res, func(item tenantcoremodel.Asset, index int) bool {
			return lo.Contains(bgassIDs, item.ID)
		})
	}

	return &res, nil
}

func (o *AssetEngine) GetAssets(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.Where != nil {
		filters := make([]*dbflex.Filter, len(payload.Where.Items))
		for i, item := range payload.Where.Items {
			switch item.Field {
			case "AssetType":
				assets := []tenantcoremodel.Asset{}
				err := h.Gets(new(tenantcoremodel.Asset), dbflex.NewQueryParam().SetWhere(
					dbflex.In("AssetType", item.Value.([]interface{})...),
				), &assets)
				if err != nil {
					return nil, fmt.Errorf("error when get asset: %s", err.Error())
				}

				ids := lo.Map(assets, func(m tenantcoremodel.Asset, index int) interface{} {
					return m.ID
				})

				filters[i] = dbflex.In("_id", ids...)
			default:
				// handle Hull no & Police No
				if len(item.Items) > 0 && item.Op == dbflex.OpOr {
					assets := []bagongmodel.Asset{}
					err := h.Gets(new(bagongmodel.Asset), dbflex.NewQueryParam().SetWhere(
						dbflex.Or(
							dbflex.Contains("DetailUnit.PoliceNum", item.Items[0].Value.([]interface{})[0].(string)),
							dbflex.Contains("DetailUnit.HullNum", item.Items[0].Value.([]interface{})[0].(string)),
						),
					), &assets)
					if err != nil {
						return nil, fmt.Errorf("error when get asset: %s", err.Error())
					}

					ids := lo.Map(assets, func(m bagongmodel.Asset, index int) interface{} {
						return m.ID
					})

					filters[i] = dbflex.In("_id", ids...)
				} else {
					filters[i] = item
				}
			}
		}
		payload.Where = dbflex.And(filters...)
	}

	assets := []bagongmodel.Asset{}
	err := h.Gets(new(bagongmodel.Asset), payload, &assets)
	if err != nil {
		return nil, fmt.Errorf("error when get asset: %s", err.Error())
	}

	var count int
	count, err = h.Count(new(bagongmodel.Asset), payload)
	if err != nil {
		return nil, fmt.Errorf("error when get count asset: %s", err.Error())
	}

	return codekit.M{"data": assets, "count": count}, nil
}

type GenerateDepreciationRequest struct {
	AcquisitionCost    float64
	AssetDuration      float64
	DepreciationDate   *time.Time
	DepreciationPeriod string
	ResidualAmount     float64
}

func (o *AssetEngine) GenerateDepreciation(ctx *kaos.Context, payload *GenerateDepreciationRequest) ([]bagongmodel.DepreciationActivity, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.AcquisitionCost == 0 || payload.AssetDuration == 0 ||
		payload.DepreciationDate == nil {
		return nil, fmt.Errorf("please check Acquisition Cost, Asset Duration, Residual Amount, Depreciation Date field")
	}

	period := new(tenantcoremodel.MasterData)
	err := h.GetByID(period, payload.DepreciationPeriod)
	if err != nil {
		return nil, fmt.Errorf("error when get master depreciation period: %s", err.Error())
	}

	deprication := (payload.AcquisitionCost - payload.ResidualAmount) / payload.AssetDuration
	depDate := *payload.DepreciationDate
	details := make([]bagongmodel.DepreciationActivity, int(payload.AssetDuration))
	for i := 0; i < int(payload.AssetDuration); i++ {
		dep := bagongmodel.DepreciationActivity{}
		dep.ID = primitive.NewObjectID().Hex()
		dep.Date = depDate
		dep.Activity = "Depreciation"
		dep.DepreciationAmount = deprication
		dep.NetBookValue = payload.AcquisitionCost - (float64(i+1) * deprication)

		if period.Name == "Monthly" {
			depDate = depDate.AddDate(0, 1, 0)
		} else {
			depDate = depDate.AddDate(1, 0, 0)
		}

		details[i] = dep
	}

	return details, nil
}

type AssetAcquireRequest struct {
	Assets          []AssetData
	AcquisitionDate time.Time
}

type AssetData struct {
	ID      string
	GroupID string
	Name    string
}

func (o *AssetEngine) Acquire(ctx *kaos.Context, payload *AssetAcquireRequest) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	if len(payload.Assets) == 0 {
		return res, fmt.Errorf("no IDs provided")
	}

	if payload.AcquisitionDate.IsZero() {
		return res, fmt.Errorf("no AcquisitionDate provided")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: connection")
	}

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	assetIDs := []string{}
	paramAssetM := map[string]AssetData{}
	for _, a := range payload.Assets {
		assetIDs = append(assetIDs, a.ID)
		paramAssetM[a.ID] = a
	}

	assets := []tenantcoremodel.Asset{}
	if err = db.GetsByFilter(new(tenantcoremodel.Asset), dbflex.In("_id", assetIDs...), &assets); err != nil {
		return res, err
	}

	assetM := lo.SliceToMap(assets, func(d tenantcoremodel.Asset) (string, tenantcoremodel.Asset) {
		return d.ID, d
	})

	bgAssets := []bagongmodel.Asset{}
	if err = db.GetsByFilter(new(bagongmodel.Asset), dbflex.In("_id", assetIDs...), &bgAssets); err != nil {
		return res, err
	}

	bgAssetM := lo.SliceToMap(bgAssets, func(d bagongmodel.Asset) (string, bagongmodel.Asset) {
		return d.ID, d
	})

	for pAssetID, assData := range paramAssetM {
		if ass, existAss := assetM[pAssetID]; existAss {
			// update existing Assets
			ass.Name = assData.Name
			if err = db.Save(&ass); err != nil {
				return res, err
			}

			// update existing BGAssets
			if bgass, existBgass := bgAssetM[pAssetID]; existBgass {
				bgass.Depreciation.AcquisitionDate = &payload.AcquisitionDate

				ref, idx, existRef := lo.FindIndexOf(bgass.References, func(item codekit.M) bool {
					return item.Has("Acquisition date")
				})
				if existRef {
					ref.Set("Key", "Acquisition date").
						Set("Value", payload.AcquisitionDate.Format(time.RFC3339Nano))
					bgass.References[idx] = ref
				} else {
					bgass.References = append(bgass.References, codekit.M{"Key": "Acquisition date", "Value": payload.AcquisitionDate.Format(time.RFC3339Nano)})
				}

				if err = db.Save(&bgass); err != nil {
					return res, err
				}
			}
		} else {
			// create new asset
			if err = db.Save(&tenantcoremodel.Asset{
				ID:      pAssetID,
				GroupID: assData.GroupID,
				Name:    assData.Name,
			}); err != nil {
				return res, err
			}

			if err = db.Save(&bagongmodel.Asset{
				ID: pAssetID,
				Depreciation: bagongmodel.Depreciation{
					AcquisitionDate: &payload.AcquisitionDate,
				},
				References: []codekit.M{{"Key": "Acquisition date", "Value": payload.AcquisitionDate.Format(time.RFC3339Nano)}},
			}); err != nil {
				return res, err
			}
		}
	}

	res = "ok"
	return res, nil
}

type GetDepreciationRequest struct {
	SiteID        string
	Period        string
	JournalTypeID string
}

func (engine *AssetEngine) GetAssetDetail(ctx *kaos.Context, payload codekit.M) (interface{}, error) {
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
	// get bagong Asset
	bagongAsset := new(bagongmodel.Asset)
	if e := hub.GetByID(bagongAsset, siteEntryAsset.AssetID); e != nil {
		return "", errors.New(fmt.Sprintf("Asset not found: %s", siteEntryAsset.AssetID))
	}

	tenantCustomer := new(tenantcoremodel.Customer)
	if len(bagongAsset.UserInfo) > 0 {
		sort.Slice(bagongAsset.UserInfo, func(i, j int) bool {
			return bagongAsset.UserInfo[i].AssetDateTo.After(bagongAsset.UserInfo[j].AssetDateTo)
		})
		// get tenantcore Customer
		if e := hub.GetByID(tenantCustomer, bagongAsset.UserInfo[0].CustomerID); e != nil {
			tenantCustomer.Name = ""
			// return "", errors.New(fmt.Sprintf("Customer not found: %s", bagongAsset.DetailUnit.OtherInfo.CustomerID))
		}
	}

	if siteEntry.Purpose == "BTS" {
		siteEntryDetail := new(bagongmodel.SiteEntryBTSDetail)
		if e := hub.GetByID(siteEntryDetail, siteEntryAsset.ID); e != nil {
			siteEntryDetail.ID = siteEntryAsset.ID
			siteEntryDetail.RitaseDetail = []bagongmodel.RitaseDetail{}
			siteEntryDetail.RitaseFuelUsage = []bagongmodel.RitaseFuelUsage{}
			siteEntryDetail.Expense = []bagongmodel.SiteExpense{}
			if e := hub.Insert(siteEntryDetail); e != nil {
				return "", errors.New(fmt.Sprintf("create SiteEntryBTSDetail error: %s.", e.Error()))
			}
		}
		//collect ids from response
		vendorJournalIDs := []interface{}{}
		for _, c := range siteEntryDetail.Expense {
			vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
		}

		//get list of vendor account by list of ids
		vendorJournal := []ficomodel.VendorJournal{}
		e := hub.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
		if e != nil {
			ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
		}

		//convert list vendor journal to map[string]VendorJournal
		mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
			return detail.ID, detail
		})

		// get list of vendor trx by list of ids
		cashSchedule := []ficomodel.CashSchedule{}
		e = hub.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
		if e != nil {
			ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
		}

		// join with vendor journal to get status and voucher no
		siteExpenses := []bagongmodel.SiteExpense{}

		for _, d := range siteEntryDetail.Expense {
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

		siteEntryDetail.Expense = siteExpenses

		result := bagongmodel.SiteEntryBTSDetailAsset{
			SiteEntryBTSDetail: *siteEntryDetail,
			SumRevenue:         siteEntryAsset.Revenue,
			SumIncome:          siteEntryAsset.Income,
			SumExpense:         siteEntryAsset.Expense,
			TrxDate:            siteEntry.TrxDate,
			PoliceNum:          bagongAsset.DetailUnit.PoliceNum,
			CustomerName:       tenantCustomer.Name,
		}

		return result, nil
	}

	if siteEntry.Purpose == "Mining" {
		siteEntryDetail := new(bagongmodel.SiteEntryMiningDetail)
		if e := hub.GetByID(siteEntryDetail, siteEntryAsset.ID); e != nil {
			siteEntryDetail.ID = siteEntryAsset.ID
			siteEntryDetail.Expense = []bagongmodel.SiteExpense{}
			siteEntryDetail.Attachment = []bagongmodel.SiteAttachment{}
			if e := hub.Insert(siteEntryDetail); e != nil {
				return "", errors.New(fmt.Sprintf("create SiteEntryMiningDetail error: %s.", e.Error()))
			}
		}

		siteEntryUsage := new(bagongmodel.SiteEntryMiningUsage)
		if e := hub.GetByID(siteEntryUsage, siteEntryAsset.ID); e != nil {
			siteEntryUsage.ID = siteEntryAsset.ID
			if e := hub.Insert(siteEntryUsage); e != nil {
				return "", errors.New(fmt.Sprintf("create SiteEntryMiningUsage error: %s.", e.Error()))
			}
		}

		//collect ids from response
		vendorJournalIDs := []interface{}{}
		for _, c := range siteEntryDetail.Expense {
			vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
		}

		//get list of vendor account by list of ids
		vendorJournal := []ficomodel.VendorJournal{}
		e := hub.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
		if e != nil {
			ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
		}

		//convert list vendor journal to map[string]VendorJournal
		mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
			return detail.ID, detail
		})

		// get list of vendor trx by list of ids
		cashSchedule := []ficomodel.CashSchedule{}
		e = hub.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
		if e != nil {
			ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
		}

		// join with vendor journal to get status and voucher no
		siteExpenses := []bagongmodel.SiteExpense{}

		for _, d := range siteEntryDetail.Expense {
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

		siteEntryDetail.Expense = siteExpenses

		result := bagongmodel.SiteEntryMiningDetailAsset{
			SiteEntryMiningDetail: *siteEntryDetail,
			SumRevenue:            siteEntryAsset.Revenue,
			SumIncome:             siteEntryAsset.Income,
			SumExpense:            siteEntryAsset.Expense,
			TrxDate:               siteEntry.TrxDate,
			PoliceNum:             bagongAsset.DetailUnit.PoliceNum,
			CustomerName:          tenantCustomer.Name,
		}

		return result, nil
	}

	if siteEntry.Purpose == "Trayek" {
		siteEntryDetail := new(bagongmodel.SiteEntryTrayekDetail)
		if e := hub.GetByID(siteEntryDetail, siteEntryAsset.ID); e != nil {
			siteEntryDetail.ID = siteEntryAsset.ID
			if e := hub.Insert(siteEntryDetail); e != nil {
				return "", errors.New(fmt.Sprintf("create SiteEntryTrayekDetail error: %s.", e.Error()))
			}
		}
		result := bagongmodel.SiteEntryTrayekDetailAsset{
			SiteEntryTrayekDetail: *siteEntryDetail,
			SumRevenue:            siteEntryAsset.Revenue,
			SumIncome:             siteEntryAsset.Income,
			SumExpense:            siteEntryAsset.Expense,
			TrxDate:               siteEntry.TrxDate,
			PoliceNum:             bagongAsset.DetailUnit.PoliceNum,
			CustomerName:          tenantCustomer.Name,
		}

		return result, nil
	}

	if siteEntry.Purpose == "Tourism" {
		siteEntryDetail := new(bagongmodel.SiteEntryTourismDetail)
		if e := hub.GetByID(siteEntryDetail, siteEntryAsset.ID); e != nil {
			siteEntryDetail.ID = siteEntryAsset.ID
			siteEntryDetail.OperationalExpense = []bagongmodel.SiteExpense{}
			siteEntryDetail.OtherExpense = []bagongmodel.SiteExpense{}
			if e := hub.Insert(siteEntryDetail); e != nil {
				return "", errors.New(fmt.Sprintf("create SiteEntryTourismDetail error: %s.", e.Error()))
			}
		}

		//collect ids from response
		vendorJournalIDs := []interface{}{}
		for _, c := range siteEntryDetail.OperationalExpense {
			vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
		}
		for _, c := range siteEntryDetail.OtherExpense {
			vendorJournalIDs = append(vendorJournalIDs, c.JournalID)
		}

		//get list of vendor account by list of ids
		vendorJournal := []ficomodel.VendorJournal{}
		e := hub.GetsByFilter(new(ficomodel.VendorJournal), dbflex.In("_id", vendorJournalIDs...), &vendorJournal)
		if e != nil {
			ctx.Log().Errorf("Failed populate data vendorJournal: %s", e.Error())
		}

		//convert list vendor journal to map[string]VendorJournal
		mapVendorJournal := lo.Associate(vendorJournal, func(detail ficomodel.VendorJournal) (string, ficomodel.VendorJournal) {
			return detail.ID, detail
		})

		// get list of vendor trx by list of ids
		cashSchedule := []ficomodel.CashSchedule{}
		e = hub.GetsByFilter(new(ficomodel.CashSchedule), dbflex.In("SourceJournalID", vendorJournalIDs...), &cashSchedule)
		if e != nil {
			ctx.Log().Errorf("Failed populate data cashSchedule: %s", e.Error())
		}

		// join with vendor journal to get status and voucher no
		opExpenses := []bagongmodel.SiteExpense{}
		otExpenses := []bagongmodel.SiteExpense{}

		for _, d := range siteEntryDetail.OperationalExpense {
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
		for _, d := range siteEntryDetail.OtherExpense {
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

		siteEntryDetail.OperationalExpense = opExpenses
		siteEntryDetail.OtherExpense = otExpenses

		result := bagongmodel.SiteEntryTourismDetailAsset{
			SiteEntryTourismDetail: *siteEntryDetail,
			SumRevenue:             siteEntryAsset.Revenue,
			SumIncome:              siteEntryAsset.Income,
			SumExpense:             siteEntryAsset.Expense,
			TrxDate:                siteEntry.TrxDate,
			PoliceNum:              bagongAsset.DetailUnit.PoliceNum,
			CustomerName:           tenantCustomer.Name,
		}

		return result, nil
	}

	return "", errors.New(fmt.Sprintf("Purpose not found: %s", siteEntry.Purpose))
}

// GetDepreciation take current month depreciation and show them in ledger journal line
func (m *AssetEngine) GetDepreciation(ctx *kaos.Context, payload *GetDepreciationRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	journalType := new(ficomodel.LedgerJournalType)
	err := h.GetByID(journalType, payload.JournalTypeID)
	if err != nil {
		return nil, fmt.Errorf("error when get journal type: %s", err.Error())
	}

	tenantAssets := []tenantcoremodel.Asset{}
	// err = h.Gets(new(tenantcoremodel.Asset), dbflex.NewQueryParam().SetWhere(
	// 	dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", payload.SiteID)),
	// ), &tenantAssets)
	err = h.Gets(new(tenantcoremodel.Asset), nil, &tenantAssets)
	if err != nil {
		return nil, fmt.Errorf("err when get tenant asset: %s", err.Error())
	}

	ids := lo.Map(tenantAssets, func(t tenantcoremodel.Asset, index int) string {
		return t.ID
	})

	mapDimensionAsset := lo.Associate(tenantAssets, func(ta tenantcoremodel.Asset) (string, tenantcoremodel.Dimension) {
		return ta.ID, ta.Dimension
	})

	start, err := time.Parse("2006-01", payload.Period)
	if err != nil {
		return nil, fmt.Errorf("error parsing period: %s", err.Error())
	}
	end := start.AddDate(0, 1, 0)
	assets := []bagongmodel.Asset{}
	err = h.Gets(new(bagongmodel.Asset), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", ids...),
			dbflex.ElemMatch("DepreciationActivity", dbflex.Gte("Date", start), dbflex.Lt("Date", end)),
		),
	), &assets)
	if err != nil {
		return nil, fmt.Errorf("err when get asset: %s", err.Error())
	}

	type DepreciationLine struct {
		ficomodel.JournalLine
		Debit  float64
		Credit float64
	}

	lines := []DepreciationLine{}
	i := 0
	for _, a := range assets {
		for _, d := range a.DepreciationActivity {
			if (d.Date.After(start) || d.Date.Equal(start)) && d.Date.Before(end) {
				references := tenantcoremodel.References{}.Set("DepreciationID", d.ID)
				journalLine := ficomodel.JournalLine{
					LineNo:        i,
					Qty:           1,
					TagObjectID1:  ficomodel.NewSubAccount("ASSET", a.ID),
					CurrencyID:    "IDR",
					Amount:        d.DepreciationAmount,
					PriceEach:     d.DepreciationAmount,
					Text:          d.Activity,
					UnitID:        "Each",
					References:    references,
					Dimension:     mapDimensionAsset[a.ID],
					OffsetAccount: journalType.DefaultOffset,
					Account:       ficomodel.NewSubAccount("ASSET", a.ID),
				}

				line := DepreciationLine{
					JournalLine: journalLine,
					Debit:       journalLine.Amount,
				}

				lines = append(lines, line)
				i++

				break
			}
		}
	}

	return lines, nil
}

type AddAssetRequest struct {
	ID          []string
	TenantAsset []tenantcoremodel.Asset
	IsActive    bool
}

func (o *AssetEngine) AddAsset(ctx *kaos.Context, payload *AddAssetRequest) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	if len(payload.ID) == 0 {
		return res, fmt.Errorf("no IDs provided")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return res, errors.New("missing: connection")
	}

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	// get asset group
	tenantAssetsGroup := []tenantcoremodel.AssetGroup{}
	mapAssetGroup := map[string]tenantcoremodel.AssetGroup{}
	err = h.Gets(new(tenantcoremodel.AssetGroup), nil, &tenantAssetsGroup)
	if err != nil {
		return "", fmt.Errorf("err when get tenant asset: %s", err.Error())
	}
	if len(tenantAssetsGroup) > 0 {
		for _, val := range tenantAssetsGroup {
			mapAssetGroup[val.ID] = val
		}
	}

	for _, val := range payload.TenantAsset {
		vAssetDuration := mapAssetGroup[val.GroupID].AssetDuration
		newAssets := bagongmodel.Asset{
			ID:        val.ID,
			Name:      val.Name,
			IsActive:  payload.IsActive,
			Dimension: val.Dimension,
			Depreciation: bagongmodel.Depreciation{
				AssetDuration: vAssetDuration,
			},
		}

		// create hull number when group id is unt
		if val.GroupID != "ELC" && val.GroupID != "PRT" {
			vHullNUmber := newAssets.ID[len(newAssets.ID)-5 : len(newAssets.ID)]
			newAssets.DetailUnit.HullNum = "BG-" + vHullNUmber
		}

		if err = db.Save(&newAssets); err != nil {
			return res, err
		}
	}

	res = "ok"
	return res, nil
}

type GetsAssetActivePerSiteRequest struct {
	AssetIDs []interface{}
}

type GetsAssetActivePerSiteResponse struct {
	AssetDateFrom time.Time
	AssetDateTo   time.Time
}

func (o *AssetEngine) GetsAssetActivePerSite(ctx *kaos.Context, payload *GetsAssetActivePerSiteRequest) (map[string]GetsAssetActivePerSiteResponse, error) {
	siteAssetActive := map[string]GetsAssetActivePerSiteResponse{}

	var (
		db  *datahub.Hub
		err error
	)

	if payload == nil || len(payload.AssetIDs) == 0 {
		return siteAssetActive, fmt.Errorf("missing payload")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return siteAssetActive, errors.New("missing: connection")
	}

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	assets := []bagongmodel.Asset{}

	db.GetsByFilter(new(bagongmodel.Asset), dbflex.In("_id", payload.AssetIDs...), &assets)
	lo.ForEach(assets, func(asset bagongmodel.Asset, index int) {
		lo.ForEach(asset.UserInfo, func(userInfo bagongmodel.UserInfo, index int) {
			key := fmt.Sprintf("%s|%s", asset.ID, userInfo.SiteID)
			if _, ok := siteAssetActive[key]; !ok {
				siteAssetActive[key] = GetsAssetActivePerSiteResponse{}
			}

			siteAssetActive[key] = GetsAssetActivePerSiteResponse{
				AssetDateFrom: userInfo.AssetDateFrom,
				AssetDateTo:   userInfo.AssetDateTo,
			}
		})
	})

	return siteAssetActive, nil
}
