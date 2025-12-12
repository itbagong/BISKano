package ficologic

import (
	"errors"
	"fmt"
	"strconv"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type FixedAssetNumberHandler struct {
}

type FixedAssetNumberRequest struct {
	AssetGroup       string
	FixedAssetNumber string
	IsUsed           bool
}

type FixedAssetNumberResponse struct {
	InitialAssetNumber int
	FixedAssetNumber   string
}

func (obj *FixedAssetNumberHandler) GetInitialNumber(ctx *kaos.Context, fanr *FixedAssetNumberRequest) (*FixedAssetNumberResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}
	result := FixedAssetNumberResponse{
		InitialAssetNumber: 1,
	}

	if fanr == nil {
		return nil, errors.New("payload is asset group required")
	} else {
		if fanr.AssetGroup == "" {
			return nil, errors.New("payload is asset group required")
		}
	}

	// get last sequence by fixed assed group
	fixedAssetNumberCodes := []ficomodel.FixedAssetNumberCode{}
	e := h.Gets(new(ficomodel.FixedAssetNumberCode), dbflex.NewQueryParam().SetWhere(dbflex.Eq("_id", fanr.AssetGroup)), &fixedAssetNumberCodes)
	if e != nil {
		return &result, errors.New("Failed populate data fixed asset number code: " + e.Error())
	}

	isCheckAsset := false
	if len(fixedAssetNumberCodes) > 0 {
		if fixedAssetNumberCodes[0].LastSequence == 0 {
			isCheckAsset = true
		} else {
			result.InitialAssetNumber = fixedAssetNumberCodes[0].LastSequence + 1
		}
	} else {
		isCheckAsset = true
	}

	if isCheckAsset {
		// get last asset by group id
		assetCounts, e := h.Count(new(tenantcoremodel.Asset), dbflex.NewQueryParam().SetWhere(
			dbflex.And(
				dbflex.Eq("GroupID", fanr.AssetGroup),
			)))
		if e != nil {
			return &result, errors.New("Failed populate data asset: " + e.Error())
		}

		// get last number
		result.InitialAssetNumber = assetCounts + 1
	}

	return &result, nil
}

func (obj *FixedAssetNumberHandler) GetReadyFixedAssetNumber(ctx *kaos.Context, fanr *FixedAssetNumberRequest) (*FixedAssetNumberResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}
	result := FixedAssetNumberResponse{}

	if fanr == nil {
		return nil, errors.New("payload is asset group required")
	} else {
		if fanr.AssetGroup == "" {
			return nil, errors.New("payload is asset group required")
		}
	}

	// get last fixed asset number list by fixed assed group
	fixedAssetNumberLists := []ficomodel.FixedAssetNumberList{}
	e := h.Gets(new(ficomodel.FixedAssetNumberList), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("FixedAssetGrup", fanr.AssetGroup),
			dbflex.Eq("IsUsed", false),
		)), &fixedAssetNumberLists)
	if e != nil {
		return &result, errors.New("Failed populate data fixed asset number list: " + e.Error())
	}

	if len(fixedAssetNumberLists) > 0 {
		firstReadyNumber := ""
		seqTemp := 0
		for i, val := range fixedAssetNumberLists {
			if i == 0 {
				seqTemp = val.Sequence
				firstReadyNumber = val.ID
				continue
			}

			if val.Sequence < seqTemp {
				seqTemp = val.Sequence
				firstReadyNumber = val.ID
			}
		}

		result.FixedAssetNumber = firstReadyNumber
	}

	return &result, nil
}

func (obj *FixedAssetNumberHandler) Insert(ctx *kaos.Context, fan *ficomodel.FixedAssetNumber) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if fan == nil {
		return nil, errors.New("payload fixed asset number is required")
	}

	ev, _ := ctx.DefaultEvent()

	e := obj.SaveFixedAssetNumber(h, fan, ev, true)
	if e != nil {
		return nil, errors.New("error save fixed asset number: " + e.Error())
	}

	return fan, nil
}

func (obj *FixedAssetNumberHandler) Save(ctx *kaos.Context, fan *ficomodel.FixedAssetNumber) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if fan == nil {
		return nil, errors.New("payload fixed asset number is required")
	}

	// delete data fixed asset number
	_, e := DeleteFixedAssetNumberList(h, fan)
	if e != nil {
		return nil, errors.New("error delete fixed asset number: " + e.Error())
	}

	ev, _ := ctx.DefaultEvent()

	e = obj.SaveFixedAssetNumber(h, fan, ev, false)
	if e != nil {
		if e.Error() == "error save fixed asset number" {
			// if error save fixed asset number, delete all data fixed asset number list by fixed number id
			_, e := DeleteFixedAssetNumberList(h, fan)
			if e != nil {
				return nil, errors.New("error delete fixed asset number list: " + e.Error())
			}
		}

		return nil, errors.New("error save fixed asset number: " + e.Error())
	}

	return fan, nil
}

func (obj *FixedAssetNumberHandler) Delete(ctx *kaos.Context, fan *ficomodel.FixedAssetNumber) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if fan == nil {
		return nil, errors.New("payload fixed asset number is required")
	}

	// delete data fixed asset number
	_, e := DeleteFixedAssetNumberList(h, fan)
	if e != nil {
		return nil, errors.New("error delete fixed asset number: " + e.Error())
	}

	// delete data fixed asset number
	e = h.Delete(fan)
	if e != nil {
		return nil, errors.New("error delete fixed asset number")
	}

	return fan, nil
}

func (obj *FixedAssetNumberHandler) UpdateIsUsed(ctx *kaos.Context, fanr *FixedAssetNumberRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if fanr == nil {
		return nil, errors.New("payload fixed asset number is required")
	} else {
		if fanr.FixedAssetNumber == "" {
			return nil, errors.New("ID asset is required")
		}
	}

	// get fixed asset number list by id
	fixedAssetNumberLists := []ficomodel.FixedAssetNumberList{}
	e := h.Gets(new(ficomodel.FixedAssetNumberList), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("_id", fanr.FixedAssetNumber),
		)), &fixedAssetNumberLists)
	if e != nil {
		return fanr, errors.New("Failed populate data fixed asset number list: " + e.Error())
	}

	if len(fixedAssetNumberLists) > 0 {
		fixedAssetNumberList := fixedAssetNumberLists[0]
		fixedAssetNumberList.IsUsed = fanr.IsUsed

		e = h.Save(&fixedAssetNumberList)
		if e != nil {
			return fanr, errors.New("Failed update data fixed asset number list: " + e.Error())
		}
	}

	return fanr, nil
}

func GenerateSeqNumber(value int, numberRange int) (string, error) {
	result := ""

	x := 10
	for i := 1; i <= numberRange; i++ {
		if value < x {
			numberRange = numberRange - i
			for i := 0; i < numberRange; i++ {
				result = result + "0"
			}
			result = result + strconv.Itoa(value)
			break
		}
		x = x * 10
	}

	return result, nil
}

func (obj *FixedAssetNumberHandler) SaveFixedAssetNumber(h *datahub.Hub, fan *ficomodel.FixedAssetNumber, ev kaos.EventHub, isInsert bool) error {
	var (
		db  *datahub.Hub
		err error
	)

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	// save data fixed asset number
	err = db.Save(fan)
	if err != nil {
		return errors.New("error save fixed asset number")
	}

	// get fixed asset number code
	fixedAssetNumberCodes := []ficomodel.FixedAssetNumberCode{}
	mapFixedAssetNumberCode := map[string]ficomodel.FixedAssetNumberCode{}
	err = db.Gets(new(ficomodel.FixedAssetNumberCode), nil, &fixedAssetNumberCodes)
	if err != nil {
		return errors.New("Failed populate data fixed asset number code: " + err.Error())
	}

	// create map fixed asset number code
	if len(fixedAssetNumberCodes) > 0 {
		for _, val := range fixedAssetNumberCodes {
			mapFixedAssetNumberCode[val.ID] = val
		}
	}

	if len(fan.Details) > 0 {
		// preparation save data fixed asset number list
		resultList := []ficomodel.FixedAssetNumberList{}
		for _, val := range fan.Details {
			fixedAssetNumberCode := ficomodel.FixedAssetNumberCode{}
			// get value for mapFixedAssetNumberCode
			if v, ok := mapFixedAssetNumberCode[val.FixedAssetGrup]; ok {
				fixedAssetNumberCode = v
			} else {
				err = errors.New("please register code for asset group")
				return errors.New("please register code for asset group")
			}

			if val.NumberAsset == 0 {
				continue
			}

			seq := 0
			for i := 0; i < val.NumberAsset; i++ {
				seq = val.InitialAssetNumber + i
				seqString, _ := GenerateSeqNumber(seq, 5)

				result := ficomodel.FixedAssetNumberList{}
				result.ID = fixedAssetNumberCode.GroupCode + "-" + seqString
				result.AssetName = val.AssetName
				result.FixedAssetNumberID = fan.ID
				result.FixedAssetGrup = val.FixedAssetGrup
				result.GroupCode = fixedAssetNumberCode.GroupCode
				result.Sequence = seq
				result.IsUsed = true

				resultList = append(resultList, result)
			}

			// update last sequence in fixed Asset Number Code
			fixedAssetNumberCode.LastSequence = seq
			err = db.Save(&fixedAssetNumberCode)
			if err != nil {
				return errors.New("error update sequence fixed asset number code.")
			}
		}

		// save data fixed asset number list and asset
		if len(resultList) > 0 {
			IDs := []string{}
			tenantAsset := []tenantcoremodel.Asset{}
			for _, val := range resultList {
				// save data fixed asset number list
				err = db.Save(&val)
				if err != nil {
					// if error delete all data by fixed number id
					// h.DeleteQuery(new(ficomodel.FixedAssetNumberList), dbflex.Eq("FixedAssetNumberID", fan.ID))
					return errors.New("error generate fixed asset number.")
				}

				// save data asset tenant
				newAsset := tenantcoremodel.Asset{
					ID:        val.ID,
					GroupID:   val.FixedAssetGrup,
					Name:      val.AssetName,
					Dimension: fan.Dimension,
				}
				err = db.Save(&newAsset)
				if err != nil {
					return errors.New("error insert new asset.")
				}

				// save id for payload to create bg asset where isInsert true
				IDs = append(IDs, val.ID)
				tenantAsset = append(tenantAsset, newAsset)
			}

			if isInsert {
				payload := struct {
					ID          []string
					TenantAsset []tenantcoremodel.Asset
					IsActive    bool
				}{
					ID:          IDs,
					TenantAsset: tenantAsset,
					IsActive:    true,
				}

				// save data asset bagong
				if ev == nil {
					return errors.New("nil: EventHub")
				}
				res := ""
				e := ev.Publish("/v1/bagong/asset/add-asset", &payload, &res, nil)
				fmt.Printf("asset id: %s | asset/insert e: %s\n", IDs, e)
			}
		}
	}

	return nil
}

func DeleteFixedAssetNumberList(h *datahub.Hub, fan *ficomodel.FixedAssetNumber) ([]ficomodel.FixedAssetNumberList, error) {
	// get fixed asset number list by fixed assed number id
	fixedAssetNumberLists := []ficomodel.FixedAssetNumberList{}
	e := h.Gets(new(ficomodel.FixedAssetNumberList), dbflex.NewQueryParam().SetWhere(dbflex.Eq("FixedAssetNumberID", fan.ID)), &fixedAssetNumberLists)
	if e != nil {
		return nil, errors.New("Failed populate data fixed asset number list: " + e.Error())
	}

	// validation for edit, check isused or not
	if len(fixedAssetNumberLists) > 0 {
		for _, val := range fixedAssetNumberLists {
			if val.IsUsed {
				return nil, errors.New("cannot edit or delete this transaction. because fixed asset number already use in other transaction")
			}
		}
	}

	// mapping group by FixedAssetGrup
	mapfixedAssetNumberList := map[string]int{}
	if len(fixedAssetNumberLists) > 0 {
		for _, val := range fixedAssetNumberLists {
			if v, ok := mapfixedAssetNumberList[val.FixedAssetGrup]; ok {
				mapfixedAssetNumberList[val.FixedAssetGrup] = v + 1
			} else {
				mapfixedAssetNumberList[val.FixedAssetGrup] = 1
			}

			// delete exist
			h.Delete(&val)
		}
	}

	fmt.Println("mapfixedAssetNumberList", mapfixedAssetNumberList)

	// get fixed asset number code
	fixedAssetNumberCodes := []ficomodel.FixedAssetNumberCode{}
	mapFixedAssetNumberCode := map[string]ficomodel.FixedAssetNumberCode{}
	e = h.Gets(new(ficomodel.FixedAssetNumberCode), nil, &fixedAssetNumberCodes)
	if e != nil {
		return nil, errors.New("Failed populate data fixed asset number code: " + e.Error())
	}

	// create map fixed asset number code
	if len(fixedAssetNumberCodes) > 0 {
		for _, val := range fixedAssetNumberCodes {
			mapFixedAssetNumberCode[val.ID] = val
		}
	}

	if len(mapfixedAssetNumberList) > 0 {
		for key, val := range mapfixedAssetNumberList {
			fixedAssetNumberCode := ficomodel.FixedAssetNumberCode{}
			// get value for mapFixedAssetNumberCode
			if v, ok := mapFixedAssetNumberCode[key]; ok {
				fixedAssetNumberCode = v
			} else {
				return nil, errors.New("please register code for asset group")
			}

			// update last sequence in fixed Asset Number Code
			fixedAssetNumberCode.LastSequence = fixedAssetNumberCode.LastSequence - val
			e := h.Save(&fixedAssetNumberCode)
			if e != nil {
				return nil, errors.New("error update sequence fixed asset number code.")
			}
		}
	}

	return fixedAssetNumberLists, nil
}
