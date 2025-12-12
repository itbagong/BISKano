package scmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockOpnameJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	Name              string                    `form_required:"1" form_section:"General"`
	JournalTypeID     string                    `form_section:"General" grid:"hide"`
	PostingProfileID  string                    `form_section:"General" grid:"hide" form:"hide"`
	TrxDate           time.Time                 `form_kind:"date" form_section:"General"`
	TrxType           InventTrxType             `form_section:"General2" grid:"hide" form:"hide" form_items:"Movement In|Movement Out|Transfer|Stock Opname"`
	Status            JournalStatus             `form_section:"General2" form_read_only:"1"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" grid:"hide"`
	InventDim         InventDimension           `form_section:"InventDim" grid:"hide"`
	Lines             []StockOpnameJournalLine  `grid:"hide" form:"hide"`
	CompanyID         string                    `form_section:"General2" form_lookup:"/tenant/company/find|_id|_id,Name" form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" form_section:"General2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

type StockOpnameJournalLine struct {
	InventJournalLine
	Description string
	QtyInSystem float64                 `form_label:"Qty In System"`
	QtyActual   float64                 `form_label:"Qty Actual"`
	Gap         float64                 // QtyActual - QtyInSystem
	Remarks     StockOpnameDetailRemark `form_multi_row:"2"`
	Note        string                  `form_multi_row:"2"`
	NoteStaff   string                  `form_multi_row:"2" form_label:"Staff Note"`
}

func (o *StockOpnameJournal) TableName() string {
	return "StockOpnameJournals"
}

func (o *StockOpnameJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *StockOpnameJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *StockOpnameJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *StockOpnameJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *StockOpnameJournal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	o.InventDim = *o.InventDim.Calc()
	// o.InventDimTo = *o.InventDimTo.Calc()
	// o.Lines = lo.Map(o.Lines, func(line StockOpnameJournalLine, index int) StockOpnameJournalLine {
	// 	line.InventDim = *line.InventDim.Calc()
	// 	return line
	// })
	return nil
}

func (o *StockOpnameJournal) PostSave(dbflex.IConnection) error {
	return nil
}
