package rbacmodel

import (
	"path"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppMenu struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" grid_keyword:"1"`
	AppID             string `form_lookup:"/admin/app/find|_id|_id"`
	ParentMenuID      string `grid:"hide" form_lookup:"/admin/menu/find|_id|PathLabel"`
	Label             string `grid_keyword:"1"`
	PathLabel         string `grid:"hide" form_read_only:"1" grid_sortable:"1"`
	FeatureID         string `form_lookup:"/admin/feature/find|_id|_id,Name"`
	Icon              string `grid:"hide"`
	Uri               string
	Section           string `form_items:"Master|Transaction|Report|Setting"`
	Priority          int
	Expand            bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AppMenu) TableName() string {
	return "AppMenus"
}

func (o *AppMenu) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AppMenu) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AppMenu) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AppMenu) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AppMenu) PreSave(conn dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}

	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	if o.PathLabel == "" || o.ParentMenuID == "" {
		parentPathLabel := o.getParentLabel(conn, o.ParentMenuID, "")
		if parentPathLabel == "" {
			o.PathLabel = o.ID
		} else {
			o.PathLabel = path.Join(parentPathLabel, o.ID)
		}
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *AppMenu) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *AppMenu) getParentLabel(conn dbflex.IConnection, id, ids string) string {
	cmd := dbflex.From(o.TableName()).Select("_id", "Label", "ParentMenuID", "AppID").Where(dbflex.Eq("_id", id))
	parentMenu := new(AppMenu)
	e := conn.Cursor(cmd, nil).Fetch(parentMenu).Close()
	if e != nil {
		return ids
	}

	if parentMenu.AppID != o.AppID {
		return ids
	}

	if ids == "" {
		ids = parentMenu.ID
	} else {
		ids = path.Join(parentMenu.ID, ids)
	}

	if parentMenu.ParentMenuID != "" {
		ids = o.getParentLabel(conn, parentMenu.ParentMenuID, ids)
	}

	return ids
}
