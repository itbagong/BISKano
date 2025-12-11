package bagonglogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type AssetMovementEngine struct{}

func (o *AssetMovementEngine) GetAssets(ctx *kaos.Context, payload *dbflex.QueryParam) ([]bagongmodel.AssetMovementLine, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	query := dbflex.NewQueryParam()
	// if payload.Skip != 0 {
	// 	query = query.SetSkip(payload.Skip)
	// }

	// if payload.Take != 0 {
	// 	query = query.SetTake(payload.Take)
	// }

	if len(payload.Sort) > 0 {
		query = query.SetSort(payload.Sort...)
	}

	filters := []*dbflex.Filter{}

	if payload != nil {
		if payload.Where != nil {
			filters2 := []*dbflex.Filter{}
			opValWhere := dbflex.OpOr
			if len(payload.Where.Items) > 0 {
				vItems := payload.Where.Items
				if len(vItems) > 0 {
					opValWhere = string(payload.Where.Op)
					for _, val := range vItems {
						fieldVal := val.Field
						opVal := val.Op
						if opVal == dbflex.OpContains {
							aInterface := val.Value.([]interface{})
							aString := make([]string, len(aInterface))
							for i, v := range aInterface {
								aString[i] = v.(string)
							}
							if len(aString) > 0 {
								if aString[0] != "" {
									filters2 = append(filters2, dbflex.Contains(fieldVal, aString[0]))
								}
							}
						} else if opVal == dbflex.OpIn {
							aInterface := val.Value.([]interface{})
							aString := []string{}
							for _, v := range aInterface {
								aString = append(aString, v.(string))
							}
							if len(aInterface) > 0 {
								filters2 = append(filters2, dbflex.In(fieldVal, aString...))
							}
						} else if opVal == dbflex.OpEq {
							aInterface := val.Value.(string)
							filters2 = append(filters2, dbflex.Eq(fieldVal, aInterface))
						}
					}
				}
			} else {
				fieldVal := payload.Where.Field
				opVal := payload.Where.Op
				if opVal == dbflex.OpEq {
					aInterface := payload.Where.Value.(string)
					filters2 = append(filters2, dbflex.Eq(fieldVal, aInterface))
				}
			}
			if len(filters2) > 0 {
				if opValWhere == string(dbflex.OpAnd) {
					filters = append(filters, dbflex.And(filters2...))
				} else {
					filters = append(filters, dbflex.Or(filters2...))
				}
			}
		}
	}

	query = query.SetWhere(dbflex.And(filters...))

	assets := []bagongmodel.Asset{}
	err := h.Gets(new(bagongmodel.Asset), query, &assets)
	if err != nil {
		return nil, fmt.Errorf("error when get asset: %s", err.Error())
	}

	custs := []tenantcoremodel.Customer{}
	err = h.Gets(new(tenantcoremodel.Customer), nil, &custs)
	if err != nil {
		return nil, fmt.Errorf("error when get customer: %s", err.Error())
	}
	mapCusts := lo.Associate(custs, func(detail tenantcoremodel.Customer) (string, string) {
		return detail.ID, detail.Name
	})

	// fmt.Println("len(assets)", len(assets))
	results := []bagongmodel.AssetMovementLine{}
	if len(assets) > 0 {
		for i, val := range assets {
			result := bagongmodel.AssetMovementLine{}

			result.LineNo = i
			result.AssetID = val.ID
			result.AssetName = val.Name
			result.SiteFrom = val.Dimension.Get("Site")
			result.PCFrom = val.Dimension.Get("PC")

			if len(val.UserInfo) > 0 {
				vCustID := val.UserInfo[len(val.UserInfo)-1].CustomerID
				result.CustomerFrom = vCustID
				result.ProjectFrom = val.UserInfo[len(val.UserInfo)-1].ProjectID

				if v, ok := mapCusts[vCustID]; ok {
					result.CustomerFromName = v
				}
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func PostAssets(h *datahub.Hub, journalID string) error {
	assetMovements := []bagongmodel.AssetMovement{}
	err := h.Gets(new(bagongmodel.AssetMovement), dbflex.NewQueryParam().SetWhere(dbflex.Eq("_id", journalID)), &assetMovements)
	if err != nil {
		return fmt.Errorf("error when get asset movement: %s", err.Error())
	}

	if len(assetMovements) > 0 {
		assetMovement := assetMovements[0]
		// change asset
		if len(assetMovement.Lines) > 0 {
			assetIDs := []string{}
			mapAsset := map[string]bagongmodel.AssetMovementLine{}
			for _, v := range assetMovement.Lines {
				assetIDs = append(assetIDs, v.AssetID)

				mapAsset[v.AssetID] = v
			}

			assets := []bagongmodel.Asset{}
			err := h.Gets(new(bagongmodel.Asset), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", assetIDs...)), &assets)
			if err != nil {
				return fmt.Errorf("error when get asset: %s", err.Error())
			}

			if len(assets) > 0 {
				for _, val := range assets {
					assetMove := mapAsset[val.ID]

					newDim := tenantcoremodel.Dimension{}
					newDim = append(newDim, tenantcoremodel.DimensionItem{
						Key:   "PC",
						Value: assetMove.PCFrom,
					})
					newDim = append(newDim, tenantcoremodel.DimensionItem{
						Key:   "CC",
						Value: "OPS",
					})
					newDim = append(newDim, tenantcoremodel.DimensionItem{
						Key:   "Site",
						Value: assetMove.SiteTo,
					})
					newDim = append(newDim, tenantcoremodel.DimensionItem{
						Key:   "Asset",
						Value: "",
					})

					// add to user info
					userInfo := bagongmodel.UserInfo{
						ProjectID:      assetMove.ProjectTo,
						AssetDateFrom:  assetMove.DateFrom,
						AssetDateTo:    assetMove.DateTo,
						SiteID:         assetMove.SiteTo,
						CustomerID:     assetMove.CustomerTo,
						NoHullCustomer: assetMove.NoHullCustomer,
					}

					val.UserInfo = append(val.UserInfo, userInfo)

					// change asset dimension
					val.Dimension = newDim

					// update bg asset
					if e := h.Save(&val); e != nil {
						return errors.New("error update Asset: " + e.Error())
					}

					// update tenant asset
					if e := h.UpdateField(&tenantcoremodel.Asset{Dimension: newDim}, dbflex.Eq("_id", val.ID), "Dimension"); e != nil {
						return e
					}

					line := struct {
						Index        uint32
						AssetUnitID  string
						IsItem       bool
						StartDate    time.Time
						EndDate      time.Time
						Uom          string
						Duration     uint32
						Qty          uint32
						Descriptions string
					}{
						Index:        0,
						AssetUnitID:  assetMove.AssetID,
						IsItem:       false,
						StartDate:    assetMove.DateFrom,
						EndDate:      assetMove.DateTo,
						Uom:          "MONTH",
						Qty:          0,
						Descriptions: "",
						Duration:     uint32(MonthsBetween(assetMove.DateFrom, assetMove.DateTo)),
					}

					// create unit calendar
					unitCalendar := sdpmodel.UnitCalendar{
						ProjectID: assetMove.ProjectTo,
						Customer:  assetMove.CustomerTo,
						Dimension: newDim,
						Site:      assetMove.SiteTo,
					}
					unitCalendar.Lines = append(unitCalendar.Lines, line)

					// insert unit calendar
					if e := h.Save(&unitCalendar); e != nil {
						return errors.New("error insert unit calendar: " + e.Error())
					}
				}
			}
		}
	}

	return nil
}
