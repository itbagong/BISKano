package scmlogic

import (
	"fmt"

	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
)

func ValidateBalance(h *datahub.Hub, qty float64, unitID string, itemID string, companyID string, inventDim scmmodel.InventDimension) error {
	opt := ItemBalanceOpt{
		CompanyID: companyID,
		ItemIDs:   []string{itemID},
		InventDim: inventDim,
	}

	ibs, e := NewItemBalanceHub(h).Get(nil, opt)
	if e != nil || len(ibs) == 0 {
		return fmt.Errorf("Sorry, Qty not enough, no balance found")
	}

	isMore, defUnit, e := MoreThanDefaultUnit(h, qty, unitID, ibs[0].Qty, itemID)
	if e != nil {
		return fmt.Errorf("error check item balance: %s", e.Error())
	}
	if isMore {
		return fmt.Errorf("Sorry, Qty not enough, balance: %v %s", ibs[0].Qty, defUnit.Name)
	}

	return nil
}
