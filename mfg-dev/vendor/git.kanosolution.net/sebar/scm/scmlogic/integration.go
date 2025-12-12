package scmlogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type PurchaseRequestInsertParam struct {
	ReffNo []string
	Name   string
	Lines  []scmmodel.PurchaseJournalLine

	DocumentDate *time.Time // optional
	TrxDate      time.Time  // optional
	PRDate       *time.Time // optional
	ExpectedDate *time.Time // optional
	Status       string     // optional
	CompanyID    string     // optional

	Location  scmmodel.InventDimension
	Dimension tenantcoremodel.Dimension
}

type PurchaseRequestLineParam struct {
	PurchaseRequestID string
	ItemID            string
	SKU               string
	QtyRequested      float64
	UoM               string
	WarehouseID       string
}

func PurchaseRequestInsert(param *PurchaseRequestInsertParam, companyID, userID string) (err error) {
	url := "/v1/scm/purchase/request/insert"

	now := time.Now()
	param.DocumentDate = lo.Ternary(param.DocumentDate != nil, param.DocumentDate, &now)
	param.TrxDate = lo.Ternary(param.TrxDate.IsZero() == false, param.TrxDate, now)
	param.Status = lo.Ternary(param.Status != "", param.Status, string(ficomodel.JournalStatusDraft))
	param.CompanyID = lo.Ternary(param.CompanyID != "", param.CompanyID, companyID)

	var apiResult interface{}
	err = scmconfig.Config.EventHub().Publish(
		url,
		param,
		&apiResult,
		&kaos.PublishOpts{Headers: codekit.M{"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}},
	)
	fmt.Printf("%s error: %s | res: %s\n", url, err, codekit.JsonStringIndent(apiResult, "\t"))

	return
}

func VendorJournalInsert(payload *ficomodel.VendorJournal, companyID, userID string) (err error) {
	url := "/v1/fico/vendorjournal/insert"
	var apires interface{}

	err = scmconfig.Config.EventHub().Publish(
		url,
		&payload,
		&apires,
		&kaos.PublishOpts{Headers: codekit.M{"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}},
	)
	fmt.Printf("%s error: %s | res: %s\n", url, err, codekit.JsonStringIndent(apires, "\t"))

	return
}

type VendorGetRes struct {
	ID     string `json:"_id"`
	Name   string
	Detail VendorGetResDetail
}

type VendorGetResDetail struct {
	ID                 string `json:"_id"`
	TenantCoreVendorID string
	Terms              VendorGetResDetailTerm
}

type VendorGetResDetailTerm struct {
	Name   string
	Taxes1 string
	Taxes2 string
}

func VendorGet(vendorID string) (vendor *VendorGetRes, err error) {
	url := "/v1/bagong/vendor/get"
	vendor = new(VendorGetRes)

	err = scmconfig.Config.EventHub().Publish(
		url,
		[]interface{}{vendorID},
		vendor,
		nil,
	)
	fmt.Printf("%s error: %s | res: %s\n", url, err, codekit.JsonStringIndent(vendor, "\t"))

	return
}

func GetSignatureByID(h *datahub.Hub, Company, JournalType, JournalID string) ([]tenantcoremodel.Signature, error) {
	result := []tenantcoremodel.Signature{}

	resp := new(ficomodel.PostingApproval)
	approvalFilter := dbflex.Eqs("CompanyID", Company, "SourceType", JournalType, "SourceID", JournalID)
	_, err := datahub.GetByParm(h, resp, dbflex.NewQueryParam().SetWhere(approvalFilter).SetSort("-Created"))
	if err != nil {
		return nil, fmt.Errorf("error when get approval: %s", err.Error())
	}

	userIDs := []string{}
	for _, approver := range resp.Approvers {
		userIDs = append(userIDs, approver.UserIDs...)
	}

	users := []tenantcoremodel.Employee{}
	err = h.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", userIDs...)), &users)
	if err != nil {
		return nil, fmt.Errorf("error when get user: %s", err.Error())
	}

	usersDetail := []bagongmodel.EmployeeDetail{}
	err = h.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", userIDs...)), &usersDetail)
	if err != nil {
		return nil, fmt.Errorf("error when get user: %s", err.Error())
	}

	mapUser := lo.Associate(users, func(user tenantcoremodel.Employee) (string, string) {
		return user.ID, user.Name
	})

	mapUserGrade := lo.Associate(usersDetail, func(usersDetail bagongmodel.EmployeeDetail) (string, string) {
		return usersDetail.ID, usersDetail.Grade
	})

	for _, e := range resp.Approvers {
		for _, d := range e.UserIDs {
			isExist := false
			signature := tenantcoremodel.Signature{}
			for _, c := range resp.Approvals {
				if d == c.UserID {
					isExist = true
					header := "Diketahui"
					footer := ""
					confirmed := ""

					if len(usersDetail) > 0 {
						if v, ok := mapUserGrade[c.UserID]; ok {
							if v == "GDE001" {
								header = "Disetujui"
							}
						}
					}

					if v, ok := mapUser[c.UserID]; ok {
						footer = v
					}

					if c.Confirmed != nil {
						confirmed = "Tgl. " + c.Confirmed.Format("02-January-2006 15:04:05")
					}

					signature = tenantcoremodel.Signature{
						ID:        c.UserID,
						Header:    header,
						Footer:    footer,
						Confirmed: confirmed,
						Status:    c.Status,
					}
					break
				}

			}

			if !isExist {
				header := "Diketahui"
				footer := ""
				confirmed := ""

				if len(usersDetail) > 0 {
					if v, ok := mapUserGrade[d]; ok {
						if v == "GDE001" {
							header = "Disetujui"
						}
					}
				}

				if v, ok := mapUser[d]; ok {
					footer = v
				}

				signature = tenantcoremodel.Signature{
					ID:        d,
					Header:    header,
					Footer:    footer,
					Confirmed: confirmed,
				}
			}

			result = append(result, signature)
		}
	}

	return result, nil
}
