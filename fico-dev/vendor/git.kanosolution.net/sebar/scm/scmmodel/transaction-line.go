package scmmodel

import (
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type TransactionLine struct {
	ItemID    string
	InventDim InventDimension
	Dimension tenantcoremodel.Dimension
	TrxType   string
	Qty       float64
	UnitID    string
	Status    ItemBalanceStatus
}
