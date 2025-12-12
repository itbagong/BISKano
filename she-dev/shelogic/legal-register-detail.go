package shelogic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type LRDLogic struct {
}

type LRDRequest struct {
	RefNo   string
	LegalNo string
	SiteID  string
	Skip    int
	Take    int
	Where   *dbflex.Filter
}

func (obj *LRDLogic) Save(ctx *kaos.Context, lr *shemodel.LegalRegisterDetail) (*shemodel.LegalRegisterDetail, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// save legal register detail
	oldLegalRegisterDetail := new(shemodel.LegalRegisterDetail)
	if e := h.GetByID(oldLegalRegisterDetail, lr.ID); e != nil {
		Sites, err := GetSites(ctx, nil)
		if err != nil {
			return nil, errors.New("error GetSites: " + err.Error())
		}

		mapSites := lo.Associate(Sites, func(detail shemodel.SHESite) (string, shemodel.SHESite) {
			return detail.ID, detail
		})

		vAlias := ""
		if vSite, ok := mapSites[lr.SiteID]; ok {
			vAlias = vSite.Alias
		}

		// generate number legal register detail
		tenantcorelogic.MWPreAssignSequenceNo("SHELegal", false, "_id")(ctx, &lr)
		lr.ID = "LEGAL/" + vAlias + "/" + lr.ID

		oldLegalRegisterDetail = lr
		if e := h.Insert(lr); e != nil {
			return nil, errors.New("error insert Legal Register detail: " + e.Error())
		}
	} else {
		if e := h.Save(lr); e != nil {
			return nil, errors.New("error update Legal Register detail: " + e.Error())
		}
	}

	// update data legal register  detail
	lr.Achievement = (float64(lr.ActualCompliance) / float64(lr.PlantCompliance)) * 100
	if e := h.Save(lr); e != nil {
		return nil, errors.New("error update legal register detail: " + e.Error())
	}

	// get legal compliance
	legalCompliances := []*shemodel.LegalCompliance{}
	e := h.Gets(new(shemodel.LegalCompliance), nil, &legalCompliances)
	if e != nil {
		return nil, errors.New("Failed populate data legal compliance list: " + e.Error())
	}

	mapCompliance := lo.Associate(legalCompliances, func(detail *shemodel.LegalCompliance) (string, *shemodel.LegalCompliance) {
		return detail.SiteID, detail
	})

	// update legal compliance
	e = UpdateLegalComplianceNew(h, mapCompliance, oldLegalRegisterDetail, lr)
	if e != nil {
		return nil, errors.New("error update legal compliance")
	}

	return lr, nil
}

func (obj *LRDLogic) Delete(ctx *kaos.Context, lrd *shemodel.LegalRegisterDetail) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if lrd == nil {
		return nil, errors.New("payload legal register detail is required")
	}

	// delete data legal register detail
	e := h.Delete(lrd)
	if e != nil {
		return nil, errors.New("error delete legal register detail")
	}

	// get legal compliance
	legalCompliances := []*shemodel.LegalCompliance{}
	e = h.Gets(new(shemodel.LegalCompliance), nil, &legalCompliances)
	if e != nil {
		return nil, errors.New("Failed populate data legal compliance list: " + e.Error())
	}

	mapCompliance := lo.Associate(legalCompliances, func(detail *shemodel.LegalCompliance) (string, *shemodel.LegalCompliance) {
		return detail.SiteID, detail
	})

	// update legal compliance
	e = UpdateLegalComplianceNew(h, mapCompliance, lrd, nil)
	if e != nil {
		return nil, errors.New("error update legal compliance")
	}

	return lrd, nil
}

func (obj *LRDLogic) Gets(ctx *kaos.Context, payload *dbflex.QueryParam) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	legalRegisterDetails := []shemodel.LegalRegisterDetail{}
	legalRegisters := []shemodel.LegalRegister{}

	r := ctx.Data().Get("http_request", nil).(*http.Request)
	SiteID := strings.TrimSpace(r.URL.Query().Get("SiteID"))

	vRelatedSite := shemodel.RS_ALL_SITE
	if SiteID == "SITE020" {
		vRelatedSite = shemodel.RS_HO
	}
	payload = payload.MergeWhere(false, dbflex.Eq("RelatedSite", vRelatedSite))

	if e := h.Gets(new(shemodel.LegalRegister), payload, &legalRegisters); e != nil {
		return nil, fmt.Errorf("error when get legal register: %s", e.Error())
	}

	if e := h.Gets(new(shemodel.LegalRegisterDetail), dbflex.NewQueryParam().SetWhere(dbflex.Eq("SiteID", SiteID)), &legalRegisterDetails); e != nil {
		return nil, fmt.Errorf("error when get legal register detail: %s", e.Error())
	}
	mapLRDs := lo.Associate(legalRegisterDetails, func(legalRegisterDetail shemodel.LegalRegisterDetail) (string, shemodel.LegalRegisterDetail) {
		return legalRegisterDetail.LegalNo, legalRegisterDetail
	})

	resultData := []shemodel.LegalRegisterDetail{}
	for _, val := range legalRegisters {

		if _, ok := mapLRDs[val.LegalNo]; !ok {
			legalDetail := shemodel.LegalRegisterDetail{
				Date:             val.Date,
				LegalNo:          val.LegalNo,
				Type:             val.Type,
				Category:         val.Category,
				SiteID:           SiteID,
				RelatedSite:      val.RelatedSite,
				Fields:           val.Fields,
				Link:             val.Link,
				Reference:        val.Reference,
				Status:           val.StatusDoc,
				PlantCompliance:  val.PlantCompliance,
				ActualCompliance: val.ActualCompliance,
				Achievement:      val.Achievement,
				LegalDetails:     val.LegalDetails,
				Dimension:        val.Dimension,
			}

			resultData = append(resultData, legalDetail)
		}
	}

	return codekit.M{"data": resultData, "count": len(resultData)}, nil
}

func (obj *LRDLogic) GetLegalDetail(ctx *kaos.Context, payload *LRDRequest) (*shemodel.LegalRegisterDetail, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	res := shemodel.LegalRegisterDetail{}
	legalRegister := shemodel.LegalRegister{}

	RefNo := payload.RefNo
	LegalNo := payload.LegalNo
	SiteID := payload.SiteID

	if RefNo == "" {
		if LegalNo != "" {
			h.GetsByFilter(new(shemodel.LegalRegister), dbflex.Eq("LegalNo", LegalNo), &legalRegister)

			res.SiteID = SiteID
			res.Date = legalRegister.Date
			res.LegalNo = legalRegister.LegalNo
			res.Type = legalRegister.Type
			res.Category = legalRegister.Category
			res.RelatedSite = legalRegister.RelatedSite
			res.Fields = legalRegister.Fields
			res.Link = legalRegister.Link
			res.Reference = legalRegister.Reference
			res.Status = legalRegister.StatusDoc
			res.PlantCompliance = legalRegister.PlantCompliance
			res.ActualCompliance = legalRegister.ActualCompliance
			res.Achievement = legalRegister.Achievement
			res.LegalDetails = legalRegister.LegalDetails
			res.Dimension = legalRegister.Dimension
			res.Created = legalRegister.Created
			res.LastUpdate = legalRegister.LastUpdate
		}
	} else {
		h.GetsByFilter(new(shemodel.LegalRegisterDetail), dbflex.Eq("_id", RefNo), &res)

		res.SiteID = SiteID
	}

	return &res, nil
}
