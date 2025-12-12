package tenantcoremodel

type CustomerSetting struct {
	MainBalanceAccount string
	DepositAccount     string
}

type VendorSetting struct {
	MainBalanceAccount string
	DepositAccount     string
}

type CashbookSetting struct {
	MainBalanceAccount string
	CurrencyID         string
	Treshold           float64
	AutoReplenish      bool
}

type AssetSetting struct {
	AcquisitionAccount  string
	DepreciationAccount string
	AdjustmentAccount   string
	DisposalAccount     string
}
