package mfglogic

import (
	"fmt"
	"io"
	"math"
	"strconv"
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
	"git.kanosolution.net/sebar/sdp/sdpmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type workOrderPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *mfgmodel.WorkOrderPlan
	trxType string

	lines []mfgmodel.WorkOrderSummaryMaterial
	items *sebar.MapRecord[*tenantcoremodel.Item]

	jt                mfgmodel.WorkOrderJournalType
	inventTrxs        []*scmmodel.InventTrx  // for Reserving purpose
	requestedItems    []ItemRequestLineParam // for Item Request creation
	materialInventTrx []materialInventTrx    // for update Reserved in Summary Material
}

type materialInventTrx struct {
	WorkOrderSummaryMaterial mfgmodel.WorkOrderSummaryMaterial
	InvTrx                   *scmmodel.InventTrx
}

func NewWorkOrderPlanPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*mfgmodel.WorkOrderPlan, mfgmodel.WorkOrderSummaryMaterial] {
	p := new(workOrderPosting)
	p.ctx = ctx
	p.opt = &opt
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*mfgmodel.WorkOrderPlan, mfgmodel.WorkOrderSummaryMaterial](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *workOrderPosting) Header() (*mfgmodel.WorkOrderPlan, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderPlan), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.jt = *jt
	p.trxType = lo.Ternary(string(jt.TrxType) != "", string(jt.TrxType), "Work Order")

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), jt.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", jt.PostingProfileID)
	}

	p.header = j
	return j, pp, nil
}

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *workOrderPosting) Lines() ([]mfgmodel.WorkOrderSummaryMaterial, error) {
	materialLines := []mfgmodel.WorkOrderSummaryMaterial{}
	p.opt.Db.GetsByFilter(new(mfgmodel.WorkOrderSummaryMaterial), dbflex.Eq("WorkOrderPlanID", p.header.ID), &materialLines)
	p.lines = lo.Map(materialLines, func(d mfgmodel.WorkOrderSummaryMaterial, i int) mfgmodel.WorkOrderSummaryMaterial {
		d.LineNo = (i + 1)
		return d
	})

	return p.lines, nil
}

func (p *workOrderPosting) Calculate(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlan, lines []mfgmodel.WorkOrderSummaryMaterial) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var err error
	trxs := map[string][]orm.DataModel{}

	// convert lines ke InventTrx yang nantinya digunakan untuk di Reserve (di Post() func)
	for _, line := range p.lines {
		inventDim := scmlogic.NewInventDimHelper(scmlogic.InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&line.InventDim)

		bal := new(scmmodel.ItemBalance)
		bals := GetAvailableStocks(p.opt.Db, GetAvailableStocksParam{
			CompanyID: p.opt.CompanyID,
			InventDim: *inventDim,
			Items:     []GetAvailableStocksParamItem{{ItemID: line.ItemID, SKU: line.SKU}},
		})
		if len(bals) > 0 {
			bal = bals[0]
		}

		// convert line Required Qty ke default
		it, _ := p.items.Get(line.ItemID)

		requiredDefUnit, e := scmlogic.ConvertUnit(p.opt.Db, line.Required, line.UnitID, it.DefaultUnitID)
		if e != nil {
			return nil, nil, 0, e
		}

		// compare dengan item balance
		if requiredDefUnit > bal.Qty {
			// butuh request item sejumlah kurangnya
			requestQty := requiredDefUnit - bal.Qty
			p.requestedItems = append(p.requestedItems, ItemRequestLineParam{
				ItemID:       line.ItemID,
				SKU:          line.SKU,
				QtyRequested: requestQty,
				UoM:          it.DefaultUnitID, // request dengan default unit biar setara semua
				InventDimTo:  *inventDim,
				Dimension:    header.Dimension,
			})

			// requiredDefUnit = bal.Qty // reserved yang ada // di comment karena seharusnya reserve itu full qty, bukan hanya yang tidak di IR aja
		}

		requiredTrxUnit, e := scmlogic.ConvertUnit(p.opt.Db, requiredDefUnit, it.DefaultUnitID, line.UnitID)
		if e != nil {
			return nil, nil, 0, e
		}

		//calculate cost per fix QTY
		costPerUnit := 0.0
		unitCost := it.CostUnit
		if unitCost == 0 {
			costPerUnit = scmlogic.GetCostPerUnit(p.opt.Db, *it, *inventDim, &p.header.TrxDate)
			unitCost = costPerUnit * requiredTrxUnit
		} else {
			costPerUnit = unitCost / requiredTrxUnit
		}

		// invent trx untuk reserved di post
		inventTrx := new(scmmodel.InventTrx)
		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.TrxDate = p.header.TrxDate

		inventTrx.Item = *it
		inventTrx.SKU = line.SKU
		inventTrx.Qty = lo.Ternary(requiredDefUnit > 0, (-1 * requiredDefUnit), requiredDefUnit)
		inventTrx.TrxQty = lo.Ternary(requiredTrxUnit > 0, (-1 * requiredTrxUnit), requiredTrxUnit)
		inventTrx.TrxUnitID = line.UnitID
		inventTrx.AmountPhysical = unitCost * requiredDefUnit
		inventTrx.AmountPhysical = lo.Ternary(inventTrx.AmountPhysical > 0, inventTrx.AmountPhysical*-1, inventTrx.AmountPhysical)
		inventTrx.AmountFinancial = inventTrx.AmountPhysical

		inventTrx.Status = scmmodel.ItemReserved // Reserved dari Item Balance

		inventTrx.SourceType = tenantcoremodel.TrxModule(mfgmodel.JournalWorkOrderPlan)
		inventTrx.SourceJournalID = p.header.ID
		inventTrx.SourceLineNo = line.LineNo
		inventTrx.SourceTrxType = string(p.trxType)

		inventTrx.InventDim = *inventDim
		inventTrx.Dimension = p.header.Dimension

		p.inventTrxs = append(p.inventTrxs, inventTrx)
		p.materialInventTrx = append(p.materialInventTrx, materialInventTrx{WorkOrderSummaryMaterial: line, InvTrx: inventTrx})
	}

	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(p.inventTrxs)

	return p.GetPreview(opt, header, lines), trxs, 0, err
}

func (p *workOrderPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlan, lines []mfgmodel.WorkOrderSummaryMaterial) *tenantcoremodel.PreviewReport {
	preview := tenantcoremodel.PreviewReport{}

	empORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Employee))
	dimORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.DimensionMaster))
	whORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LocationWarehouse))
	assetORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Asset))
	bomORM := sebar.NewMapRecordWithORM(opt.Db, new(mfgmodel.BoM))
	uomORM := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.UoM))
	itemORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Item))
	specORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemSpec))
	mdORM := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.MasterData))

	requestorWO, _ := empORM.Get(header.RequestorWOName)
	requestorWR, _ := empORM.Get(header.RequestorName)
	cc, _ := dimORM.Get(header.Dimension.Get("CC"))
	wh, _ := whORM.Get(header.InventDim.WarehouseID)
	asset, _ := assetORM.Get(header.Asset)
	bom, _ := bomORM.Get(header.BOM)
	woTypeKindMD, _ := mdORM.Get(header.WoTypeKind)
	woTypeKindName := lo.Ternary(woTypeKindMD != nil, fmt.Sprintf(" %s", woTypeKindMD.Name), "")
	assetMD, _ := mdORM.Get(asset.DriveType)

	signatureRequestor := tenantcoremodel.Signature{
		ID:        header.RequestorWOName,
		Header:    "Pembuat",
		Footer:    requestorWO.Name,
		Confirmed: "Tgl. " + header.Created.Format("02-January-2006 15:04:05"),
	}
	preview.Signature = append(preview.Signature, signatureRequestor)

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(mfgmodel.JournalWorkOrderPlan), header.ID)
	preview.Signature = append(preview.Signature, signature...)
	current := time.Now()

	// preview header and footer
	preview.Header = codekit.M{
		"Data": [][]string{
			// Full Ga Rapi
			// {"WO No:", header.ID, "", "Requestor:", requestorWO.Name, "", "WR Information", "", "", ""},
			// {"WO Name:", header.WOName, "", "WO Type:", header.WorkRequestType, "", "Reff WR No:", header.WorkRequestID, "WR Date:", FormatDate(&header.WRDate)},
			// {"Site:", header.Dimension.Get("Site"), "", "Priority:", header.Priority, "", "WR Requestor:", requestorWR.Name, "Department:", cc.Label},
			// {"Warehouse:", wh.Name, "", "Breakdown Type:", header.BreakdownType, "", "Asset:", asset.Name, "Code Unit:", header.HullNo},
			// {"WO Date:", FormatDate(&header.TrxDate), "", "BOM:", bom.Title, "", "No Hull Customer:", header.NoHullCustomer, "Sun No:", header.SunID},
			// {"Trx Creation Date:", FormatDate(&header.TrxCreatedDate), "", "Safety instruction:", strings.Join(header.SafetyInstruction, ", "), "", "Start Downtime:", FormatDate(header.StartDownTime), "Carosery Code:", header.CaroseryCode},
			// {"Description:", "", "", "", "", "", "Expected Completed Date:", FormatDate(&header.ExpectedCompletedDate), "Kilometers:", FormatFloatDecimal2(header.Kilometers)},
			// {"", header.WorkDescription, "", "", "", "", "Merk:", header.Merk, "Unit Type:", header.UnitType},

			// Full tapi lebih rapi
			// {"WO No:", header.ID, "Requestor:", requestorWO.Name, "WR Information", "", "", ""},
			// {"WO Name:", header.WOName, "WO Type:", header.WorkRequestType, "Reff WR No:", header.WorkRequestID, "WR Date:", FormatDate(&header.WRDate)},
			// {"Site:", header.Dimension.Get("Site"), "Priority:", header.Priority, "WR Requestor:", requestorWR.Name, "Department:", cc.Label},
			// {"Warehouse:", wh.Name, "Breakdown Type:", header.BreakdownType, "Asset:", asset.Name, "Code Unit:", header.HullNo},
			// {"WO Date:", FormatDate(&header.TrxDate), "BOM:", bom.Title, "No Hull Customer:", header.NoHullCustomer, "Sun No:", header.SunID},
			// {"Trx Creation Date:", FormatDate(&header.TrxCreatedDate), "Safety instruction:", strings.Join(header.SafetyInstruction, ", "), "Start Downtime:", FormatDate(header.StartDownTime), "Carosery Code:", header.CaroseryCode},
			// {"Description:", "", "", "", "Expected Completed Date:", FormatDate(&header.ExpectedCompletedDate), "Kilometers:", FormatFloatDecimal2(header.Kilometers)},
			// {"", header.WorkDescription, "", "", "Merk:", header.Merk, "Unit Type:", header.UnitType},

			// {"WO No:", header.ID, "WR Information", "", "", ""},
			// {"WO Name:", header.WOName, "Reff WR No:", header.WorkRequestID, "WR Date:", FormatDate(&header.WRDate)},
			// {"Site:", header.Dimension.Get("Site"), "WR Requestor:", requestorWR.Name, "Department:", cc.Label},
			// {"Warehouse:", wh.Name, "Asset:", asset.Name, "Code Unit:", header.HullNo},
			// {"WO Date:", FormatDate(&header.TrxDate), "No Hull Customer:", header.NoHullCustomer, "Sun No:", header.SunID},
			// {"Trx Creation Date:", FormatDate(&header.TrxCreatedDate), "Start Downtime:", FormatDate(header.StartDownTime), "Carosery Code:", header.CaroseryCode},

			// {"Requestor:", requestorWO.Name, "", "", "", ""},
			// {"WO Type:", header.WorkRequestType, "", "", "", ""},
			// {"Breakdown Type:", header.BreakdownType, "", "", "", ""},
			// {"Priority:", header.Priority, "", "", "", ""},
			// {"BOM:", bom.Title, "", "", "", ""},
			// {"Safety instruction:", strings.Join(header.SafetyInstruction, ", "), "", "", "", ""},

			// {"Description:", "", "", "", "Expected Completed Date:", FormatDate(&header.ExpectedCompletedDate), "Kilometers:", FormatFloatDecimal2(header.Kilometers)},
			// {"", header.WorkDescription, "", "", "Merk:", header.Merk, "Unit Type:", header.UnitType},

			// Sesuai Template dari Bagong
			{"WO No:", header.ID, "Site:", header.Dimension.Get("Site")},
			{"Asset:", asset.Name, "Priority:", header.Priority},
			{"Unit Type:", header.UnitType, "WO Date:", FormatDate(&header.TrxDate)},
			{"Model Unit:", assetMD.Name, "Printed On:", FormatDate(&current)},
			{"Karoseri:", header.CaroseryCode, "Reff WR No:", header.WorkRequestID},
			{"KM Unit:", FormatFloatDecimal2(header.Kilometers)},
			{"WO Description:", header.WorkDescription},
		},
		"DataJobCard": [][]string{
			{"Job Type:", fmt.Sprintf("%s%s", p.jt.Name, woTypeKindName), "", ""},
			{"Start Downtime:", FormatDateTime(header.StartDownTime), "", ""},
			{"Safety Instruction:", strings.Join(header.SafetyInstruction, ", "), "", ""},
			{"Job Instruction:", header.WRDescription, "", ""},
			{"Note:", header.WRDescription, "", ""},
		},
		"Footer": [][]string{},
	}

	preview.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"WO No:", header.ID},
			{"WO Name:", header.WOName},
			{"Site:", header.Dimension.Get("Site")},
			{"Warehouse:", wh.Name},
			{"WO Date:", FormatDate(&header.TrxDate)},
			{"Trx Creation Date:", FormatDate(&header.TrxCreatedDate)},
			{"Description:", ""},
			{"", header.WorkDescription},

			{"Requestor:", requestorWO.Name},
			{"WO Type:", header.WorkRequestType},
			{"Priority:", header.Priority},
			{"Breakdown Type:", header.BreakdownType},
			{"BOM:", bom.Title},
			{"Safety instruction:", strings.Join(header.SafetyInstruction, ", ")},

			{"WR Information:", ""},
			{"Reff WR No:", header.WorkRequestID}, {"WR Date:", FormatDate(&header.WRDate)},
			{"WR Requestor:", requestorWR.Name}, {"Department:", cc.Label},
			{"Asset:", asset.Name}, {"Code Unit:", header.HullNo},
			{"No Hull Customer:", header.NoHullCustomer}, {"Sun No:", header.SunID},
			{"Start Downtime:", FormatDate(header.StartDownTime)}, {"Carosery Code:", header.CaroseryCode},
			{"Expected Completed Date:", FormatDate(&header.ExpectedCompletedDate)}, {"Kilometers:", FormatFloatDecimal2(header.Kilometers)},
			{"Merk:", header.Merk}, {"Unit Type:", header.UnitType},
		},
		Footer: [][]string{},
	}

	// build lines
	lineCount := 1
	previewLines := [][]string{{"Part Number", "Part Description", "UoM", "Required", "Available Stock", "Warehouse Location"}}
	prelines := lo.Map(lines, func(d mfgmodel.WorkOrderSummaryMaterial, i int) []string {
		item, _ := itemORM.Get(d.ItemID)
		u, _ := uomORM.Get(d.UnitID)
		wh, _ := whORM.Get(d.InventDim.WarehouseID)
		spec, _ := specORM.Get(d.SKU)
		tenantcorelogic.MWPreAssignItem(d.ItemID+"~~"+d.SKU, false)(p.ctx, &item)

		result := []string{strconv.Itoa(lineCount) + ". " + spec.SKU, lo.Ternary(spec.OtherName != "", spec.OtherName, item.Name), u.Name, FormatNumberNoDecimal(d.Required), FormatNumberNoDecimal(d.AvailableStock), wh.Name}
		lineCount++
		return result
	})
	previewLines = append(previewLines, prelines...)

	// preview sections
	preview.Sections = append(preview.Sections, tenantcoremodel.PreviewSection{
		Title:       "",
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items:       previewLines,
	})

	return &preview
}

// ToJournalLines adalah proses convert dari line journal kita ke ficomodel.JournalLine
func (p *workOrderPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlan, lines []mfgmodel.WorkOrderSummaryMaterial) []ficomodel.JournalLine {
	return lo.Map(lines, func(line mfgmodel.WorkOrderSummaryMaterial, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		return *jl
	})
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *workOrderPosting) Post(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderPlan, lines []mfgmodel.WorkOrderSummaryMaterial, trxs map[string][]orm.DataModel) (string, error) {
	var (
		res string
		err error
	)

	err = sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		// proses reserve qty tiap lines & sync with NewBalanceHub
		// filter hanya trx yang ada qty nya, khawatirnya ada item yang di warehouse tsb tidak ada stok sama sekali, jadinya semuanya di Item Request kan bukan di reserve disini
		p.inventTrxs = lo.Filter(p.inventTrxs, func(d *scmmodel.InventTrx, i int) bool { return d.Qty != 0 })
		trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(p.inventTrxs)

		savedTrxs, e := scmlogic.InventTrxSingleJournalSave(tx, trxs, p.opt.CompanyID, p.trxType, p.header.ID, scmmodel.ItemReserved)
		if e != nil {
			return e
		}

		res, err = ficologic.PostModelSave(tx, p.header, "WorkOrderVoucherNo", map[string][]orm.DataModel{}) // validate header only
		if err != nil {
			return err
		}

		_, err = scmlogic.NewItemBalanceHub(tx).Sync(nil, scmlogic.ItemBalanceOpt{
			CompanyID:       header.CompanyID,
			ConsiderSKU:     true,
			DisableGrouping: true,
			ItemIDs: lo.Map(savedTrxs, func(t *scmmodel.InventTrx, index int) string {
				return t.Item.ID
			}),
		})
		if err != nil {
			return err
		}

		// proses auto item request if any
		requestedItemPerWH := lo.GroupBy(p.requestedItems, func(d ItemRequestLineParam) string {
			// d.InventDimTo = *d.InventDimTo.Calc()
			return d.InventDimTo.WarehouseID
		})

		for _, lineItems := range requestedItemPerWH {
			// create Item Request
			_, err = ItemRequestInsert(p.ctx, &ItemRequestInsertParam{
				Name:        fmt.Sprintf("FROM WO: %s", p.header.ID),
				WOReff:      header.ID,
				Requestor:   header.RequestorWOName,
				Department:  header.RequestorDepartment,
				InventDimTo: lineItems[0].InventDimTo,
				Dimension:   header.Dimension,
				Lines:       lineItems,
			}, p.opt.CompanyID, p.opt.UserID)
			if err != nil {
				return err
			}
		}

		// update Summary Material
		for _, minv := range p.materialInventTrx {
			summ := minv.WorkOrderSummaryMaterial
			summ.Reserved = math.Abs(minv.InvTrx.TrxQty)
			summ.LineNo = minv.InvTrx.SourceLineNo
			tx.Update(&summ, "Reserved", "LineNo")
		}

		header.Status = ficomodel.JournalStatusPosted
		header.StatusOverall = mfgmodel.WorkOrderPlanStatusOverallInProgress
		tx.Update(header, "Status", "StatusOverall")

		if header.JournalTypeID == "WO_Production" {
			// Insert to Document Unit Checklist (SDPM)
			DUC := sdpmodel.DocumentUnitChecklist{}
			DUC.WONo = header.ID
			DUC.SUNID = header.SunID
			DUC.HullNo = header.HullNo
			DUC.AssetID = header.Asset
			DUC.Dimension = header.Dimension

			if err := tx.Save(&DUC); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return res, err
	}

	return res, err
}

func (p *workOrderPosting) Approved() error {
	return nil
}

func (p *workOrderPosting) Rejected() error {
	return nil
}

func (p *workOrderPosting) GetAccount() string {
	return ""
}

func (p *workOrderPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
	employee := new(tenantcoremodel.Employee)
	err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
		dbflex.Eq("_id", p.opt.UserID),
	))
	if err != nil && err != io.EOF {
		return fmt.Errorf("error when get email employee : %s", err.Error())
	}

	for _, app := range pa.Approvals {
		if app.Line == pa.CurrentStage {
			notification := ficomodel.Notification{
				UserSubmitter:            p.opt.UserID,
				UserSubmitterEmail:       employee.Email,
				JournalID:                p.header.ID,
				JournalType:              p.header.JournalTypeID,
				PostingProfileApprovalID: pa.ID,
				TrxDate:                  p.header.TrxDate,
				Text:                     p.header.Text,
				UserTo:                   app.UserID,
				TrxType:                  string(p.header.TrxType),
				Menu:                     string(p.header.TrxType),
				Status:                   app.Status,
				CompanyID:                p.opt.CompanyID,
			}

			employee := new(tenantcoremodel.Employee)
			err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
				dbflex.Eq("_id", app.UserID),
			))
			if err != nil && err != io.EOF {
				return fmt.Errorf("error when get employee : %s", err.Error())
			}

			notification.UserToEmail = employee.Email

			err = p.opt.Db.Save(&notification)
			if err != nil {
				return fmt.Errorf("error when save notification : %s", err.Error())
			}
		}
	}

	return nil
}

func (p *workOrderPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
	// get latest notification
	latestNotif := new(ficomodel.Notification)
	err := p.opt.Db.GetByParm(latestNotif, dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("JournalID", p.header.ID),
			dbflex.Eq("UserTo", p.opt.UserID),
		),
	).SetSort("-Created"))
	if err != nil {
		return fmt.Errorf("error when get notification : %s", err.Error())
	}

	if op == ficologic.PostOpApprove {
		latestNotif.Status = string(ficomodel.JournalStatusApproved)
	} else {
		latestNotif.Status = string(ficomodel.JournalStatusRejected)
	}

	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification : %s", err.Error())
	}

	// create notification for submitter user
	latestNotif.ID = primitive.NewObjectID().Hex()
	latestNotif.UserTo = latestNotif.UserSubmitter
	latestNotif.UserToEmail = latestNotif.UserSubmitterEmail
	err = p.opt.Db.Save(latestNotif)
	if err != nil {
		return fmt.Errorf("error when save notification for submitter user : %s", err.Error())
	}

	// get user approval stage
	userApprovals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.UserID == p.opt.UserID
	})

	approvals := lo.Filter(pa.Approvals, func(a *ficomodel.PostingProfileApprovalItem, index int) bool {
		return a.Line == userApprovals[0].Line && a.Status != "PENDING"
	})

	// check if need to send notification
	if len(approvals) >= pa.Approvers[userApprovals[0].Line-1].MinimalApproverCount &&
		userApprovals[0].Line != pa.CurrentStage {
		for _, app := range pa.Approvals {
			if app.Line == pa.CurrentStage {
				latestNotif.ID = primitive.NewObjectID().Hex()
				latestNotif.UserTo = app.UserID
				latestNotif.Status = app.Status

				employee := new(tenantcoremodel.Employee)
				err := p.opt.Db.GetByParm(employee, dbflex.NewQueryParam().SetWhere(
					dbflex.Eq("_id", app.UserID),
				))
				if err != nil && err != io.EOF {
					return fmt.Errorf("error when get employee : %s", err.Error())
				}

				latestNotif.UserToEmail = employee.Email

				err = p.opt.Db.Save(latestNotif)
				if err != nil {
					return fmt.Errorf("error when save notification for next approval : %s", err.Error())
				}
			}
		}
	}

	return nil
}
