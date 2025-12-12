package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoodIssueFrom string
type GoodIssueStatus string

const (
	GoodIssueFromMovementOut  GoodIssueFrom = "MovementOut"
	GoodIssueFromItemAssembly GoodIssueFrom = "ItemAssembly"
	GoodIssueFromSO           GoodIssueFrom = "SO"
	GoodIssueFromItemTransfer GoodIssueFrom = "ItemTransfer"

	GoodIssueStatusOpen          GoodIssueStatus = "Open"
	GoodIssueStatusPartialIssued GoodIssueStatus = "PartialIssued"
	GoodIssueStatusClosed        GoodIssueStatus = "Closed"
)

type GoodIssue struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General1"`
	GoodIssueDate      time.Time                 `form_kind:"date" form_section:"General1"`
	GoodIssueFrom      GoodIssueFrom             `form_section:"General1"`
	ReffNo             string                    `form_section:"General1"`
	JournalType        string                    `form_required:"1" form_lookup:"/scm/inventorytransactionjournaltype/find?TransactionType=Movement_In|_id|_id,Name" form_section:"General1"`
	Company            string                    `form_section:"General2"`
	Notes              string                    `form_multi_row:"5" form_section:"General2"`
	Status             GoodIssueStatus           `form_section:"General2" form_read_only:"1"`
	InventoryDimension InventDimension           `grid:"hide" form_section:"Dimension1"`
	FinancialDimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension2"`
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *GoodIssue) TableName() string {
	return "GoodIssues"
}

func (o *GoodIssue) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *GoodIssue) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *GoodIssue) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *GoodIssue) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *GoodIssue) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *GoodIssue) PostSave(dbflex.IConnection) error {
	return nil
}
