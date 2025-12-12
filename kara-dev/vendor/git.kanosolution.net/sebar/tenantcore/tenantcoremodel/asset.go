package tenantcoremodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Asset struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	GroupID             string    `form_section:"General" form_read_only_edit:"1" form_lookup:"/tenant/assetgroup/find|_id|_id,Name"`
	ID                  string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_read_only_new:"1" form_section_direction:"row" form_section:"General" form_section_size:"3"`
	Name                string    `form_required:"1" form_section:"General"`
	AssetType           string    `form_section:"General"`
	AcquisitionAccount  string    `form_section:"General2" form_section_direction:"row" form_section_size:"3" form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	DepreciationAccount string    `form_section:"General2" form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	DisposalAccount     string    `form_section:"General2" form_section_direction:"row" form_section_size:"3" form_lookup:"/tenant/ledgeraccount/find|_id|_id,Name"`
	AdjustmentAccount   string    `form_section:"General3"`
	Dimension           Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created             time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate          time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Asset) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "General2", ShowTitle: false, AutoCol: 1},
			{Title: "General3", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},

		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *Asset) TableName() string {
	return "Assets"
}

func (o *Asset) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Asset) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Asset) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Asset) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Asset) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Asset) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Asset) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "GroupIndex", Fields: []string{"AssetGroupID"}},
	}
}
