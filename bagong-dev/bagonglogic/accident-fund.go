package bagonglogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/kaos"
	bm "git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sebar"
)

type AccidentFundEngine struct{}

func (engine *AccidentFundEngine) InsertAccidentFundDetail(ctx *kaos.Context, payload *bm.AccidentFundDetail) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get SiteEntry
	accFund := new(bm.AccidentFund)
	if e := hub.GetByID(accFund, payload.AccidentFundID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("accident fund not found: %s", payload.AccidentFundID))
	}

	// check if mutation is minus & greater than balance
	if payload.Mutation < 0 {
		if -(payload.Mutation) > accFund.Balance {
			return "", fmt.Errorf("mutation is greater than balance")
		}
	}

	// save accident funddetail
	if e := hub.Save(payload); e != nil {
		return "", fmt.Errorf("error when save accident fund: %s", e.Error())
	}

	// calculate balance
	accFund.Balance += payload.Mutation
	// save accident fund
	if e := hub.Save(accFund); e != nil {
		return "", fmt.Errorf("error when save accident fund: %s", e.Error())
	}

	return "success", nil
}

func (engine *AccidentFundEngine) SaveAccidentFundDetail(ctx *kaos.Context, payload *bm.AccidentFundDetail) (interface{}, error) {
	if payload.ID == "" {
		return "", errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get accident fund detail
	oldAccFundDetail := new(bm.AccidentFundDetail)
	if e := hub.GetByID(oldAccFundDetail, payload.ID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("accident fund detail not found: %s", payload.ID))
	}

	// get SiteEntry
	accFund := new(bm.AccidentFund)
	if e := hub.GetByID(accFund, oldAccFundDetail.AccidentFundID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("accident fund not found: %s", oldAccFundDetail.AccidentFundID))
	}

	// reset balance value
	if oldAccFundDetail.Mutation < 0 {
		accFund.Balance += -(oldAccFundDetail.Mutation)
	} else {
		accFund.Balance -= oldAccFundDetail.Mutation
	}

	// check if mutation is minus & greater than balance
	if payload.Mutation < 0 {
		if -(payload.Mutation) > accFund.Balance {
			return "", fmt.Errorf("mutation is greater than balance")
		}
	}

	// save accident funddetail
	oldAccFundDetail.Mutation = payload.Mutation
	if e := hub.Save(oldAccFundDetail); e != nil {
		return "", fmt.Errorf("error when save accident fund detail: %s", e.Error())
	}

	// calculate balance
	accFund.Balance += payload.Mutation
	// save accident fund
	if e := hub.Save(accFund); e != nil {
		return "", fmt.Errorf("error when save accident fund: %s", e.Error())
	}

	return "success", nil
}

func (engine *AccidentFundEngine) DeleteAccidentFundDetail(ctx *kaos.Context, payload *bm.AccidentFundDetail) (interface{}, error) {
	if payload.ID == "" {
		return "", errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}

	// get accident fund detail
	accFundDetail := new(bm.AccidentFundDetail)
	if e := hub.GetByID(accFundDetail, payload.ID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("accident fund detail not found: %s", payload.ID))
	}

	// get SiteEntry
	accFund := new(bm.AccidentFund)
	if e := hub.GetByID(accFund, accFundDetail.AccidentFundID); e != nil {
		return "", fmt.Errorf(fmt.Sprintf("accident fund not found: %s", accFundDetail.AccidentFundID))
	}

	// delete accident fund detail
	if e := hub.Delete(accFundDetail); e != nil {
		return "", fmt.Errorf("error when delete accident fund: %s", e.Error())
	}

	// calculate balance
	if accFundDetail.Mutation < 0 {
		accFund.Balance += -(accFundDetail.Mutation)
	} else {
		accFund.Balance -= accFundDetail.Mutation
	}

	// save accident fund
	if e := hub.Save(accFund); e != nil {
		return "", fmt.Errorf("error when save accident fund: %s", e.Error())
	}

	return "success", nil
}
