package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MovementStatus string

const (
	MovementStatusDraft     MovementStatus = "Draft"
	MovementStatusSubmitted MovementStatus = "Submitted"
	MovementStatusApproved  MovementStatus = "Approved"
	// MovementStatusPartialReceived MovementStatus = "PartialReceived"
	// MovementStatusPartialIssued   MovementStatus = "PartialIssued"
	MovementStatusRejected MovementStatus = "Rejected"
	MovementStatusClosed   MovementStatus = "Closed"
)

type MovementIn struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section_size:"4" form_section:"General1"`
	TrxDate            *time.Time                `form_kind:"date" form_section:"General1"`
	JournalType        string                    `grid:"hide" form_required:"1" form_lookup:"/scm/inventorytransactionjournaltype/find?TransactionType=Movement_In|_id|_id,Name"  form_section:"General1"`
	Departement        string                    `form_section:"General1"`
	Notes              string                    `form_multi_row:"3" form_section:"General2"`
	ReasonReject       string                    `form_multi_row:"3" form_section:"General2"`
	Status             MovementStatus            `form_section:"General2" form_read_only:"1"`
	InventoryDimension InventDimension           `grid:"hide" form_section:"Dimension1"`
	FinancialDimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension2"`
	RevisionNo         int                       `grid:"hide" form:"hide"` // TODO: increment this when revision
	Created            time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate         time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *MovementIn) TableName() string {
	return "MovementIns"
}

func (o *MovementIn) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *MovementIn) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *MovementIn) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *MovementIn) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *MovementIn) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *MovementIn) PostSave(dbflex.IConnection) error {
	return nil
}
