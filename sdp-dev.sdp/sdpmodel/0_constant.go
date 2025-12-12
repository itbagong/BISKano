package sdpmodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

// const (
// 	ModuleInventory tenantcoremodel.TrxModule = "INVENTORY"
// 	ModulePurchase  tenantcoremodel.TrxModule = "PURCHASE"
// 	ModuleWorkorder tenantcoremodel.TrxModule = "WORKORDER"
// )

type JournalStatus string

const (
	JournalDraft     JournalStatus = "DRAFT"
	JournalSubmitted JournalStatus = "SUBMITTED"
	JournalReady     JournalStatus = "READY"
	JournalPosted    JournalStatus = "POSTED"
)

type InventTrxType string

const (
	JournalMovementIn  InventTrxType = "Movement In"
	JournalMovementOut InventTrxType = "Movement Out"
	JournalTransfer    InventTrxType = "Transfer"
	JournalOpname      InventTrxType = "Stock Opname"

	PurchQuote InventTrxType = "Purchase Quotation"
	PurchOrder InventTrxType = "Purchase Order"

	SalsOrder     tenantcoremodel.TrxModule = "Sales Order"
	SalsQuotation tenantcoremodel.TrxModule = "Sales Quotation"
	// SalsOrder InventTrxType = "Sales Order"

	InventReceive  InventTrxType = "Inventory Receive"
	InventIssuance InventTrxType = "Inventory Issuance"
)

type ItemBalanceStatus string

const (
	ItemConfirmed ItemBalanceStatus = "Confirmed"
	ItemReserved  ItemBalanceStatus = "Reserved"
	ItemPlanned   ItemBalanceStatus = "Planned"
)

var SourceTypeURLMap = map[string]string{
	string(SalsOrder):              "scm/salesorder",
	string(SalsQuotation):          "scm/salesquotation",
}
