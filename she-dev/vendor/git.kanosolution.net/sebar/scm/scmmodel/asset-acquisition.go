package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetAcquisitionJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`

	CompanyID        string    `form_section:"General" form_lookup:"tenant/company/find|_id|_id,Name" form:"hide"`
	TrxDate          time.Time `form_kind:"date" form_section:"General"`
	JournalTypeID    string    `form_required:"1" form_lookup:"/scm/asset-acquisition/journal/type/find|_id|_id,Name" form_section:"General"`
	PostingProfileID string    `form_section:"General" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`

	TransferName string                    `form_section:"General2"`
	TransferDate time.Time                 `form_kind:"date" form_section:"General2"`
	Status       ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	Dimension    tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension1"`
	TransferFrom InventDimension           `grid:"hide" form_section:"Dimension2"`

	ItemTranfers   []AssetItemTransfer `grid:"hide" form:"hide"`
	AssetRegisters []AssetRegister     `grid:"hide" form:"hide"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`

	Text string // biar ga error aja karena ada penambahan baru
}

type AssetItemTransfer struct {
	InventJournalLine
	Item         tenantcoremodel.Item
	SourceType   InventTrxType `form_items:"Purchase Request|Work Order"`
	SourceReffNo string
}

type AssetRegister struct {
	LineNo                      int `grid:"hide"`
	ItemID                      string
	SKU                         string `label:"SKU"`
	AssetGroup                  string
	DoesFixedAssetNumberIsExist bool
	AssetID                     string
}

func (o *AssetAcquisitionJournal) TableName() string {
	return "AssetAcquisitions"
}

func (o *AssetAcquisitionJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetAcquisitionJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetAcquisitionJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetAcquisitionJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetAcquisitionJournal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetAcquisitionJournal) PostSave(dbflex.IConnection) error {
	return nil
}
