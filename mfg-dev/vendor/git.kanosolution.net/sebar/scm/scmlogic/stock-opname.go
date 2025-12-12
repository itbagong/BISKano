package scmlogic

import (
	"fmt"
	"sync"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type StockOpnameEngine struct{}

type StockOpnameGeneralRequest struct {
	StockOpname scmmodel.StockOpname
	Lines       []scmmodel.StockOpnameDetail
}

func (o *StockOpnameEngine) SaveOpname(ctx *kaos.Context, payload *StockOpnameGeneralRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	if e := h.Save(&payload.StockOpname); e != nil {
		return nil, e
	}

	if e := h.DeleteByFilter(new(scmmodel.StockOpnameDetail), dbflex.Eq("StockOpnameID", payload.StockOpname.ID)); e != nil {
		return nil, fmt.Errorf("error clear item serials: " + e.Error())
	}

	for _, l := range payload.Lines {
		if e := h.Save(&l); e != nil {
			return nil, e
		}
	}

	return payload, nil
}

func (o *StockOpnameEngine) Submit(ctx *kaos.Context, payload *StockOpnameGeneralRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	if len(payload.Lines) == 0 {
		return nil, fmt.Errorf("missing: lines data")
	}

	// validate
	for _, l := range payload.Lines {
		if l.QtyInSystem == 0 || l.UnitID == "" {
			return nil, fmt.Errorf("not found item balance for item id: %s, sku: %s, warehouse id: %s, aisle id: %s, section id: %s, box id: %s",
				l.ItemID, l.SKU, l.InventDim.WarehouseID, l.InventDim.AisleID, l.InventDim.SectionID, l.InventDim.BoxID)
		}
	}

	payload.StockOpname.Status = scmmodel.StockOpnameStatusSubmitted
	if e := h.Save(&payload.StockOpname); e != nil {
		return nil, e
	}

	for _, l := range payload.Lines {
		if e := h.Save(&l); e != nil {
			return nil, e
		}
	}

	return payload, nil
}

type StockOpnameProcessRequest struct {
	StockOpnameID   string
	StockOpnameDate *time.Time
}

func (o *StockOpnameEngine) Process(ctx *kaos.Context, payload *StockOpnameProcessRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	if payload.StockOpnameDate == nil {
		return nil, fmt.Errorf("missing: StockOpnameDate")
	}

	so := new(scmmodel.StockOpname)
	if e := h.GetByID(so, payload.StockOpnameID); e != nil {
		return nil, e
	}

	so.StockOpnameDate = payload.StockOpnameDate
	so.Status = scmmodel.StockOpnameStatusOnProgress
	if e := h.Save(so); e != nil {
		return nil, e
	}

	// TODO: need to lock item balance mechanism? so that item balance qty doesn't change via movement / transfer

	return payload, nil
}

func (o *StockOpnameEngine) AskReview(ctx *kaos.Context, payload *StockOpnameGeneralRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	payload.StockOpname.InputDate = time.Now()
	payload.StockOpname.Status = scmmodel.StockOpnameStatusWaitingForReview
	if e := h.Save(&payload.StockOpname); e != nil {
		return nil, e
	}

	for li, l := range payload.Lines {
		l.Gap = l.QtyActual - l.QtyInSystem

		switch true {
		case l.Gap == 0:
			l.Remarks = scmmodel.StockOpnameDetailRemarkOK
		case l.Gap > 0:
			l.Remarks = scmmodel.StockOpnameDetailRemarkOver
		case l.Gap < 0:
			l.Remarks = scmmodel.StockOpnameDetailRemarkMinus
		}

		if e := h.Save(&l); e != nil {
			return nil, e
		}

		payload.Lines[li] = l
	}

	return payload, nil
}

type StockOpnameApproveRequest struct {
	StockOpnameID string
}

func (o *StockOpnameEngine) Approve(ctx *kaos.Context, payload *StockOpnameApproveRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	if payload == nil {
		return nil, fmt.Errorf("missing: payload")
	}

	so := new(scmmodel.StockOpname)
	if e := h.GetByID(so, payload.StockOpnameID); e != nil {
		return nil, e
	}

	details := []scmmodel.StockOpnameDetail{}
	if e := h.Gets(new(scmmodel.StockOpnameDetail), dbflex.NewQueryParam().SetWhere(dbflex.Eq("StockOpnameID", payload.StockOpnameID)), &details); e != nil {
		return nil, e
	}

	gapLines := []scmmodel.StockOpnameDetail{}
	for _, d := range details {
		if d.Gap != 0 {
			gapLines = append(gapLines, d)
		}
	}

	if len(gapLines) > 0 {
		inv := scmmodel.InventoryAdjustment{
			StockOpnameID:      so.ID,
			AdjustmentDate:     time.Now(),
			Company:            so.Company,
			Status:             scmmodel.InventoryAdjustmentStatusNeedToReview,
			FinancialDimension: so.FinancialDimension,
			InventoryDimension: so.InventoryDimension,
		}
		if e := h.Save(&inv); e != nil {
			return nil, e
		}

		for _, gl := range gapLines {
			invDetail := scmmodel.InventoryAdjustmentDetail{
				ID:                    gl.ID,
				InventoryAdjustmentID: inv.ID,
				ItemID:                gl.ItemID,
				SKU:                   gl.SKU,
				Description:           gl.Description,
				InventoryDimension:    gl.InventDim,
				UoM:                   gl.UnitID,
				QtyInSystem:           gl.QtyInSystem,
				QtyActual:             gl.QtyActual,
				Gap:                   gl.Gap,
				Remarks:               gl.Remarks,
				Note:                  gl.Note,
				NoteStaff:             gl.NoteStaff,
			}
			if e := h.Save(&invDetail); e != nil {
				return nil, e
			}
		}
	}

	so.Status = scmmodel.StockOpnameStatusDone
	if e := h.Save(so); e != nil {
		return nil, e
	}

	return payload, nil
}

func (o *StockOpnameEngine) GetLines(ctx *kaos.Context, payload *scmmodel.InventDimension) ([]scmmodel.StockOpnameDetail, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: connection")
	}

	dfs := []*dbflex.Filter{}
	if payload.WarehouseID != "" {
		dfs = append(dfs, dbflex.Eq("InventDim.WarehouseID", payload.WarehouseID))
	}

	if payload.AisleID != "" {
		dfs = append(dfs, dbflex.Eq("InventDim.AisleID", payload.AisleID))
	}

	if payload.SectionID != "" {
		dfs = append(dfs, dbflex.Eq("InventDim.SectionID", payload.SectionID))
	}

	if payload.BoxID != "" {
		dfs = append(dfs, dbflex.Eq("InventDim.BoxID", payload.BoxID))
	}

	ibs := []scmmodel.ItemBalance{}
	if e := h.GetsByFilter(new(scmmodel.ItemBalance), dbflex.And(dfs...), &ibs); e != nil {
		return nil, e
	}

	if len(ibs) == 0 {
		return []scmmodel.StockOpnameDetail{}, nil
	}

	// get items data
	itemIDs := lo.Uniq(lo.Map(ibs, func(d scmmodel.ItemBalance, index int) string {
		return d.ItemID
	}))

	items := []tenantcoremodel.Item{}
	if e := h.GetsByFilter(new(tenantcoremodel.Item), dbflex.In("_id", codekit.ToInterfaceArray(itemIDs)...), &items); e != nil {
		return nil, e
	}

	itemM := lo.Associate(items, func(item tenantcoremodel.Item) (string, tenantcoremodel.Item) {
		return item.ID, item
	})

	// get spec data
	specFilters := lo.Map(ibs, func(d scmmodel.ItemBalance, index int) *dbflex.Filter {
		return dbflex.And(dbflex.Eq("ItemID", d.ItemID), dbflex.Eq("SKU", d.SKU))
	})

	specs := []tenantcoremodel.ItemSpec{}
	if e := h.GetsByFilter(new(tenantcoremodel.ItemSpec), dbflex.Or(specFilters...), &specs); e != nil {
		return nil, e
	}

	specM := lo.Associate(specs, func(d tenantcoremodel.ItemSpec) (string, tenantcoremodel.ItemSpec) {
		return fmt.Sprintf("%s|%s", d.ItemID, d.SKU), d
	})

	res := []scmmodel.StockOpnameDetail{}
	for _, ib := range ibs {
		uom := ""
		if it, ok := itemM[ib.ItemID]; ok {
			uom = it.DefaultUnitID
		}

		desc := ""
		if spec, ok := specM[fmt.Sprintf("%s|%s", ib.ItemID, ib.SKU)]; ok {
			desc = formatItemDescription(h, spec)
		}

		res = append(res, scmmodel.StockOpnameDetail{
			ItemID:      ib.ItemID,
			SKU:         ib.SKU,
			InventDim:   ib.InventDim,
			QtyInSystem: ib.QtyAvail,
			UnitID:      uom,
			Description: desc,
		})
	}

	return res, nil
}

func formatItemDescription(h *datahub.Hub, spec tenantcoremodel.ItemSpec) string {
	itemName := ""
	variantName := ""
	sizeName := ""
	gradeName := ""
	var wg sync.WaitGroup

	if spec.ItemID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sd := new(tenantcoremodel.Item)
			h.GetByID(sd, spec.ItemID)
			itemName = sd.Name
		}()
	}

	if spec.SpecVariantID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sd := new(tenantcoremodel.SpecVariant)
			h.GetByID(sd, spec.SpecVariantID)
			variantName = sd.Name
		}()
	}

	if spec.SpecSizeID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sd := new(tenantcoremodel.SpecSize)
			h.GetByID(sd, spec.SpecSizeID)
			sizeName = sd.Name
		}()
	}

	if spec.SpecGradeID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sd := new(tenantcoremodel.SpecGrade)
			h.GetByID(sd, spec.SpecGradeID)
			gradeName = sd.Name
		}()
	}

	wg.Wait()

	separator := "-"
	desc := itemName
	if variantName != "" {
		desc += fmt.Sprintf("%s%s", separator, variantName)
	}
	if sizeName != "" {
		desc += fmt.Sprintf("%s%s", separator, sizeName)
	}
	if gradeName != "" {
		desc += fmt.Sprintf("%s%s", separator, gradeName)
	}

	return desc
}
