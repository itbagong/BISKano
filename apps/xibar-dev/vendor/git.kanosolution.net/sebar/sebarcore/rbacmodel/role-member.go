package rbacmodel

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleScope string
type DimensionScope string

const (
	RoleScopeGlobal RoleScope = "GLOBAL"
	RoleScopeTenant RoleScope = "TENANT"

	DimensionNone   DimensionScope = "None"
	DimensionUser   DimensionScope = "User"
	DimensionCustom DimensionScope = "Custom"
)

type RoleMember struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string         `bson:"_id" json:"_id" key:"1" form:"hide" grid:"hide"`
	UserID            string         `form_lookup:"admin/user/find|_id|LoginID" form_label_field:"LoginID" grid:"hide" form_pos:"1,1"`
	LoginID           string         `form:"hide" grid_keyword:"1" grid_sortable:"1"`
	Scope             RoleScope      `form_items:"GLOBAL|TENANT" form_pos:"1,1"`
	TenantID          string         `form_lookup:"/admin/tenant/find|_id|Name" form_pos:"1,2"`
	RoleID            string         `form_lookup:"admin/role/find|_id|Name" form_label_field:"RoleName" grid:"hide" form_pos:"1,3" grid_keyword:"1" grid_sortable:"1"`
	RoleName          string         `form:"hide" grid_keyword:"1"`
	DimensionScope    DimensionScope `form_items:"None|User|Custom" form_section:"Dimension" form_section_auto_col:"2"`
	Dimension         Dimension      `form_section:"Dimension"`
	Hash              string
}

func (o *RoleMember) TableName() string {
	return "RbacRoleMembers"
}

func (o *RoleMember) FK() []*orm.FKConfig {
	return []*orm.FKConfig{
		{FieldID: "UserID", RefTableName: new(User).TableName(), RefField: "_id", Map: codekit.M{"LoginID": "LoginID"}},
		{FieldID: "RoleID", RefTableName: new(Role).TableName(), RefField: "_id", Map: codekit.M{"RoleName": "Name"}},
	}
}

func (o *RoleMember) ReverseFK() []*orm.ReverseFKConfig {
	return []*orm.ReverseFKConfig{}
}

func (o *RoleMember) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *RoleMember) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *RoleMember) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.DimensionScope != DimensionCustom {
		o.Dimension = Dimension{}
	}
	o.Hash = o.Dimension.Hash()
	return nil
}

func (o *RoleMember) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *RoleMember) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "userid_index", IsUnique: false, Fields: []string{"UserID", "RoleID", "Dimension.Items.Kind", "Dimension.Items.Value"}},
		{Name: "roleid_index", IsUnique: false, Fields: []string{"RoleID"}},
	}
}
