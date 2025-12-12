package mfglogic

import (
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type ItemRequestInsertParam struct {
	ID          string `json:"_id" bson:"_id"`
	Name        string
	WOReff      string
	Requestor   string
	Department  string
	Priority    string
	Dimension   tenantcoremodel.Dimension
	InventDimTo scmmodel.InventDimension
	Lines       []ItemRequestLineParam

	RequestDate  *time.Time // optional
	DocumentDate *time.Time // optional
	TrxDate      time.Time  // optional
	TrxType      string     // optional
	Status       string     // optional
	CompanyID    string     // optional
}

type ItemRequestLineParam struct {
	ItemRequestID string
	ItemID        string
	SKU           string
	QtyRequested  float64
	UoM           string
	WarehouseID   string
	Dimension     tenantcoremodel.Dimension
	RequestedBy   string

	// helper fields
	InventDimTo scmmodel.InventDimension
}

type NumSeqClaimRespond struct {
	Number string
}

type NumSeqClaimPayload struct {
	NumberSequenceID string
	Date             time.Time
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

func ItemRequestInsert(ctx *kaos.Context, param *ItemRequestInsertParam, companyID, userID string) (irID string, err error) {
	url := "/v1/scm/item/request/insert"
	urlLine := "/v1/scm/item/request/detail/insert"

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: connection")
	}

	//get number sequence setup
	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return "", nil
	}

	setup := tenantcorelogic.GetSequenceSetup(h, "ItemRequest", companyID)
	resp := new(NumSeqClaimRespond)
	if e := ev.Publish("/v1/tenant/numseq/claim", &NumSeqClaimPayload{NumberSequenceID: setup.NumSeqID, Date: time.Now()}, resp, &kaos.PublishOpts{Headers: codekit.M{
		"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}}); e != nil {
		return "", e
	}

	// id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx)
	now := time.Now()
	param.ID = resp.Number
	param.Priority = lo.Ternary(param.Priority != "", param.Priority, "P1")
	param.RequestDate = lo.Ternary(param.RequestDate != nil, param.RequestDate, &now)
	param.DocumentDate = lo.Ternary(param.DocumentDate != nil, param.DocumentDate, &now)
	param.TrxDate = lo.Ternary(param.TrxDate.IsZero() == false, param.TrxDate, now)
	param.TrxType = lo.Ternary(param.TrxType != "", param.TrxType, "Item Request")
	param.Status = lo.Ternary(param.Status != "", param.Status, string(ficomodel.JournalStatusDraft))
	param.CompanyID = lo.Ternary(param.CompanyID != "", param.CompanyID, companyID)

	var irResult interface{}
	err = Config.EventHub.Publish(
		url,
		param,
		&irResult,
		&kaos.PublishOpts{Headers: codekit.M{"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}},
	)

	if err != nil {
		fmt.Printf("%s Error: %s | res: %s\n", url, err.Error(), codekit.JsonStringIndent(irResult, "\t"))
		return "", err
	}

	irM, _ := codekit.ToM(irResult)
	irID = irM.GetString("_id")

	for _, line := range param.Lines {
		line.ItemRequestID = irID

		err = Config.EventHub.Publish(
			urlLine,
			&line,
			nil,
			&kaos.PublishOpts{Headers: codekit.M{"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}},
		)
		if err != nil {
			fmt.Printf("%s Error: %s\n", urlLine, err)
		}
	}

	return
}

func Posting(params []ficologic.PostRequest, companyID, userID string) (err error) {
	url := "/v1/scm/new/postingprofile/post"

	var posRes interface{}
	err = Config.EventHub.Publish(
		url,
		&params,
		&posRes,
		&kaos.PublishOpts{Headers: codekit.M{"CompanyID": companyID, sebar.CtxJWTReferenceID: userID}},
	)
	fmt.Printf("%s Error: %s | res: %s\n", url, err, codekit.JsonStringIndent(posRes, "\t"))

	return
}

func TemplateToPDF(param *tenantcorelogic.PDFFromTemplateRequest) ([]byte, error) {
	url := "/v1/tenant/pdf/from-template"

	res := tenantcorelogic.PDFByteResponse{}
	err := Config.EventHub.Publish(
		url,
		param,
		&res,
		nil,
	)
	if err != nil {
		fmt.Println("TemplateToPDF | error publish:", err)
		return nil, err
	}

	return res.PDFByte, nil
}

// BagongAssetRes grow fields as necessary
type BagongAssetRes struct {
	ID              string `json:"_id"`
	Name            string
	GroupID         string
	IsActive        bool
	AssetType       string
	LatestCustomer  string
	Depreciation    BagongAssetResDepreciation
	Dimension       tenantcoremodel.Dimension
	UserInfo        []BagongAssetResUserInfo
	CurrentUserInfo []BagongAssetResUserInfo // user info yang dipakai, sesuai param date, bila kosong, pakai tgl sekarang (will be set dynamically in helper func)
	References      []codekit.M
}

type BagongAssetResDepreciation struct {
	AssetDuration      int
	DepreciationPeriod string
	DepreciationDate   *time.Time
	AcquisitionDate    *time.Time
	User               string
	ResidualAmount     float64
}

type BagongAssetResUserInfo struct {
	ProjectID      string
	AssetDateFrom  time.Time
	AssetDateTo    time.Time
	SiteID         string
	UserID         string
	CustomerID     string
	NoHullCustomer string
	Description    string
}

func (o *BagongAssetRes) GetCurrentUserInfos(date ...time.Time) []BagongAssetResUserInfo {
	execDate := time.Now()
	if len(date) > 0 && !date[0].IsZero() {
		execDate = date[0]
	}

	res := lo.Filter(o.UserInfo, func(d BagongAssetResUserInfo, i int) bool {
		from := d.AssetDateFrom.Before(execDate) || d.AssetDateFrom.Equal(execDate)
		to := d.AssetDateTo.After(execDate) || d.AssetDateTo.Equal(execDate)

		return from && to
	})

	return res
}

func BagongAssetGet(assetID string, date ...time.Time) (bgAsset *BagongAssetRes, err error) {
	url := "/v1/bagong/asset/get"
	bgAsset = new(BagongAssetRes)

	err = Config.EventHub.Publish(
		url,
		[]interface{}{assetID},
		bgAsset,
		nil,
	)
	// fmt.Printf("%s error: %s | res: %s\n", url, err, codekit.JsonStringIndent(bgAsset, "\t"))

	bgAsset.CurrentUserInfo = bgAsset.GetCurrentUserInfos(date...)

	return
}

func BagongAssetGets(assetIDs []string, date ...time.Time) (bgAssets []BagongAssetRes, err error) {
	url := "/v1/bagong/asset/gets"
	bgAssets = []BagongAssetRes{}

	if len(assetIDs) == 0 {
		return []BagongAssetRes{}, nil
	}

	payload := scmlogic.GeneralRequest{
		Sort: []string{"-_id"},
		Where: &dbflex.Filter{
			Op: "$and",
			Items: []*dbflex.Filter{{
				Op:    "$in",
				Field: "_id",
				Value: assetIDs,
			}},
		},
	}

	apiRes := codekit.M{}
	err = Config.EventHub.Publish(
		url,
		&payload,
		&apiRes,
		nil,
	)
	// fmt.Printf("%s error: %s\n", url, err)

	for _, res := range codekit.ToInterfaceArray(apiRes["data"]) {
		bgAss := BagongAssetRes{}
		serde.Serde(res, &bgAss)
		bgAssets = append(bgAssets, bgAss)
	}

	bgAssets = lo.Map(bgAssets, func(ass BagongAssetRes, i int) BagongAssetRes {
		ass.CurrentUserInfo = ass.GetCurrentUserInfos(date...)
		return ass
	})

	return
}
