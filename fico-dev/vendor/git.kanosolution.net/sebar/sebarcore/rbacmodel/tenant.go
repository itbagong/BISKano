package rbacmodel

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/sebar"
)

type CreateTenantRequest struct {
	FID  string `label:"Friendly ID" form_required:"1"`
	Name string `form_required:"1"`
}

type Tenant struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string   `bson:"_id" json:"_id" key:"1" form_read_only:"1" grid:"hide"`
	FID               string   `label:"Friendly ID" form_required:"1"`
	Name              string   `form_required:"1"`
	Dimensions        []string `form_use_list:"1" form_allow_add:"1"`
	LicenseKey        string   `grid:"hide"`
	Enable            bool
	OwnerID           string
	Use2FA            bool      `label:"Use 2FA"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1"`
}

func (o *Tenant) TableName() string {
	return "Tenants"
}

func (o *Tenant) FK() []*orm.FKConfig {
	return []*orm.FKConfig{}
}

func (o *Tenant) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *Tenant) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Tenant) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Tenant) PreSave(conn dbflex.IConnection) error {
	tenants := []Tenant{}
	cmd := dbflex.From(o.TableName()).Select().Where(dbflex.Eq("FID", o.FID)).Take(1)
	if err := conn.Cursor(cmd, nil).Fetchs(&tenants, 0).Close(); err == nil && len(tenants) > 0 {
		if tenants[0].ID != o.ID {
			return fmt.Errorf("tenant with friendly ID %s already exist", o.FID)
		}
	}

	if o.ID == "" {
		o.ID = sebar.MakeID("", sebar.PrecisionMilli, 32)
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Tenant) PostSave(dbflex.IConnection) error {
	return nil
}
