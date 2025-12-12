package ficomodel

import "git.kanosolution.net/sebar/tenantcore/tenantcoremodel"

type CustomerTrxType tenantcoremodel.TrxType
type VendorTrxType tenantcoremodel.TrxType
type CashBankTrxType tenantcoremodel.TrxType
type AssetTrxType tenantcoremodel.TrxType
type InventoryTrxType tenantcoremodel.TrxType
type JournalStatus tenantcoremodel.TrxType

const (
	CustomerPurchase   CustomerTrxType = "Customer Purchase"
	CustomerPayment    CustomerTrxType = "Customer Payment"
	CustomerAdjustment CustomerTrxType = "Customer Adjustment"

	VendorPurchase   VendorTrxType = "Vendor Purchase"
	VendorPayment    VendorTrxType = "Vendor Payment"
	VendorAdjustment VendorTrxType = "Vendor Adjustment"

	AssetAcquistion  AssetTrxType = "Asset Acquisition"
	AssetDeprecation AssetTrxType = "Asset Depreciation"
	AssetAdjustment  AssetTrxType = "Asset Adjustment"
	AssetDisposal    AssetTrxType = "Asset Disposal"

	CashIn       CashBankTrxType = "Cash In"
	CashOut      CashBankTrxType = "Cash Out"
	CastInterest CashBankTrxType = "Interest"
	CashFee      CashBankTrxType = "Fee"

	JournalStatusDraft     JournalStatus = "DRAFT"
	JournalStatusSubmitted JournalStatus = "SUBMITTED"
	JournalStatusReady     JournalStatus = "READY"
	JournalStatusPosted    JournalStatus = "POSTED"
	JournalStatusApproved  JournalStatus = "APPROVED"
	JournalStatusRejected  JournalStatus = "REJECTED"
)
