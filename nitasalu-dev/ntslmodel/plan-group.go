package ntslmodel

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlanGroup struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	PlanID            []string
	MemberCount       int
	OwnerID           string
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PlanGroup) TableName() string {
	return "PlanGroups"
}

func (o *PlanGroup) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PlanGroup) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PlanGroup) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PlanGroup) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PlanGroup) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if len(o.PlanID) == 0 {
		o.MemberCount = 0
	} else {
		cmd := dbflex.From(new(Plan).TableName()).Aggr(dbflex.Sum("MemberCount")).Where(dbflex.In("_id", o.PlanID))
		m := codekit.M{}
		e := conn.Cursor(cmd, nil).Fetch(&m)
		if e != nil {
			return fmt.Errorf("calc member count: %s", e.Error())
		}
		o.MemberCount = m.GetInt("MemberCount")
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PlanGroup) PostSave(dbflex.IConnection) error {
	return nil
}
