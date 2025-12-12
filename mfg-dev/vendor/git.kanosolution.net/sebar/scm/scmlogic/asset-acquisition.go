package scmlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type AssetAcquisitionEngine struct{}

type AcquisitionAfterSaveRequest struct {
	AssetAcquisitionID string
}

func (e *AssetAcquisitionEngine) AfterSave(ctx *kaos.Context, payload *AcquisitionAfterSaveRequest) (string, error) {
	userID := sebar.GetUserIDFromCtx(ctx)
	if userID == "" {
		userID = "SYSTEM"
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return "NOK", errors.New("missing: connection")
	}

	if payload.AssetAcquisitionID == "" {
		return "NOK", errors.New("missing: payload")
	}

	//get asset acquisition by id
	acquisition, err := datahub.GetByID(h, new(scmmodel.AssetAcquisitionJournal), payload.AssetAcquisitionID)
	if err != nil {
		return "NOK", errors.New("Asset acquisition not found")
	}

	lo.ForEach(acquisition.ItemTranfers, func(item scmmodel.AssetItemTransfer, index int) {
		//update stock
		balance := NewInventBalanceCalc(h)
		balance.Update(&scmmodel.ItemBalance{
			CompanyID: acquisition.CompanyID,
			ItemID:    item.ItemID,
			InventDim: acquisition.TransferFrom,
			Qty:       (-1 * item.Qty),
		})
	})

	assetIDs := lo.FilterMap(acquisition.AssetRegisters, func(asset scmmodel.AssetRegister, index int) (string, bool) {
		return asset.AssetID, asset.DoesFixedAssetNumberIsExist == true // UI: dicentang
	})

	if len(assetIDs) > 0 {
		payload := struct {
			IDs             []string
			AcquisitionDate time.Time
		}{
			IDs:             assetIDs,
			AcquisitionDate: acquisition.Created,
		}

		e := scmconfig.Config.EventHub().Publish(
			"/v1/bagong/asset/acquire",
			&payload,
			nil,
			&kaos.PublishOpts{Headers: codekit.M{"CompanyID": acquisition.CompanyID, sebar.CtxJWTReferenceID: userID}},
		)
		fmt.Println("asset/acquire e:", e)
	}

	fanIDs := lo.FilterMap(acquisition.AssetRegisters, func(asset scmmodel.AssetRegister, index int) (string, bool) {
		return asset.AssetID, asset.DoesFixedAssetNumberIsExist == false // UI: tidak dicentang
	})

	if len(fanIDs) > 0 {
		payload := struct {
			IDs    []string
			IsUsed bool
		}{
			IDs:    fanIDs,
			IsUsed: true,
		}

		e := scmconfig.Config.EventHub().Publish(
			"/v1/fico/fixedassetnumberlist/use",
			&payload,
			nil,
			&kaos.PublishOpts{Headers: codekit.M{"CompanyID": acquisition.CompanyID, sebar.CtxJWTReferenceID: userID}},
		)
		fmt.Println("fixedassetnumberlist/use e:", e)
	}

	return "OK", nil
}
