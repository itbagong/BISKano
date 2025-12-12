package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
)

type TransactionHistory struct {
	orm.DataModelBase `bson:"-" json:"-"`
	TrxDate           time.Time `label:"Transaction Date"`
	SourceType        string    `label:"Type"`
	SourceJournalID   string
	Text              string
	Amount            int
	Balance           int
}

type TransactionHistoryCashBank struct {
	orm.DataModelBase `bson:"-" json:"-"`
	TrxDate           time.Time `label:"Transaction Date"`
	TrxType           string
	SourceJournalID   string
	Text              string
	Amount            int
	Balance           int
}

type TransactionHistoryCOA struct {
	orm.DataModelBase `bson:"-" json:"-"`
	TrxDate           time.Time `label:"Transaction Date"`
	SourceTrxType     string    `label:"Type"`
	SourceJournalID   string
	Text              string
	Amount            int
	Balance           int
}
