package bagonglogic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type VendorEngine struct{}

type VendorGetResponse struct {
	tenantcoremodel.Vendor
	Detail bagongmodel.BGVendor
}

func (engine *VendorEngine) Get(ctx *kaos.Context, req []interface{}) (*VendorGetResponse, error) {
	if len(req) == 0 {
		return nil, errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	res := new(VendorGetResponse)

	vendorID := req[0]
	tenantcoreVendor := new(tenantcoremodel.Vendor)
	e := hub.GetByID(tenantcoreVendor, vendorID)
	if e == nil {
		res.Vendor = *tenantcoreVendor

		//get detail of vendor
		bagongVendor := new(bagongmodel.BGVendor)
		e = hub.GetByFilter(bagongVendor, dbflex.Eq("TenantCoreVendorID", tenantcoreVendor.ID))
		if e == nil {
			res.Detail = *bagongVendor
		}
	}

	return res, nil
}

type VendorSaveRequest struct {
	tenantcoremodel.Vendor
	Detail bagongmodel.BGVendor
}

func (engine *VendorEngine) Save(ctx *kaos.Context, req *VendorSaveRequest) (interface{}, error) {
	var e error

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	if req.Vendor.ID == "" {
		//save to vendor tenantmodel
		tenantcorelogic.MWPreAssignSequenceNo("Vendor", false, "_id")(ctx, &req.Vendor)

		if e := hub.GetByID(new(tenantcoremodel.Vendor), req.Vendor.ID); e == nil {
			ctx.Log().Errorf("error duplicate key: %s", e.Error())
			return nil, errors.New("error duplicate key: " + e.Error())
		}
	}

	if e = hub.Save(&req.Vendor); e != nil {
		return nil, errors.New("error update vendor: " + e.Error())
	}

	req.Detail.TenantCoreVendorID = req.ID

	//save to bagong
	e = hub.Save(&req.Detail)
	if e != nil {
		return nil, errors.New("error update bagong vendor: " + e.Error())
	}

	return req, nil
}

func (engine *VendorEngine) GetAssetDetail(ctx *kaos.Context, payload SiteEntryReq) ([]tenantcoremodel.Vendor, error) {

	// hub := sebar.GetTenantDBFromContext(ctx)
	// if hub == nil {
	// 	return nil, errors.New("missing: connection")
	// }

	// // get SiteEntryAsset
	// siteEntryAsset := new(bagongmodel.SiteEntryAsset)
	// if e := hub.GetByID(siteEntryAsset, "668bb72e60677bb53917dc63"); e != nil {
	// 	return nil, errors.New(fmt.Sprintf("SiteEntryAsset not found: %s", "s"))
	// }

	return []tenantcoremodel.Vendor{}, nil
}

func (engine *VendorEngine) GetVendorActive(ctx *kaos.Context, payload *dbflex.QueryParam) ([]tenantcoremodel.Vendor, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	query := dbflex.NewQueryParam()
	if payload.Skip != 0 {
		query = query.SetSkip(payload.Skip)
	}

	if payload.Take != 0 {
		query = query.SetTake(payload.Take)
	}

	if len(payload.Sort) > 0 {
		query = query.SetSort(payload.Sort...)
	}

	filters := []*dbflex.Filter{}

	if payload != nil {
		if payload.Where != nil {
			filters2 := []*dbflex.Filter{}
			if len(payload.Where.Items) > 0 {
				vItems := payload.Where.Items
				if len(vItems) > 0 {
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
						}
					}
				}
			} else {
				fieldVal := payload.Where.Field
				opVal := payload.Where.Op
				if opVal == dbflex.OpEq {
					if payload.Where.Value != nil {
						aInterface := payload.Where.Value.(string)
						filters2 = append(filters2, dbflex.Eq(fieldVal, aInterface))
					}
				}
			}
			if len(filters2) > 0 {
				filters = append(filters, dbflex.Or(filters2...))
			}
		}
	}

	if len(filters) > 0 {
		query = query.SetWhere(dbflex.And(filters...))
	}

	vendors := []tenantcoremodel.Vendor{}
	err := hub.Gets(new(tenantcoremodel.Vendor), query, &vendors)
	if err != nil {
		return nil, err
	}

	return vendors, nil
}

type GetVendorRequest struct {
	Where struct {
		SiteID string
	}
}

func (engine *VendorEngine) GetVendorBySite(ctx *kaos.Context, payload GetVendorRequest) ([]tenantcoremodel.Vendor, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	vendors := []tenantcoremodel.Vendor{}
	filtered := []tenantcoremodel.Vendor{}
	err := hub.GetsByFilter(new(tenantcoremodel.Vendor), dbflex.Eq("IsActive", true), &vendors)
	if err != nil {
		return nil, err
	}

	for _, c := range vendors {
		if ok := StringInSlice(payload.Where.SiteID, c.Sites); ok {
			filtered = append(filtered, c)
		}
	}

	return filtered, nil
}

func (engine *VendorEngine) GetVendorBank(ctx *kaos.Context, payload *dbflex.QueryParam) ([]bagongmodel.VendorBank, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	r := ctx.Data().Get("http_request", nil).(*http.Request)
	VendorID := strings.TrimSpace(r.URL.Query().Get("VendorID"))

	if VendorID == "" {
		return nil, errors.New("vendor id is empty")
	}

	filters := []string{}
	if payload != nil {
		if payload.Where != nil {
			if len(payload.Where.Items) > 0 {
				vItems := payload.Where.Items
				if len(vItems) > 0 {
					for _, val := range vItems {
						opVal := val.Op
						if opVal == dbflex.OpContains {
							aInterface := val.Value.([]interface{})
							aString := make([]string, len(aInterface))
							for i, v := range aInterface {
								aString[i] = v.(string)
							}
							if len(aString) > 0 {
								if aString[0] != "" {
									filters = append(filters, aString[0])
								}
							}
						} else if opVal == dbflex.OpIn {
							aInterface := val.Value.([]interface{})
							aString := []string{}
							for _, v := range aInterface {
								aString = append(aString, v.(string))
							}
							if len(aInterface) > 0 {
								filters = append(filters, aString...)
							}
						} else if opVal == dbflex.OpEq {
							aInterface := val.Value.(string)
							filters = append(filters, aInterface)
						}
					}
				}
			} else {
				opVal := payload.Where.Op
				if opVal == dbflex.OpEq {
					aInterface := payload.Where.Value.(string)
					filters = append(filters, aInterface)
				}
			}
		}
	}

	bgVendors := []bagongmodel.BGVendor{}
	newBGVendorBanks := []bagongmodel.VendorBank{}
	err := h.Gets(new(bagongmodel.BGVendor), dbflex.NewQueryParam().SetWhere(dbflex.Eq("_id", VendorID)), &bgVendors)
	if err != nil {
		return nil, fmt.Errorf("error when get asset: %s", err.Error())
	}

	if len(bgVendors) > 0 {
		bgVendorBanks := bgVendors[0].VendorBank

		if len(bgVendorBanks) > 0 {
			for _, val := range bgVendorBanks {
				if len(filters) > 0 {
					for _, filter := range filters {
						if strings.Contains(strings.ToLower(val.ID), strings.ToLower(filter)) || strings.Contains(strings.ToLower(val.BankName), strings.ToLower(filter)) {
							newBGVendorBanks = append(newBGVendorBanks, val)
							break
						}
					}
				} else {
					newBGVendorBanks = append(newBGVendorBanks, val)
				}
			}
		}
	}

	return newBGVendorBanks, nil
}
