package ntslconfig

type ModConfig struct {
	MaxMemberPerGroup int `json:"max_member_per_group"`
}

var (
	Config = &ModConfig{}
)
