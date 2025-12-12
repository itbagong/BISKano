package tenantcoremodel

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NumberSequence struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string
	OutFormat         string
	UseDate           string
	LastNo            int
}

func (o *NumberSequence) Format(num int, dt *time.Time) string {
	if o.OutFormat == "" {
		return ""
	}
	if o.UseDate == "" {
		return fmt.Sprintf(o.OutFormat, num)
	} else {
		return fmt.Sprintf(o.OutFormat, dt.Format(o.UseDate), num)
	}

}
func (o *NumberSequence) TableName() string {
	return "NumberSequence"
}

func (o *NumberSequence) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *NumberSequence) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *NumberSequence) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *NumberSequence) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *NumberSequence) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}

	return nil
}

func (o *NumberSequence) PostSave(dbflex.IConnection) error {
	return nil
}
