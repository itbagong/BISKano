package shelogic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

type MeetingLogic struct {
}

func (obj *MeetingLogic) Save(ctx *kaos.Context, sc *shemodel.Meeting) (*shemodel.Meeting, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// generate number Meeting
	tenantcorelogic.MWPreAssignSequenceNo("Meeting", false, "_id")(ctx, &sc)
	vNewTag := []string{}
	vNewTag = append(vNewTag, string(shemodel.MODULE_ATTACHMENT_MEETING)+"_"+sc.ID)

	// if len(sc.Result) > 0 {
	// 	// set status
	// 	sc.PicaStatus = string(ficomodel.JournalStatusDraft)

	// 	for _, val := range sc.Result {
	// 		if val.Pica != nil && val.Pica.EmployeeID != "" {
	// 			//generate number pica
	// 			tenantcorelogic.MWPreAssignSequenceNo("Pica", false, "_id")(ctx, &val.Pica)
	// 			vNewTag = append(vNewTag, string(shemodel.MODULE_ATTACHMENT_PICA)+"_"+val.Pica.ID)

	// 			//save to pica
	// 			if e := h.GetByID(new(shemodel.Pica), val.Pica.ID); e != nil {
	// 				val.Pica.JournalTypeID = "SHE-001"
	// 				val.Pica.PostingProfileID = "PP-SHE"
	// 				val.Pica.SourceNumber = sc.ID
	// 				val.Pica.SourceModule = shemodel.MODULE_MEETING
	// 				val.Pica.Status = string(ficomodel.JournalStatusDraft)
	// 				if e := h.Insert(val.Pica); e != nil {
	// 					return nil, errors.New("error insert Pica: " + e.Error())
	// 				}
	// 			} else {
	// 				if e := h.Save(val.Pica); e != nil {
	// 					return nil, errors.New("error update Pica: " + e.Error())
	// 				}
	// 			}
	// 		}
	// 	}
	// }

	if e := h.GetByID(new(shemodel.Meeting), sc.ID); e != nil {
		if e := h.Insert(sc); e != nil {
			return nil, errors.New("error insert Meeting: " + e.Error())
		}
	} else {
		if e := h.Save(sc); e != nil {
			return nil, errors.New("error update Meeting: " + e.Error())
		}
	}

	// update attachment
	_, e := UpdateTagByJournal(ctx, string(shemodel.MODULE_ATTACHMENT_MEETING), sc.ID, []string{}, vNewTag)
	if e != nil {
		return nil, e
	}

	if sc.Status == string(shemodel.SHEStatusCompleted) {
		vST := shemodel.SummaryTransaction{
			Module:    shemodel.MODULE_MEETING,
			RefID:     sc.ID,
			CreatedBy: "",
		}

		if e := h.Insert(&vST); e != nil {
			return nil, errors.New("error insert Summary Transaction: " + e.Error())
		}
	}

	return sc, nil
}
