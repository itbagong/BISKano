package scmlogic

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"gopkg.in/mgo.v2/bson"
)

type ItemRequestDetailEngine struct{}

type ItemRequestDetailSaveMultipleRequest struct {
	ItemRequestID      string
	ItemRequestDetails []scmmodel.ItemRequestDetail
}

func (o *ItemRequestDetailEngine) SaveMultiple(ctx *kaos.Context, payload *ItemRequestDetailSaveMultipleRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ItemRequestID == "" {
		return nil, errors.New("missing: payload")
	}

	p := new(scmmodel.ItemRequest)
	if e := h.GetByID(p, payload.ItemRequestID); e != nil {
		return nil, errors.New("no item request data found: " + e.Error())
	}

	if e := h.DeleteByFilter(new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", payload.ItemRequestID)); e != nil {
		return nil, errors.New("error clear pr details: " + e.Error())
	}

	for _, dt := range payload.ItemRequestDetails {
		dt.ItemRequestID = payload.ItemRequestID
		if e := h.Save(&dt); e != nil {
			return nil, errors.New("error update Movement  Detail: " + e.Error())
		}
	}

	return payload, nil
}

type FulfillmentRequest struct {
	ItemRequestID string
	UserID        string
}

func (o *ItemRequestDetailEngine) Fulfillment(ctx *kaos.Context, h *datahub.Hub, payload *FulfillmentRequest) (interface{}, error) {
	var e error
	var referenceID string

	if payload == nil && payload.ItemRequestID == "" {
		return nil, errors.New("missing: payload")
	}

	//get Item Request By ID
	itemRequest, err := datahub.GetByID(h, new(scmmodel.ItemRequest), payload.ItemRequestID)
	if err != nil {
		return nil, errors.New("missing: item request")
	}

	//get item request detail
	itemDetails := []scmmodel.ItemRequestDetail{}
	e = h.GetsByFilter(new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", itemRequest.ID), &itemDetails)
	if e != nil {
		return nil, errors.New("missing: item details")
	}

	type TransferGroup struct {
		InventDimID   string
		InventDimFrom scmmodel.InventDimension
		Lines         []scmmodel.InventJournalLine
	}

	prLines := []scmmodel.PurchaseJournalLine{}
	movInLines := []scmmodel.InventJournalLine{}
	movOutLines := []scmmodel.InventJournalLine{}
	transGroup := []TransferGroup{}

	lo.ForEach(itemDetails, func(detail scmmodel.ItemRequestDetail, index int) {
		lo.ForEach(detail.DetailLines, func(detailLine scmmodel.ItemRequestDetailLine, indexLine int) {
			switch detailLine.FulfillmentType {
			case "Movement In", "Movement Out":
				line := scmmodel.InventJournalLine{
					LineNo:    1,
					ItemID:    detail.ItemID,
					SKU:       detail.SKU,
					Text:      detail.Description,
					Qty:       detailLine.QtyFulfilled,
					UnitID:    detailLine.UoM,
					Dimension: itemRequest.Dimension,
					InventDim: itemRequest.InventDimTo,
				}

				if detailLine.FulfillmentType == "Movement In" {
					movInLines = append(movInLines, line)
				} else if detailLine.FulfillmentType == "Movement Out" {
					movOutLines = append(movOutLines, line)
				}
			case "Item Transfer":
				line := scmmodel.InventJournalLine{
					LineNo:    1,
					ItemID:    detail.ItemID,
					SKU:       detail.SKU,
					Text:      detail.Description,
					Qty:       detailLine.QtyFulfilled,
					UnitID:    detailLine.UoM,
					Dimension: itemRequest.Dimension,
					InventDim: itemRequest.InventDimTo,
				}

				inventDimFrom := detailLine.InventDimFrom.Calc()
				inventDimID := inventDimFrom.InventDimID

				_, idx, found := lo.FindIndexOf(transGroup, func(d TransferGroup) bool { return d.InventDimID == inventDimID })
				if found {
					transGroup[idx].Lines = append(transGroup[idx].Lines, line)
				} else {
					transGroup = append(transGroup, TransferGroup{
						InventDimID:   inventDimID,
						InventDimFrom: *inventDimFrom,
						Lines:         []scmmodel.InventJournalLine{line},
					})
				}

			case "Purchase Request":
				line := scmmodel.PurchaseJournalLine{
					InventJournalLine: scmmodel.InventJournalLine{
						LineNo:       1,
						ItemID:       detail.ItemID,
						SKU:          detail.SKU,
						Text:         detail.Description,
						Qty:          detailLine.QtyFulfilled,
						UnitID:       detailLine.UoM,
						Dimension:    itemRequest.Dimension,
						InventDim:    itemRequest.InventDimTo,
						RemainingQty: detailLine.QtyFulfilled,
					},
				}

				prLines = append(prLines, line)
			case "Assembly":
				//not implemented yet
				// wo := struct{
				// 	Name string
				// Dimension tenantcoremodel.Dimension
				// InventDim scmmodel.InventDimension
				// Lines:     []scmmodel.InventReceiveIssueLine,
				// }{}
				// 	scmconfig.Config.EventHub().Publish("/v1/mfg/workorder/save", mfgmodel.Model{})
			}
		})
	})

	now := time.Now()

	if len(movInLines) > 0 {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "InventJournal") // TODO: seharusnya pake MWPreAssignSequenceNo

		movement := &scmmodel.InventJournal{
			ID:        id,
			TrxType:   scmmodel.JournalMovementIn,
			Status:    ficomodel.JournalStatusDraft,
			Text:      itemRequest.Name,
			Dimension: itemRequest.Dimension,
			InventDim: itemRequest.InventDimTo,
			Lines:     movInLines,
			ReffNo:    []string{itemRequest.ID},
			TrxDate:   now,
		}

		referenceID = id
		if e = h.Save(movement); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	if len(movOutLines) > 0 {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "InventJournal") // TODO: seharusnya pake MWPreAssignSequenceNo

		movement := &scmmodel.InventJournal{
			ID:        id,
			TrxType:   scmmodel.JournalMovementOut,
			Status:    ficomodel.JournalStatusDraft,
			Text:      itemRequest.Name,
			Dimension: itemRequest.Dimension,
			InventDim: itemRequest.InventDimTo,
			Lines:     movOutLines,
			TrxDate:   now,
			ReffNo:    []string{itemRequest.ID},
		}

		referenceID = id
		if e = h.Save(movement); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	for _, tg := range transGroup {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "InventJournal") // TODO: seharusnya pake MWPreAssignSequenceNo

		transfer := &scmmodel.InventJournal{
			ID:          id,
			Text:        itemRequest.Name,
			Status:      ficomodel.JournalStatusDraft,
			Dimension:   itemRequest.Dimension,
			InventDim:   tg.InventDimFrom,
			InventDimTo: itemRequest.InventDimTo,
			Lines:       tg.Lines,
			TrxType:     scmmodel.JournalTransfer,
			ReffNo:      []string{itemRequest.ID},
			TrxDate:     now,
		}

		referenceID = id
		if e = h.Save(transfer); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	if len(prLines) > 0 {
		// Note: entah kenapa nats tidak bisa di trigger disini, ajax jadi loading lama tidak selesai2
		// purchase := &PurchaseRequestInsertParam{
		// 	Name:      itemRequest.Name,
		// 	Dimension: itemRequest.Dimension,
		// 	Location:  itemRequest.InventDimTo,
		// 	ReffNo:    []string{itemRequest.ID},
		// 	CompanyID: itemRequest.CompanyID,
		// 	Lines:     prLines,
		// }
		// PurchaseRequestInsert(purchase, itemRequest.CompanyID, payload.UserID)

		id, err := tenantcorelogic.GenerateIDFromNumSeq(ctx, "PurchaseRequest") // TODO: seharusnya pake MWPreAssignSequenceNo
		if err != nil {
			return nil, err
		}

		purchase := &scmmodel.PurchaseRequestJournal{
			ID:           id,
			Name:         itemRequest.Name,
			Status:       ficomodel.JournalStatusDraft,
			Dimension:    itemRequest.Dimension,
			Location:     itemRequest.InventDimTo,
			TrxDate:      now,
			DocumentDate: &now,
			PRDate:       &now,
			ExpectedDate: &now,
			ReffNo:       []string{payload.ItemRequestID},
			CompanyID:    itemRequest.CompanyID,
			Priority:     itemRequest.Priority,
			Lines:        prLines,
		}

		referenceID = id
		if e = h.Save(purchase); e != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return nil, e
		}
	}

	return referenceID, nil
}

type ItemRequestDetailGetLinesRequest struct {
	FulfillmentType string
	Keyword         string
	Status          string
	Site            string
	Warehouse       string
	Skip            int
	Take            int
}

type ItemRequestDetailGetLinesResult struct {
	ID            string `grid:"hide" bson:"_id" json:"_id"`
	ItemRequestID string
	ItemID        string
	SKU           string
	Description   string
	ItemType      string
	UoM           string
	QtyRequested  float64
	QtyFulfilled  float64
	QtyAvailable  float64
	QtyRemaining  float64
	Complete      bool
	Remarks       string
	WarehouseID   string
	Created       time.Time
	LastUpdate    time.Time
	Dimension     tenantcoremodel.Dimension
	DetailLines   []scmmodel.ItemRequestDetailLine

	HeaderStatus string
	ItemName     string
	SKUName      string
	ItemRequest  scmmodel.ItemRequest
}

func (o *ItemRequestDetailEngine) GetLines(ctx *kaos.Context, payload *ItemRequestDetailGetLinesRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil {
		return nil, errors.New("missing: payload")
	}

	payload.Keyword = strings.TrimSpace(payload.Keyword)

	if payload.Skip < 0 {
		payload.Skip = 0
	}

	if payload.Take <= 0 {
		payload.Take = 10
	}

	// TODO: bisa lebih di optimasi lagi dengan lookup filter, bila payload.Keyword is provided
	pipe := []bson.M{}
	if payload.FulfillmentType != "" {
		pipe = append(pipe, bson.M{"$match": bson.M{"DetailLines.FulfillmentType": payload.FulfillmentType}})
	}

	pipeLookup := []bson.M{}
	e := Deserialize(fmt.Sprintf(`
    [
		{
			"$lookup": {
				"from": "ItemRequests",
				"localField": "ItemRequestID",
				"foreignField": "_id",
				"as": "requestHeader"
			}
		},
		{
			"$unwind": "$requestHeader"
		},
		{
			"$set": {
				"HeaderStatus": "$requestHeader.Status",
				"ItemRequest": "$requestHeader"
			}
		}
	]
    `), &pipeLookup)
	if e != nil {
		return nil, e
	}
	pipe = append(pipe, pipeLookup...)

	// Filter Match after Lookup
	pipeMatch := []bson.M{}

	if payload.Keyword != "" {
		escapedKeyword := regexp.QuoteMeta(payload.Keyword)
		pipeMatch = append(pipeMatch, bson.M{"$or": []bson.M{
			{"ItemRequestID": bson.M{"$regex": escapedKeyword, "$options": "i"}},
		}})
	}

	if payload.Status != "" {
		pipeMatch = append(pipeMatch, bson.M{"requestHeader.Status": payload.Status})
	}

	if payload.Site != "" {
		pipeMatch = append(pipeMatch, bson.M{"$and": []bson.M{
			{"Dimension.Key": "Site"},
			{"Dimension.Value": payload.Site},
		}})
	}

	if payload.Warehouse != "" {
		pipeMatch = append(pipeMatch, bson.M{"DetailLines.InventDimFrom.WarehouseID": payload.Warehouse})
	}

	if len(pipeMatch) > 0 {
		pipe = append(pipe, bson.M{"$match": bson.M{"$and": pipeMatch}})
	}

	// Sorting
	pipe = append(pipe, bson.M{"$sort": bson.M{"Created": -1, "ItemName": 1, "SKU": 1}})

	pipeSkipTake := make([]bson.M, len(pipe))
	copy(pipeSkipTake, pipe)
	pipeSkipTake = append(pipeSkipTake, bson.M{"$skip": payload.Skip}, bson.M{"$limit": payload.Take})

	results := []ItemRequestDetailGetLinesResult{}
	cmd := dbflex.From(new(scmmodel.ItemRequestDetail).TableName()).Command("pipe", pipeSkipTake)
	if _, e := h.Populate(cmd, &results); e != nil {
		return nil, fmt.Errorf("error when fetching data : %s", e)
	}

	countCmd := dbflex.From(new(scmmodel.ItemRequestDetail).TableName()).Command("pipe", append(pipe, bson.M{"$count": "totalCount"}))
	countResult := []codekit.M{}
	if _, e := h.Populate(countCmd, &countResult); e != nil {
		return nil, fmt.Errorf("error when counting data : %s", e)
	}

	var totalCount int
	if len(countResult) > 0 {
		totalCount = countResult[0].GetInt("totalCount")
	} else {
		totalCount = 0
	}

	skuIDs := make([]string, len(results))
	itemIDs := make([]string, len(results))
	for i, r := range results {
		skuIDs[i] = r.SKU
		itemIDs[i] = r.ItemID
	}

	items := []tenantcoremodel.Item{}
	err := h.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", itemIDs...),
	), &items)
	if err != nil {
		return nil, fmt.Errorf("error when get item: %s", err.Error())
	}

	mapItem := lo.Associate(items, func(v tenantcoremodel.Item) (string, string) {
		return v.ID, v.Name
	})

	specs := []tenantcoremodel.ItemSpec{}
	err = h.Gets(new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", skuIDs...),
	), &specs)
	if err != nil {
		return nil, fmt.Errorf("error when get item spec: %s", err.Error())
	}

	mapSKU := lo.Associate(specs, func(v tenantcoremodel.ItemSpec) (string, string) {
		return v.ID, v.SKU
	})

	mapItemVariant, err := ItemVariantName(h, itemIDs, skuIDs)
	if err != nil {
		return nil, fmt.Errorf("error when get item variant name: %s", err.Error())
	}

	// Normalize Name
	for resI, res := range results {
		itemName := mapItemVariant[res.ItemID+res.SKU]
		results[resI].ItemName = lo.Ternary(itemName != "", itemName, mapItem[res.ItemID])
		results[resI].SKUName = mapSKU[res.SKU]
	}

	res := codekit.M{
		"data":  results,
		"count": totalCount,
	}

	return res, nil
}

type ItemRequestDetailReferenceRequest struct {
	ItemRequestID string
}

func (o *ItemRequestDetailEngine) Reference(ctx *kaos.Context, payload *ItemRequestDetailReferenceRequest) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if payload == nil || payload.ItemRequestID == "" {
		return nil, errors.New("missing: payload")
	}

	requestDetails, _ := datahub.FindByFilter(h, new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", payload.ItemRequestID))
	if len(requestDetails) == 0 {
		return nil, nil
	}

	data := []codekit.M{}
	fulfilmentAlreadyTrace := map[string]bool{}

	//trace fulfilment
	lo.ForEach(requestDetails, func(requestDetail *scmmodel.ItemRequestDetail, _ int) {
		lo.ForEach(requestDetail.DetailLines, func(line scmmodel.ItemRequestDetailLine, _ int) {
			if _, ok := fulfilmentAlreadyTrace[string(line.FulfillmentType)]; !ok {
				switch line.FulfillmentType {
				case scmmodel.ItemRequestFulfillmentTypeItemTransfer:
					//find transfer journal
					journals, _ := datahub.FindByFilter(h, new(scmmodel.InventJournal), dbflex.Eq("ReffNo", requestDetail.ItemRequestID))
					if len(journals) > 0 {
						lo.ForEach(journals, func(journal *scmmodel.InventJournal, _ int) {
							data = append(data, codekit.M{"JournalType": journal.TrxType, "ReffNo": journal.ID, "Status": journal.Status})
							//find gr with reff no transfer
							receiveIssueJournals, _ := datahub.FindByFilter(h, new(scmmodel.InventReceiveIssueJournal), dbflex.Eq("ReffNo", journal.ID))
							if len(receiveIssueJournals) > 0 {
								lo.ForEach(receiveIssueJournals, func(journal *scmmodel.InventReceiveIssueJournal, _ int) {
									data = append(data, codekit.M{"JournalType": journal.TrxType, "ReffNo": journal.ID, "Status": journal.Status})
								})
							}
						})
					}

					//find gi with reff no transfer
				case scmmodel.ItemRequestFulfillmentTypeMovementOut:
					//not yet implementation
				case scmmodel.ItemRequestFulfillmentTypePurchaseRequest:
					//find pr journal
					prJournals, _ := datahub.FindByFilter(h, new(scmmodel.PurchaseRequestJournal), dbflex.Eq("ReffNo", requestDetail.ItemRequestID))
					if len(prJournals) > 0 {
						lo.ForEach(prJournals, func(prJournal *scmmodel.PurchaseRequestJournal, _ int) {
							data = append(data, codekit.M{"JournalType": "Purchase Request", "ReffNo": prJournal.ID, "Status": prJournal.Status})
							//find po journal
							poJournals, _ := datahub.FindByFilter(h, new(scmmodel.PurchaseOrderJournal), dbflex.Eq("ReffNo", prJournal.ID))
							if len(poJournals) > 0 {
								lo.ForEach(poJournals, func(poJournal *scmmodel.PurchaseOrderJournal, _ int) {
									data = append(data, codekit.M{"JournalType": "Purchase Order", "ReffNo": poJournal.ID, "Status": poJournal.Status})
									//find gr journal
									grJournals, _ := datahub.FindByFilter(h, new(scmmodel.InventReceiveIssueJournal), dbflex.Eq("ReffNo", poJournal.ID))
									if len(grJournals) > 0 {
										lo.ForEach(grJournals, func(grJournal *scmmodel.InventReceiveIssueJournal, _ int) {
											data = append(data, codekit.M{"JournalType": grJournal.TrxType, "ReffNo": grJournal.ID, "Status": grJournal.Status})
										})
									}
								})
							}
						})

					}
				default:
				}

				fulfilmentAlreadyTrace[string(line.FulfillmentType)] = true
			}
		})
	})

	//get details
	return codekit.M{"Data": data}, nil
}

type PostSubmitRequest struct {
	ItemRequestID string
}

func (o *ItemRequestDetailEngine) PostSubmit(ctx *kaos.Context, payload *PostSubmitRequest) (string, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "NOK", errors.New("missing: connection")
	}

	if payload == nil {
		return "NOK", errors.New("missing: payload")
	}

	//get item request
	itemRequest, err := datahub.GetByID(h, new(scmmodel.ItemRequest), payload.ItemRequestID)
	if err != nil && MongoNotFound(err) == false {
		return "NOK", fmt.Errorf("Error when get item request: %s", err.Error())
	}

	//get posting profile approval
	param := dbflex.NewQueryParam()
	param.SetSort("-LastUpdate")
	param.SetWhere(dbflex.Eq("SourceID", payload.ItemRequestID))

	postingApproval, err := datahub.GetByParm(h, new(ficomodel.PostingApproval), param)
	if err != nil {
		return "NOK", fmt.Errorf("Error when get posting approval with error: %s", err.Error())
	}

	if (postingApproval.Status == string(ficomodel.JournalStatusApproved) || postingApproval.Status == string(ficomodel.JournalStatusPosted)) && itemRequest.Status == ficomodel.JournalStatusPosted {
		//check pr line already generate or not
		prs, err := datahub.FindByFilter(h, new(scmmodel.PurchaseRequestJournal), dbflex.Eq("ReffNo", itemRequest.ID))
		if err != nil && MongoNotFound(err) == false {
			return "NOK", fmt.Errorf("Error when get pr data: %s", err.Error())
		}

		if len(prs) == 0 {
			err = o.generatePR(h, itemRequest, ctx)
			if err != nil {
				return "NOK", err
			}
		}
	}

	return "OK", nil
}

func (o *ItemRequestDetailEngine) generatePR(h *datahub.Hub, itemRequest *scmmodel.ItemRequest, ctx *kaos.Context) error {
	//get item request detail
	irLines, err := datahub.FindByFilter(h, new(scmmodel.ItemRequestDetail), dbflex.Eq("ItemRequestID", itemRequest.ID))
	if err != nil && MongoNotFound(err) == false {
		return fmt.Errorf("Error when get item request detail: %s", err.Error())
	}

	prLines := []scmmodel.PurchaseJournalLine{}

	lo.ForEach(irLines, func(line *scmmodel.ItemRequestDetail, _ int) {
		lo.ForEach(line.DetailLines, func(lineDetail scmmodel.ItemRequestDetailLine, _ int) {
			if lineDetail.FulfillmentType == scmmodel.ItemRequestFulfillmentTypePurchaseRequest {
				line := scmmodel.PurchaseJournalLine{
					InventJournalLine: scmmodel.InventJournalLine{
						LineNo:       1,
						ItemID:       line.ItemID,
						SKU:          line.SKU,
						Text:         line.Description,
						Qty:          lineDetail.QtyFulfilled,
						UnitID:       lineDetail.UoM,
						Dimension:    itemRequest.Dimension,
						InventDim:    itemRequest.InventDimTo,
						RemainingQty: lineDetail.QtyFulfilled,
					},
				}

				prLines = append(prLines, line)
			}
		})
	})

	if len(prLines) > 0 {
		id, err := tenantcorelogic.GenerateIDFromNumSeq(ctx, "PurchaseRequest") // TODO: seharusnya pake MWPreAssignSequenceNo
		if err != nil {
			return fmt.Errorf("Error when generate ID PR: %s", err.Error())
		}

		now := time.Now()

		purchase := &scmmodel.PurchaseRequestJournal{
			ID:           id,
			Name:         itemRequest.Name,
			Status:       ficomodel.JournalStatusDraft,
			Dimension:    itemRequest.Dimension,
			Location:     itemRequest.InventDimTo,
			TrxDate:      now,
			DocumentDate: &now,
			PRDate:       &now,
			ExpectedDate: &now,
			ReffNo:       []string{itemRequest.ID},
			CompanyID:    itemRequest.CompanyID,
			Priority:     itemRequest.Priority,
			Lines:        prLines,
		}

		if err = h.Save(purchase); err != nil { // TODO: seharusnya pakai nats insert agar semua MW kepanggil
			return fmt.Errorf("Error when save purchase request with error: %s", err.Error())
		}
	}

	return nil
}
