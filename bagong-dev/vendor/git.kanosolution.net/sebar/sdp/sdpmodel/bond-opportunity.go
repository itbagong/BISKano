package sdpmodel

import (
	"time"
)

type BondOpportunity struct {
	TypeBond    string    `form_kind:"text" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOTB|_id|Name"` // get from master type Bond
	Amount      float64   `form_kind:"number"`
	SubmitDate  time.Time `form_kind:"date"`
	StatusBond  string    `form_kind:"text" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=SOBS|_id|Name"` // Get from master status bond
	ExpiredDate time.Time `form_kind:"date"`
	Guarantor   string    `form_kind:"text"`
}
