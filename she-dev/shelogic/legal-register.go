package shelogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type LRLogic struct {
}

func (obj *LRLogic) Save(ctx *kaos.Context, lr *shemodel.LegalRegister) (*shemodel.LegalRegister, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// save legal register
	oldLegalRegister := new(shemodel.LegalRegister)
	if e := h.GetByID(oldLegalRegister, lr.ID); e != nil {
		if e := h.Insert(lr); e != nil {
			return nil, errors.New("error insert Legal Register: " + e.Error())
		}

		return lr, nil
	}

	if e := h.Save(lr); e != nil {
		return nil, errors.New("error update Legal Register: " + e.Error())
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

	// get legal register detail by legal no
	legalRegisterDetails := []shemodel.LegalRegisterDetail{}
	e = h.Gets(new(shemodel.LegalRegisterDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("LegalNo", oldLegalRegister.LegalNo),
		)), &legalRegisterDetails)
	if e != nil {
		return nil, errors.New("Failed populate data legal register detail list: " + e.Error())
	}

	if len(legalRegisterDetails) > 0 {
		for _, lrd := range legalRegisterDetails {
			newLRD := new(shemodel.LegalRegisterDetail)
			if oldLegalRegister.RelatedSite != lr.RelatedSite {
				newLRD = nil
				// delete base on legal no
				e = h.Delete(&lrd)
				if e != nil {
					return nil, errors.New("error delete legal register detail")
				}
			} else {
				// update legal register detail
				lrd.Date = lr.Date
				lrd.LegalNo = lr.LegalNo
				lrd.Type = lr.Type
				lrd.Category = lr.Category
				lrd.RelatedSite = lr.RelatedSite
				lrd.Fields = lr.Fields
				lrd.Link = lr.Link
				lrd.Reference = lr.Reference
				lrd.Status = lr.StatusDoc
				lrd.PlantCompliance = lr.PlantCompliance
				lrd.ActualCompliance = lr.ActualCompliance
				lrd.Achievement = lr.Achievement
				lrd.Dimension = lr.Dimension
				lrd.Created = lr.Created
				lrd.LastUpdate = lr.LastUpdate
				lrd.LegalDetails = lr.LegalDetails

				if e := h.Save(&lrd); e != nil {
					return nil, errors.New("error update Legal Register Detail: " + e.Error())
				}

				newLRD.SiteID = lrd.SiteID
				newLRD.ActualCompliance = lr.ActualCompliance
				newLRD.PlantCompliance = lr.PlantCompliance
			}

			// update or insert legal compliance
			e = UpdateLegalComplianceNew(h, mapCompliance, &lrd, newLRD)
			if e != nil {
				return nil, errors.New("error UpdateLegalCompliance: " + e.Error())
			}
		}
	}

	return lr, nil
}

func UpdateLegalCompliance(h *datahub.Hub, mapCompliance map[string]*shemodel.LegalCompliance, siteID, typeTrans string, LegalDetail, newLegalDetail []shemodel.LegalDetail) error {
	if vComp, ok := mapCompliance[siteID]; ok {
		vCountPlant := 0
		vCountActual := 0
		if len(LegalDetail) > 0 {
			for _, val := range LegalDetail {
				vCountPlant = vCountPlant - len(val.ActivityPoints)
				if len(val.ActivityPoints) > 0 {
					for _, valAP := range val.ActivityPoints {
						if valAP.IsComply {
							vCountActual--
						}
					}
				}
			}
		}
		if len(newLegalDetail) > 0 {
			for _, val := range newLegalDetail {
				vCountPlant = vCountPlant + len(val.ActivityPoints)
				if typeTrans == "save" {
					if len(val.ActivityPoints) > 0 {
						for _, valAP := range val.ActivityPoints {
							if valAP.IsComply {
								vCountActual--
							}
						}
					}
				}
			}
		}

		vTotalPlant := vComp.PlantCompliance + vCountPlant
		vTotalActual := vComp.ActualCompliance + vCountActual
		vComp.PlantCompliance = vTotalPlant
		vComp.ActualCompliance = vTotalActual

		vTotal := 0.0
		if vTotalPlant > 0 {
			vTotal = (float64(vTotalActual) / float64(vTotalPlant)) * 100
		}

		fieldUpdate := []string{
			"PlantCompliance",
			"ActualCompliance",
			"Compliance",
		}
		if e := h.UpdateField(&shemodel.LegalCompliance{
			PlantCompliance:  vTotalPlant,
			ActualCompliance: vTotalActual,
			Compliance:       vTotal,
		}, dbflex.Eq("SiteID", siteID), fieldUpdate...); e != nil {
			return e
		}
	} else {
		if typeTrans == "save" {
			legalCompliance := shemodel.LegalCompliance{
				SiteID:           siteID,
				PlantCompliance:  len(LegalDetail),
				ActualCompliance: 0.0,
				Compliance:       0.0,
			}
			if e := h.Insert(&legalCompliance); e != nil {
				return errors.New("error insert legal compliance: " + e.Error())
			}
		}
	}

	return nil
}
