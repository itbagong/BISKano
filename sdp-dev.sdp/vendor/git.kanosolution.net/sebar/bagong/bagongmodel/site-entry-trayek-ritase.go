package bagongmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RitasePassenger struct {
	Ritase   int
	From     string
	FromName string
	To       string
	ToName   string
	Total    int
	Tariff   float64
}

type RitaseIncome struct {
	Ritase         int
	TerminalID     string
	TerminalName   string
	Amount         float64
	TotalPassenger int
}

type SiteExpenseTerminal struct {
	TerminalID      string
	TerminalName    string
	ExpenseID       string
	ExpenseName     string
	ExpenseCategory string
	ExpenseValue    float64
	Value           float64
	Amount          float64
	ApprovalStatus  string
	JournalID       string
	VoucherID       string
}

type RitaseSummary struct {
	TotalRitasePassenger float64
	TotalRitaseIncome    float64
	TotalOtherIncome     float64
	TotalTerminalExpense float64
	TotalOtherExpense    float64
	TotalFixExpense      float64
	TotalBonus           float64
	TotalMethod          float64
}

type TicketNumber struct {
	Start string
	End   string
}

type SiteEntryTrayekRitase struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string            `form:"hide" bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"1" form_section_show_title:"0"`
	SiteEntryAssetID  string            `grid:"hide" form:"hide" form_section:"General"`
	Income            float64           `form_section:"General" form:"hide"`
	Expense           float64           `form_section:"General" form:"hide"`
	Revenue           float64           `form_section:"General" form:"hide"`
	RevenueType       string            `form_section:"General" form_items:"Premi|Setoran"  form_pos:"1,1"`
	ConfigDeposit     ConfigDeposit     `form_section:"General2" form_pos:"2,1" `
	TrayekName        string            `grid:"hide" form:"hide"`
	CategoryDeposit   string            `form_section:"General2" form_items:"Flat|Non Flat|Rent"  form_pos:"2,2"  form_label:"Type"`
	RitaseDeposit     int               `form_section:"General2" form_pos:"2,3" form_label:"Ritase"`
	AmountDeposit     float64           `form_section:"General2" form_pos:"3,1"`
	IsFullDeposit     bool              `form_section:"General2" form_pos:"3,2"`
	ConfigPremi       ConfigPremi       `form:"hide"`
	RitaseSummary     RitaseSummary     `grid:"hide" form:"hide"`
	RitasePremi       []int             `form_section:"General" form:"hide"`
	ShiftID           string            `form_section:"General"`
	TicketNumbers     []TicketNumber    `form_section:"Ticket Number" grid:"hide" form_section_show_title:"1"`
	Kurs              int               `grid:"hide" form:"hide"`
	StartKM           int               `grid:"hide" form_section:"Kilometer" form_section_auto_col:"2" form_label:"Start KM" form_required:"1"`
	EndKM             int               `grid:"hide" form_section:"Kilometer" form_label:"End KM" form_required:"1"`
	RitasePassenger   []RitasePassenger `grid:"hide" form_section:"Terminal Passenger" form_section_auto_col:"1"  form_section_show_title:"1"`
	RitaseIncome      []RitaseIncome    `grid:"hide" form:"hide"  form_section:"Terminal Passenger" form_section_auto_col:"1"  form_section_show_title:"0"`
	OtherIncome       []SiteIncome      `grid:"hide" form_section:"Other Income" form_section_auto_col:"1"  form_section_show_title:"1"`
	PassengerIncome   SiteIncome        `grid:"hide" form:"hide" form_section:"General" form_section_auto_col:"1"  form_section_show_title:"1"`
	DepositIncome     SiteIncome        `grid:"hide" form:"hide" form_section:"General" form_section_auto_col:"1"  form_section_show_title:"1"`
	// TerminalExpense   []SiteExpenseTerminal `grid:"hide" form:"hide" form_section:"Terminal Expense" form_section_auto_col:"1"`
	OtherExpense []SiteExpense `grid:"hide" form_section:"Expense" form_section_auto_col:"1"  form_section_show_title:"1"`
	FixExpense   []SiteExpense `grid:"hide" form_section:"Fix Expense" form_section_auto_col:"1"  form_section_show_title:"1"`
	Created      time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"  form_section_show_title:"1"`
	LastUpdate   time.Time     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *SiteEntryTrayekRitase) TableName() string {
	return "BGSiteEntryTrayekRitases"
}

func (o *SiteEntryTrayekRitase) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *SiteEntryTrayekRitase) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *SiteEntryTrayekRitase) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *SiteEntryTrayekRitase) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *SiteEntryTrayekRitase) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *SiteEntryTrayekRitase) PostSave(dbflex.IConnection) error {
	return nil
}
