package rbacmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppIconType string

const (
	AppIconNone  AppIconType = "NONE"
	AppIconImage AppIconType = "IMAGE"
	AppIconMDI   AppIconType = "MDI"
)

type App struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	Public            bool
	UseRoleToAccess   bool
	RoleID            string
	Address           string
	IconType          AppIconType `form_items:"NONE|IMAGE|MDI"`
	IconValue         string
	Enable            bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *App) TableName() string {
	return "Apps"
}

func (o *App) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *App) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *App) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *App) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *App) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *App) PostSave(dbflex.IConnection) error {
	return nil
}
