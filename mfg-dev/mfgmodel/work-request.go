package mfgmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WRStatusApproval string
type WRSourceType string

var (
	WRApproved               WRStatusApproval = "Approved"
	WRWaiting                WRStatusApproval = "Waiting"
	WRSourceTypeAssembly     WRSourceType     = "Assembly"
	WRSourceTypeRoutineCheck WRSourceType     = "Routine Check"
	WRSourceTypeSalesOrder   WRSourceType     = "Sales Order"
	WRSourceTypeNonRoutine   WRSourceType     = "Non-Routine"
)

type WorkRequest struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" label:"WR No" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"1" form_section_size:"4" `
	Name              string `form_section:"General" form_required:"1" label:"Requestor" form_lookup:"/tenant/employee/find|_id|Name"` // Requestor ID
	// Department        string                            `form_section:"General" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=DME|_id|Name" form_read_only:"1"`
	TrxDate          *time.Time                        `form_kind:"date" form_section:"General" label:"WR Date"`
	TrxType          scmmodel.InventTrxType            `form_section:"General" grid:"hide" form:"hide"`
	CompanyID        string                            `form:"hide" grid:"hide" form_lookup:"/tenant/company/find|_id|_id,Name"`
	SourceType       WRSourceType                      `form_section:"General" grid:"hide" form_items:"Item Request|Routine Check|Sales Order|Non-Routine"`
	SourceID         string                            `form_section:"General" grid:"hide" label:"Ref No."`
	JournalTypeID    string                            `form_section:"General2" grid:"hide" form_lookup:"/mfg/workrequestor/journal/type/find|_id|_id,Name" form_read_only:"1"`
	PostingProfileID string                            `form:"hide" form_section:"General2" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	WorkRequestType  string                            `form_section:"General2" grid:"hide" form_items:"Production|Service"`
	EquipmentType    string                            `form_section:"General2" label:"Asset Type" form_lookup:"/tenant/assetgroup/find|_id|Name"`
	EquipmentNo      string                            `form_section:"General2" form_lookup:"/tenant/asset/find?GroupID=UNT|_id|_id,Name" label:"Asset No"`
	Priority         string                            `form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IRPriority|_id|Name"`
	StatusApproval   WRStatusApproval                  `form_section:"General2" grid:"hide" form:"hide" form_read_only:"1" form_items:"Approved|Waiting"`
	StartDownTime    *time.Time                        `form_kind:"datetime-local" form_section:"General3" grid:"hide"`
	TargetFinishTime *time.Time                        `form_kind:"datetime-local" form_section:"General3" grid:"hide"`
	Kilometers       float64                           `form_section:"General3" grid:"hide"`
	Description      string                            `form_section:"General3" grid:"hide" form_multi_row:"4"`
	Dimension        tenantcoremodel.Dimension         `form_section:"Dimension" label:"Site"`
	Lines            []scmmodel.InventReceiveIssueLine `grid:"hide" form:"hide"` //dummy
	Status           ficomodel.JournalStatus           `form_section:"General3" form_read_only:"1"`
	Created          time.Time                         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate       time.Time                         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`

	Text string `grid:"hide" form:"hide"` // biar ga error aja karena ada penambahan baru
}

type WorkRequestGrid struct {
	ID               string                            `bson:"_id" json:"_id" key:"1" label:"WR No" form_read_only_edit:"1" form_section_size:"4" form_section:"General1"`
	Name             string                            `form_section:"General1" form_required:"1" label:"Requestor"` // Requestor ID
	Department       string                            `form_section:"General1"`
	TrxDate          *time.Time                        `form_kind:"date" form_section:"General1" label:"WR Date"`
	TrxType          scmmodel.InventTrxType            `form_section:"General1" grid:"hide" form:"hide"`
	SourceType       WRSourceType                      `form_section:"General1" grid:"hide" form_items:"Item Request|Routine Check|Sales Order|Non-Routine"`
	SourceID         string                            `form_section:"General1" label:"Ref No."`
	JournalTypeID    string                            `form_section:"General2" grid:"hide"`
	PostingProfileID string                            `form_section:"General2" grid:"hide"`
	WorkRequestType  string                            `form_section:"General2" grid:"hide" form_items:"Production|Service"`
	EquipmentType    string                            `form_section:"General2" label:"Asset Type" form_lookup:"/tenant/assetgroup/find|_id|Name"`
	EquipmentNo      string                            `form_section:"General2" label:"Police No"`
	Priority         string                            `form_section:"General2" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=IRPriority|_id|Name"`
	StatusApproval   WRStatusApproval                  `form_section:"General2" form_read_only:"1" form_items:"Approved|Waiting" grid:"hide"`
	StartDownTime    *time.Time                        `form_kind:"datetime-local" form_section:"General3" grid:"hide"`
	TargetFinishTime *time.Time                        `form_kind:"datetime-local" form_section:"General3" grid:"hide"`
	Kilometers       float64                           `form_section:"General3" grid:"hide"`
	Description      string                            `form_section:"General3" grid:"hide" form_multi_row:"4"`
	Dimension        tenantcoremodel.Dimension         `form_section:"Dimension" label:"Site"`
	Status           ficomodel.JournalStatus           `form_section:"General2" form_read_only:"1"`
	Lines            []scmmodel.InventReceiveIssueLine `grid:"hide" form:"hide"` //dummy
	Created          time.Time                         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate       time.Time                         `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *WorkRequest) TableName() string {
	return "WorkRequests"
}

func (o *WorkRequest) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkRequest) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkRequest) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkRequest) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkRequest) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkRequest) PostSave(dbflex.IConnection) error {
	return nil
}
