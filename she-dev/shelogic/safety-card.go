package shelogic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
)

type SCLogic struct {
}

func (obj *SCLogic) Save(ctx *kaos.Context, sc *shemodel.SafetyCard) (*shemodel.SafetyCard, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	// generate number safety card
	tenantcorelogic.MWPreAssignSequenceNo("SafetyCard", false, "_id")(ctx, &sc)
	vNewTag := []string{}
	vNewTag = append(vNewTag, string(shemodel.MODULE_ATTACHMENT_SAFETYCARD)+"_"+sc.ID)

	// if sc.Pica != nil && sc.Pica.EmployeeID != "" {
	// 	//generate number pica
	// 	tenantcorelogic.MWPreAssignSequenceNo("Pica", false, "_id")(ctx, &sc.Pica)
	// 	vNewTag = append(vNewTag, string(shemodel.MODULE_ATTACHMENT_PICA)+"_"+sc.Pica.ID)

	// 	//save to pica
	// 	if e := h.GetByID(new(shemodel.Pica), sc.Pica.ID); e != nil {
	// 		sc.Pica.JournalTypeID = "SHE-001"
	// 		sc.Pica.PostingProfileID = "PP-SHE"
	// 		sc.Pica.SourceNumber = sc.ID
	// 		sc.Pica.SourceModule = shemodel.MODULE_SAFETYCARD
	// 		sc.Pica.Status = string(ficomodel.JournalStatusDraft)
	// 		if e := h.Insert(sc.Pica); e != nil {
	// 			return nil, errors.New("error insert Pica: " + e.Error())
	// 		}
	// 	} else {
	// 		if e := h.Save(sc.Pica); e != nil {
	// 			return nil, errors.New("error update Pica: " + e.Error())
	// 		}
	// 	}
	// }

	if sc.Pica != nil {
		sc.IsPica = true
	} else {
		sc.IsPica = false
	}

	if e := h.GetByID(new(shemodel.SafetyCard), sc.ID); e != nil {
		if e := h.Insert(sc); e != nil {
			return nil, errors.New("error insert Safety Card: " + e.Error())
		}
	} else {
		if e := h.Save(sc); e != nil {
			return nil, errors.New("error update Safety Card: " + e.Error())
		}
	}

	// update attachment
	_, e := UpdateTagByJournal(ctx, string(shemodel.MODULE_ATTACHMENT_SAFETYCARD), sc.ID, []string{}, vNewTag)
	if e != nil {
		return nil, e
	}

	if sc.DetailsFinding.Status == shemodel.SHEStatusCompleted {
		vST := shemodel.SummaryTransaction{
			Module:    shemodel.MODULE_SAFETYCARD,
			RefID:     sc.ID,
			CreatedBy: "",
		}

		if e := h.Insert(&vST); e != nil {
			return nil, errors.New("error insert Summary Transaction: " + e.Error())
		}
	}

	return sc, nil
}
