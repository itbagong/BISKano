package flowconfig

type ModuleConfig struct {
	AssetEndPoint  string
	AssetRegion    string
	AssetBucket    string
	AssetKey       string
	AssetSecret    string
	AddrUserFind   string
	AddrUserAccess string
}

var (
	Config = new(ModuleConfig)
)
