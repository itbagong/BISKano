package karamodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
)

type UserStatus string

const (
	UserRegistered UserStatus = "Registered"
	UserLinked     UserStatus = "Linked"
)

type UserProfile struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id" key:"1" form_read_only:"1" grid:"hide" form_section:"General" form_section_auto_col:"2"`
	UserID             string `form_lookup:"/iam/user/find|_id|DisplayName,Email" form_read_only_edit:"1"`
	Enable             bool
	HolidayProfileID   string `form_lookup:"/kara/holiday/find|_id|_id,Name"`
	MinimumWorkingHour int
	Created            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type UserProfileExtended struct {
	orm.DataModelBase  `bson:"-" json:"-"`
	ID                 string `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form_section_auto_col:"2"`
	UserID             string
	Enable             bool
	HolidayProfileID   string
	Username           string `json:"Name"`
	Email              string
	MinimumWorkingHour int
	Created            time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate         time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *UserProfileExtended) TableName() string {
	return "KaraUserProfiles"
}
func (o *UserProfile) TableName() string {
	return "KaraUserProfiles"
}

func (o *UserProfile) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *UserProfile) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *UserProfile) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *UserProfile) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *UserProfile) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		if o.UserID == "" {
			return errors.New("missing: user")
		}
		o.ID = o.UserID
	}

	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *UserProfile) PostSave(dbflex.IConnection) error {
	return nil
}
