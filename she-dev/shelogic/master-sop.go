package shelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

type MasterSOPLogic struct {
}

func (obj *MasterSOPLogic) Save(ctx *kaos.Context, sop *shemodel.MasterSOP) (*shemodel.MasterSOP, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// save data asset bagong
	ev, _ := ctx.DefaultEvent()
	if ev == nil {
		return nil, errors.New("nil: EventHub")
	}

	// generate number Meeting
	tenantcorelogic.MWPreAssignSequenceNo("SOP", false, "_id")(ctx, &sop)

	// save master SOP
	oldMasterSOP := new(shemodel.MasterSOP)
	if e := h.GetByID(oldMasterSOP, sop.ID); e != nil {
		if e := h.Insert(sop); e != nil {
			return nil, errors.New("error insert master sop: " + e.Error())
		}

		return sop, nil
	}

	if sop.Status == string(shemodel.SHEStatusApproved) {
		masterSOPSummary := new(shemodel.MasterSOPSummary)
		if sop.NatureOfChange == shemodel.NOC_PEMBUATAN {
			// generate number master SOP
			tenantcorelogic.MWPreAssignSequenceNo("SOPSummary", false, "_id")(ctx, &masterSOPSummary)

			masterSOPSummary.DocumentType = sop.DocumentType
			masterSOPSummary.TitleOfDocument = sop.TitleOfDocument
			masterSOPSummary.EffectiveDate = sop.EffectiveDate
			masterSOPSummary.IsActive = true
			if e := h.Insert(masterSOPSummary); e != nil {
				return nil, errors.New("error insert master sop summary: " + e.Error())
			}

			sop.DocumentRefno = masterSOPSummary.ID

			// add tag
			// vTag := "SHE_SOP_" + sop.ID
			newTag := "SHE_SOP_SUMMARY_" + masterSOPSummary.ID + "_" + masterSOPSummary.EffectiveDate.Format("2006-01-02T15:04:05")
			payload := []UpdateAttachmentTags{{
				JournalType: "SHE_SOP",
				JournalID:   sop.ID,
				Tags:        []string{},
				NewTags:     []string{newTag},
			}}

			vRes := ""
			err := ev.Publish("/v1/asset/update-tag-by-journal", payload, &vRes, nil)
			if err != nil {
				return nil, fmt.Errorf("error : update attachment tags: %s", err.Error())
			}
		} else {
			if e := h.GetByID(masterSOPSummary, sop.DocumentRefno); e != nil {
				return nil, errors.New("error populate data master sop summary: " + e.Error())
			} else {
				if sop.NatureOfChange == shemodel.NOC_OBSOLETE {
					if e := h.UpdateField(&shemodel.MasterSOPSummary{
						IsActive: false,
					}, dbflex.Eq("_id", masterSOPSummary.ID), "IsActive"); e != nil {
						return nil, errors.New("error update data master sop summary: " + e.Error())
					}
				} else if sop.NatureOfChange == shemodel.NOC_REVISI {
					if e := h.UpdateField(&shemodel.MasterSOPSummary{
						EffectiveDate: sop.EffectiveDate,
					}, dbflex.Eq("_id", masterSOPSummary.ID), "EffectiveDate"); e != nil {
						return nil, errors.New("error update data master sop summary: " + e.Error())
					}

					// update tag
					// vTag := "SHE_SOP_" + sop.ID
					newTag := "SHE_SOP_SUMMARY_" + masterSOPSummary.ID + "_" + sop.EffectiveDate.Format("2006-01-02T15:04:05")
					payload := []UpdateAttachmentTags{{
						JournalType: "SHE_SOP",
						JournalID:   sop.ID,
						Tags:        []string{},
						NewTags:     []string{newTag},
					}}
					vRes := ""
					err := ev.Publish("/v1/asset/update-tag-by-journal", payload, &vRes, nil)
					if err != nil {
						return nil, fmt.Errorf("error : update attachment tags: %s", err.Error())
					}
				}
			}
		}
	}

	if e := h.Save(sop); e != nil {
		return nil, errors.New("error update master sop: " + e.Error())
	}

	return sop, nil
}
