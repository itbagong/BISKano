package ficomodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
)

type SubmissionEmployeeLedgerJournal struct {
	orm.DataModelBase    `bson:"-" json:"-"`
	CompanyID            string                     `form_lookup:"/tenant/company/find|_id|Name" grid:"hide" form:"hide" form_read_only_edit:"1" form_section:"General"`
	ID                   string                     `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General"`
	JournalTypeID        string                     `form_lookup:"/fico/ledgerjournaltype/find|_id|_id,Name" form_read_only:"1"  form_section:"General" form_required:"1" grid:"hide"`
	VendorID             string                     `form_lookup:"/tenant/vendor/find|_id|_id,Name" form_read_only:"1"  form_section:"General" form_required:"1"`
	TrxDate              time.Time                  `form_kind:"date" form_section:"Other Information" grid_sortable:"1" form_space_after:"1"`
	EmployeeID           string                     `grid:"hide" form:"hide" label:"Employee" form_required:"1" form_section:"Other Information" form_multi_row:"5"  form_lookup:"/tenant/employee/find|_id|Name"`
	Text                 string                     `form_required:"1" form_section:"General" form_multi_row:"5"  form_section_size:"3"`
	TaxCodes             []string                   `form_lookup:"/fico/taxcode/find|_id|Name" form_section_show_title:"1" grid:"hide"`
	DefaultOffset        SubledgerAccount           `grid:"hide" form:"hide" form_section:"Other Information"`
	ReferenceTemplateID  string                     `grid:"hide" form:"hide"  form_lookup:"/tenant/referencetemplate/find|_id|_id,Name" form_section:"Other Information"`
	ChecklistTemplateID  string                     `grid:"hide" form:"hide" form_lookup:"/tenant/checklisttemplate/find|_id|_id,Name" form_section:"Other Information"`
	References           tenantcoremodel.References `grid:"hide" form:"hide" form_section:"General3" form_section_size:"3"`
	Dimension            tenantcoremodel.Dimension  `grid:"hide" form_section:"Dimension" form_section_size:"4"`
	HeaderDiscountType   string                     `form_section:"Other Information" form_items:"fixed|percent" grid:"hide"`
	HeaderDiscountValue  float64                    `form_section:"Other Information" grid:"hide"`
	Checklists           tenantcoremodel.Checklists `grid:"hide" form:"hide" form_section:"General3"`
	Lines                []JournalLine              `form:"hide" grid:"hide" form_section:"General3"`
	PostingProfileID     string                     `form:"hide" grid:"hide"`
	TotalAmount          float64                    `form_section:"Amount" label:"Grand total amount" form_read_only:"1" form_section_auto_col:"2" form_section_show_title:"1" grid:"hide"`
	SubtotalAmount       float64                    `form_section:"Amount" form_read_only:"1" grid:"hide"`
	PriceTotalAmount     float64                    `form_section:"Amount" form_read_only:"1" grid:"hide"`
	TaxAmount            float64                    `form_section:"Amount" form_read_only:"1" grid:"hide"`
	DiscountAmount       float64                    `form_section:"Amount" label:"Line discount amount" form_read_only:"1" grid:"hide"`
	PPNAmount            float64                    `form_section:"Amount" label:"PPN Amount" form_read_only:"1" grid:"hide"`
	HeaderDiscountAmount float64                    `form_section:"Amount" form_read_only:"1" grid:"hide"`
	PPHAmount            float64                    `form_section:"Amount" label:"PPh Amount" form_read_only:"1" grid:"hide"`
	Status               JournalStatus              `form_read_only:"1" form_section:"Info" form_section_auto_col:"2"`
	Created              time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
	LastUpdate           time.Time                  `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Info"`
}

func (o *SubmissionEmployeeLedgerJournal) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Other Information", ShowTitle: true, AutoCol: 2},
			{Title: "Info", ShowTitle: true, AutoCol: 2},
		}},
		{Sections: []suim.FormSection{
			{Title: "Dimension", ShowTitle: true, AutoCol: 1},
			{Title: "Amount", ShowTitle: true, AutoCol: 2},
		}},
	}
}
