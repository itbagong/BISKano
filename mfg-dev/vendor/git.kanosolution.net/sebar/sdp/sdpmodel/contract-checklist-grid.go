package sdpmodel

import "time"

type ContractChecklistGridTab struct {
	UnitType           string
	SunID              string
	AssetID            string
	UnitStatus         string
	DocumentStatus     string
	DeliveryStatus     string
	CommisioningDate   time.Time
	CommisioningResult string
	CommisioningStatus string
}
