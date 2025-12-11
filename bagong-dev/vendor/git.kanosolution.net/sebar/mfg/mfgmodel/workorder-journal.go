package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type WorkOrderJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                            `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General"  form_section_auto_col:"2"`
	CompanyID         string                            `form_section:"General" form_lookup:"/tenant/company/find|_id|_id,Name" form:"hide"`
	Name              string                            `form_required:"1" form_section:"General" form_section_direction:"row"`
	JournalTypeID     string                            `form_section:"General" form_lookup:"/mfg/workorder/journal/type/find|_id|_id,Name"`
	PostingProfileID  string                            `form_section:"General" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	ItemUsage         []scmmodel.InventReceiveIssueLine `grid:"hide" form:"hide"`
	//TODO: ini dirubah ke type data untuk manpower usage dan machine usage
	TrxDate    *time.Time                `form_kind:"date" form_section:"General2"`
	TrxType    scmmodel.InventTrxType    `form_section:"General2" grid:"hide" form:"hide"`
	Status     ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	InventDim  scmmodel.InventDimension  `grid:"hide" form_section:"InventDim"`
	Dimension  tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension"`
	Created    time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *WorkOrderJournal) TableName() string {
	return "WorkOrderJournals"
}

func (o *WorkOrderJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkOrderJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkOrderJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkOrderJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkOrderJournal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkOrderJournal) PostSave(dbflex.IConnection) error {
	return nil
}
