package scmlogic

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stockOpnamePosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.StockOpnameJournal
	jt      *scmmodel.InventJournalType
	trxType string

	lines []scmmodel.StockOpnameJournalLine
	trxs  map[string][]orm.DataModel
	items *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewStockOpnamePosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.StockOpnameJournal, scmmodel.StockOpnameJournalLine] {
	p := new(stockOpnamePosting)
	p.opt = &opt
	p.ctx = ctx
	p.items = sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	pvd := ficologic.PostingProvider[*scmmodel.StockOpnameJournal, scmmodel.StockOpnameJournalLine](p)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *stockOpnamePosting) Header() (*scmmodel.StockOpnameJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.StockOpnameJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	p.jt, err = datahub.GetByID(p.opt.Db, new(scmmodel.InventJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}

	p.trxType = string(p.jt.TransactionType)
	if p.trxType == "" {
		return nil, nil, fmt.Errorf("invalid transaction type")
	}

	j.PostingProfileID = tenantcorelogic.TernaryString(j.PostingProfileID, p.jt.PostingProfileID)
	if j.PostingProfileID == "" {
		return nil, nil, fmt.Errorf("missing: posting profile")
	}

	pp, err := datahub.GetByID(p.opt.Db, new(ficomodel.PostingProfile), j.PostingProfileID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: posting profile: %s", j.PostingProfileID)
	}

	p.header = j
	return j, pp, nil
}

// Lines adalah proses untuk pengisian field-field dari line journal kita
func (p *stockOpnamePosting) Lines() ([]scmmodel.StockOpnameJournalLine, error) {
	p.header.Lines = lo.Map(p.header.Lines, func(line scmmodel.StockOpnameJournalLine, index int) scmmodel.StockOpnameJournalLine {
		line.ID = primitive.NewObjectID().Hex()
		return line
	})
	return p.header.Lines, nil
}

func (p *stockOpnamePosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.StockOpnameJournal, lines []scmmodel.StockOpnameJournalLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.StockOpnameJournalLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.OffsetAccount = p.jt.DefaultOffset
		jl.Amount = line.UnitCost * line.Qty
		return *jl
	})
}

func (p *stockOpnamePosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.StockOpnameJournal, lines []scmmodel.StockOpnameJournalLine) (*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		err error
	)

	preview := tenantcoremodel.PreviewReport{}
	ledgers := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LedgerAccount))
	itemORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Item))
	itemGroups := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemGroup))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}
	trxs := map[string][]orm.DataModel{}
	inventTrxs := []*scmmodel.InventTrx{}

	if err != nil {
		return &preview, trxs, 0, err
	}

	for index, line := range lines {
		item, _ := itemORM.Get(line.ItemID)
		inventTrxLine := scmmodel.InventTrxLine{
			InventJournalLine: scmmodel.InventJournalLine{
				ItemID:    line.ItemID,
				Qty:       line.Qty,
				Text:      line.Text,
				Dimension: line.Dimension,
				LineNo:    line.LineNo,
				SKU:       line.SKU,
				InventDim: line.InventDim,
			},
			JournalID: p.header.ID,
			TrxType:   scmmodel.InventTrxType(p.trxType),
			Item:      item,
			TrxDate:   p.header.TrxDate,
		}

		inventTrx, err := trxLineToInventTrx(&inventTrxLine, opt.Db)
		if err != nil {
			return &preview, trxs, 0, fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}

		inventTrx.CompanyID = p.header.CompanyID
		inventTrxs = append(inventTrxs, inventTrx)

		itemGroup, _ := itemGroups.Get(inventTrx.Item.ItemGroupID)
		ledgerAccount, err := ledgers.Get(inventTrx.Item.LedgerAccountIDStock)
		if err != nil {
			ledgerAccount, err = ledgers.Get(itemGroup.LedgerAccountIDStock)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: main inventory account for item %s: %s, %s", line.ItemID, p.jt.DefaultOffset.AccountID, err.Error())
			}
		}

		costPerUnit := GetCostPerUnit(p.opt.Db, inventTrx.Item, inventTrx.InventDim, &p.header.TrxDate)
		totalCost := costPerUnit * line.Qty
		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        scmmodel.ModuleInventory,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     p.trxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Status:            ficomodel.AmountConfirmed,
			Account:           *ledgerAccount,
			Amount:            totalCost,
		}

		offset, err := ledgers.Get(p.jt.DefaultOffset.AccountID)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("invalid: offset account on journal type: %s, %s", p.jt.DefaultOffset.AccountID, err.Error())
		}
		ltOffset := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        scmmodel.ModuleInventory,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     p.trxType,
			SourceLineNo:      line.LineNo,
			TrxDate:           p.header.TrxDate,
			Text:              line.Text,
			Account:           *offset,
			Status:            ficomodel.AmountConfirmed,
			Amount:            -totalCost,
		}

		ledgerTrxs = append(ledgerTrxs, ltMain, ltOffset)
	}

	if len(inventTrxs) > 0 {
		trxs[inventTrxs[0].TableName()] = ficologic.ToDataModels(inventTrxs)
	}

	if len(ledgerTrxs) > 0 {
		trxs[ledgerTrxs[0].TableName()] = ficologic.ToDataModels(ledgerTrxs)
	}

	// err = p.validate(inventTrxs)
	p.trxs = trxs

	return p.GetPreview(opt, header, lines), trxs, lo.SumBy(inventTrxs, func(i *scmmodel.InventTrx) float64 {
		return i.AmountPhysical
	}), err
}

// Post proses me-reserve inventory dan set status ke Posted
func (p *stockOpnamePosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.StockOpnameJournal, lines []scmmodel.StockOpnameJournalLine, trxs map[string][]orm.DataModel) (string, error) {
	err := sebar.Tx(p.opt.Db, true, func(tx *datahub.Hub) error {
		var err error

		gapLines := lo.Filter(lines, func(line scmmodel.StockOpnameJournalLine, index int) bool {
			return line.Gap != 0
		})

		if len(gapLines) > 0 {
			id, e := tenantcorelogic.GenerateIDFromNumSeq(p.ctx, "InventoryAdjustment")
			if e != nil {
				id = "" // dikosongin, biar pake object id
			}

			inv := scmmodel.InventoryAdjustment{
				ID:             id,
				StockOpnameID:  header.ID,
				AdjustmentDate: time.Now(),
				CompanyID:      header.CompanyID,
				Status:         scmmodel.InventoryAdjustmentStatusDone, // langsung Done, karena processing langsung dilakukan di code di bawah
				Dimension:      header.Dimension,
				InventDim:      header.InventDim,
			}
			if e := tx.Save(&inv); e != nil {
				return e
			}

			invDetails := []scmmodel.InventoryAdjustmentDetail{}
			for gli, gl := range gapLines {
				qtyActualNumb, err := strconv.ParseFloat(gl.QtyActual, 64)
				if err != nil {
					return err
				}

				invDetail := scmmodel.InventoryAdjustmentDetail{
					ID:                    gl.ID,
					InventoryAdjustmentID: inv.ID,
					ItemID:                gl.ItemID,
					SKU:                   gl.SKU,
					Description:           gl.Description,
					InventoryDimension:    gl.InventDim,
					UoM:                   gl.UnitID,

					ItemName:    gl.ItemName,
					UnitName:    gl.UnitName,
					AisleName:   gl.AisleName,
					SectionName: gl.SectionName,
					BoxName:     gl.BoxName,

					QtyInSystem: gl.QtyInSystem,
					QtyActual:   qtyActualNumb,
					Gap:         gl.Gap,
					Remarks:     gl.Remarks,
					Note:        gl.Note,
					// NoteStaff:             gl.NoteStaff,
					LineNo: gli + 1,
				}

				if e := tx.Save(&invDetail); e != nil {
					return e
				}

				invDetails = append(invDetails, invDetail)
			}

			err = new(InventoryAdjustmentEngine).process(tx, &inv, invDetails)
		}

		return err
	})
	if err != nil {
		return "", err
	}

	return "", nil
}

func (p *stockOpnamePosting) Approved() error {
	return nil
}

func (p *stockOpnamePosting) Rejected() error {
	return nil
}

func (p *stockOpnamePosting) GetAccount() string {
	return p.header.Name // TODO: seharusnya return header.Text kalau field Name sudah diganti dengan Text
}

func (p *stockOpnamePosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.StockOpnameJournal, lines []scmmodel.StockOpnameJournalLine) *tenantcoremodel.PreviewReport {
	pv := new(tenantcoremodel.PreviewReport)

	itemORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.Item))
	specORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemSpec))
	uomORM := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.UoM))
	siteORM := sebar.NewMapRecordWithORM(opt.Db, new(bagongmodel.Site))
	site, _ := siteORM.Get(header.Dimension.Get("Site"))

	signature, _ := GetSignatureByID(opt.Db, header.CompanyID, string(scmmodel.ModuleInventory), header.ID)
	pv.Signature = append(pv.Signature, signature...)

	pv.Header = codekit.M{}.Set("Data", [][]string{
		{"ID:", header.ID},
		{"Name:", header.Text},
		{"Site:", site.Name},
		{"Date:", FormatDate(&header.TrxDate)},
	}).Set("Footer", [][]string{})

	pv.HeaderMobile = tenantcoremodel.PreviewReportHeaderMobile{
		Data: [][]string{
			{"ID:", header.ID},
			{"Name:", header.Text},
			{"Site:", site.Name},
			{"Date:", FormatDate(&header.TrxDate)},
		},
		Footer: [][]string{},
	}

	sectionLine := tenantcoremodel.PreviewSection{
		HideTitle:   false,
		SectionType: tenantcoremodel.PreviewAsGrid,
		Items: [][]string{
			{"No", "Part Number", "Part Description", "Qty", "UoM", "Remarks"},
		},
	}

	lineCount := 1
	sectionLine.Items = append(sectionLine.Items, lo.Map(lines, func(line scmmodel.StockOpnameJournalLine, index int) []string {
		item, _ := itemORM.Get(line.ItemID)
		spec, _ := specORM.Get(line.SKU)
		unit, _ := uomORM.Get(line.UnitID)

		tenantcorelogic.MWPreAssignItem(line.ItemID+"~~"+line.SKU, false)(p.ctx, &item)

		res := make([]string, 6)
		res[0] = strconv.Itoa(lineCount)
		res[1] = spec.SKU
		res[2] = lo.Ternary(item.ID != "", item.ID, item.Name)
		res[3] = fmt.Sprintf("%.2f", line.Qty)
		res[4] = unit.Name
		res[5] = string(line.Remarks)

		lineCount++
		return res
	})...)

	pv.Sections = append(pv.Sections, sectionLine)

	return pv
}

func (p *stockOpnamePosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
	empORM := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Employee))
	employee, err := empORM.Get(p.opt.UserID)
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

			employee, err = empORM.Get(app.UserID)
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

func (p *stockOpnamePosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
