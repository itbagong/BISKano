package scmlogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

type GoodIssueBatchEngine struct{}

type GoodIssueBatchMultipleSaveRequest struct {
	GoodIssueID string
	Batchs      []scmmodel.GoodIssueItemBatch
}

func (o *GoodIssueBatchEngine) SaveMultiple(ctx *kaos.Context, payload *GoodIssueBatchMultipleSaveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.GoodIssueID == "" {
		return nil, errors.New("missing: payload")
	}

	//clear data by movement in id before save
	goodIssueID := payload.GoodIssueID
	e := h.DeleteByFilter(new(scmmodel.GoodIssueItemBatch), dbflex.Eq("GoodIssueID", goodIssueID))
	if e != nil {
		return nil, errors.New("error clear item batchs: " + e.Error())
	}

	for _, itemBatch := range payload.Batchs {
		itemBatch.GoodIssueID = payload.GoodIssueID
		if e := h.GetByID(new(scmmodel.GoodIssueItemBatch), itemBatch.ID); e != nil {
			if e := h.Insert(&itemBatch); e != nil {
				return nil, errors.New("error insert Good Issue Batch: " + e.Error())
			}
		} else {
			if e := h.Save(&itemBatch); e != nil {
				return nil, errors.New("error update Good Issue Batch: " + e.Error())
			}
		}
	}

	return payload, nil
}
