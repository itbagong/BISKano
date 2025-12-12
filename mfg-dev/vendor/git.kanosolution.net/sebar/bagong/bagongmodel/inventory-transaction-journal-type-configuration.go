package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventoryTransactionJournalTypeConfiguration struct {
	orm.DataModelBase                 `bson:"-" json:"-"`
	ID                                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	InventoryTransactionJournalTypeID string    `form:"hide"`
	BreakdownHourDivider              string    `form_required:"1" form_section:"Unit Rate" label:"Breakdown Hour Divider" form_pos:"1,1" form_lookup:"/bagong/masterdata/find?MasterDataTypeID=BRHD|_id|Name"`
	BreakdownRate                     int       `form_required:"1" form_section:"Unit Rate" label:"Breakdown Rate" form_pos:"1,2"`
	Created                           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
	LastUpdate                        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *InventoryTransactionJournalTypeConfiguration) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "Unit Rate", ShowTitle: true, AutoCol: 1},
		}},
	}
}

func (o *InventoryTransactionJournalTypeConfiguration) TableName() string {
	return "InventoryTransactionJournalTypeConfigurations"
}

func (o *InventoryTransactionJournalTypeConfiguration) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventoryTransactionJournalTypeConfiguration) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventoryTransactionJournalTypeConfiguration) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventoryTransactionJournalTypeConfiguration) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventoryTransactionJournalTypeConfiguration) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *InventoryTransactionJournalTypeConfiguration) PostSave(dbflex.IConnection) error {
	return nil
}
