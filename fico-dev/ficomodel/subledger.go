package ficomodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

const (
	SubledgerVendor     tenantcoremodel.TrxModule = "VENDOR"
	SubledgerCustomer   tenantcoremodel.TrxModule = "CUSTOMER"
	SubledgerAsset      tenantcoremodel.TrxModule = "ASSET"
	SubledgerPayroll    tenantcoremodel.TrxModule = "PAYROLL"
	SubledgerExpense    tenantcoremodel.TrxModule = "EXPENSE"
	SubledgerCashBank   tenantcoremodel.TrxModule = "CASHBANK"
	SubledgerTax        tenantcoremodel.TrxModule = "TAX"
	SubledgerNone       tenantcoremodel.TrxModule = "LEDGERACCOUNT"
	SubledgerAccounting tenantcoremodel.TrxModule = "LEDGERACCOUNT"

	// SCM
	Inventory        tenantcoremodel.TrxModule = "INVENTORY" // Movement In / Movement Out
	GoodReceive      tenantcoremodel.TrxModule = "Inventory Receive"
	GoodIssuance     tenantcoremodel.TrxModule = "Inventory Issuance"
	Transfer         tenantcoremodel.TrxModule = "Transfer"
	ItemRequest      tenantcoremodel.TrxModule = "Item Request"
	PurchOrder       tenantcoremodel.TrxModule = "Purchase Order"
	PurchRequest     tenantcoremodel.TrxModule = "Purchase Request"
	AssetAcquisition tenantcoremodel.TrxModule = "Asset Acquisition"

	// MFG
	WorkRequest                tenantcoremodel.TrxModule = "Work Request"
	WorkOrder                  tenantcoremodel.TrxModule = "Work Order"
	WorkOrderReportConsumption tenantcoremodel.TrxModule = "Work Order Report Consumption"
	WorkOrderReportResource    tenantcoremodel.TrxModule = "Work Order Report Resource"
	WorkOrderReportOutput      tenantcoremodel.TrxModule = "Work Order Report Output"
)

var SourceTypeURLMap = map[string]string{
	"CASH IN":                   "bagong/cashin",
	"CASH OUT":                  "bagong/cashout",
	string(SubledgerCustomer):   "fico/CustomerTransaction",
	string(SubledgerVendor):     "fico/vendortransaction",
	string(SubledgerAccounting): "fico/ledgerjournal",
}
