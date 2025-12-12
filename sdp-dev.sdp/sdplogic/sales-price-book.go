package sdplogic

import (
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
)

type SalesPriceBookEngine struct{}

type SalesPriceBookSearchRequest struct {
	SalesPriceBookName string
	DateFrom           string
	DateTo             string
}

func (o *SalesPriceBookEngine) GetsFilter(ctx *kaos.Context, payload *SalesPriceBookSearchRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	responseData := struct {
		Count int         `json:"count"`
		Data  interface{} `json:"data"`
	}{}

	// dateFilter := dbflex.Or(dbflex.Eq("Name", payload.SalesPriceBookName), dbflex.Gte("StartPeriod", payload.DateFrom), dbflex.Lte("EndPeriod", payload.DateTo))
	filters := []*dbflex.Filter{}
	filters = append(filters, dbflex.Contains("Name", payload.SalesPriceBookName))
	filters = append(filters, dbflex.And(dbflex.Gte("StartPeriod", payload.DateFrom), dbflex.Lte("EndPeriod", payload.DateTo)))

	salesPriceBook := []sdpmodel.SalesPriceBook{}
	err := h.Gets(new(sdpmodel.SalesPriceBook), dbflex.NewQueryParam().SetWhere(dbflex.Or(filters...)), &salesPriceBook)
	if err != nil {
		return nil, fmt.Errorf("Search Price Book: %s", err.Error())
	}

	fmt.Println(salesPriceBook)
	count := len(salesPriceBook)
	responseData.Count = count
	responseData.Data = salesPriceBook
	// //clear data by movement in id before save

	return responseData, nil
}
