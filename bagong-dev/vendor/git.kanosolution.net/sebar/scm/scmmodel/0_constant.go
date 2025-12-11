package scmmodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

const (
	ModuleInventory tenantcoremodel.TrxModule = "INVENTORY"
	ModulePurchase  tenantcoremodel.TrxModule = "PURCHASE"
	ModuleWorkorder tenantcoremodel.TrxModule = "WORKORDER"
)

type JournalStatus string

const (
	JournalDraft     JournalStatus = "DRAFT"
	JournalSubmitted JournalStatus = "SUBMITTED"
	JournalReady     JournalStatus = "READY"
	JournalPosted    JournalStatus = "POSTED"
	JournalRejected  JournalStatus = "REJECTED"
)

type InventTrxType string

const (
	JournalMovementIn  InventTrxType = "Movement In"
	JournalMovementOut InventTrxType = "Movement Out"
	JournalTransfer    InventTrxType = "Transfer"
	JournalOpname      InventTrxType = "Stock Opname"
	JournalWorkOrder   InventTrxType = "Work Order"
	JournalWorkRequest InventTrxType = "Work Request"

	PurchQuote   InventTrxType = "Purchase Quotation"
	PurchOrder   InventTrxType = "Purchase Order"
	PurchRequest InventTrxType = "Purchase Request"

	InventReceive   InventTrxType = "Inventory Receive"
	InventIssuance  InventTrxType = "Inventory Issuance"
	ItemRequestType InventTrxType = "Item Request"

	AssetAcquisitionTrxType InventTrxType = "Asset Acquisition"
)

type ItemBalanceStatus string

const (
	ItemConfirmed ItemBalanceStatus = "Confirmed"
	ItemReserved  ItemBalanceStatus = "Reserved"
	ItemPlanned   ItemBalanceStatus = "Planned"
)

type PurchaseType string

const (
	PurchaseTypeStock   PurchaseType = "STOCK"
	PurchaseTypeVirtual PurchaseType = "VIRTUAL"
	PurchaseTypeService PurchaseType = "SERVICE"
	PurchaseTypeAsset   PurchaseType = "ASSET"
)

var SourceTypeURLMap = map[string]string{
	string(PurchOrder):              "scm/PurchaseOrder",
	string(PurchRequest):            "scm/PurchaseRequest",
	string(JournalMovementIn):       "scm/InventoryJournal?type=Movement%20In&title=Movement%20In",
	string(JournalMovementOut):      "scm/InventoryJournal?type=Movement%20Out&title=Movement%20Out",
	string(InventReceive):           "scm/InventTrx?type=Inventory%20Receive&title=Inventory%20Receive",
	string(InventIssuance):          "scm/InventTrx?type=Inventory%20Issuance&title=Inventory%20Issuance",
	string(JournalTransfer):         "scm/InventoryJournal?type=Transfer&title=Item%20Transfer",
	string(ItemRequestType):         "scm/ItemRequest",
	string(AssetAcquisitionTrxType): "scm/AssetAcquisition",
}

type InventTrxReferenceKey string

const (
	RefKeyMovementInID      InventTrxReferenceKey = "MovementInID"
	RefKeyMovementOutID     InventTrxReferenceKey = "MovementOutID"
	RefKeyPurchaseOrderID   InventTrxReferenceKey = "PurchaseOrderID"
	RefKeyPurchaseRequestID InventTrxReferenceKey = "PurchaseRequestID"
	RefKeyItemRequestID     InventTrxReferenceKey = "ItemRequestID"
	RefKeyTransferID        InventTrxReferenceKey = "TransferID"
	RefKeyGoodReceive       InventTrxReferenceKey = "GRID"
	RefKeyPOLineID          InventTrxReferenceKey = "POLineID"
)

func GetRefKey(trxType string) string {
	// tidak perlu semua di map, sebutuhnya saja
	refKeyMap := map[string]InventTrxReferenceKey{
		string(JournalMovementIn):  RefKeyMovementInID,
		string(JournalMovementOut): RefKeyMovementOutID,
		string(PurchOrder):         RefKeyPurchaseOrderID,
		string(PurchRequest):       RefKeyPurchaseRequestID,
		string(ItemRequestType):    RefKeyItemRequestID,
		string(JournalTransfer):    RefKeyTransferID,
	}

	key, exist := refKeyMap[trxType]
	if !exist {
		key = "UNKNOWN"
	}

	return string(key)
}
