package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChequeGiroBook struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string                    `bson:"_id" json:"_id" key:"1" grid:"hide" form_read_only_new:"1" form_read_only_edit:"1" form_section:"General" form_section_direction:"row" form_section_size:"3"`
	Name              string                    `form_required:"1" form_section:"General" grid:"hide"`
	Kind              CGType                    `form_label:"Sequence Type" grid_label:"Sequence Type" grid_pos:"3" form_lookup:"/tenant/masterdata/find?MasterDataTypeID=CGT|_id|Name"`
	ReceiveDate       *time.Time                `form:"hide" grid:"hide"`
	From              string                    `form_label:"Sequence Initial No." grid_label:"Sequence Initial No." grid_pos:"4"`
	To                string                    `form_label:"Sequence Last No." grid_label:"Sequence Last No." grid_pos:"5"`
	CashBookID        string                    `form_label:"Bank Account No" grid_label:"Bank Account No" grid_pos:"2" form_lookup:"/tenant/cashbank/find|_id|Name"`
	Qty               int                       `grid:"hide" form:"hide"`
	CompanyID         string                    `form:"hide" grid:"hide"`
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" form_section_direction:"row" grid:"hide" form_section_size:"3"`
	Created           time.Time                 `form_kind:"datetime" form_read_only:"1" grid_pos:"1" form_section:"Time Info" form_section_auto_col:"2" form_section_size:"1"`
	LastUpdate        time.Time                 `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *ChequeGiroBook) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 1},
			{Title: "Dimension", ShowTitle: false, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *ChequeGiroBook) TableName() string {
	return "CGBooks"
}

func (o *ChequeGiroBook) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *ChequeGiroBook) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *ChequeGiroBook) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *ChequeGiroBook) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *ChequeGiroBook) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *ChequeGiroBook) PostSave(dbflex.IConnection) error {
	return nil
}
