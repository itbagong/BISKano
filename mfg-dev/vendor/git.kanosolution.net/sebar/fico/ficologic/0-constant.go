package ficologic

var (
	uiEndPoints   = []string{"formconfig", "gridconfig", "listconfig", "new"}
	readEndPoints = []string{"get", "gets", "find"}
	//updateEndPoints = []string{"insert", "update", "save"}
	deleteEndPoints = []string{"delete"}
)

type AssetTrxType string

const (
	AssetAcquisition  AssetTrxType = "Acquisition"
	AssetDepreciation AssetTrxType = "Depreciation"
	AssetDispoal      AssetTrxType = "Disposal"
)
