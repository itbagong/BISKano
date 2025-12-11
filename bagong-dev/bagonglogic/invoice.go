package bagonglogic

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ekokurniadi/terbilang"
	"github.com/leekchan/accounting"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type TaxInvoiceRequest struct {
	Prefix string `json:"prefix"`
	From   int    `json:"from"`
	To     int    `json:"to"`
}

type InvoiceHandler struct {
}

type GetSalesOrderRequest struct {
	Where struct {
		SiteID     string
		CustomerID string
	}
	Skip int
	Take int
}

func (m *InvoiceHandler) GetSalesOrder(ctx *kaos.Context, payload *GetSalesOrderRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	filters := []*dbflex.Filter{
		dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", payload.Where.SiteID)),
		dbflex.Eq("CustomerID", payload.Where.CustomerID),
		dbflex.Eq("Status", ficomodel.JournalStatusPosted),
	}
	// get sales order
	orders := []sdpmodel.SalesOrder{}
	err := hub.Gets(new(sdpmodel.SalesOrder), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filters...),
	).SetSkip(payload.Skip).SetTake(payload.Take), &orders)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when sales order: %s", err.Error()))
	}

	// get count site entry
	count, err := hub.Count(new(sdpmodel.SalesOrder), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filters...),
	))
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when count sales order: %s", err.Error()))
	}

	return codekit.M{"count": count, "data": orders}, nil
}

type GeneralInvoiceRequest struct {
	JournalID    string
	SalesOrderID []string
}

func (m *InvoiceHandler) GenerateGeneralInvoice(ctx *kaos.Context, payload *GeneralInvoiceRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get customer journal
	journal := new(ficomodel.CustomerJournal)
	if e := hub.GetByID(journal, payload.JournalID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("customer journal not found: %s", payload.JournalID))
	}

	// get sales order
	orders := []sdpmodel.SalesOrder{}
	err := hub.Gets(new(sdpmodel.SalesOrder), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", payload.SalesOrderID...),
	), &orders)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when sales order: %s", err.Error()))
	}

	// get customer
	cust := new(tenantcoremodel.Customer)
	if err := hub.GetByID(cust, journal.CustomerID); err != nil {
		return nil, fmt.Errorf("error when get customer : %s", err.Error())
	}

	// get customer group
	custGroup := new(tenantcoremodel.CustomerGroup)
	if err := hub.GetByID(custGroup, cust.GroupID); err != nil {
		return nil, fmt.Errorf("error when get customer group: %s", err.Error())
	}

	// get tax setup
	taxSetups := []ficomodel.TaxSetup{}
	err = hub.Gets(new(ficomodel.TaxSetup), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("IsActive", true),
	), &taxSetups)
	if err != nil {
		return nil, fmt.Errorf("error when get tax setup: %s", err.Error())
	}

	mapTaxSetup := lo.Associate(taxSetups, func(tax ficomodel.TaxSetup) (string, ficomodel.TaxSetup) {
		return tax.ID, tax
	})

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := hub.GetByID(curr, journal.CurrencyID); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("currency not found: %s", journal.CurrencyID))
	}

	section := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       make([][]string, 0),
	}

	preview, _ := tenantcorelogic.GetPreviewBySource(hub, "CUSTOMER", payload.JournalID, "", journal.TransactionType)
	preview.Name = journal.TransactionType
	preview.PreviewReport = &tenantcoremodel.PreviewReport{}
	preview.PreviewReport.Header = codekit.M{}
	preview.PreviewReport.Sections = make([]tenantcoremodel.PreviewSection, 0)

	headerInfo := make([][]string, 4)
	headerInfo[0] = []string{"", "", "", "", journal.TrxDate.Format("2 January 2006"), ""}
	headerInfo[1] = []string{"", "", "", "", "To:", ""}
	headerInfo[2] = []string{"No", journal.ID, "", "", cust.Name, ""}
	headerInfo[3] = []string{"PO", journal.References.Get("PO No.", "").(string), "", "", journal.AddressAndTax.BillingAddress, ""}
	preview.PreviewReport.Header["Data"] = headerInfo

	mapTaxLine := map[string]float64{}
	mapAmountTaxHeader := map[string]float64{}
	section.Items = append(section.Items, []string{"Item", "Description", "Quantity", "UoM", "Unit Price", "Total"})
	lineNo := 0
	totalInvoice := float64(0)
	// reset lines
	journal.Lines = make([]ficomodel.JournalLine, 0)
	for _, o := range orders {
		for _, l := range o.Lines {
			totalInvoice += float64(l.Amount)
			quantity := codekit.ToString(int(l.Qty))
			unitPrice := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.UnitPrice, 0, ",", "."))
			amount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(l.Amount, 0, ",", "."))

			// calculate tax only if taxable is true
			if l.Taxable {
				for _, code := range l.TaxCodes {
					if v, ok := mapTaxSetup[code]; ok {
						if v.CalcMethod == ficomodel.TaxCalcLine {
							mapTaxLine[v.ID] += float64(l.Amount) * v.Rate
						} else {
							mapAmountTaxHeader[v.ID] += float64(l.Amount)
						}
					}
				}
			}
			section.Items = append(section.Items, []string{l.Asset, l.Description, quantity, "Each", unitPrice, amount})

			line := ficomodel.JournalLine{
				LineNo: lineNo,
				Account: ficomodel.SubledgerAccount{
					AccountType: "LEDGERACCOUNT",
					AccountID:   custGroup.Setting.MainBalanceAccount,
				},
				TagObjectID1: ficomodel.SubledgerAccount{
					AccountType: "ASSET",
					AccountID:   l.Asset,
				},
				Qty:          float64(l.Qty),
				UnitID:       "Each",
				PriceEach:    float64(l.UnitPrice),
				Amount:       float64(l.Amount),
				DiscountType: l.DiscountType,
				Discount:     float64(l.Discount),
				Text:         l.Description,
				Taxable:      true,
			}
			journal.Lines = append(journal.Lines, line)
			lineNo++
		}
	}

	grandTotal := totalInvoice
	taxes := make([]codekit.M, len(mapAmountTaxHeader)+len(mapTaxLine))
	i := 0
	for k, v := range mapTaxLine {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = fmt.Sprintf("%s%d", curr.Symbol2, int(v))
				if mapTaxSetup[k].IncludeInInvoice {
					grandTotal += v
				}
			} else {
				tax["Value"] = fmt.Sprintf("-%s%d", curr.Symbol2, int(v))
				if mapTaxSetup[k].IncludeInInvoice {
					grandTotal -= v
				}
			}
		}

		taxes[i] = tax
		i++
	}

	for k, v := range mapAmountTaxHeader {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			val := v * taxSetup.Rate
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = fmt.Sprintf("%s%d", curr.Symbol2, int(val))
				if mapTaxSetup[k].IncludeInInvoice {
					grandTotal += val
				}
			} else {
				tax["Value"] = fmt.Sprintf("-%s%d", curr.Symbol2, int(val))
				if mapTaxSetup[k].IncludeInInvoice {
					grandTotal -= val
				}
			}
		}

		taxes[i] = tax
		i++
	}

	section.Items = append(section.Items, []string{"", "", "", "", "TOTAL INVOICE", fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(totalInvoice, 0, ",", "."))})
	for _, t := range taxes {
		section.Items = append(section.Items, []string{"", "", "", "", t["Key"].(string), t["Value"].(string)})
	}
	section.Items = append(section.Items, []string{"", "", "", "", "GRAND TOTAL", fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(grandTotal, 0, ",", "."))})

	preview.PreviewReport.Sections = append(preview.PreviewReport.Sections, section)
	hub.Save(preview)

	err = hub.Save(journal)
	if err != nil {
		return nil, fmt.Errorf("error when save customer journal: %s", err.Error())
	}

	return journal, nil
}

type GetSiteEntryRequest struct {
	Where struct {
		SiteID     string
		Start      time.Time
		End        time.Time
		CustomerID string
		ProjectID  string
	}
	Take int
	Skip int
}

func (m *InvoiceHandler) GetSiteEntry(ctx *kaos.Context, payload *GetSiteEntryRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get site entry
	siteEntrys := []bagongmodel.SiteEntry{}
	err := hub.Gets(new(bagongmodel.SiteEntry), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("SiteID", payload.Where.SiteID),
			dbflex.Gte("TrxDate", payload.Where.Start),
			dbflex.Lte("TrxDate", payload.Where.End),
		),
	), &siteEntrys)
	if err != nil {
		return nil, fmt.Errorf("error when get site entry: %s", err.Error())
	}

	// get site entry id
	siteEntryIDs := lo.Map(siteEntrys, func(m bagongmodel.SiteEntry, index int) interface{} {
		return m.ID
	})

	// get site entry
	siteEntryAssets := []bagongmodel.SiteEntryAsset{}
	err = hub.Gets(new(bagongmodel.SiteEntryAsset), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("SiteEntryID", siteEntryIDs...),
			dbflex.Eq("CustomerID", payload.Where.CustomerID),
			dbflex.Eq("ProjectID", payload.Where.ProjectID),
		),
	), &siteEntryAssets)
	if err != nil {
		return nil, fmt.Errorf("error when get site entry assets: %s", err.Error())
	}

	result := []bagongmodel.SiteEntryAsset{}
	mapAsset := map[string]bool{}
	for _, asset := range siteEntryAssets {
		if _, ok := mapAsset[asset.AssetID]; !ok {
			result = append(result, asset)
			mapAsset[asset.AssetID] = true
		}
	}

	return codekit.M{"count": len(result), "data": result}, nil
}

type GenerateDetailMining struct {
	No                int
	PoliceNo          string
	HullNo            string
	VehicleType       string
	Price             float64
	StartUsePeroid    time.Time
	EndUsePeroid      time.Time
	StartPeriod       time.Time
	StandbyDiscount   float64
	StandbyDay        float64
	BreakdownDay      float64
	BreakdownDiscount float64
	TotalDiscount     float64
	Total             float64
	ProductionYear    int
	UnitCondition     string
	SONumber          string
}

func (m *InvoiceHandler) buildDetailMining(hub *datahub.Hub, payload *GenerateDetailMiningRequest) ([]*GenerateDetailMining, error) {
	// get site entry
	siteEntrys := []bagongmodel.SiteEntry{}
	err := hub.Gets(new(bagongmodel.SiteEntry), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("SiteID", payload.SiteID),
			dbflex.Gte("TrxDate", payload.Start),
			dbflex.Lte("TrxDate", payload.End),
		),
	), &siteEntrys)
	if err != nil {
		return nil, fmt.Errorf("error when get site entry: %s", err.Error())
	}

	// build site entry map
	mapSiteEntry := map[string]time.Time{}
	siteEntryIDs := make([]string, len(siteEntrys))
	for i, s := range siteEntrys {
		mapSiteEntry[s.ID] = s.TrxDate
		siteEntryIDs[i] = s.ID
	}

	// get site entry
	siteEntryAssets := []bagongmodel.SiteEntryAsset{}
	err = hub.Gets(new(bagongmodel.SiteEntryAsset), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("SiteEntryID", siteEntryIDs...),
			dbflex.In("AssetID", payload.AssetID...),
			dbflex.Eq("CustomerID", payload.CustomerID),
			dbflex.Eq("ProjectID", payload.ProjectID),
		),
	).SetSort("Created"), &siteEntryAssets)
	if err != nil {
		return nil, fmt.Errorf("error when get site entry assets: %s", err.Error())
	}

	assetIDs := make([]string, len(siteEntryAssets))
	siteEntryAssetIDs := make([]interface{}, len(siteEntryAssets))
	vehicleTypeIDs := []string{}
	for i, d := range siteEntryAssets {
		siteEntryAssetIDs[i] = d.ID
		assetIDs[i] = d.AssetID
		vehicleTypeIDs = append(vehicleTypeIDs, d.UnitType)
	}

	// get vehicle type
	vehicles := []tenantcoremodel.MasterData{}
	err = hub.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", vehicleTypeIDs...),
	), &vehicles)
	if err != nil {
		return nil, fmt.Errorf("error when get vehicle data: %s", err.Error())
	}

	mapVehicle := lo.Associate(vehicles, func(v tenantcoremodel.MasterData) (string, string) {
		return v.ID, v.Name
	})

	// get site entry mining detail
	details := []bagongmodel.SiteEntryMiningDetail{}
	err = hub.Gets(new(bagongmodel.SiteEntryMiningDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", siteEntryAssetIDs...),
	), &details)
	if err != nil {
		return nil, fmt.Errorf("error when get site entry mining detail: %s", err.Error())
	}

	mapMiningDetail := lo.Associate(details, func(detail bagongmodel.SiteEntryMiningDetail) (string, bagongmodel.SiteEntryMiningDetail) {
		return detail.ID, detail
	})

	// get asset
	assets := []bagongmodel.Asset{}
	err = hub.Gets(new(bagongmodel.Asset), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", assetIDs...),
	), &assets)
	if err != nil {
		return nil, fmt.Errorf("error when get asset: %s", err.Error())
	}

	salesOrderNo := []string{}
	mapAsset := map[string]bagongmodel.Asset{}
	for _, a := range assets {
		mapAsset[a.ID] = a

		for _, ui := range a.UserInfo {
			salesOrderNo = append(salesOrderNo, ui.SONumber)
		}
	}

	// get sales order
	salesOrders := []sdpmodel.SalesOrder{}
	err = hub.Gets(new(sdpmodel.SalesOrder), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", salesOrderNo...),
	), &salesOrders)
	if err != nil {
		return nil, fmt.Errorf("error when get sales order: %s", err.Error())
	}

	mapSalesOrder := lo.Associate(salesOrders, func(detail sdpmodel.SalesOrder) (string, sdpmodel.SalesOrder) {
		return detail.ID, detail
	})

	// get customer detail
	custDetail := new(bagongmodel.CustomerDetail)
	if err := hub.GetByParm(custDetail, dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("CustomerID", payload.Journal.CustomerID),
	)); err != nil {
		return nil, fmt.Errorf("err when get customer detail: %s", err.Error())
	}

	site := payload.Journal.Dimension.Get("Site")
	// get sales price book
	salesBooks := new(sdpmodel.SalesPriceBook)
	err = hub.GetByParm(salesBooks, dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("_id", custDetail.DefaultPriceBook...),
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", site)),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("err when get sales price book: %s", err.Error())
	}

	mapLineSales := map[string]float64{}
	for _, d := range salesBooks.Lines {
		mapLineSales[d.AssetID] = float64(d.SalesPrice)
	}

	// find so based on asset id
	findSOLine := func(lines []sdpmodel.SalesOrderLine, assetID string) (sdpmodel.SalesOrderLine, bool) {
		line := sdpmodel.SalesOrderLine{}

		for _, l := range lines {
			// match asset id
			if l.Asset == assetID {
				return l, true
			}
		}

		return line, false
	}

	type invoice struct {
		StandbyDay float64
		BDDay      float64
		Standby    float64
		Breakdown  float64
	}

	// calculate detail invoice
	calculateInvoice := func(siteEntry *bagongmodel.SiteEntryAsset, price float64) (invoice, string, time.Time) {
		detail := invoice{}
		var startPeriod time.Time
		soNumber := ""
		// mapping other info from asset
		if asset, ok := mapAsset[siteEntry.AssetID]; ok {
			for _, ui := range asset.UserInfo {
				// match customer id
				if ui.CustomerID == payload.Journal.CustomerID {
					if ui.SOStartDate != nil && ui.SOEndDate != nil {
						// check sales order date
						if (payload.Start.After(*ui.SOStartDate) || payload.Start.Equal(*ui.SOStartDate)) &&
							(payload.End.Before(*ui.SOEndDate) || payload.End.Equal(*ui.SOEndDate)) {
							soNumber = mapSalesOrder[ui.SONumber].SalesOrderNo
							// mapping site entry asset mining detail
							if mining, ok := mapMiningDetail[siteEntry.ID]; ok {
								// find sales order line
								soLine, isFound := findSOLine(mapSalesOrder[ui.SONumber].Lines, siteEntry.AssetID)
								if isFound {
									workingHour := float64(soLine.References.Get("Working Hour", 0).(int64))
									standbyRate := float64(soLine.References.Get("Standby Rate", 0).(int))
									divider := float64(0)
									if soLine.References.Get("Divider Type", "").(string) == "Auto" {
										start := float64(time.Date(payload.Journal.TrxDate.Year(), payload.Journal.TrxDate.Month()+1, 0, 0, 0, 0, 0, payload.Journal.TrxDate.Location()).Day())
										divider = start * workingHour
									} else {
										divider = float64(soLine.References.Get("Divider", 0).(int64))
									}

									if divider != 0 {
										if mining.Status == "Standby" {
											detail.Standby = workingHour / divider * price
											detail.StandbyDay = 1
										} else if mining.Status == "Standby Ready" {
											standby := float64(mining.StandbyHour)
											detail.Standby = standby / divider * price * standbyRate / 100
											detail.StandbyDay = standby / workingHour
										} else if (mining.Status == "Partial" || mining.Status == "Breakdown") && mining.SpareAsset == "" {
											bd := float64(mining.BreakdownHour)
											detail.BDDay = bd / workingHour
											detail.Breakdown = bd / divider * price
										} else if mining.Status == "Breakdown No Driver" {
											detail.BDDay = 1
											detail.Breakdown = workingHour / divider * price
										}
									}
								}
							}

							// set start period
							startPeriod = ui.AssetDateFrom
						}
					}
				}
			}
		}

		return detail, soNumber, startPeriod
	}

	// build map result
	mapResult := map[string]*GenerateDetailMining{}
	i := 1
	for _, d := range siteEntryAssets {
		if v, ok := mapResult[d.AssetID]; ok {
			v.EndUsePeroid = mapSiteEntry[d.SiteEntryID]

			detail, _, _ := calculateInvoice(&d, v.Price)
			v.Total = v.Total - detail.Standby - detail.Breakdown
			v.StandbyDiscount += detail.Standby
			v.StandbyDay += detail.StandbyDay
			v.BreakdownDay += detail.BDDay
			v.BreakdownDiscount += detail.Breakdown
			v.TotalDiscount += detail.Breakdown + detail.Standby
		} else {
			// set price
			price := float64(0)
			if v, ok := mapLineSales[d.AssetID]; ok {
				price = v
			} else {
				price = mapLineSales[""]
			}

			total := price
			detail, soNumber, startPeriod := calculateInvoice(&d, price)
			mapResult[d.AssetID] = &GenerateDetailMining{
				No:                i,
				PoliceNo:          d.PoliceNo,
				HullNo:            d.HullNo,
				VehicleType:       mapVehicle[d.UnitType],
				Price:             price,
				Total:             total - detail.Standby - detail.Breakdown,
				StandbyDiscount:   detail.Standby,
				StandbyDay:        detail.StandbyDay,
				BreakdownDay:      detail.BDDay,
				BreakdownDiscount: detail.Breakdown,
				TotalDiscount:     detail.Breakdown + detail.Standby,
				StartUsePeroid:    mapSiteEntry[d.SiteEntryID],
				EndUsePeroid:      mapSiteEntry[d.SiteEntryID],
				StartPeriod:       startPeriod,
				ProductionYear:    mapAsset[d.AssetID].DetailUnit.ProductionYear,
				UnitCondition:     mapAsset[d.AssetID].DetailUnit.UnitCondition,
				SONumber:          soNumber,
			}

			i++
		}
	}

	result := make([]*GenerateDetailMining, len(mapResult))
	i = 0
	for _, v := range mapResult {
		result[i] = v
		i++
	}

	return result, nil
}

type GenerateDetailMiningRequest struct {
	AssetID    []string
	SiteID     string
	CustomerID string
	ProjectID  string
	Start      time.Time
	End        time.Time
	Journal    *ficomodel.CustomerJournal
}

type GenerateDetailMiningResponse struct {
	Asset         string
	Text          string
	Quantity      int64
	UnitID        string
	PriceEach     float64
	Amount        float64
	OffsetAccount string
	Tax           string
	Dimension     string
}

func (m *InvoiceHandler) GenerateDetailMining(ctx *kaos.Context, payload *GenerateDetailMiningRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	detail, err := m.buildDetailMining(hub, payload)
	if err != nil {
		return "", err
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := hub.GetByID(curr, payload.Journal.CurrencyID); err != nil {
		return "", fmt.Errorf("currency not found")
	}

	// get customer configuration
	custConf := new(bagongmodel.CustomerConfiguration)
	if e := hub.GetByParm(custConf, dbflex.NewQueryParam().SetWhere(dbflex.Eq("CustomerID", payload.Journal.CustomerID))); e != nil && e != io.EOF {
		return nil, fmt.Errorf("error when get customer configuration: %s", e.Error())
	}

	// get & set customer name
	cust := new(tenantcoremodel.Customer)
	if err := hub.GetByID(cust, payload.Journal.CustomerID); err != nil {
		return "", fmt.Errorf("customer not found")
	}

	// reset lines
	payload.Journal.Lines = make([]ficomodel.JournalLine, len(detail))
	soNumber := ""
	for i, d := range detail {
		soNumber = d.SONumber
		usedPeriod := fmt.Sprintf("%d - %s", d.StartUsePeroid.Day(), d.EndUsePeroid.Format("02 Jan 2006"))
		// set references
		references := tenantcoremodel.References{}
		references = references.Set("Standby (Day)", d.StandbyDay)
		references = references.Set("Standby (Potongan)", d.StandbyDiscount)
		references = references.Set("Breakdown (Day)", d.BreakdownDay)
		references = references.Set("Breakdown (Potongan)", d.BreakdownDiscount)
		references = references.Set("UserPeriod", usedPeriod)
		references = references.Set("HullNo", d.HullNo)
		references = references.Set("VehicleType", d.VehicleType)
		references = references.Set("StartPeriod", d.StartPeriod.Format("02-Jan-06"))
		references = references.Set("ProductionYear", d.ProductionYear)
		references = references.Set("UnitCondition", d.UnitCondition)

		// build lines
		line := ficomodel.JournalLine{
			TagObjectID1: ficomodel.SubledgerAccount{
				AccountType: "ASSET",
				AccountID:   d.PoliceNo,
			},
			Qty:        1,
			UnitID:     "Each",
			PriceEach:  d.Price,
			Text:       d.PoliceNo,
			Amount:     d.Total,
			Taxable:    true,
			TaxCodes:   payload.Journal.TaxCodes,
			LineNo:     i,
			References: references,
		}

		payload.Journal.Lines[i] = line
	}
	payload.Journal.References = payload.Journal.References.Set("SO No.", soNumber)

	return payload.Journal, nil
}

func (m *InvoiceHandler) customerToStringConvert(amount float64, symbol string, custConf *bagongmodel.CustomerConfiguration) string {
	switch custConf.Round {
	case "Up":
		amount = codekit.ToFloat64(amount, custConf.Decimals, codekit.RoundingUp)
	case "Down":
		amount = codekit.ToFloat64(amount, custConf.Decimals, codekit.RoundingDown)
	default:
		amount = codekit.ToFloat64(amount, custConf.Decimals, codekit.RoundingAuto)
	}

	return fmt.Sprintf("%s%s", symbol, accounting.FormatNumber(amount, custConf.Decimals, ",", "."))
}

func (m *InvoiceHandler) taxConvert(amount float64, taxSetup *ficomodel.TaxSetup) float64 {
	val := float64(0)
	if taxSetup.Rounding == "Up" {
		val = codekit.ToFloat64(amount, taxSetup.Decimal, codekit.RoundingUp)
	} else if taxSetup.Rounding == "Down" {
		val = codekit.ToFloat64(amount, taxSetup.Decimal, codekit.RoundingDown)
	} else {
		val = codekit.ToFloat64(amount, taxSetup.Decimal, codekit.RoundingAuto)
	}
	return val
}

func (m *InvoiceHandler) taxToStringConvert(amount float64, symbol string, taxSetup *ficomodel.TaxSetup) string {
	if taxSetup.Rounding == "Up" {
		amount = codekit.ToFloat64(amount, taxSetup.Decimal, codekit.RoundingUp)
	} else if taxSetup.Rounding == "Down" {
		amount = codekit.ToFloat64(amount, taxSetup.Decimal, codekit.RoundingDown)
	} else {
		amount = codekit.ToFloat64(amount, taxSetup.Decimal, codekit.RoundingAuto)
	}

	return fmt.Sprintf("%s%s", symbol, accounting.FormatNumber(amount, taxSetup.Decimal, ",", "."))
}

func (m *InvoiceHandler) GenerateTaxInvoice(ctx *kaos.Context, payload *TaxInvoiceRequest) (interface{}, error) {
	var err error

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	db, _ := hub.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	for i := payload.From; i <= payload.To; i++ {
		tax := bagongmodel.TaxInvoice{
			ID:     payload.Prefix + fmt.Sprint(i),
			FPNo:   payload.Prefix + fmt.Sprint(i),
			Status: "Open",
		}
		db.Save(&tax)
	}

	return "Success", nil
}

func (m *InvoiceHandler) SyncMiningInvoice(ctx *kaos.Context, payload *ficomodel.CustomerJournal) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	err := hub.Save(payload)
	if err != nil {
		return nil, fmt.Errorf("error when save customer journal: %s", err.Error())
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := hub.GetByID(curr, payload.CurrencyID); err != nil {
		return nil, fmt.Errorf("currency not found")
	}

	// get customer configuration
	custConf := new(bagongmodel.CustomerConfiguration)
	if err = hub.GetByParm(custConf, dbflex.NewQueryParam().SetWhere(dbflex.Eq("CustomerID", payload.CustomerID))); err != nil && err != io.EOF {
		return nil, fmt.Errorf("error when get customer configuration: %s", err.Error())
	}

	// get & set customer name
	cust := new(tenantcoremodel.Customer)
	if err := hub.GetByID(cust, payload.CustomerID); err != nil {
		return nil, fmt.Errorf("customer not found")
	}

	// get payterm
	term := new(ficomodel.PaymentTerm)
	if err := hub.GetByID(term, payload.PaymentTermID); err != nil && err != io.EOF {
		return nil, fmt.Errorf("payterm not found")
	}

	// get tax setup
	taxSetups := []ficomodel.TaxSetup{}
	err = hub.Gets(new(ficomodel.TaxSetup), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("IsActive", true),
	), &taxSetups)
	if err != nil {
		return nil, fmt.Errorf("error when get tax setup: %s", err.Error())
	}
	mapTaxSetup := lo.Associate(taxSetups, func(tax ficomodel.TaxSetup) (string, ficomodel.TaxSetup) {
		return tax.ID, tax
	})

	preview, _ := tenantcorelogic.GetPreviewBySource(hub, "CUSTOMER", payload.ID, "", payload.TransactionType)
	preview.Name = payload.TransactionType
	preview.PreviewReport = &tenantcoremodel.PreviewReport{}
	preview.PreviewReport.Header = codekit.M{}
	preview.PreviewReport.Sections = make([]tenantcoremodel.PreviewSection, 0)

	var data [][]string
	// set general info
	data = append(data, []string{"No : ", payload.ID, "", "", "", payload.TrxDate.Format("2 January 2006")})
	data = append(data, []string{"Perihal : ", payload.Text, "", "", "", "To :"})
	data = append(data, []string{"PO : ", payload.References.Get("SO No.", "").(string), "", "", "", cust.Name})
	data = append(data, []string{"Termin : ", term.Name, "", "", "", payload.AddressAndTax.BillingAddress})
	data = append(data, []string{"Jatuh Tempo : ", payload.ExpectedDate.Format("2 January 2006"), "", "", "", ""})
	data = append(data, []string{"Faktur Pajak : ", payload.TaxInvoiceNo, "", "", "", ""})
	preview.PreviewReport.Header["Data"] = data

	previewDetail := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       make([][]string, 0),
	}

	// set header
	previewDetail.Items = append(previewDetail.Items, []string{"No.", "Nopol", "ID", "Tahun", "Jenis", "Kondisi", "Awal Jalan", "Harga",
		"Periode", "BD", "Potongan", "Total Harga"})

	mapTaxLine := map[string]float64{}
	mapAmountTaxHeader := map[string]float64{}
	totalInvoice := float64(0)
	for i, line := range payload.Lines {
		refStandbyDiscount, _ := line.References.Get("Standby (Potongan)", 0).(json.Number).Float64()
		// refStandbyDay, _ := line.References.Get("Standby (Day)", 0).(json.Number).Int64()
		refBreakdownDay, _ := line.References.Get("Breakdown (Day)", 0).(json.Number).Int64()
		refBreakdownDiscount, _ := line.References.Get("Breakdown (Potongan)", 0).(json.Number).Float64()
		productionYear, _ := line.References.Get("ProductionYear", 0).(json.Number).Int64()

		index := i + 1
		stringIndex := codekit.ToString(index)
		price := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(line.PriceEach, 0, ",", "."))
		// standbyDiscount := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(refStandbyDiscount, 0, ",", "."))
		// standbyDay := codekit.ToString(refStandbyDay)
		breakdownDay := codekit.ToString(refBreakdownDay)
		// breakdownDiscount := m.customerToStringConvert(refBreakdownDiscount, curr.Symbol2, custConf)
		totalDiscount := m.customerToStringConvert(refBreakdownDiscount+refStandbyDiscount, curr.Symbol2, custConf)
		total := line.PriceEach - refBreakdownDiscount - refStandbyDiscount
		strTotal := m.customerToStringConvert(line.PriceEach-refBreakdownDiscount-refStandbyDiscount, curr.Symbol2, custConf)

		usedPeriod := line.References.Get("UserPeriod", "").(string)
		hullNo := line.References.Get("HullNo", "").(string)
		vehicleType := line.References.Get("VehicleType", "").(string)
		startPeriod := line.References.Get("StartPeriod", "").(string)
		unitCondition := line.References.Get("UnitCondition", "").(string)

		previewDetail.Items = append(previewDetail.Items, []string{stringIndex, line.TagObjectID1.AccountID, hullNo,
			codekit.ToString(productionYear), vehicleType, codekit.ToString(unitCondition), startPeriod, price, usedPeriod,
			breakdownDay, totalDiscount, strTotal})

		totalInvoice += total
		for _, code := range payload.TaxCodes {
			if v, ok := mapTaxSetup[code]; ok {
				if v.CalcMethod == ficomodel.TaxCalcLine {
					mapTaxLine[v.ID] += m.taxConvert(total*v.Rate, &v)
				} else {
					mapAmountTaxHeader[v.ID] += total
				}
			}
		}
	}

	grantTotal := totalInvoice
	taxes := make([]codekit.M, len(mapAmountTaxHeader)+len(mapTaxLine))
	i := 0
	for k, v := range mapTaxLine {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = m.taxToStringConvert(v, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal += v
				}
			} else {
				tax["Value"] = "-" + m.taxToStringConvert(v, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal -= v
				}
			}
		}

		taxes[i] = tax
		i++
	}

	for k, v := range mapAmountTaxHeader {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			val := m.taxConvert(v*taxSetup.Rate, &taxSetup)
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = m.taxToStringConvert(val, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal += val
				}
			} else {
				tax["Value"] = "-" + m.taxToStringConvert(val, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal -= val
				}
			}
		}

		taxes[i] = tax
		i++
	}

	totalInv := m.customerToStringConvert(totalInvoice, curr.Symbol2, custConf)
	grantTot := m.customerToStringConvert(grantTotal, curr.Symbol2, custConf)
	previewDetail.Items = append(previewDetail.Items, []string{"", "", "", "", "", "", "", "", "", "", "", "Total Invoice", totalInv})
	for _, t := range taxes {
		previewDetail.Items = append(previewDetail.Items, []string{"", "", "", "", "", "", "", "", "", "", "", t["Key"].(string), t["Value"].(string)})
	}
	previewDetail.Items = append(previewDetail.Items, []string{"", "", "", "", "", "", "", "", "", "", "", "Grand Total", grantTot})

	preview.PreviewReport.Sections = append(preview.PreviewReport.Sections, previewDetail)

	// keterangan & terbilang
	format := terbilang.Init()
	preview.PreviewReport.Header["Footer"] = [][]string{
		{"Terbilang"},
		{title(format.Convert(int64(grantTotal)))},
		{""},
		{"Keterangan"},
		{payload.References.Get("Keterangan", "").(string)},
	}
	hub.Save(preview)

	return payload, nil
}

type SyncMiningInvoiceRequest struct {
	ID string
}

func (m *InvoiceHandler) SyncMiningInvoice2(ctx *kaos.Context, p *SyncMiningInvoiceRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	payload := new(ficomodel.CustomerJournal)
	err := hub.GetByID(payload, p.ID)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when get customer journal: %s", err.Error()))
	}

	// get currency
	curr := new(tenantcoremodel.Currency)
	if err := hub.GetByID(curr, payload.CurrencyID); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("currency not found: %s", payload.ID))
	}

	// get customer configuration
	custConf := new(bagongmodel.CustomerConfiguration)
	if err = hub.GetByParm(custConf, dbflex.NewQueryParam().SetWhere(dbflex.Eq("CustomerID", payload.CustomerID))); err != nil && err != io.EOF {
		return nil, fmt.Errorf("error when get customer configuration: %s", err.Error())
	}

	// get & set customer name
	cust := new(tenantcoremodel.Customer)
	if err := hub.GetByID(cust, payload.CustomerID); err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("customer not found: %s", payload.CustomerID))
	}

	// get tax setup
	taxSetups := []ficomodel.TaxSetup{}
	err = hub.Gets(new(ficomodel.TaxSetup), dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("IsActive", true),
	), &taxSetups)
	if err != nil {
		return nil, fmt.Errorf("error when get tax setup: %s", err.Error())
	}
	mapTaxSetup := lo.Associate(taxSetups, func(tax ficomodel.TaxSetup) (string, ficomodel.TaxSetup) {
		return tax.ID, tax
	})

	preview, err := tenantcorelogic.GetPreviewBySource(hub, "CUSTOMER", payload.ID, "", payload.TransactionType)
	if err != nil {
		return nil, fmt.Errorf("error when get preview: %s", err.Error())
	}

	mapTaxLine := map[string]float64{}
	mapAmountTaxHeader := map[string]float64{}
	totalInvoice := float64(0)
	for i, line := range payload.Lines {
		totalInvoice += line.Amount

		// edit item
		// price
		price := fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(line.PriceEach, 0, ",", "."))
		preview.PreviewReport.Sections[0].Items[i+1][4] = price

		// standby day
		standbyDay := codekit.ToString(line.References.Get("Standby (Day)", 0).(int))
		preview.PreviewReport.Sections[0].Items[i+1][7] = standbyDay

		// standby discount
		standbyDiscount := line.References.Get("Standby (Potongan)", 0).(float64)
		preview.PreviewReport.Sections[0].Items[i+1][8] = fmt.Sprintf("%s%s", curr.Symbol2, accounting.FormatNumber(standbyDiscount, 0, ",", "."))

		// breakdown day
		breakdownDay := codekit.ToString(line.References.Get("Breakdown (Day)", 0).(int))
		preview.PreviewReport.Sections[0].Items[i+1][9] = breakdownDay

		// breakdown discount
		breakdownDiscount := line.References.Get("Breakdown (Potongan)", 0).(float64)
		preview.PreviewReport.Sections[0].Items[i+1][10] = m.customerToStringConvert(breakdownDiscount, curr.Symbol2, custConf)

		// total discount
		preview.PreviewReport.Sections[0].Items[i+1][11] = m.customerToStringConvert(breakdownDiscount+standbyDiscount, curr.Symbol2, custConf)

		// total
		preview.PreviewReport.Sections[0].Items[i+1][12] = m.customerToStringConvert(line.PriceEach-breakdownDiscount-standbyDiscount, curr.Symbol2, custConf)

		for _, code := range payload.TaxCodes {
			if v, ok := mapTaxSetup[code]; ok {
				if v.CalcMethod == ficomodel.TaxCalcLine {
					mapTaxLine[v.ID] += m.taxConvert(line.Amount*v.Rate, &v)
				} else {
					mapAmountTaxHeader[v.ID] += line.Amount
				}
			}
		}
	}

	grantTotal := totalInvoice
	taxes := make([]codekit.M, len(mapAmountTaxHeader)+len(mapTaxLine))
	i := 0
	for k, v := range mapTaxLine {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = m.taxToStringConvert(v, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal += v
				}
			} else {
				tax["Value"] = "-" + m.taxToStringConvert(v, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal -= v
				}
			}
		}

		taxes[i] = tax
		i++
	}

	for k, v := range mapAmountTaxHeader {
		tax := codekit.M{}
		if taxSetup, ok := mapTaxSetup[k]; ok {
			val := m.taxConvert(v*taxSetup.Rate, &taxSetup)
			tax["Key"] = taxSetup.Name

			if taxSetup.InvoiceOperation == ficomodel.TaxIncreaseAmount {
				tax["Value"] = m.taxToStringConvert(val, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal += val
				}
			} else {
				tax["Value"] = "-" + m.taxToStringConvert(val, curr.Symbol2, &taxSetup)
				if mapTaxSetup[k].IncludeInInvoice {
					grantTotal -= val
				}
			}
		}

		taxes[i] = tax
		i++
	}

	totalInv := m.customerToStringConvert(totalInvoice, curr.Symbol2, custConf)
	grantTot := m.customerToStringConvert(grantTotal, curr.Symbol2, custConf)
	totalLine := len(payload.Lines) + 1
	preview.PreviewReport.Sections[0].Items[totalLine] = []string{"", "", "", "", "", "", "", "", "", "", "", "Total Invoice", totalInv}
	for _, t := range taxes {
		totalLine++
		preview.PreviewReport.Sections[0].Items[totalLine] = []string{"", "", "", "", "", "", "", "", "", "", "", t["Key"].(string), t["Value"].(string)}
	}
	preview.PreviewReport.Sections[0].Items[totalLine+1] = []string{"", "", "", "", "", "", "", "", "", "", "", "Grant Total", grantTot}
	hub.Save(preview)

	return payload, nil
}

type GetProjectRequest struct {
	Where struct {
		CustomerID string
		SiteID     string
		Start      time.Time
		End        time.Time
	}
}

func (m *InvoiceHandler) GetProject(ctx *kaos.Context, payload *GetProjectRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get site entry
	siteEntrys := []bagongmodel.SiteEntry{}
	err := hub.Gets(new(bagongmodel.SiteEntry), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("SiteID", payload.Where.SiteID),
			dbflex.Gte("TrxDate", payload.Where.Start),
			dbflex.Lte("TrxDate", payload.Where.End),
		),
	), &siteEntrys)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when get site entry: %s", err.Error()))
	}

	// get site entry id
	siteEntryIDs := lo.Map(siteEntrys, func(m bagongmodel.SiteEntry, index int) interface{} {
		return m.ID
	})

	// get site entry
	siteEntryAssets := []bagongmodel.SiteEntryAsset{}
	err = hub.Gets(new(bagongmodel.SiteEntryAsset), dbflex.NewQueryParam().SetWhere(
		dbflex.In("SiteEntryID", siteEntryIDs...),
	), &siteEntryAssets)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when get site entry assets: %s", err.Error()))
	}

	assetIDs := make([]string, len(siteEntryAssets))
	projectIDs := make([]string, len(siteEntryAssets))
	for i, se := range siteEntryAssets {
		assetIDs[i] = se.AssetID
		projectIDs[i] = se.ProjectID
	}

	assets := []bagongmodel.Asset{}
	err = hub.GetsByFilter(new(bagongmodel.Asset), dbflex.In("_id", assetIDs...), &assets)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when get asset: %s", err.Error()))
	}
	mapAssets := lo.Associate(assets, func(asset bagongmodel.Asset) (string, []bagongmodel.UserInfo) {
		return asset.ID, asset.UserInfo
	})

	measureProjects := []sdpmodel.MeasuringProject{}
	err = hub.GetsByFilter(new(sdpmodel.MeasuringProject), dbflex.In("_id", projectIDs...), &measureProjects)
	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error when get measuring project: %s", err.Error()))
	}
	mapProjects := lo.Associate(measureProjects, func(measureProjects sdpmodel.MeasuringProject) (string, string) {
		return measureProjects.ID, measureProjects.ProjectAlias + "|" + measureProjects.ProjectName
	})

	mapResult := map[string]string{}
	for _, se := range siteEntryAssets {
		if v, ok := mapAssets[se.AssetID]; ok {
			if len(v) > 0 {
				userInfo := v[0]

				for i := 1; i < len(v); i++ {
					dateFrom1 := userInfo.AssetDateFrom

					dateFrom2 := v[i].AssetDateFrom

					if dateFrom2.After(dateFrom1) {
						userInfo = v[i]
					}
				}

				if userInfo.CustomerID == payload.Where.CustomerID {
					mapResult[se.ProjectID] = mapProjects[se.ProjectID]
				}
			}
		}
	}

	result := []codekit.M{}
	for k, v := range mapResult {
		if k != "" {
			result = append(result, codekit.M{
				"_id":  k,
				"Name": v,
			})
		}
	}

	return result, nil
}
