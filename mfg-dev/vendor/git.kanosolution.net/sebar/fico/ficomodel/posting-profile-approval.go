package ficomodel

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficoconfig"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostingProfileApprovalItem struct {
	Line      int
	UserID    string
	Status    string
	Text      string
	Assigned  time.Time
	Confirmed *time.Time
}

type PostingApproval struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	PostingProfileID  string
	CompanyID         string
	SourceType        string
	SourceID          string
	Version           int
	CurrentStage      int
	Status            string
	AccountID         string
	Text              string
	TrxDate           time.Time
	Amount            float64
	Confirmed         *time.Time
	Dimension         tenantcoremodel.Dimension
	Approvers         []PostingUsers
	Postingers        []PostingUsers
	Submitters        []PostingUsers
	Approvals         []*PostingProfileApprovalItem
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *PostingApproval) TableName() string {
	return "PostingProfileApprovals"
}

func (o *PostingApproval) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *PostingApproval) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *PostingApproval) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *PostingApproval) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *PostingApproval) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *PostingApproval) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *PostingApproval) UpdateStage(h *datahub.Hub) error {
	o.CurrentStage++
	if len(o.Approvers) < o.CurrentStage {
		return fmt.Errorf("approvers stage %d out of range", o.CurrentStage)
	}
	o.Approvals = lo.Filter(o.Approvals, func(a *PostingProfileApprovalItem, index int) bool {
		return a.Line != o.CurrentStage
	})
	approvers := o.Approvers[o.CurrentStage-1]
	newApprovals := lo.Map(approvers.UserIDs, func(userID string, index int) *PostingProfileApprovalItem {
		return &PostingProfileApprovalItem{
			Line:     o.CurrentStage,
			UserID:   userID,
			Status:   "PENDING",
			Assigned: time.Now(),
		}
	})

	if o.Status == "PENDING" {
		o.sendEmailToApprovers(h) // send only when postingProfile.NeedApproval == true, which is status is still Pending
	}

	o.Approvals = append(o.Approvals, newApprovals...)
	return nil
}

func (o *PostingApproval) UpdateApproval(h *datahub.Hub, approverUserID, approvalState, text string) error {
	if approvalState == "Reject" && text == "" {
		return errors.New("text is mandatory for status REJECTED")
	}

	approver := o.Approvers[o.CurrentStage-1]
	approvals := lo.Filter(o.Approvals, func(a *PostingProfileApprovalItem, index int) bool {
		return a.Line == o.CurrentStage && a.UserID == approverUserID
	})
	if len(approvals) == 0 {
		return fmt.Errorf("user %s is not registered as approval", approverUserID)
	}

	switch approvalState {
	case "Approve":
		dt := time.Now()
		approvals[0].Confirmed = &dt
		approvals[0].Text = text
		approvals[0].Status = "APPROVED"

		approvedCount := lo.Filter(o.Approvals, func(a *PostingProfileApprovalItem, index int) bool {
			return a.Line == o.CurrentStage && a.Status == "APPROVED"
		})
		// update all approver status with same line when minimal approver count has been achieved
		if approver.MinimalApproverCount == len(approvedCount) {
			for _, ap := range o.Approvals {
				if ap.Line == o.CurrentStage && ap.UserID != approverUserID {
					ap.Status = "APPROVED"
				}
			}
		}

		if len(approvedCount) >= o.Approvers[o.CurrentStage-1].MinimalApproverCount {
			if o.CurrentStage == len(o.Approvers) {
				o.Status = "APPROVED"
				o.Confirmed = &dt
			} else {
				return o.UpdateStage(h)
			}
		}

	case "Reject":
		dt := time.Now()
		approvals[0].Confirmed = &dt
		approvals[0].Text = text
		approvals[0].Status = "REJECTED"

		// update all approver status with same line
		for _, ap := range o.Approvals {
			if ap.Line == o.CurrentStage && ap.UserID != approverUserID {
				ap.Status = "REJECTED"
			}
		}

		o.Status = "REJECTED"
		o.Confirmed = &dt
	}

	return nil
}

type SourceURLRequest struct {
	SourceType string
	SourceID   string
}

type SourceURL struct {
	Menu      string
	URL       string
	JournalID string // used in mfg only
}

func (o *PostingApproval) sendEmailToApprovers(h *datahub.Hub) {
	urlRes := new(SourceURL)
	urlAPI := ""

	if o.SourceType == "Purchase Order" ||
		o.SourceType == "Purchase Request" ||
		o.SourceType == "INVENTORY" ||
		o.SourceType == "Inventory Receive" ||
		o.SourceType == "Inventory Issuance" ||
		o.SourceType == "Transfer" ||
		o.SourceType == "Item Request" ||
		o.SourceType == "Asset Acquisition" {
		urlAPI = "/v1/scm/new/postingprofile/map-source-data-to-url"
	}

	if o.SourceType == "WORKORDER" ||
		o.SourceType == "Work Order" ||
		o.SourceType == "Work Order Report Consumption" ||
		o.SourceType == "Work Order Report Output" ||
		o.SourceType == "Work Order Report Resource" ||
		o.SourceType == "Work Request" {
		urlAPI = "/v1/mfg/postingprofile/map-source-data-to-url"
	}

	if o.SourceType == "CASHBANK" ||
		o.SourceType == "CUSTOMER" ||
		o.SourceType == "LEDGERACCOUNT" ||
		o.SourceType == "VENDOR" {
		urlAPI = "/v1/fico/postingprofile/map-source-data-to-url"
	}

	if o.SourceType == "Sales Order" ||
		o.SourceType == "Sales Quotation" {
		urlAPI = "/v1/sdp/postingprofile/map-source-data-to-url"
	}

	if urlAPI != "" {
		ficoconfig.Config.EventHub.Publish(
			urlAPI,
			SourceURLRequest{SourceType: o.SourceType, SourceID: o.SourceID},
			urlRes,
			&kaos.PublishOpts{Headers: codekit.M{}},
		)
	}

	empORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Employee))

	for _, stage := range o.Approvers {
		for _, approverUserID := range stage.UserIDs {
			emp, _ := empORM.Get(approverUserID)
			if emp.ID == "" {
				continue
			}

			go func() {
				if _, e := SendEmailByTemplate(emp.Email, "posting-profile-approval", "en-us", codekit.M{
					"Name":     emp.Name,
					"Menu":     lo.Ternary(urlRes.Menu != "", urlRes.Menu, o.SourceType),
					"SourceID": lo.Ternary(urlRes.JournalID != "", urlRes.JournalID, o.SourceID),
					"URL":      urlRes.URL,
				}); e != nil {
					ficoconfig.Config.EventHub.Service().Log().Errorf("SendEmailByTemplate: %s", e.Error())
				}
			}()
		}
	}
}
