package mfglogic

import (
	"fmt"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type RoutineEngine struct{}

func (o *RoutineEngine) AddNew(ctx *kaos.Context, payload *mfgmodel.Routine) (*mfgmodel.Routine, error) {
	coID, _, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	rt := new(mfgmodel.Routine)
	db.GetByFilter(rt, dbflex.And(
		dbflex.Eq("SiteID", payload.SiteID),
		dbflex.Eq("ExecutionDate", payload.ExecutionDate),
	))

	if rt.ID != "" {
		return rt, nil // directly return routine without creating the routine details
	}

	err = sebar.Tx(db, true, func(tx *datahub.Hub) error {
		if e := tx.Save(payload); e != nil {
			return e
		}

		setup := tenantcorelogic.GetSequenceSetup(tx, "RoutineDetail", coID)

		assets := []tenantcoremodel.Asset{}
		if e := tx.GetsByFilter(new(tenantcoremodel.Asset), dbflex.And(
			dbflex.Eq("GroupID", "UNT"),
			dbflex.Eq("Dimension.Key", "Site"),
			dbflex.Eq("Dimension.Value", payload.SiteID),
		), &assets); e != nil {
			return e
		}

		tassetM := lo.SliceToMap(assets, func(d tenantcoremodel.Asset) (string, tenantcoremodel.Asset) {
			return d.ID, d
		})

		assetIDs := lo.Map(assets, func(asset tenantcoremodel.Asset, index int) string {
			return asset.ID
		})

		bgAssets, e := BagongAssetGets(assetIDs)
		if e != nil {
			return e
		}

		// payloadSite := struct {
		// 	AssetIDs []interface{}
		// }{
		// 	AssetIDs: codekit.ToInterfaceArray(assetIDs),
		// }

		// responseSite := map[string]struct {
		// 	AssetDateFrom time.Time
		// 	AssetDateTo   time.Time
		// }{}

		// e = Config.EventHub.Publish("/v1/bagong/asset/gets-asset-active-per-site", &payloadSite, &responseSite, nil)
		// if e != nil {
		// 	fmt.Println(e)
		// }

		for _, as := range bgAssets {
			// key := fmt.Sprintf("%s|%s", as.ID, as.Dimension.Get("Site"))
			// if _, ok := responseSite[key]; !ok {
			// 	continue
			// }

			// if payload.ExecutionDate.Before(responseSite[key].AssetDateFrom) || payload.ExecutionDate.After(responseSite[key].AssetDateTo) {
			// 	continue
			// }

			id := ""
			if setup != nil || setup.NumSeqID != "" {
				resp := new(tenantcorelogic.NumSeqClaimRespond)
				Config.EventHub.Publish("/v1/tenant/numseq/claim", &tenantcorelogic.NumSeqClaimPayload{NumberSequenceID: setup.NumSeqID, Date: time.Now()}, resp, nil)
				id = resp.Number
			}

			rd := mfgmodel.RoutineDetail{
				ID:              id,
				RoutineID:       payload.ID,
				AssetID:         as.ID,
				AssetType:       as.AssetType,
				DriveType:       tassetM[as.ID].DriveType,
				Dimension:       tassetM[as.ID].Dimension,
				StatusCondition: mfgmodel.RoutineDetailStatusConditionNotChecked,
			}

			if rd.AssetType == "" {
				rd.AssetType = tassetM[as.ID].AssetType
			}

			if len(as.CurrentUserInfo) > 0 {
				rd.CustomerID = as.CurrentUserInfo[0].CustomerID
			}

			if e := tx.Insert(&rd); e != nil {
				return e
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return payload, nil
}

type CreateForWORequest struct {
	RoutineID                  string
	RoutineChecklistID         string
	EquipmentNo                string
	Kilometers                 float64
	Description                string
	Site                       string
	RoutineChecklistCategories []RoutineChecklistCategoryResponse
}

func (o *RoutineEngine) CreateForWO(ctx *kaos.Context, payload *CreateForWORequest) (interface{}, error) {
	companyID, userID, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	description := ""
	lo.ForEach(payload.RoutineChecklistCategories, func(item RoutineChecklistCategoryResponse, index int) {
		lo.ForEach(item.RoutineChecklistDetails, func(detail RoutineChecklistDetailResponse, index int) {
			if detail.Status == mfgmodel.RoutineChecklistStatusDamaged {
				description += fmt.Sprintf("%s:%s;", detail.Name, detail.Note)
			}
		})
	})

	routine, e := datahub.GetByID(h, new(mfgmodel.Routine), payload.RoutineID)
	if e != nil {
		return nil, fmt.Errorf("Routine not found")
	}

	routineChecklist, e := datahub.GetByID(h, new(mfgmodel.RoutineChecklist), payload.RoutineChecklistID)
	if e != nil {
		return nil, fmt.Errorf("Routine Checklist not found")
	}

	if routineChecklist.IsAlreadyRequest == true {
		return nil, fmt.Errorf("Already request WR before")
	}

	// date := carbon.CreateFromStdTime(routine.ExecutionDate, carbon.Local).ToDateString()
	// utc, _ := time.LoadLocation(carbon.UTC)
	// downTime := carbon.Parse(fmt.Sprintf("%s %s", date, routineChecklist.TimeBreakdown), carbon.Local).ToStdTime().UTC()

	tbSplits := strings.Split(routineChecklist.TimeBreakdown, ":")
	tbH := 0
	tbM := 0
	if len(tbSplits) == 2 {
		tbH = codekit.ToInt(tbSplits[0], codekit.RoundingAuto)
		tbM = codekit.ToInt(tbSplits[1], codekit.RoundingAuto)
	}
	jakartaLoc, _ := time.LoadLocation("Asia/Jakarta")
	execDate := routine.ExecutionDate.In(jakartaLoc) // jadikan WIB karena jam dan menit nya juga WIB
	downTime := time.Date(execDate.Year(), execDate.Month(), execDate.Day(), tbH, tbM, 0, 0, jakartaLoc)

	fmt.Println("routine.ExecutionDate.Location():", routine.ExecutionDate.Location().String())
	fmt.Println("execDate.Location():", execDate.Location().String())
	fmt.Println("downTime.Location():", downTime.Location().String())

	res := new(mfgmodel.WorkRequest)
	if description != "" {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "WorkRequest")

		now := time.Now()
		res.ID = id
		res.TrxDate = &now
		res.EquipmentNo = payload.EquipmentNo
		res.Kilometers = routineChecklist.KmToday
		res.Description = description
		res.Status = ficomodel.JournalStatusDraft
		res.SourceType = mfgmodel.WRSourceTypeRoutineCheck
		res.SourceID = payload.RoutineID
		res.WorkRequestType = "Service"
		res.Dimension = tenantcoremodel.Dimension{}.Set("Site", routine.SiteID)
		res.StartDownTime = &downTime
		res.Name = routineChecklist.Name
		// res.Department = routineChecklist.Department
		// res.TargetFinishTime = &now
		res.TargetFinishTime = &downTime // mba Fanny: untuk target finish time bisa disamakan juga kayak downtimenya
		res.EquipmentType = "UNT"

		rd, _ := datahub.GetByID(h, new(mfgmodel.RoutineDetail), payload.RoutineChecklistID)
		if rd.AssetID != "" {
			tenantAsset, e := datahub.GetByID(h, new(tenantcoremodel.Asset), rd.AssetID)
			if e == nil && tenantAsset.ID != "" {
				res.Dimension = tenantAsset.Dimension
			}
		}

		var reply interface{}
		e = Config.EventHub.Publish("/v1/mfg/work/request/save", res, reply, &kaos.PublishOpts{
			Headers: codekit.M{
				"CompanyID":             companyID,
				sebar.CtxJWTReferenceID: userID,
			},
		})

		if e != nil {
			return res, nil
		}

		// e := h.Save(&res)
		// if e != nil {
		// 	return res, e
		// }

		routineChecklist.IsAlreadyRequest = true
		e = h.Save(routineChecklist)
		if e != nil {
			return res, e
		}
	}

	return res, nil
}
