package bagongmodel

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tarif struct {
	From string
	To   string
	Rate float64
}

type ConfigDeposit struct {
	TargetFlat    float64 `form_section:"General"`
	TargetNonFlat float64 `form_section:"General"`
	TargetRent    float64 `form_section:"General"`
	IsToll        bool    `form_section:"General"`
}

type ConfigPremi struct {
	Method   int     `form_items:"1|2|3"`
	Target1  float64 `form_section:"General"`
	Target2  float64 `form_section:"General"`
	Percent1 float64 `form_section:"General"`
	Percent2 float64 `form_section:"General"`
}

type Trayek struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string   `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string   `form_required:"1" form_section:"General"`
	Description       string   `form_multi_row:"4"`
	Terminals         []string `form_lookup:"/bagong/terminal/find|_id|Name" form_section:"General"`
	Tarifs            []Tarif  `grid:"hide" form:"hide"`
	ConfigDeposit     ConfigDeposit
	ConfigPremi       ConfigPremi
	Expense           []tenantcoremodel.ExpenseType `grid:"hide" form:"hide"`
	ExpenseTypeID     string                        `grid_label:"Expense Type" form_lookup:"/tenant/expensetype/find|_id|Name"`
	Dimension         tenantcoremodel.Dimension     `grid:"hide"`
	IsActive          bool                          `form_section:"General" form_section_auto_col:"2"`
	Created           time.Time                     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time                     `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *Trayek) TableName() string {
	return "BGTrayeks"
}

func (o *Trayek) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Trayek) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Trayek) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Trayek) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Trayek) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *Trayek) PostSave(dbflex.IConnection) error {
	return nil
}

// GetTariff func
// with assumptions: o.Terminals must be in correct order (lowest price to highest price)
func (o *Trayek) GetTariff(fromTerminalID, toTerminalID string, debug ...bool) (float64, error) {
	sep := "||"
	fromTID := ""
	toTID := ""
	usedTariff := float64(0)
	defer func() {
		if len(debug) > 0 && debug[0] {
			fmt.Printf("# GetTariff FROM: %s, TO: %s | Will Use Tariff FROM: %s, TO: %s | TARIFF: %v\n", fromTerminalID, toTerminalID, fromTID, toTID, usedTariff)
		}
	}()

	tarifMap := make(map[string]float64)
	for _, t := range o.Tarifs {
		tarifMap[strings.ToLower(fmt.Sprintf("%s%s%s", t.From, sep, t.To))] = t.Rate
	}

	// loop backwards Destination Terminals
	currentToTID := toTerminalID
	for i := len(o.Terminals) - 1; i >= 0; i-- {
		if currentToTID == o.Terminals[i] {
			if rate, ok := tarifMap[strings.ToLower(fmt.Sprintf("%s%s%s", fromTerminalID, sep, currentToTID))]; ok && rate > 0 {
				fromTID = fromTerminalID
				toTID = currentToTID
				usedTariff = rate
				return rate, nil // designated rate already set
			}

			if rate, ok := tarifMap[strings.ToLower(fmt.Sprintf("%s%s%s", currentToTID, sep, fromTerminalID))]; ok && rate > 0 {
				fromTID = currentToTID
				toTID = fromTerminalID
				usedTariff = rate
				return rate, nil // the opposite rate
			}

			if i > 0 {
				currentToTID = o.Terminals[(i - 1)] // get previous terminal
			}
		}
	}

	// loop backwards Departure Terminals
	currentFromTID := fromTerminalID
	for i := len(o.Terminals) - 1; i >= 0; i-- {
		if currentFromTID == o.Terminals[i] {
			if rate, ok := tarifMap[strings.ToLower(fmt.Sprintf("%s%s%s", currentFromTID, sep, toTerminalID))]; ok && rate > 0 {
				fromTID = currentFromTID
				toTID = toTerminalID
				usedTariff = rate
				return rate, nil // designated rate already set
			}

			if rate, ok := tarifMap[strings.ToLower(fmt.Sprintf("%s%s%s", toTerminalID, sep, currentFromTID))]; ok && rate > 0 {
				fromTID = toTerminalID
				toTID = currentFromTID
				usedTariff = rate
				return rate, nil // the opposite rate
			}

			if i > 0 {
				currentFromTID = o.Terminals[(i - 1)] // get previous terminal
			}
		}
	}

	return 0, nil
}
