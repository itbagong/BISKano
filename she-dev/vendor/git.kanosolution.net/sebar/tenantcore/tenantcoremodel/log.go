package tenantcoremodel

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/datahub"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogModule string
type LogMenu string
type LogAction string

const (
	// Modules
	LogModuleInventory LogModule = "INVENTORY"
	LogModulePurchase  LogModule = "PURCHASE"
	LogModuleWorkOrder LogModule = "WORKORDER"

	// Menus
	LogMenuItem             LogMenu = "Item"
	LogMenuItemSpec         LogMenu = "ItemSpec"
	LogMenuGoodReceive      LogMenu = "Good Receive"
	LogMenuGoodIssuance     LogMenu = "Good Issuance"
	LogMenuTransfer         LogMenu = "Transfer"
	LogMenuItemRequest      LogMenu = "Item Request"
	LogMenuAssetAcquisition LogMenu = "Asset Acquisition"
	LogMenuMovementInOut    LogMenu = "Movement In/Out"

	LogMenuPurchaseOrder   LogMenu = "Purchase Order"
	LogMenuPurchaseRequest LogMenu = "Purchase Request"

	// Actions
	LogActionCreate LogAction = "create"
	LogActionUpdate LogAction = "update"
	LogActionDelete LogAction = "delete"
)

var (
	// LogMenuMap untuk convert dari Journal Type ke nama menu (optional)
	LogMenuMap = map[string]LogMenu{
		// INVENTORY
		"Inventory Receive":  LogMenuGoodReceive,
		"Inventory Issuance": LogMenuGoodIssuance,
		"Item Request":       LogMenuItemSpec,
		"INVENTORY":          LogMenuMovementInOut,
	}

	// LogModuleMap untuk menentukan module dari menu
	LogModuleMap = map[LogMenu]LogModule{
		// INVENTORY
		LogMenuItem:             LogModuleInventory,
		LogMenuItemSpec:         LogModuleInventory,
		LogMenuGoodReceive:      LogModuleInventory,
		LogMenuGoodIssuance:     LogModuleInventory,
		LogMenuTransfer:         LogModuleInventory,
		LogMenuItemRequest:      LogModuleInventory,
		LogMenuAssetAcquisition: LogModuleInventory,
		LogMenuMovementInOut:    LogModuleInventory,

		// PURCHASE
		LogMenuPurchaseOrder:   LogModulePurchase,
		LogMenuPurchaseRequest: LogModulePurchase,
	}
)

type Log struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Module            LogModule `form_section:"General" form_section_auto_col:"2"`
	Menu              LogMenu   `form_section:"General" form_section_auto_col:"2"`
	Action            LogAction `form_section:"General" form_section_auto_col:"2"`
	ValueData         string    `form_section:"General" form_section_auto_col:"2"`
	Datetime          time.Time `form_kind:"datetime" form_section:"General" form_section_auto_col:"2"`
	UserLogin         string    `form_section:"General" form_section_auto_col:"2"`
}

type LogParam struct {
	Hub           *datahub.Hub
	Module        string
	Menu          string
	Action        string
	TransactionID string
	Name          string
	UserLogin     string
}

func (o *Log) TableName() string {
	return "Logs"
}

func (o *Log) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *Log) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *Log) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *Log) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *Log) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Datetime.IsZero() {
		o.Datetime = time.Now()
	}
	return nil
}

func (o *Log) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *Log) Indexes() []dbflex.DbIndex {
	return []dbflex.DbIndex{
		{Name: "Module", Fields: []string{"Module"}},
		{Name: "Menu", Fields: []string{"Menu"}},
	}
}

func (or *Log) Add(param LogParam) error {
	l := new(Log)
	l.Action = LogAction(param.Action)
	l.UserLogin = param.UserLogin

	// translate menu to LogMenu if any
	m, ok := LogMenuMap[param.Menu]
	if !ok {
		m = LogMenu(param.Menu)
	}
	l.Menu = m

	if param.Module == "" {
		// get module from menu if param module is empty
		l.Module = LogModuleMap[LogMenu(l.Menu)]
	}

	valDatas := []string{}
	if param.TransactionID != "" {
		valDatas = append(valDatas, param.TransactionID)
	}
	if param.Name != "" {
		valDatas = append(valDatas, param.Name)
	}

	l.ValueData = fmt.Sprintf("%s", strings.Join(valDatas, "&&"))

	return param.Hub.Save(l)
}

func (o *Log) AddMultiple(params []LogParam) error {
	for _, param := range params {
		err := o.Add(param)
		if err != nil {
			return err
		}
	}
	return nil
}
