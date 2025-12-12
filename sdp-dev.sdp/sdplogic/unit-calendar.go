package sdplogic

import (
	"errors"
	"fmt"
	"io"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/sebarcode/codekit"
)

type UnitCalendarEngine struct{}

func (o *UnitCalendarEngine) upsertucsite(h *datahub.Hub, model *sdpmodel.UnitCalendarSite) error {
	if e := h.GetByID(new(sdpmodel.UnitCalendarSite), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Unit Calendar Site : " + e.Error())
		}
	} else {
		if e := h.Update(model); e != nil {
			return errors.New("error update Unit Calendar  Site  : " + e.Error())
		}
	}

	return nil
}

func (o *UnitCalendarEngine) upsert(h *datahub.Hub, model *sdpmodel.UnitCalendar) error {
	if e := h.GetByID(new(sdpmodel.UnitCalendar), model.ID); e != nil {
		if e := h.Insert(model); e != nil {
			return errors.New("error insert Unit Calendar : " + e.Error())
		}
	} else {
		if e := h.Update(model); e != nil {
			return errors.New("error update Unit Calendar : " + e.Error())
		}
	}

	return nil
}

func (o *UnitCalendarEngine) Insert(ctx *kaos.Context, payload *sdpmodel.UnitCalendar) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	eventhub, err := ctx.DefaultEvent()
	if err != nil {
		return nil, errors.New("missing: connection")
	}

	// userID := sebar.GetUserIDFromCtx(ctx)

	if payload.Dimension.Get("Site") == "" || len(payload.Lines) <= 0 {
		return nil, errors.New("missing: payload")
	}

	payload.Site = payload.Dimension.Get("Site")

	LastUCs := []sdpmodel.UnitCalendar{}
	err = h.Gets(new(sdpmodel.UnitCalendar), dbflex.NewQueryParam().SetWhere(dbflex.Eq("SORefNo", payload.SORefNo)).SetSort("-No"), &LastUCs)
	if err != nil {
		return nil, fmt.Errorf("Error: %v Last Unit Calendar", err)
	}

	LastSO := sdpmodel.SalesOrder{}

	if payload.SORefNo != "" {
		err = h.GetByID(&LastSO, payload.SORefNo)
		if err != nil {
			return nil, fmt.Errorf("Error: %v Last SO", err)
		}

		type UCbatch struct {
			Index        uint32
			AssetUnitID  string
			IsItem       bool
			StartDate    time.Time
			EndDate      time.Time
			Uom          string
			Duration     uint32
			Qty          uint32
			Descriptions string
		}

		UCbatchLines := []UCbatch{}
		for _, lastUC := range LastUCs {
			for _, line := range lastUC.Lines {
				if len(UCbatchLines) == 0 {
					UCbatchLines = append(UCbatchLines, line)
					continue
				}

				var UCbatchLine any = nil

				for _, batch := range UCbatchLines {
					if batch.Index == line.Index {
						UCbatchLine = batch
						break
					}
				}

				if UCbatchLine != nil {
					for index, UCbatchLine := range UCbatchLines {
						if UCbatchLine.Index == line.Index {
							UCbatchLines[index].Qty += line.Qty
							UCbatchLines[index].AssetUnitID = line.AssetUnitID
						}
					}
				} else {
					UCbatchLines = append(UCbatchLines, line)
				}
			}
		}

		lines := []struct {
			Index  uint32
			maxQty uint32
		}{}

		for index, line := range LastSO.Lines {
			var ucline any = nil
			for _, UCbatchLine := range UCbatchLines {
				if UCbatchLine.Index == uint32(index) {
					ucline = UCbatchLine
				}
			}

			if ucline != nil {
				lines = append(lines, struct {
					Index  uint32
					maxQty uint32
				}{
					Index:  uint32(index),
					maxQty: uint32(line.Qty) - ucline.(UCbatch).Qty,
				})
			}
		}

		// Check QTY
		for _, line := range payload.Lines {
			for _, ln := range lines {
				if line.Index == ln.Index && ln.maxQty-line.Qty < 0 {
					return nil, errors.New("Qty greater than max Qty")
				}
			}
		}

	}
	ucItems := []sdpmodel.UnitCalendar{}
	err = h.GetsByFilter(new(sdpmodel.UnitCalendar), dbflex.And(
		dbflex.Eq("Customer", payload.Customer),
		dbflex.Eq("ProjectID", payload.ProjectID),
		dbflex.Eq("Site", payload.Site),
	), &ucItems)

	if err != io.EOF {
		for _, c := range payload.Lines {
			for _, cc := range ucItems {
				for _, d := range cc.Lines {
					if c.AssetUnitID == d.AssetUnitID &&
						c.Duration == d.Duration &&
						c.StartDate.Format("2006-01-02") == d.StartDate.Format("2006-01-02") &&
						c.EndDate.Format("2006-01-02") == d.EndDate.Format("2006-01-02") {
						return nil, fmt.Errorf("data already exist")
					}
				}
			}
		}
	}

	err = o.upsert(h, payload)
	if err != nil {
		return nil, err
	}

	LastUCs = []sdpmodel.UnitCalendar{}
	err = h.Gets(new(sdpmodel.UnitCalendar), dbflex.NewQueryParam().SetWhere(dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Dimension.Get("Site")))), &LastUCs)
	if err != nil {
		return nil, fmt.Errorf("Error: %v Last Unit Calendar", err)
	}

	startdate := time.Time{}
	enddate := time.Time{}

	for _, LastUC := range LastUCs {
		for _, line := range LastUC.Lines {
			if startdate.IsZero() || startdate.After(line.StartDate) {
				startdate = line.StartDate
			}

			if enddate.IsZero() || enddate.Before(line.EndDate) {
				enddate = line.EndDate
			}
		}
	}

	LastUCSites := []sdpmodel.UnitCalendarSite{}
	LastUCSite := sdpmodel.UnitCalendarSite{}
	err = h.Gets(new(sdpmodel.UnitCalendarSite), dbflex.NewQueryParam().SetWhere(dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Dimension.Get("Site")))), &LastUCSites)
	if err != nil {
		return nil, fmt.Errorf("Error: %v Last Unit Calendar Site", err)
	}

	if len(LastUCSites) > 0 {
		LastUCSite = LastUCSites[0]

		LastUCSite.StartDate = startdate
		LastUCSite.EndDate = enddate
	} else {
		LastUCSite = sdpmodel.UnitCalendarSite{
			StartDate: startdate,
			EndDate:   enddate,
			Dimension: payload.Dimension,
		}
	}

	err = o.upsertucsite(h, &LastUCSite)
	if err != nil {
		return nil, err
	}

	for _, line := range payload.Lines {
		bagongasset := bagongmodel.Asset{}
		send := []string{line.AssetUnitID}

		err = eventhub.Publish("/v1/bagong/asset/get", &send, &bagongasset, &kaos.PublishOpts{})
		if err != nil {
			return nil, fmt.Errorf("Error get data bagong: %v", err)
		}

		userinfo := bagongmodel.UserInfo{
			ProjectID:     payload.ProjectID,
			AssetDateFrom: GetStartOfDay(line.StartDate),
			AssetDateTo:   GetStartOfDay(line.EndDate),
			SiteID:        payload.Dimension.Get("Site"),
			CustomerID:    payload.Customer,
			Description:   line.Descriptions,
		}

		if payload.SORefNo != "" {
			var LineSO struct {
				Asset          string
				Item           string
				Description    string
				Shift          int64
				Qty            uint
				UoM            string
				ContractPeriod int
				UnitPrice      int64
				Amount         int64
				DiscountType   string
				Discount       int
				Taxable        bool
				TaxCodes       []string
				Spesifications []string
				StartDate      time.Time
				EndDate        time.Time
				Checklists     tenantcoremodel.Checklists
				References     tenantcoremodel.References
			}

			for indexSO, lineSO := range LastSO.Lines {
				if indexSO == int(line.Index) {
					LineSO = lineSO
					break
				}
			}

			tmpStartDate := GetStartOfDay(LineSO.StartDate)
			tmpEndDate := GetStartOfDay(LineSO.EndDate)
			userinfo.SONumber = payload.SORefNo
			userinfo.SOStartDate = &tmpStartDate
			userinfo.SOEndDate = &tmpEndDate
		}

		bagongasset.UserInfo = append(bagongasset.UserInfo, userinfo)
		bagongasset.Dimension = payload.Dimension

		err = eventhub.Publish("/v1/bagong/asset/save", &bagongasset, nil, &kaos.PublishOpts{Headers: codekit.M{"FnResult": bagongasset}})
		if err != nil {
			return nil, fmt.Errorf("Error get data bagong: %v", err)
		}
	}

	return payload, nil
}

func (o *UnitCalendarEngine) Update(ctx *kaos.Context, payload *sdpmodel.UnitCalendar) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	eventhub, err := ctx.DefaultEvent()
	if err != nil {
		return nil, errors.New("missing: connection")
	}

	if payload.ID == "" || len(payload.Lines) <= 0 {
		return nil, errors.New("missing: payload")
	}

	unitcalendar := sdpmodel.UnitCalendar{}
	err = h.GetByID(&unitcalendar, payload.ID)
	if err != nil {
		return nil, fmt.Errorf("Error: %v Get Unit Calendar", err)
	}

	payload.Site = payload.Dimension.Get("Site")

	LastSO := sdpmodel.SalesOrder{}
	if payload.SORefNo != "" {
		err = h.GetByID(&LastSO, payload.SORefNo)
		if err != nil {
			return nil, fmt.Errorf("Error: %v Last SO", err)
		}
	}

	editedLine := []struct {
		Index        uint32
		AssetUnitID  string
		IsItem       bool
		StartDate    time.Time
		EndDate      time.Time
		Uom          string
		Duration     uint32
		Qty          uint32
		Descriptions string
	}{}

	for index, ucline := range unitcalendar.Lines {
		for _, palline := range payload.Lines {
			if ucline.Index == palline.Index {
				editedLine = append(editedLine, palline)
				unitcalendar.Lines[index] = palline
			}
		}
	}

	err = o.upsert(h, &unitcalendar)
	if err != nil {
		return nil, err
	}

	LastUCs := []sdpmodel.UnitCalendar{}
	err = h.Gets(new(sdpmodel.UnitCalendar), dbflex.NewQueryParam().SetWhere(dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Dimension.Get("Site")))), &LastUCs)
	if err != nil {
		return nil, fmt.Errorf("Error: %v Last Unit Calendar", err)
	}

	startdate := time.Time{}
	enddate := time.Time{}

	for _, LastUC := range LastUCs {
		for _, line := range LastUC.Lines {
			if startdate.IsZero() || startdate.After(line.StartDate) {
				startdate = line.StartDate
			}

			if enddate.IsZero() || enddate.Before(line.EndDate) {
				enddate = line.EndDate
			}
		}
	}

	LastUCSites := []sdpmodel.UnitCalendarSite{}
	LastUCSite := sdpmodel.UnitCalendarSite{}
	err = h.Gets(new(sdpmodel.UnitCalendarSite), dbflex.NewQueryParam().SetWhere(dbflex.And(dbflex.Eq("Dimension.Key", "Site"), dbflex.Eq("Dimension.Value", payload.Dimension.Get("Site")))), &LastUCSites)
	if err != nil {
		return nil, fmt.Errorf("Error: %v Last Unit Calendar Site", err)
	}

	if len(LastUCSites) > 0 {
		LastUCSite = LastUCSites[0]

		LastUCSite.StartDate = startdate
		LastUCSite.EndDate = enddate
	} else {
		LastUCSite = sdpmodel.UnitCalendarSite{
			StartDate: startdate,
			EndDate:   enddate,
			Dimension: payload.Dimension,
		}
	}

	for _, line := range editedLine {
		bagongasset := bagongmodel.Asset{}
		send := []string{line.AssetUnitID}

		err = eventhub.Publish("/v1/bagong/asset/get", &send, &bagongasset, &kaos.PublishOpts{})
		if err != nil {
			return nil, fmt.Errorf("Error get data bagong: %v", err)
		}

		// err = h.GetByID(&bagongasset, line.AssetUnitID)
		// if err != nil {
		// 	return nil, fmt.Errorf("Error get data bagong: %v", err)
		// }
		userinfo := bagongmodel.UserInfo{
			ProjectID:     payload.ProjectID,
			AssetDateFrom: GetStartOfDay(line.StartDate),
			AssetDateTo:   GetStartOfDay(line.EndDate),
			SiteID:        payload.Dimension.Get("Site"),
			CustomerID:    payload.Customer,
			Description:   line.Descriptions,
		}

		if payload.SORefNo != "" {

			var LineSO struct {
				Asset          string
				Item           string
				Description    string
				Shift          int64
				Qty            uint
				UoM            string
				ContractPeriod int
				UnitPrice      int64
				Amount         int64
				DiscountType   string
				Discount       int
				Taxable        bool
				TaxCodes       []string
				Spesifications []string
				StartDate      time.Time
				EndDate        time.Time
				Checklists     tenantcoremodel.Checklists
				References     tenantcoremodel.References
			}

			for indexSO, lineSO := range LastSO.Lines {
				if indexSO == int(line.Index) {
					LineSO = lineSO
					break
				}
			}

			tmpStartDate := GetStartOfDay(LineSO.StartDate)
			tmpEndDate := GetStartOfDay(LineSO.EndDate)
			userinfo.SONumber = payload.SORefNo
			userinfo.SOStartDate = &tmpStartDate
			userinfo.SOEndDate = &tmpEndDate
		}

		bagongasset.UserInfo = append(bagongasset.UserInfo, userinfo)
		bagongasset.Dimension = payload.Dimension

		err = eventhub.Publish("/v1/bagong/asset/save", &bagongasset, nil, &kaos.PublishOpts{Headers: codekit.M{"FnResult": bagongasset}})
		if err != nil {
			return nil, fmt.Errorf("Error get data bagong: %v", err)
		}

		// err := h.Update(&bagongasset)
		// if err != nil {
		// 	return nil, fmt.Errorf("Error save to bagong: %v", err)
		// }
	}

	return payload, nil
}
