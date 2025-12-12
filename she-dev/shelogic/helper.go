package shelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type UpdateAttachmentTags struct {
	JournalType string
	JournalID   string
	Tags        []string
	NewTags     []string
}

func GetCompanyIDFromContext(ctx *kaos.Context) string {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		return "DEMO"
	}
	return coID
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
}

func GetSites(ctx *kaos.Context, vFilter *dbflex.Filter) ([]shemodel.SHESite, error) {
	res := struct {
		Site []shemodel.SHESite
	}{}

	payload := struct {
		Filter *dbflex.Filter
	}{
		Filter: vFilter,
	}

	// get data site
	ev, _ := ctx.DefaultEvent()

	if ev == nil {
		return nil, errors.New("nil: EventHub")
	}

	ev.Publish("/v1/bagong/sitesetup/get-sites", &payload, &res, nil)

	return res.Site, nil
}

func UpdateLegalComplianceNew(h *datahub.Hub, mapCompliance map[string]*shemodel.LegalCompliance, LegalDetail, newLegalDetail *shemodel.LegalRegisterDetail) error {
	// get legal compliance
	legalCompliances := []*shemodel.LegalCompliance{}
	e := h.Gets(new(shemodel.LegalCompliance), dbflex.NewQueryParam().SetWhere(dbflex.Eq("SiteID", LegalDetail.SiteID)), &legalCompliances)
	if e != nil {
		return errors.New("Failed populate data legal compliance list: " + e.Error())
	}

	if len(legalCompliances) > 0 {
		legalCompliance := legalCompliances[0]
		totalActual := -LegalDetail.ActualCompliance
		totalPlant := -LegalDetail.PlantCompliance
		if newLegalDetail != nil {
			totalActual = totalActual + newLegalDetail.ActualCompliance
			totalPlant = totalPlant + newLegalDetail.PlantCompliance
		}

		legalCompliance.PlantCompliance = legalCompliance.PlantCompliance + totalPlant
		legalCompliance.ActualCompliance = legalCompliance.ActualCompliance + totalActual
		legalCompliance.Compliance = (float64(legalCompliance.ActualCompliance) / float64(legalCompliance.PlantCompliance)) * 100
		if e := h.Save(legalCompliance); e != nil {
			return errors.New("error update legal compliance: " + e.Error())
		}
	} else {
		if newLegalDetail != nil {
			legalCompliance := shemodel.LegalCompliance{
				SiteID:           newLegalDetail.SiteID,
				PlantCompliance:  newLegalDetail.PlantCompliance,
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

func GetURLQueryParams(ctx *kaos.Context) map[string]string {
	r, ok := sebar.GetHTTPRequest(ctx)
	if !ok {
		return map[string]string{}
	}

	res := map[string]string{}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			res[key] = values[0]
		}
	}

	return res
}

func UpdateTagByJournal(ctx *kaos.Context, journalType, journalID string, tags, newTags []string) (string, error) {
	evHub, _ := ctx.DefaultEvent()

	updateAttachmentTags := []UpdateAttachmentTags{}
	updateAttachmentTags = append(updateAttachmentTags, UpdateAttachmentTags{
		JournalType: journalType,
		JournalID:   journalID,
		Tags:        tags,
		NewTags:     newTags,
	})

	vRes := ""
	err := evHub.Publish("/v1/asset/update-tag-by-journal", updateAttachmentTags, &vRes, nil)
	if err != nil {
		return vRes, fmt.Errorf("error : update attachment tags: %s", err.Error())
	}

	return vRes, nil
}
