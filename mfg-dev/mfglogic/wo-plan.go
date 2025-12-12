package mfglogic

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type WorkOrderPlanEngine struct{}

type WorkOrderPlanSaveReq struct {
	mfgmodel.WorkOrderPlan
	WorkOrderSummaryMaterial []mfgmodel.WorkOrderSummaryMaterial
	WorkOrderSummaryResource []mfgmodel.WorkOrderSummaryResource
	WorkOrderSummaryOutput   []mfgmodel.WorkOrderSummaryOutput
}

func (o *WorkOrderPlanEngine) Save(ctx *kaos.Context, req *WorkOrderPlanSaveReq) (*WorkOrderPlanSaveReq, error) {
	var (
		h   *datahub.Hub
		err error
	)

	coID, _, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: connection")
	}

	h, _ = db.BeginTx()
	defer func() {
		if h.IsTx() {
			if err == nil {
				h.Commit()
			} else {
				h.Rollback()
			}
		}
	}()

	if req.ID == "" {
		id, _ := tenantcorelogic.GenerateIDFromNumSeq(ctx, "WorkOrder")
		req.ID = id
	}

	req.WorkOrderPlan.CompanyID = lo.Ternary(req.WorkOrderPlan.CompanyID != "", req.WorkOrderPlan.CompanyID, coID)
	req.WorkOrderPlan.TrxType = mfgmodel.JournalWorkOrderPlan

	err = h.Save(&req.WorkOrderPlan)
	if err != nil {
		return nil, err
	}

	// validate WorkOrderSummaryMaterial, tidak boleh ada yang sama ItemID + SKU (Unique)
	// itemORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
	// itemSpecORM := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
	sumMatGroup := lo.GroupBy(req.WorkOrderSummaryMaterial, func(d mfgmodel.WorkOrderSummaryMaterial) string {
		return fmt.Sprintf("%s||%s", d.ItemID, d.SKU)
	})

	for key, mats := range sumMatGroup {
		if len(mats) > 1 {
			keys := strings.Split(key, "||")
			item, err := datahub.GetByID(h, new(tenantcoremodel.Item), keys[0])
			if err != nil {
				item = new(tenantcoremodel.Item)
			}

			spec, err := datahub.GetByID(h, new(tenantcoremodel.ItemSpec), keys[0])
			if err != nil {
				spec = new(tenantcoremodel.ItemSpec)
			}

			return nil, fmt.Errorf("Summary Material is duplicated | Item: %s | SKU: %s", item.Name, spec.SKU)
		}
	}

	// Material
	matExs := []mfgmodel.WorkOrderSummaryMaterial{}
	h.GetsByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", req.WorkOrderPlan.ID), &matExs)

	matNewIDs := []string{}
	for _, d := range req.WorkOrderSummaryMaterial {
		d.WorkOrderPlanID = req.WorkOrderPlan.ID
		err = h.Save(&d)
		if err != nil {
			return nil, err
		}
		matNewIDs = append(matNewIDs, d.ID)
	}

	matDeletedIDs := lo.FilterMap(matExs, func(d mfgmodel.WorkOrderSummaryMaterial, i int) (string, bool) {
		return d.ID, lo.Contains(matNewIDs, d.ID) == false // hanya delete id yang tidak ada pada saat save gelondongan diatas
	})
	if len(matDeletedIDs) > 0 {
		h.DeleteByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.In("_id", matDeletedIDs...))
	}

	// Resource
	resExs := []mfgmodel.WorkOrderSummaryResource{}
	h.GetsByFilter(new(mfgmodel.WorkOrderSummaryResource), dbflex.Eq("WorkOrderPlanID", req.WorkOrderPlan.ID), &resExs)

	resNewIDs := []string{}
	for _, d := range req.WorkOrderSummaryResource {
		d.WorkOrderPlanID = req.WorkOrderPlan.ID
		err = h.Save(&d)
		if err != nil {
			return nil, err
		}
		resNewIDs = append(resNewIDs, d.ID)
	}

	resDeletedIDs := lo.FilterMap(resExs, func(d mfgmodel.WorkOrderSummaryResource, i int) (string, bool) {
		return d.ID, lo.Contains(resNewIDs, d.ID) == false // hanya delete id yang tidak ada pada saat save gelondongan diatas
	})
	if len(resDeletedIDs) > 0 {
		h.DeleteByFilter(new(mfgmodel.WorkOrderSummaryResource), dbflex.In("_id", resDeletedIDs...))
	}

	// Output
	outExs := []mfgmodel.WorkOrderSummaryOutput{}
	h.GetsByFilter(new(mfgmodel.WorkOrderSummaryOutput), dbflex.Eq("WorkOrderPlanID", req.WorkOrderPlan.ID), &outExs)

	outNewIDs := []string{}
	for _, d := range req.WorkOrderSummaryOutput {
		d.WorkOrderPlanID = req.WorkOrderPlan.ID
		err = h.Save(&d)
		if err != nil {
			return nil, err
		}
		outNewIDs = append(outNewIDs, d.ID)
	}

	outDeletedIDs := lo.FilterMap(outExs, func(d mfgmodel.WorkOrderSummaryOutput, i int) (string, bool) {
		return d.ID, lo.Contains(outNewIDs, d.ID) == false // hanya delete id yang tidak ada pada saat save gelondongan diatas
	})
	if len(outDeletedIDs) > 0 {
		h.DeleteByFilter(new(mfgmodel.WorkOrderSummaryOutput), dbflex.In("_id", outDeletedIDs...))
	}

	return req, nil
}

type WOPGetsMaterialLineReq struct {
	WorkOrderPlanID string
	Type            string // Material / Resource / Output
	DateFrom        *time.Time
	DateTo          *time.Time
	ItemIDs         []string // khusus kalo Type nya Material
	UseID           bool

	// TODO: implement skip take
	// Skip            int
	// Take            int
}

func (o *WorkOrderPlanEngine) GetsMaterialLine(ctx *kaos.Context, p *WOPGetsMaterialLineReq) (codekit.M, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	if p.WorkOrderPlanID == "" {
		return nil, fmt.Errorf("WorkOrderPlanID is required")
	}

	filters := []*dbflex.Filter{}

	if p.WorkOrderPlanID != "" {
		filters = append(filters, dbflex.Eq("WorkOrderPlanID", p.WorkOrderPlanID))
	}

	if p.DateFrom != nil && !p.DateFrom.IsZero() {
		filters = append(filters, dbflex.Gte("WorkDate", p.DateFrom))
	}

	if p.DateTo != nil && !p.DateTo.IsZero() {
		filters = append(filters, dbflex.Lte("WorkDate", p.DateTo))
	}

	worps := []mfgmodel.WorkOrderPlanReport{}
	if e := h.GetsByFilter(new(mfgmodel.WorkOrderPlanReport), dbflex.And(filters...), &worps); e != nil {
		return nil, e
	}

	// items := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
	// ledgers := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LedgerAccount))
	// specs := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemSpec))
	// groups := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemGroup))
	// units := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.UoM))
	// exps := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ExpenseType))
	// employees := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Employee))

	switch p.Type {
	case "Material":
		ids := lo.Map(worps, func(worp mfgmodel.WorkOrderPlanReport, i int) string {
			return worp.WorkOrderPlanReportConsumptionID
		})

		subs := []mfgmodel.WorkOrderPlanReportConsumption{}
		if e := h.GetsByFilter(new(mfgmodel.WorkOrderPlanReportConsumption), dbflex.And(
			dbflex.In("_id", ids...),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted), // request mba Fanny, hanya tampilkan yang statusnya posted
		), &subs); e != nil {
			return nil, e
		}

		lines := lo.FlatMap(subs, func(s mfgmodel.WorkOrderPlanReportConsumption, i int) []mfgmodel.WorkOrderMaterialItem {
			// di math.Abs biar hasil qty nya ga minus
			s.Lines = lo.Map(s.Lines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
				d.Qty = math.Abs(d.Qty)
				return d
			})
			s.AdditionalLines = lo.Map(s.AdditionalLines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
				d.Qty = math.Abs(d.Qty)
				return d
			})
			return append(s.Lines, s.AdditionalLines...)
		})

		lines = lo.Filter(lines, func(d mfgmodel.WorkOrderMaterialItem, i int) bool {
			return d.Qty > 0 // filter hanya yang Qty nya terisi saja
		})

		if len(p.ItemIDs) > 0 {
			lines = lo.Filter(lines, func(d mfgmodel.WorkOrderMaterialItem, i int) bool {
				return lo.Contains(p.ItemIDs, d.ItemID)
			})
		}

		if p.UseID == false {
			// normalize data from id to name

			itemIDs := make([]interface{}, len(lines))
			specIDs := make([]interface{}, len(lines))
			unitIDs := make([]interface{}, len(lines))
			employeeIDs := make([]interface{}, len(lines))

			lo.ForEach(lines, func(line mfgmodel.WorkOrderMaterialItem, index int) {
				itemIDs[index] = line.ItemID
				specIDs[index] = line.SKU
				unitIDs[index] = line.UnitID
				employeeIDs[index] = line.RequestedBy
			})

			// itemsKV := map[string]*tenantcoremodel.Item{}
			// items, _ := datahub.Find(h, new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", itemIDs...)))
			// itemsKV = lo.Associate(items, func(item *tenantcoremodel.Item) (string, *tenantcoremodel.Item) {
			// 	return item.ID, item
			// })

			specKV := map[string]*tenantcoremodel.ItemSpec{}
			specs, _ := datahub.Find(h, new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", specIDs...)))
			specKV = lo.Associate(specs, func(spec *tenantcoremodel.ItemSpec) (string, *tenantcoremodel.ItemSpec) {
				return spec.ID, spec
			})

			unitKV := map[string]*tenantcoremodel.UoM{}
			units, _ := datahub.Find(h, new(tenantcoremodel.UoM), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", unitIDs...)))
			unitKV = lo.Associate(units, func(unit *tenantcoremodel.UoM) (string, *tenantcoremodel.UoM) {
				return unit.ID, unit
			})

			employeeKV := map[string]*tenantcoremodel.Employee{}
			employees, _ := datahub.Find(h, new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", employeeIDs...)))
			employeeKV = lo.Associate(employees, func(employee *tenantcoremodel.Employee) (string, *tenantcoremodel.Employee) {
				return employee.ID, employee
			})

			lines = lo.Map(lines, func(d mfgmodel.WorkOrderMaterialItem, i int) mfgmodel.WorkOrderMaterialItem {
				// item, ok := itemsKV[d.ItemID]
				// if !ok {
				// 	item = new(tenantcoremodel.Item)
				// }

				spec, ok := specKV[d.SKU]
				if !ok {
					spec = new(tenantcoremodel.ItemSpec)
				}

				unit, ok := unitKV[d.UnitID]
				if !ok {
					unit = new(tenantcoremodel.UoM)
				}

				employee, ok := employeeKV[d.RequestedBy]
				if !ok {
					employee = new(tenantcoremodel.Employee)
				}

				// d.ItemID = tenantcorelogic.ItemVariantName(h, d.ItemID, spec.ID)
				// tenantcorelogic.MWPreAssignItem(d.ItemID+"~~"+d.SKU, false)(ctx, &item)
				d.ItemID = spec.OtherName
				d.SKU = spec.SKU
				d.UnitID = unit.Name
				d.RequestedBy = employee.Name

				// d.UnitCost = lo.Ternary(d.UnitCost != 0, d.UnitCost, item.CostUnit)
				d.Total = d.Qty * d.UnitCost
				return d
			})
		}

		sort.Slice(lines, func(i, j int) bool {
			return lines[i].Date.After(lines[j].Date)
		})

		return codekit.M{
			"data":  lines,
			"count": len(lines),
			"total": lo.SumBy(lines, func(d mfgmodel.WorkOrderMaterialItem) float64 {
				return d.Total
			}),
		}, nil

	case "Resource":
		ids := lo.Map(worps, func(worp mfgmodel.WorkOrderPlanReport, i int) string {
			return worp.WorkOrderPlanReportResourceID
		})

		subs := []mfgmodel.WorkOrderPlanReportResource{}
		if e := h.GetsByFilter(new(mfgmodel.WorkOrderPlanReportResource), dbflex.And(
			dbflex.In("_id", ids...),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted), // request mba Fanny, hanya tampilkan yang statusnya posted
		), &subs); e != nil {
			return nil, e
		}

		lines := lo.FlatMap(subs, func(s mfgmodel.WorkOrderPlanReportResource, i int) []mfgmodel.WorkOrderResourceItem {
			// calculate total
			s.Lines = lo.Map(s.Lines, func(l mfgmodel.WorkOrderResourceItem, li int) mfgmodel.WorkOrderResourceItem {
				l.Total = l.WorkingHour * l.RatePerHour
				return l
			})

			return s.Lines
		})

		if p.UseID == false {
			// normalize data from id to name
			expIDs := make([]interface{}, len(lines))
			employeeIDs := make([]interface{}, len(lines))

			lo.ForEach(lines, func(line mfgmodel.WorkOrderResourceItem, index int) {
				expIDs[index] = line.ExpenseType
				employeeIDs[index] = line.Employee
			})

			expsKV := map[string]*tenantcoremodel.ExpenseType{}
			exps, _ := datahub.Find(h, new(tenantcoremodel.ExpenseType), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", expIDs...)))
			expsKV = lo.Associate(exps, func(exp *tenantcoremodel.ExpenseType) (string, *tenantcoremodel.ExpenseType) {
				return exp.ID, exp
			})

			employeeKV := map[string]*tenantcoremodel.Employee{}
			employees, _ := datahub.Find(h, new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", employeeIDs...)))
			employeeKV = lo.Associate(employees, func(employee *tenantcoremodel.Employee) (string, *tenantcoremodel.Employee) {
				return employee.ID, employee
			})

			lines = lo.Map(lines, func(d mfgmodel.WorkOrderResourceItem, i int) mfgmodel.WorkOrderResourceItem {
				exp, ok := expsKV[d.ExpenseType]
				if !ok {
					exp = new(tenantcoremodel.ExpenseType)
				}

				d.ExpenseType = exp.Name

				emp, ok := employeeKV[d.Employee]
				if !ok {
					emp = new(tenantcoremodel.Employee)
				}

				d.Employee = emp.Name
				return d
			})
		}

		sort.Slice(lines, func(i, j int) bool {
			return lines[i].Date.After(lines[j].Date)
		})

		return codekit.M{
			"data":  lines,
			"count": len(lines),
			"total": lo.SumBy(lines, func(d mfgmodel.WorkOrderResourceItem) float64 {
				return d.Total
			}),
		}, nil

	case "Output":
		ids := lo.Map(worps, func(worp mfgmodel.WorkOrderPlanReport, i int) string {
			return worp.WorkOrderPlanReportOutputID
		})

		subs := []mfgmodel.WorkOrderPlanReportOutput{}
		if e := h.GetsByFilter(new(mfgmodel.WorkOrderPlanReportOutput), dbflex.And(
			dbflex.In("_id", ids...),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted), // request mba Fanny, hanya tampilkan yang statusnya posted
		), &subs); e != nil {
			return nil, e
		}

		lines := lo.FlatMap(subs, func(s mfgmodel.WorkOrderPlanReportOutput, i int) []mfgmodel.WorkOrderOutputItem {
			return s.Lines
		})

		if p.UseID == false {
			// normalize data from id to name
			itemIDs := []interface{}{}
			specIDs := make([]interface{}, len(lines))
			unitIDs := make([]interface{}, len(lines))
			groupIDs := make([]interface{}, len(lines))
			ledgerIDs := []interface{}{}

			lo.ForEach(lines, func(line mfgmodel.WorkOrderOutputItem, index int) {
				// itemIDs[index] = line.ItemID
				specIDs[index] = line.SKU
				unitIDs[index] = line.UnitID
				groupIDs[index] = line.GroupID
				switch line.Type {
				case mfgmodel.WorkOrderOutputTypeWasteLedger:
					ledgerIDs = append(ledgerIDs, line.InventoryLedgerAccID)
				default:
					itemIDs = append(itemIDs, line.InventoryLedgerAccID)
				}
				// employeeIDs[index] = line.RequestedBy
			})

			specsKV := map[string]*tenantcoremodel.ItemSpec{}
			specs, _ := datahub.Find(h, new(tenantcoremodel.ItemSpec), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", specIDs...)))
			specsKV = lo.Associate(specs, func(spec *tenantcoremodel.ItemSpec) (string, *tenantcoremodel.ItemSpec) {
				return spec.ID, spec
			})

			unitsKV := map[string]*tenantcoremodel.UoM{}
			units, _ := datahub.Find(h, new(tenantcoremodel.UoM), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", unitIDs...)))
			unitsKV = lo.Associate(units, func(unit *tenantcoremodel.UoM) (string, *tenantcoremodel.UoM) {
				return unit.ID, unit
			})

			groupsKV := map[string]*tenantcoremodel.ItemGroup{}
			groups, _ := datahub.Find(h, new(tenantcoremodel.ItemGroup), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", groupIDs...)))
			groupsKV = lo.Associate(groups, func(group *tenantcoremodel.ItemGroup) (string, *tenantcoremodel.ItemGroup) {
				return group.ID, group
			})

			itemsKV := map[string]*tenantcoremodel.Item{}
			if len(itemIDs) > 0 {
				items, _ := datahub.Find(h, new(tenantcoremodel.Item), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", itemIDs...)))
				itemsKV = lo.Associate(items, func(item *tenantcoremodel.Item) (string, *tenantcoremodel.Item) {
					return item.ID, item
				})
			}

			ledgersKV := map[string]*tenantcoremodel.LedgerAccount{}
			if len(ledgerIDs) > 0 {
				ledgers, _ := datahub.Find(h, new(tenantcoremodel.LedgerAccount), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", ledgerIDs...)))
				ledgersKV = lo.Associate(ledgers, func(ledger *tenantcoremodel.LedgerAccount) (string, *tenantcoremodel.LedgerAccount) {
					return ledger.ID, ledger
				})
			}

			lines = lo.Map(lines, func(d mfgmodel.WorkOrderOutputItem, i int) mfgmodel.WorkOrderOutputItem {
				spec, ok := specsKV[d.SKU]
				if !ok {
					spec = new(tenantcoremodel.ItemSpec)
				}

				switch d.Type {
				// case mfgmodel.WorkOrderOutputTypeWOOutput, mfgmodel.WorkOrderOutputTypeWasteItem:

				case mfgmodel.WorkOrderOutputTypeWasteLedger:
					lg, ok := ledgersKV[d.InventoryLedgerAccID]
					if !ok {
						lg = new(tenantcoremodel.LedgerAccount)
					}

					d.InventoryLedgerAccID = tenantcorelogic.ItemVariantName(h, "", spec.ID)
					itemDetails := strings.Split(d.InventoryLedgerAccID, "-")
					if len(itemDetails) == 0 {
						d.InventoryLedgerAccID = lg.Name
					} else {
						itemDetails = lo.Map(itemDetails, func(item string, index int) string {
							//sku
							if index == 0 {
								//inject name afer sku
								item = fmt.Sprintf("%s - %s", item, lg.Name)
							}

							return item
						})
					}

					d.InventoryLedgerAccID = strings.Join(itemDetails, "-")
				default:
					it, ok := itemsKV[d.InventoryLedgerAccID]
					if !ok {
						it = new(tenantcoremodel.Item)
					}

					d.InventoryLedgerAccID = tenantcorelogic.ItemVariantName(h, it.ID, spec.ID)
				}

				gr, ok := groupsKV[d.GroupID]
				if !ok {
					gr = new(tenantcoremodel.ItemGroup)
				}

				unit, ok := unitsKV[d.UnitID]
				if !ok {
					unit = new(tenantcoremodel.UoM)
				}

				d.SKU = spec.SKU
				d.GroupID = gr.Name
				d.UnitID = unit.Name
				return d
			})
		}

		sort.Slice(lines, func(i, j int) bool {
			return lines[i].Date.After(lines[j].Date)
		})

		return codekit.M{
			"data":  lines,
			"count": len(lines),
		}, nil
	}

	return nil, fmt.Errorf("Type is unknown")
}

func (o *WorkOrderPlanEngine) GetsAvailableStock(ctx *kaos.Context, p *GetAvailableStocksParam) ([]scmmodel.ItemBalance, error) {
	coID, _, err := GetCompanyAndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if p.CompanyID == "" {
		p.CompanyID = coID
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	bals := GetAvailableStocks(h, *p)

	res := lo.Map(bals, func(d *scmmodel.ItemBalance, i int) scmmodel.ItemBalance {
		return *d
	})

	return res, nil
}

type GetAvailableStocksParam struct {
	CompanyID     string
	GroupBy       []string
	InventDim     scmmodel.InventDimension
	Items         []GetAvailableStocksParamItem
	BalanceFilter struct {
		WarehouseIDs []interface{}
		SectionIDs   []interface{}
		SKUs         []interface{}
	}
}

type GetAvailableStocksParamItem struct {
	ItemID string
	SKU    string
}

// GetAvailableStocks have to always return exact len with parm.Items len
// hanya akan mengembalikan Item Balance sesuai dengan masing2 Invent Dim (WH, Sec, Ais, Box)
func GetAvailableStocks(h *datahub.Hub, parm GetAvailableStocksParam) []*scmmodel.ItemBalance {
	res := []*scmmodel.ItemBalance{}

	for _, item := range parm.Items {
		// inventDim := scmlogic.NewInventDimHelper(scmlogic.InventDimHelperOpt{DB: h, SKU: item.SKU}).TernaryInventDimension(&parm.InventDim) DEPLOY AGAIN

		// bals, _ := scmlogic.NewInventBalanceCalc(h).Get(&scmlogic.InventBalanceCalcOpts{
		// 	CompanyID:   parm.CompanyID,
		// 	ItemID:      []string{item.ItemID},
		// 	InventDim:   *inventDim,
		// 	BalanceDate: nil,
		// })

		opt := scmlogic.ItemBalanceOpt{
			CompanyID:   parm.CompanyID,
			InventDim:   parm.InventDim,
			ItemIDs:     []string{item.ItemID},
			GroupBy:     []string{"WarehouseID", "SectionID", "AisleID", "BoxID"},
			ConsiderSKU: true,
		}

		if len(parm.GroupBy) > 0 {
			opt.GroupBy = parm.GroupBy
		}

		if parm.BalanceFilter.WarehouseIDs != nil {
			opt.BalanceFilter.WarehouseIDs = parm.BalanceFilter.WarehouseIDs
		}

		if parm.BalanceFilter.SectionIDs != nil {
			opt.BalanceFilter.SectionIDs = parm.BalanceFilter.SectionIDs
		}

		if parm.BalanceFilter.SKUs != nil {
			opt.BalanceFilter.SKUs = parm.BalanceFilter.SKUs
		}

		bals, _ := scmlogic.NewItemBalanceHub(h).Get(nil, opt)

		bals = lo.Filter(bals, func(d *scmmodel.ItemBalance, i int) bool {
			return d.SKU == item.SKU
		})

		if len(bals) == 0 {
			bals = append(bals, new(scmmodel.ItemBalance))
		}

		res = append(res, bals...)
	}

	return res
}

func (o *WorkOrderPlanEngine) End(ctx *kaos.Context, id string) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	wop := new(mfgmodel.WorkOrderPlan)
	if e := h.GetByID(wop, id); e != nil {
		return nil, e
	}

	ledgerTrx := map[string][]orm.DataModel{}

	//get waste output
	ledgers, e := o.calcWasteOutput(h, *wop)
	if e != nil {
		return nil, e
	}

	if len(ledgers) > 0 {
		ledgerTrx[ledgers[0].TableName()] = ficologic.ToDataModels(ledgers)
	}

	_, e = ficologic.PostModelSave(h, wop, "WorkOrderVoucherNo", ledgerTrx) // validate only
	if e != nil {
		return nil, e
	}

	//get wo output
	ledgerTrx = map[string][]orm.DataModel{}

	//get waste output
	ledgers, e = o.calcWasteOutput(h, *wop)
	if e != nil {
		return nil, e
	}

	if len(ledgers) > 0 {
		ledgerTrx[ledgers[0].TableName()] = ficologic.ToDataModels(ledgers)
	}

	_, e = ficologic.PostModelSave(h, wop, "WorkOrderVoucherNo", ledgerTrx) // validate only
	if e != nil {
		return nil, e
	}

	wop.StatusOverall = mfgmodel.WorkOrderPlanStatusOverallEnd
	if e := h.Update(wop, "StatusOverall"); e != nil {
		return nil, e
	}

	return "ok", nil
}

func (o *WorkOrderPlanEngine) calcWasteOutput(h *datahub.Hub, woplan mfgmodel.WorkOrderPlan) ([]*ficomodel.LedgerTransaction, error) {
	ledgers := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LedgerAccount))
	items := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.Item))
	itemGroups := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.ItemGroup))
	journalTypes := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkOrderJournalType))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}

	outputs := []mfgmodel.WorkOrderSummaryOutput{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrderSummaryOutput), dbflex.And(
		dbflex.Eq("WorkOrderPlanID", woplan.ID),
		dbflex.Eq("Group", "Waste Item"),
	), &outputs)
	if e != nil {
		return ledgerTrxs, nil
	}

	lo.ForEach(outputs, func(waste mfgmodel.WorkOrderSummaryOutput, index int) {
		item, e := items.Get(waste.InventoryLedgerAccID)
		if e != nil {
			e = e
			return
		}

		itemGroup, e := itemGroups.Get(item.ItemGroupID)
		if e != nil {
			e = e
			return
		}

		ledgerAccount, err := ledgers.Get(item.LedgerAccountIDStock)
		if err != nil {
			ledgerAccount, err = ledgers.Get(itemGroup.LedgerAccountIDStock)
			if err != nil {
				err = fmt.Errorf("invalid: main inventory account for item %s: %s", item.ID, err.Error())
				return
			}
		}

		// costPerUnit := 0.0
		unitCost := item.CostUnit
		if unitCost == 0 {
			costPerUnit := scmlogic.GetCostPerUnit(h, *item, woplan.InventDim, &woplan.TrxDate)
			unitCost = costPerUnit * waste.AchievedQtyAmount
		}

		totalCost := unitCost * waste.AchievedQtyAmount
		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         woplan.CompanyID,
			Dimension:         woplan.Dimension,
			SourceType:        scmmodel.ModuleWorkorder,
			SourceJournalID:   woplan.ID,
			SourceJournalType: woplan.JournalTypeID,
			SourceTrxType:     string(woplan.TrxType),
			SourceLineNo:      0,
			TrxDate:           woplan.TrxDate,
			Text:              woplan.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            totalCost,
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", woplan.ID),
		}

		var offsetLedger *tenantcoremodel.LedgerAccount
		jt, _ := journalTypes.Get(woplan.JournalTypeID)
		offsetLedger, err = ledgers.Get(jt.DefaultOffsiteConsumption.AccountID)
		if err != nil {
			return
		}

		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         woplan.CompanyID,
			Dimension:         woplan.Dimension,
			SourceType:        scmmodel.ModuleWorkorder,
			SourceJournalID:   woplan.ID,
			SourceJournalType: woplan.JournalTypeID,
			SourceTrxType:     string(woplan.TrxType),
			SourceLineNo:      0,
			TrxDate:           woplan.TrxDate,
			Text:              woplan.Text,
			Account:           *offsetLedger,
			Status:            ficomodel.AmountConfirmed,
			Amount:            (-1 * totalCost),
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", woplan.ID),
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	})
	return ledgerTrxs, nil
}

func (o *WorkOrderPlanEngine) calcWoOutput(h *datahub.Hub, woplan mfgmodel.WorkOrderPlan) ([]*ficomodel.LedgerTransaction, error) {
	ledgers := sebar.NewMapRecordWithORM(h, new(tenantcoremodel.LedgerAccount))
	journalTypes := sebar.NewMapRecordWithORM(h, new(mfgmodel.WorkOrderJournalType))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}

	outputs := []mfgmodel.WorkOrderSummaryOutput{}
	e := h.GetsByFilter(new(mfgmodel.WorkOrderSummaryOutput), dbflex.And(
		dbflex.Eq("WorkOrderPlanID", woplan.ID),
		dbflex.Eq("Group", "WO Output"),
	), &outputs)
	if e != nil {
		return ledgerTrxs, nil
	}

	lo.ForEach(outputs, func(output mfgmodel.WorkOrderSummaryOutput, index int) {
		ledgerAccount, err := ledgers.Get(output.InventoryLedgerAccID)
		if err != nil {
			e = err
			return
		}

		unitCost := WOCalcOutputValuePerQty(h, woplan.CompanyID, woplan.ID)
		totalCost := unitCost * output.AchievedQtyAmount
		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         woplan.CompanyID,
			Dimension:         woplan.Dimension,
			SourceType:        scmmodel.ModuleWorkorder,
			SourceJournalID:   woplan.ID,
			SourceJournalType: woplan.JournalTypeID,
			SourceTrxType:     string(woplan.TrxType),
			SourceLineNo:      0,
			TrxDate:           woplan.TrxDate,
			Text:              woplan.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            totalCost,
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", woplan.ID),
		}

		var offsetLedger *tenantcoremodel.LedgerAccount
		jt, _ := journalTypes.Get(woplan.JournalTypeID)
		offsetLedger, err = ledgers.Get(jt.DefaultOffsiteConsumption.AccountID)
		if err != nil {
			return
		}

		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         woplan.CompanyID,
			Dimension:         woplan.Dimension,
			SourceType:        scmmodel.ModuleWorkorder,
			SourceJournalID:   woplan.ID,
			SourceJournalType: woplan.JournalTypeID,
			SourceTrxType:     string(woplan.TrxType),
			SourceLineNo:      0,
			TrxDate:           woplan.TrxDate,
			Text:              woplan.Text,
			Account:           *offsetLedger,
			Status:            ficomodel.AmountConfirmed,
			Amount:            (-1 * totalCost),
			References: tenantcoremodel.References{}.
				Set("WorkOrderID", woplan.ID),
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	})

	return ledgerTrxs, nil
}
