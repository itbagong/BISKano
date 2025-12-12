package scmmodel

import (
	"errors"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	Text              string                    `form_required:"1" form_section:"General"`
	JournalTypeID     string                    `form_section:"General" grid:"hide"`
	PostingProfileID  string                    `form_section:"General" grid:"hide" form:"hide"`
	TrxDate           time.Time                 `form_kind:"date" form_section:"General"`
	ETA               time.Time                 `form_kind:"date" form_section:"General" label:"ETA"`
	TrxType           InventTrxType             `form_section:"General2" grid:"hide" form:"hide" form_items:"Movement In|Movement Out|Transfer|Stock Opname"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" grid:"hide"`
	InventDim         InventDimension           `form_section:"InventDim" grid:"hide"`
	InventDimTo       InventDimension           `form_section:"InventDim" grid:"hide"`
	ReffNo            []string                  `form_section:"General" form_read_only:"1"`
	Lines             []InventJournalLine       `grid:"hide" form:"hide"`
	AttachmentID      string                    `grid:"hide" form:"hide"`
	DeliveryService   string                    `form_section:"General2"`
	Status            ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	CompanyID         string                    `form_section:"General2" form_lookup:"/tenant/company/find|_id|_id,Name" grid:"hide"  form:"hide"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

type InventJournalGrid struct {
	ID               string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	Text             string                    `form_required:"1" form_section:"General"`
	JournalTypeID    string                    `form_section:"General" grid:"hide"`
	PostingProfileID string                    `form_section:"General" grid:"hide"`
	TrxDate          time.Time                 `form_kind:"date" form_section:"General"`
	TrxType          InventTrxType             `form_section:"General2" grid:"hide" form:"hide" form_items:"Movement In|Movement Out|Transfer|Stock Opname"`
	Dimension        tenantcoremodel.Dimension `form_section:"Dimension" grid:"hide"`
	InventDim        InventDimension           `form_section:"InventDim" grid:"hide"`
	InventDimTo      InventDimension           `form_section:"InventDim" grid:"hide"`
	Lines            []InventJournalLine       `grid:"hide" form:"hide"`
	AttachmentID     string                    `grid:"hide" form:"hide"`
	ETA              time.Time                 `form_kind:"date" form_section:"General2" label:"ETA"`
	DeliveryService  string                    `form_section:"General2"`
	CompanyID        string                    `form_section:"General2"  form:"hide"`
	Status           ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	Created          time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
	LastUpdate       time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General2"`
}

func (o *InventJournal) TableName() string {
	return "InventJournals"
}

func (o *InventJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventJournal) PreDelete(dbflex.IConnection) error {
	if o.Status != "DRAFT" {
		return errors.New("protected record, could not be deleted")
	}

	return nil
}

func (o *InventJournal) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	o.InventDim = *o.InventDim.Calc()
	o.InventDimTo = *o.InventDimTo.Calc()
	o.Lines = lo.Map(o.Lines, func(line InventJournalLine, index int) InventJournalLine {
		line.InventDim = *line.InventDim.Calc()
		return line
	})
	return nil
}

func (o *InventJournal) PostSave(dbflex.IConnection) error {
	return nil
}
