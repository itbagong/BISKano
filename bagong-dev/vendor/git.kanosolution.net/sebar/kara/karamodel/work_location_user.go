package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkLocationUser_UserEntry struct {
	ID             string `grid:"hide" form:"hide" bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	WorkLocationID string `grid:"hide" form:"hide" form_required:"1"`
	UserID         string `grid:"hide" form_lookup:"/iam/user/find|_id|DisplayName,Email"`
	UserName       string
	Email          string
	RuleID         string    `form_required:"1" form_section:"General" form_lookup:"kara/rule/find|_id|Name"`
	RosterID       string    `form_required:"1" form_section:"General" form_lookup:"tenant/masterdata/find?MasterDataTypeID=Roster|_id|Name"`
	From           time.Time `form_kind:"date" form_required:"1"`
	To             time.Time `form_kind:"date" form_required:"1"`
}

type WorkLocationUser struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" grid:"hide" form:"hide" form_section:"General" form_section_auto_col:"2"`
	WorkLocationID    string    `form_required:"1" form_section:"General" form_lookup:"admin/worklocation/find|_id|Name"`
	UserID            string    `form_read_only:"1" grid:"hide" form:"hide"`
	From              time.Time `form_kind:"date"`
	To                time.Time `form_kind:"date"`
	RuleID            string    `form_required:"1" form_section:"General" form_lookup:"admin/rule/find|_id|Name"`
	RosterID          string    `form_required:"1" form_section:"General" form_lookup:"tenant/masterdata/find?MasterDataTypeID=Roster|_id|Name"`
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *WorkLocationUser) TableName() string {
	return "KaraWorkLocationUsers"
}

func (o *WorkLocationUser) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkLocationUser) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkLocationUser) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkLocationUser) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkLocationUser) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkLocationUser) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *WorkLocationUser) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "User_Rule_WorkLocation_Index", Fields: []string{"RuleID", "UserID", "WorkLocationID"}},
	}
}
