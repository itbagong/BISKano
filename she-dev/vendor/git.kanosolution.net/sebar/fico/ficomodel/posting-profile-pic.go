package ficomodel

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostingUsers struct {
	UserIDs              []string `json:"UserIDs" bson:"UserIDs" form_lookup:"/iam/user/find|_id|DisplayName,Email" label:"User ID"`
	GroupIds             []string `json:"GroupIDs" bson:"GroupIDs"`
	MinimalApproverCount int
}

type PostingProfileAccountSelector string

type PostingProfileAmountType string

const (
	AmountAll             PostingProfileAmountType = "Total amount"
	AmountMaxPricePerUnit PostingProfileAmountType = "Max price"
)

type PostingProfileAccount struct {
	AccountType tenantcoremodel.TrxModule
	IsRange     bool
	AccountIDs  []string
}

func (p *PostingProfileAccount) IsValid(id string) bool {
	if p.IsRange {
		if len(p.AccountIDs) < 2 {
			return false
		}
		return strings.Compare(id, p.AccountIDs[0]) >= 0 && strings.Compare(id, p.AccountIDs[1]) <= 0
	}

	if len(p.AccountIDs) == 0 {
		return true
	}

	return lo.IndexOf(p.AccountIDs, id) >= 0
}

type PostingProfilePIC struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only:"1" form_section:"General" form:"hide" form_section_auto_col:"2" grid:"hide"`
	Name              string
	PostingProfileID  string `grid:"hide" form:"hide" form_read_only:"1" form_section_size:"2"`
	Priority          int
	Account           PostingProfileAccount
	Dimension         tenantcoremodel.Dimension `form_section:"Dimension" form_section_show_title:"1"`
	Submitters        []PostingUsers            `form_section:"Approvers" grid:"hide" form:"hide" label:"User to submit"`
	Postingers        []PostingUsers            `form_section:"Approvers" grid:"hide" form:"hide" label:"User to post"`
	Approvers         []PostingUsers            `form_section:"Approvers" grid:"hide" form:"hide" label:"User to review"`
	UseRange          bool                      `form_section:"Range" form_section_show_title:"1"`
	AmountType        PostingProfileAmountType  `form_section:"Range" form_items:"Total amount|Max price"`
	LowAmount         float64                   `form_section:"Range"`
	HiAmount          float64                   `form_section:"Range"`
	Exclusive         bool
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PostingProfilePIC) TableName() string {
	return "PostingProfilePICs"
}

func (o *PostingProfilePIC) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PostingProfilePIC) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PostingProfilePIC) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PostingProfilePIC) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PostingProfilePIC) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PostingProfilePIC) PostSave(dbflex.IConnection) error {
	return nil
}

/*
func (o *PostingProfilePIC) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General"},
			{Title: "Dimension", ShowTitle: true},
			{Title: "Range", ShowTitle: true},
		}},
		{Sections: []suim.FormSection{
			{Title: "Approvers", ShowTitle: true},
			{Title: "Time Info"},
		}},
	}
}
*/

func (o *PostingProfilePIC) KxPostDelete(ctx *kaos.Context, mdl orm.DataModel) error {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("PostDelete: missing connection")
	}

	o = mdl.(*PostingProfilePIC)
	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
		if ctx.Data().Get("UserID", "").(string) != "" {
			userID = ctx.Data().Get("UserID", "").(string)
		}
	}

	log := PostingProfilePICLog{
		UpdatedBy:         userID,
		PostingProfilePIC: o,
		Action:            "delete",
	}

	pp := new(PostingProfile)
	err := h.GetByParm(pp, dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("_id", o.PostingProfileID),
	))
	if err != nil && err != io.EOF {
		return fmt.Errorf("PostDelete: error when get posting profile: %s", err.Error())
	}

	log.PostingProfileName = pp.Name
	err = h.Save(&log)
	if err != nil {
		return fmt.Errorf("PostDelete: error when save posting profile pic log: %s", err.Error())
	}

	return nil
}
