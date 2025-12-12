package shelogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
)

type PicaLogic struct {
}

func (obj *PicaLogic) TakeAction(ctx *kaos.Context, pi *shemodel.Pica) (*shemodel.Pica, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}
	vNewTag := []string{}
	vNewTag = append(vNewTag, string(shemodel.MODULE_ATTACHMENT_PICA)+"_"+pi.ID)

	// update attachment
	_, e := UpdateTagByJournal(ctx, string(shemodel.MODULE_ATTACHMENT_PICA), pi.ID, []string{}, vNewTag)
	if e != nil {
		return nil, e
	}

	// update pica
	pi.Status = string(shemodel.SHEStatusCompleted)
	if e := h.Save(pi); e != nil {
		return nil, errors.New("error update pica: " + e.Error())
	}

	// update source status
	if pi.SourceModule == shemodel.MODULE_SAFETYCARD {
		// get safetycard
		safetyCard := new(shemodel.SafetyCard)
		if e := h.GetByID(safetyCard, pi.SourceNumber); e != nil {
			return nil, fmt.Errorf(fmt.Sprintf("safety card not found: %s", pi.SourceNumber))
		}

		// update status pica safety card to completed
		// safetyCard.DetailsFinding.Status = shemodel.SHECompleted
		safetyCard.Pica.Status = string(shemodel.SHEStatusCompleted)
		if e := h.Save(safetyCard); e != nil {
			return nil, errors.New("error update safety card: " + e.Error())
		}
	}

	return pi, nil
}
