package sdplogic_test

import (
	"encoding/json"
	"testing"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/smartystreets/goconvey/convey"
)

func TestSalesQuotationNoOpportunity(t *testing.T) {
	ctx := kaos.NewContextFromService(svc, nil)
	db := sebar.GetTenantDBFromContext(ctx)

	prepareCtxData(ctx, "random-user", testCoID1)
	spb := sdpmodel.SalesPriceBook{}
	customer := map[string]any{}
	// Item := map[string]any{}
	// Asset := map[string]any{}

	convey.Convey("insert master salespricebook", t, func() {
		spb = insertmastersalespricebookSalesQuotation(ctx)
	})

	convey.Convey("insert Customer", t, func() {
		customer = insertCustomerSalesQuotation(ctx)
	})

	// convey.Convey("insert Asset", t, func() {
	// 	Asset = insertAssetLinesSalesQuotation(ctx)
	// })

	// convey.Convey("insert Item", t, func() {
	// 	Item = insertItemLinesSalesQuotation(ctx)
	// })

	convey.Convey("Insert General", t, func() {
		salesquotation := sdpmodel.SalesQuotation{}
		err := json.Unmarshal([]byte(`{"Customer":"G0001","Address":"dk warg","AddressDelivery":"surabayasd","City":"surabaya","CityDelivery":"surabaya","Province":"jatim","ProvinceDelivery":"jatim","Country":"ID","CountryDelivery":"ID","Zipcode":"60221","ZipcodeDelivery":"60224","Name":"Customer 1","QuotationNo":"rere","QuotationName":"treyryre","SalesPriceBook":"6555eb114aff7cabad7edc0e","Lines":[{"Item":"","Description":"ewtt","Qty":10,"ContractPeriod":10,"UnitPrice":10,"Amount":100,"Discount":0,"TaxCodes":["TCM007","TCM001","PPN"],"GetTax":[],"index":0,"Asset":"AST0067","Spesifications":["05mm","BAN_DALAM1"],"Uom":"EAC","currentIndex":0,"Taxable":true,"TaxAmount":0.11,"suimRecordChange":false}],"SubTotalAmount":100,"TaxAmount":0,"DiscountAmount":0,"TotalAmount":100}`), &salesquotation)
		convey.So(err, convey.ShouldBeNil)

		salesquotation.Customer = customer["_id"].(string)
		salesquotation.SalesPriceBook = spb.ID

		sq, err := callAPI(ctx, "/v1/sdp/salesquotation/insert", &salesquotation, &salesquotation)
		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("Sales Quotation ID", sq.ID, convey.ShouldNotBeEmpty)

		dbsalesquotation, err := datahub.GetByID(db, new(sdpmodel.SalesQuotation), sq.ID)
		convey.So(err, convey.ShouldBeNil)
		convey.SoMsg("DB Sales Quotation ID", dbsalesquotation.ID, convey.ShouldNotBeEmpty)

		// submitted, _ := datahub.GetByID(db, new(scmmodel.InventJournal), j.ID)
	})

}

func insertmastersalespricebookSalesQuotation(ctx *kaos.Context) sdpmodel.SalesPriceBook {
	// spb := sdpmodel.SalesPriceBook{}
	// err := json.Unmarshal([]byte(`{"Name":"mitsubitshi","EndPeriod":"2024-01-30T12:00:00+07:00","Lines":[{"ProductionYear":2022,"AssetType":"AUT011","AssetID":"N8111BG","ItemID":"AIR_FILTER","MinPrice":10000,"MaxPrice":100000,"Quantity":0,"Unit":"DAYS","PhysicalDimension":{"IsEnabledSpecVariant":true,"IsEnabledSpecSize":false,"IsEnabledSpecGrade":false,"IsEnabledItemBatch":false,"IsEnabledItemSerial":false,"IsEnabledLocationWarehouse":true,"IsEnabledLocationAisle":false,"IsEnabledLocationSection":false,"IsEnabledLocationBox":false},"FinanceDimension":{"IsEnabledSpecVariant":true,"IsEnabledSpecSize":false,"IsEnabledSpecGrade":false,"IsEnabledItemBatch":false,"IsEnabledItemSerial":false,"IsEnabledLocationWarehouse":true,"IsEnabledLocationAisle":false,"IsEnabledLocationSection":false,"IsEnabledLocationBox":false}}],"StartPeriod":"2024-01-04T03:04:59+07:00"}`), &spb)
	// convey.So(err, convey.ShouldBeNil)

	// spbres, err := callAPI(ctx, "/v1/sdp/salespricebook/insert", &spb, spb)
	// convey.So(err, convey.ShouldBeNil)

	spb := sdpmodel.SalesPriceBook{}
	err := json.Unmarshal(
		[]byte(`
	{
    "_id": "6555f80b05b482e8b3252849",
    "Name": "12asd",
    "StartPeriod": "0001-01-01T00:00:00Z",
    "EndPeriod": "0001-01-01T00:00:00Z",
    "Dimension": null,
    "Lines": null,
    "Created": "0001-01-01T00:00:00Z",
    "LastUpdate": "0001-01-01T00:00:00Z"
}
	`), &spb)
	convey.So(err, convey.ShouldBeNil)

	return spb
}

func insertCustomerSalesQuotation(ctx *kaos.Context) map[string]any {
	// customer := map[string]any{}
	// err := json.Unmarshal([]byte(`{"_id":"sasa","Name":"ddd","IsActive":true,"GroupID":"CG01","Dimension":[{"Key":"PC","Value":"MINING"},{"Key":"CC","Value":"HQO"},{"Key":"Site","Value":"SITE0001"},{"Key":"Asset","Value":""}]}`), &customer)
	// convey.So(err, convey.ShouldBeNil)

	// customerres, err := callAPI(ctx, "/v1/bagong/customer/save", &customer, &customer)
	// convey.So(err, convey.ShouldBeNil)
	// convey.Println(customerres)

	customer := map[string]any{}
	err := json.Unmarshal([]byte(`
		{
			"_id": "G0001",
			"Name": "Customer 1",
			"GroupID": "",
			"Setting": {
					"MainBalanceAccount": "",
					"DepositAccount": ""
			},
			"Dimension": null,
			"IsActive": false,
			"Created": "0001-01-01T00:00:00Z",
			"LastUpdate": "0001-01-01T00:00:00Z"
		}
`), &customer)
	convey.So(err, convey.ShouldBeNil)

	return customer
}

func insertAssetLinesSalesQuotation(ctx *kaos.Context) map[string]any {
	asset := map[string]any{}
	err := json.Unmarshal([]byte(`
	{
    "_id": "KT8999LLL",
    "Name": "KT 8999 LLL",
    "GroupID": "",
    "AssetType": "",
    "AcquisitionAccount": "",
    "DepreciationAccount": "",
    "DisposalAccount": "",
    "AdjustmentAccount": "",
    "Dimension": null,
    "Created": "0001-01-01T00:00:00Z",
    "LastUpdate": "0001-01-01T00:00:00Z"
}
	`), &asset)
	convey.So(err, convey.ShouldBeNil)

	return asset
}

func insertItemLinesSalesQuotation(ctx *kaos.Context) map[string]any {
	item := map[string]any{}
	err := json.Unmarshal([]byte(`
	{
    "_id": "BALLPOINT",
    "Name": "BALLPOINT",
    "ItemType": "",
    "ItemGroupID": "",
    "LedgerAccountIDStock": "",
    "DefaultUnitID": "",
    "CostUnitCalcMethod": "",
    "CostUnit": 0,
    "PhysicalDimension": {
        "IsEnabledSpecVariant": false,
        "IsEnabledSpecSize": false,
        "IsEnabledSpecGrade": false,
        "IsEnabledItemBatch": false,
        "IsEnabledItemSerial": false,
        "IsEnabledLocationWarehouse": false,
        "IsEnabledLocationAisle": false,
        "IsEnabledLocationSection": false,
        "IsEnabledLocationBox": false
    },
    "FinanceDimension": {
        "IsEnabledSpecVariant": false,
        "IsEnabledSpecSize": false,
        "IsEnabledSpecGrade": false,
        "IsEnabledItemBatch": false,
        "IsEnabledItemSerial": false,
        "IsEnabledLocationWarehouse": false,
        "IsEnabledLocationAisle": false,
        "IsEnabledLocationSection": false,
        "IsEnabledLocationBox": false
    },
    "Created": "0001-01-01T00:00:00Z",
    "LastUpdate": "0001-01-01T00:00:00Z"
}
	`), &item)
	convey.So(err, convey.ShouldBeNil)

	return item
}
