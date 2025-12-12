package scmlogic

import (
	"fmt"
	"io"
	"strconv"

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

type InventoryPosting struct {
	ctx     *kaos.Context
	opt     *ficologic.PostingHubCreateOpt
	header  *scmmodel.InventJournal
	jt      *scmmodel.InventJournalType
	trxType string
	trxs    map[string][]orm.DataModel
}

func NewInventoryPosting(ctx *kaos.Context, opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*scmmodel.InventJournal, scmmodel.InventTrxLine] {
	invent := new(InventoryPosting)
	invent.ctx = ctx
	invent.opt = &opt

	pvd := ficologic.PostingProvider[*scmmodel.InventJournal, scmmodel.InventTrxLine](invent)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *InventoryPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.InventTrxLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.OffsetAccount = p.jt.DefaultOffset
		jl.Amount = line.UnitCost * line.Qty
		return *jl
	})
}

func (p *InventoryPosting) Header() (*scmmodel.InventJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(scmmodel.InventJournal), p.opt.JournalID)
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

func (p *InventoryPosting) Lines() ([]scmmodel.InventTrxLine, error) {
	var err error
	mapItems := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	lineTrxs := make([]scmmodel.InventTrxLine, len(p.header.Lines))
	for idx, headerLine := range p.header.Lines {
		if p.jt.TransactionType == scmmodel.JournalMovementOut {
			headerLine.Qty = -headerLine.Qty
		}

		lineTrx := scmmodel.InventTrxLine{}
		lineTrx.InventJournalLine = headerLine
		if lineTrx.Text == "" {
			lineTrx.Text = p.header.Text
		}
		lineTrx.JournalID = p.header.ID
		lineTrx.TrxDate = p.header.TrxDate
		lineTrx.TrxType = p.header.TrxType
		lineTrx.LineNo = idx + 1

		lineTrx.Item, err = mapItems.Get(headerLine.ItemID)
		if err != nil {
			return lineTrxs, fmt.Errorf("invalid item: %s: %s", headerLine.ItemID, err.Error())
		}

		lineTrx.Dimension = tenantcorelogic.TernaryDimension(lineTrx.Dimension, p.header.Dimension)
		lineTrx.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: headerLine.SKU}).TernaryInventDimension(&headerLine.InventDim, &p.header.InventDim)

		unitRatio, err := ConvertUnit(p.opt.Db, 1, headerLine.UnitID, lineTrx.Item.DefaultUnitID)
		if err != nil {
			return nil, fmt.Errorf("invalid: unit for item %s: %s", lineTrx.Item.ID, err.Error())
		}
		lineTrx.InventQty = unitRatio * headerLine.Qty

		if headerLine.UnitCost == 0 {
			lineTrx.CostPerUnit = GetCostPerUnit(p.opt.Db, *lineTrx.Item, lineTrx.InventDim, &p.header.TrxDate)
			lineTrx.UnitCost = lineTrx.CostPerUnit * unitRatio
		} else {
			lineTrx.CostPerUnit = headerLine.UnitCost / unitRatio
		}
		lineTrxs[idx] = lineTrx
	}

	return lineTrxs, nil
}

func (p *InventoryPosting) Calculate(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		err error
	)

	preview := tenantcoremodel.PreviewReport{}
	ledgers := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.LedgerAccount))
	itemGroups := sebar.NewMapRecordWithORM(opt.Db, new(tenantcoremodel.ItemGroup))
	ledgerTrxs := []*ficomodel.LedgerTransaction{}
	trxs := map[string][]orm.DataModel{}
	inventTrxs := []*scmmodel.InventTrx{}

	if err != nil {
		return &preview, trxs, 0, err
	}

	for index, line := range lines {
		inventTrx, err := trxLineToInventTrx(&line, opt.Db)
		if err != nil {
			return &preview, trxs, 0, fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}

		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.Status = scmmodel.ItemConfirmed
		inventTrxs = append(inventTrxs, inventTrx)

		itemGroup, _ := itemGroups.Get(inventTrx.Item.ItemGroupID)
		ledgerAccount, err := ledgers.Get(inventTrx.Item.LedgerAccountIDStock)
		if err != nil {
			ledgerAccount, err = ledgers.Get(itemGroup.LedgerAccountIDStock)
			if err != nil {
				return nil, nil, 0, fmt.Errorf("invalid: main inventory account for item %s: %s, %s", line.ItemID, p.jt.DefaultOffset.AccountID, err.Error())
			}
		}

		totalCost := line.CostPerUnit * line.InventQty
		ltMain := &ficomodel.LedgerTransaction{
			CompanyID:         p.header.CompanyID,
			Dimension:         line.Dimension,
			SourceType:        scmmodel.ModuleInventory,
			SourceJournalID:   p.header.ID,
			SourceJournalType: p.header.JournalTypeID,
			SourceTrxType:     string(line.TrxType),
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
			SourceTrxType:     string(line.TrxType),
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

	err = p.validate(inventTrxs)
	p.trxs = trxs
	return p.GetPreview(opt, header, lines), trxs, lo.SumBy(inventTrxs, func(i *scmmodel.InventTrx) float64 {
		return i.AmountPhysical
	}), err
}

func (p *InventoryPosting) GetPreview(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine) *tenantcoremodel.PreviewReport {
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
	sectionLine.Items = append(sectionLine.Items, lo.Map(lines, func(line scmmodel.InventTrxLine, index int) []string {
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
		res[5] = line.Remarks

		lineCount++
		return res
	})...)

	pv.Sections = append(pv.Sections, sectionLine)

	return pv
}

func (p *InventoryPosting) Post(opt ficologic.PostingHubExecOpt, header *scmmodel.InventJournal, lines []scmmodel.InventTrxLine, trxs map[string][]orm.DataModel) (string, error) {
	var (
		db  *datahub.Hub
		err error
		res string
	)

	db, _ = p.opt.Db.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	// track the inventory and posting it as RESVD or PLANNED
	p.opt.Db.DeleteByFilter(new(scmmodel.InventTrx),
		dbflex.Eqs(
			"CompanyID", p.header.CompanyID,
			"SourceType", scmmodel.ModuleInventory,
			"SourceJournalID", p.header.ID))
	inventTrxs := ficologic.FromDataModels(p.trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
	for _, trx := range inventTrxs {
		trx.Status = scmmodel.ItemConfirmed
		p.opt.Db.Save(trx)
	}

	delete(p.trxs, new(scmmodel.InventTrx).TableName())
	res, err = ficologic.PostModelSave(db, p.header, "MovementInVoucherNo", p.trxs)
	if err != nil {
		return res, err
	}

	balanceOpt := ItemBalanceOpt{}
	balanceOpt.CompanyID = p.header.CompanyID
	balanceOpt.DisableGrouping = true
	balanceOpt.ConsiderSKU = true
	balanceOpt.ItemIDs = lo.Map(inventTrxs, func(t *scmmodel.InventTrx, index int) string {
		return t.Item.ID
	})

	_, err = NewItemBalanceHub(db).Sync(nil, balanceOpt)

	if p.jt.TransactionType == scmmodel.JournalMovementIn {
		items := lo.GroupBy(inventTrxs, func(i *scmmodel.InventTrx) string {
			return fmt.Sprintf("%s_%s", i.Item.ID, i.InventDim.InventDimID)
		})

		for _, trx := range items {
			CalcUnitCost(p.opt.Db, trx[0].Item, trx[0].InventDim, &p.header.TrxDate)
		}
	}

	return res, err
}

func (p *InventoryPosting) Approved() error {
	// track the inventory and posting it as RESVD or PLANNED
	p.opt.Db.DeleteByFilter(new(scmmodel.InventTrx),
		dbflex.Eqs("CompanyID", p.header.CompanyID, "SourceType", scmmodel.ModuleInventory, "SourceJournalID", p.header.ID))
	inventTrxs := ficologic.FromDataModels(p.trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
	for _, trx := range inventTrxs {
		trx.Status = lo.Ternary(trx.Qty > 0, scmmodel.ItemPlanned, scmmodel.ItemReserved)
		p.opt.Db.Save(trx)
	}

	balanceOpt := ItemBalanceOpt{}
	balanceOpt.CompanyID = p.header.CompanyID
	balanceOpt.ConsiderSKU = true
	balanceOpt.DisableGrouping = true
	balanceOpt.ItemIDs = lo.Map(inventTrxs, func(t *scmmodel.InventTrx, index int) string {
		return t.Item.ID
	})
	_, err := NewItemBalanceHub(p.opt.Db).Sync(nil, balanceOpt)
	if err != nil {
		return err
	}
	return nil
}

func (p *InventoryPosting) Rejected() error {
	// do nothing
	return nil
}

func (p *InventoryPosting) validate(lines []*scmmodel.InventTrx) error {
	if len(lines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	// TODO: group lines by item balance calc opts: njagani kalo ada 2 lines dgn item dan invent dim yg sama
	for _, line := range lines {
		switch p.header.TrxType {
		case scmmodel.JournalMovementOut:
			inventDim := *NewInventDimHelper(InventDimHelperOpt{DB: p.opt.Db, SKU: line.SKU}).TernaryInventDimension(&line.InventDim, &p.header.InventDim)

			if e := ValidateBalance(p.opt.Db, line.TrxQty, line.TrxUnitID, line.Item.ID, p.header.CompanyID, inventDim); e != nil {
				return e
			}
		}
	}

	// check dimension

	// check inventory dimension

	// check item min max
	if e := ItemMinMaxValidation(p.opt.Db, lines, ""); e != nil {
		return e
	}

	return nil
}

func (p *InventoryPosting) GetAccount() string {
	return p.header.Text
}

func trxLineToInventTrx(line *scmmodel.InventTrxLine, _ *datahub.Hub) (*scmmodel.InventTrx, error) {
	trx := new(scmmodel.InventTrx)

	trx.Item = *line.Item
	trx.SKU = line.SKU // SKU perlu dihitung ulang ndak ya
	trx.Dimension = line.Dimension
	trx.InventDim = line.InventDim
	trx.TrxDate = line.TrxDate
	trx.Text = line.Text

	trx.Qty = line.InventQty
	trx.TrxQty = line.Qty
	trx.TrxUnitID = line.UnitID

	trx.SourceType = scmmodel.ModuleInventory
	trx.SourceJournalID = line.JournalID
	trx.SourceTrxType = string(line.TrxType)
	trx.SourceLineNo = line.LineNo
	trx.AmountPhysical = line.UnitCost * line.Qty
	trx.AmountFinancial = trx.AmountPhysical
	trx.SKU = line.SKU

	return trx, nil
}

func (p *InventoryPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
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

func (p *InventoryPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
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
