package sdpmodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SalesPriceBookLine struct {
	ID               string `bson:"_id" json:"_id" key:"1"`
	SalesPriceBookID string
	AssetID          string
	AssetType        string
	ProductionYear   int
	ItemID           string    `form_lookup:"/tenant/item/find|_id|Name"`
	MinPrice         float32   `form_required:"1"`
	MaxPrice         float32   `form_required:"1"`
	MaxDiscount      float32   `form_label:"Max Discount(%)"`
	SalesPrice       float32   `grid_label:"Price"`
	Quantity         int       `form_required:"1"`
	Unit             string    `form_required:"1" form_lookup:"/tenant/unit/find|_id|Name"`
	Created          time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate       time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

type SalesPriceBookLineForm struct {
	ProductionYear int                       ``
	AssetType      string                    ``
	AssetID        string                    ``
	ItemID         string                    `form_lookup:"/tenant/item/find|_id|Name"`
	MinPrice       float32                   `form_label:"Min Price" form_required:"1"`
	MaxPrice       float32                   `form_label:"Max Price" form_required:"1"`
	Quantity       int                       `form_required:"1"`
	MaxDiscount    float32                   `form_label:"Max Discount(%)"`
	SalesPrice     float32                   `form_label:"Sales Price"`
	Unit           string                    `form_required:"1" form_lookup:"/tenant/unit/find|_id|Name"`
	Dimension      tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1" form_read_only:"1"`
}

type SalesPriceBookLineGrid struct {
	ProductionYear int     ``
	AssetType      string  ``
	AssetID        string  ``
	ItemID         string  `form_lookup:"/tenant/item/find|_id|Name"`
	MinPrice       float32 `grid_label:"Min Price" form_required:"1"`
	MaxPrice       float32 `grid_label:"Max Price" form_required:"1"`
	MaxDiscount    float32 `grid_label:"Max Discount(%)"`
	SalesPrice     float32 `grid_label:"Sales Price"`
	Quantity       int     `form_required:"1"`
	Unit           string  `form_required:"1" form_lookup:"/tenant/unit/find|_id|Name"`
}

type SalesPriceBook struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" key:"1" form_read_only_edit:"1" form:"hide" form_section:"Dimension1"`
	Name              string                    `form_required:"1" form_pos:"1,1" form_section:"Dimension1"`
	StartPeriod       time.Time                 `form_kind:"date" form_pos:"1,2" form_section:"Dimension1"`
	EndPeriod         time.Time                 `form_kind:"date" form_pos:"1,3" form_section:"Dimension1"`
	Dimension         tenantcoremodel.Dimension `grid:"hide" form_pos:"1,4" form_section:"Dimension1"`
	Lines             []SalesPriceBookLine      `form:"hide" grid:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" form:"hide" grid:"hide" form_section:"Time Info"`
}

func (o *SalesPriceBook) TableName() string {
	return "SalesPriceBook"
}

func (o *SalesPriceBook) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SalesPriceBook) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SalesPriceBook) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SalesPriceBook) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SalesPriceBook) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}

	if len(o.Lines) > 0 {
		for _, line := range o.Lines {

			if line.SalesPrice < line.MinPrice && line.SalesPrice > line.MaxPrice {
				return errors.New("Alert!! Price required around Minimum Price and Maximum Price")
			}

			if line.MinPrice > line.MaxPrice {
				return errors.New("Alert!! Minimum Price can't higher than Maximum Price")
			}
		}
	}

	o.LastUpdate = time.Now()
	return nil
}

func (o *SalesPriceBook) PostSave(dbflex.IConnection) error {
	return nil
}
