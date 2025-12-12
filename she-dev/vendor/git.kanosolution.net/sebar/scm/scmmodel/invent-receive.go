package scmmodel

import (
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventReceiveIssueLine struct {
	InventJournalLine
	InventQty       float64
	CostPerUnit     float64
	Item            tenantcoremodel.Item
	SourceType      tenantcoremodel.TrxModule
	SourceJournalID string
	SourceTrxType   string
	SourceLineNo    int

	OriginalQty float64
	SettledQty  float64

	// Vendor required fields
	DiscountType    DiscountType
	DiscountValue   float64
	DiscountAmount  float64
	DiscountGeneral PurchaseDiscount
	TaxCodes        []string
	References      tenantcoremodel.References
}

type InventReceiveIssueJournal struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	CompanyID         string                    `form_section:"General"  form:"hide"`
	TrxType           InventTrxType             `form_section:"General"`
	Name              string                    `form_required:"1" form_section:"General"` // TODO: ganti dengan Text
	ReffNo            []string                  `form_section:"General"`
	WarehouseID       string                    `form_section:"General2" form_lookup:"/tenant/warehouse/find|_id|_id,Name"`
	TrxDate           time.Time                 `form_kind:"date" form_section:"General2"`
	JournalTypeID     string                    `form_section:"General2" label:"Journal Type"`
	PostingProfileID  string                    `form_section:"General2" label:"Posting Profile" form_lookup:"/fico/postingprofile/find|_id|_id,Name"`
	SectionID         string                    `form_section:"General2" form:"hide" form_lookup:"/tenant/section/find|_id|_id,Name"`
	Status            ficomodel.JournalStatus   `form_section:"General3" form_read_only:"1"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension"`
	Lines             []InventReceiveIssueLine  `grid:"hide" form:"hide" form_section:"General3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`

	IsUsedVendor bool   `grid:"hide" form:"hide"` // flag dibutuhkan vendor: apakah sudah dipakai vendor atau belum
	Text         string `grid:"hide" form:"hide"` // biar ga error aja karena ada penambahan baru
}

type InventReceiveIssueJournalGrid struct {
	ID               string                    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_size:"4"`
	CompanyID        string                    `form_section:"General" grid:"hide"  form:"hide"`
	TrxType          InventTrxType             `form_section:"General"`
	Name             string                    `form_required:"1" form_section:"General"`
	ReffNo           []string                  `form_section:"General"`
	WarehouseID      string                    `form_section:"General2"`
	SectionID        string                    `form_section:"General2"`
	Status           ficomodel.JournalStatus   `form_section:"General2" form_read_only:"1"`
	TrxDate          time.Time                 `form_kind:"date" form_section:"General3"`
	Dimension        tenantcoremodel.Dimension `form_section:"Dimension"`
	PostingProfileID string                    `form_section:"General3" grid:"hide"`
	Lines            []InventReceiveIssueLine  `grid:"hide" form:"hide" form_section:"General3"`
	Created          time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
	LastUpdate       time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"General3"`
}

func (o *InventReceiveIssueJournal) TableName() string {
	return "InventReceiveJournals"
}

func (o *InventReceiveIssueJournal) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *InventReceiveIssueJournal) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *InventReceiveIssueJournal) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *InventReceiveIssueJournal) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *InventReceiveIssueJournal) PreSave(dbflex.IConnection) error {
	o.formatID()
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	o.Lines = lo.Map(o.Lines, func(line InventReceiveIssueLine, index int) InventReceiveIssueLine {
		line.InventDim = *line.InventDim.Calc()
		return line
	})
	return nil
}

func (o *InventReceiveIssueJournal) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *InventReceiveIssueJournal) formatID() {
	o.ID = strings.Replace(o.ID, "[GR/GI]", lo.Ternary(o.TrxType == InventIssuance, "GI 4", "GR 5"), -1)
}
