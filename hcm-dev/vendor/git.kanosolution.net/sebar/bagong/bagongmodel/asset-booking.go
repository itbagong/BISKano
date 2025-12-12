package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetBookingLine struct {
	FromDate      string `form_kind:"date"`
	ToDate        string `form_kind:"date"`
	Specification string `label:"Spec"`
	Description   string
	UnitBooked    int
	UnitPrice     float64
	Total         float64
	UnitAllocated int
}

// todo: some column auto generated
type AssetBooking struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"3"`
	SiteID            string                    `form_lookup:"/bagong/sitesetup/find|_id|Name" label:"Site" grid:"hide" form:"hide"`
	CustomerID        string                    `form_section:"General" form_lookup:"/tenant/customer/find|_id|Name" grid:"hide"`
	FromDate          string                    `form_kind:"date" form_section:"General"`
	ToDate            string                    `form_kind:"date" form_section:"General"`
	Unit              int                       `grid:"hide" form:"hide"`
	Amount            float64                   `form_section:"General2" form_section_size:"3"`
	Text              string                    `form_section:"General2"`
	Status            string                    `form_section:"General2"`
	Lines             []AssetBookingLine        `grid:"hide" form:"hide"`
	UnitBooked        int                       `form_section:"General3" form_section_size:"3"`
	UnitAllocated     int                       `form_section:"General3"`
	JournalID         string                    `grid_label:"Journal No" form_section:"General3"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *AssetBooking) FormSections() []suim.FormSectionGroup {
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

func (o *AssetBooking) TableName() string {
	return "BGAssetBookings"
}

func (o *AssetBooking) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *AssetBooking) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *AssetBooking) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *AssetBooking) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *AssetBooking) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *AssetBooking) PostSave(dbflex.IConnection) error {
	return nil
}
