package karamodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrxRequest struct {
	UserID           string    `form_lookup:"/iam/user/find-by|_id|LoginID"`
	Op               string    `form_use_list:"1" form_items:"Checkin|Checkout|Transit" form_required:"1"`
	WorkLocationID   string    `form_lookup:"/kara/worklocation/find|_id|Name" form_required:"1"`
	TrxDate          time.Time `form_kind:"date" form_required:"1"`
	TrxTime          string    `form_kind:"time" form_required:"1"`
	Ref1             string
	ConfirmForReview bool
	Long             float64
	Lat              float64
}

type TrxGridView struct {
	ID             string `bson:"_id" json:"_id" key:"1" grid:"hide" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2" grid_sortable:"1"`
	UserID         string `label:"Profile" grid_sortable:"1" grid:"hide"`
	Name           string
	WorkLocationID string    `label:"Location" grid_sortable:"1"`
	TrxDate        time.Time `grid_sortable:"1"`
	Ref1           string    `label:"Asset"`
	Ref2           string    `label:"Nopol"`
	Op             OpCode
	Hours          float64
	Long           float64
	Lat            float64
	Message        string `label:"Note"`
	Status         OpStatus
}

type AttendanceTrx struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	UserID            string `form_lookup:"/kara/profile/find|_id|Name"`
	Name              string
	TrxDate           time.Time
	RuleID            string `form_lookup:"/kara/rule/find|_id|Name"`
	RuleLineID        string `form_lookup:"/kara/ruleline/find|_id|Name"`
	WorkLocationID    string `form_lookup:"/kara/location/find|_id|Name"`
	Op                OpCode
	Long              float64
	Lat               float64
	Hours             float64
	Ref1              string
	Ref2              string
	Ref3              string
	Message           string
	Status            OpStatus                  `form_items:"Need review|Cancelled|OK"`
	Dimension         tenantcoremodel.Dimension `grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AttendanceTrx) TableName() string {
	return "KaraAttendanceTrxs"
}

func (o *AttendanceTrx) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AttendanceTrx) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AttendanceTrx) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AttendanceTrx) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AttendanceTrx) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AttendanceTrx) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *AttendanceTrx) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "RuleIndex", Fields: []string{"RuleID", "RuleLineID"}},
		{Name: "UserIndex", Fields: []string{"UserID"}},
		{Name: "WorkLocationIndex", Fields: []string{"WorkLocationID"}},
	}
}
