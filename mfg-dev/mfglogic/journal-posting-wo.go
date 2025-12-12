package mfglogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmlogic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
)

type WorkOrderPosting struct {
	opt     *ficologic.PostingHubCreateOpt
	header  *mfgmodel.WorkOrderJournal
	trxType string

	inventTrxs []*scmmodel.InventTrx
}

func NewJournalPostingWorkOrder(opt ficologic.PostingHubCreateOpt) *ficologic.PostingHub[*mfgmodel.WorkOrderJournal, scmmodel.InventReceiveIssueLine] {
	wo := new(WorkOrderPosting)
	wo.opt = &opt

	pvd := ficologic.PostingProvider[*mfgmodel.WorkOrderJournal, scmmodel.InventReceiveIssueLine](wo)
	return ficologic.NewPostingHub(pvd, opt)
}

func (p *WorkOrderPosting) ToJournalLines(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderJournal, lines []scmmodel.InventReceiveIssueLine) []ficomodel.JournalLine {
	return lo.Map(lines, func(line scmmodel.InventReceiveIssueLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}

func (p *WorkOrderPosting) Header() (*mfgmodel.WorkOrderJournal, *ficomodel.PostingProfile, error) {
	j, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderJournal), p.opt.JournalID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal: %s: %s", j.ID, err.Error())
	}

	jt, err := datahub.GetByID(p.opt.Db, new(mfgmodel.WorkOrderJournalType), j.JournalTypeID)
	if err != nil {
		return nil, nil, fmt.Errorf("missing: journal type: %s: %s", j.JournalTypeID, err.Error())
	}
	p.trxType = jt.TrxType
	if p.trxType == "" {
		p.trxType = string(scmmodel.JournalWorkOrder)
	}

	j.PostingProfileID = tenantcorelogic.TernaryString(j.PostingProfileID, jt.PostingProfileID)
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

func (p *WorkOrderPosting) Lines() ([]scmmodel.InventReceiveIssueLine, error) {
	mapItems := sebar.NewMapRecordWithORM(p.opt.Db, new(tenantcoremodel.Item))
	lines := make([]scmmodel.InventReceiveIssueLine, len(p.header.ItemUsage))
	for idx, line := range p.header.ItemUsage {
		item, err := mapItems.Get(line.ItemID)
		if err != nil {
			return lines, fmt.Errorf("invalid: item: line %d, %s", idx, line.ItemID)
		}

		line.Item = *item
		line.InventDim.Calc()
		line.InventQty, _ = scmlogic.ConvertUnit(p.opt.Db, line.Qty, line.UnitID, item.DefaultUnitID)
		line.CostPerUnit = scmlogic.GetCostPerUnit(p.opt.Db, line.Item, line.InventDim, p.header.TrxDate)

		line.LineNo = idx + 1
		line.SourceJournalID = p.header.ID
		line.SourceTrxType = string(scmmodel.JournalWorkOrder)
		line.SourceType = tenantcoremodel.TrxModule(scmmodel.ModuleWorkorder)
		lines[idx] = line
	}

	return lines, nil
}

func (p *WorkOrderPosting) Calculate(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderJournal, lines []scmmodel.InventReceiveIssueLine) (
	*tenantcoremodel.PreviewReport, map[string][]orm.DataModel, float64, error) {
	var (
		err error
	)

	preview := tenantcoremodel.PreviewReport{}
	trxs := map[string][]orm.DataModel{}

	inventTrxs := []orm.DataModel{}
	for _, line := range lines {
		inventTrx := new(scmmodel.InventTrx)
		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.Status = scmmodel.ItemReserved
		inventTrx.TrxDate = *p.header.TrxDate

		inventTrx.Item = line.Item
		inventTrx.InventDim = *scmmodel.TernaryInventDimension(&line.InventDim, &p.header.InventDim)
		inventTrx.Qty = line.InventQty
		inventTrx.Status = scmmodel.ItemReserved
		inventTrx.AmountPhysical = line.CostPerUnit * inventTrx.Qty

		inventTrx.SourceType = scmmodel.ModuleWorkorder
		inventTrx.SourceJournalID = p.header.ID
		inventTrx.SourceLineNo = line.LineNo
		inventTrx.SourceTrxType = string(p.header.TrxType)

		p.inventTrxs = append(p.inventTrxs, inventTrx)
		inventTrxs = append(inventTrxs, inventTrx)
	}

	if len(inventTrxs) > 0 {
		trxs[inventTrxs[0].TableName()] = inventTrxs
	}

	return &preview, trxs, 0, err
}

func (p *WorkOrderPosting) Post(opt ficologic.PostingHubExecOpt, header *mfgmodel.WorkOrderJournal, lines []scmmodel.InventReceiveIssueLine, trxs map[string][]orm.DataModel) (string, error) {
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

	res, err = ficologic.PostModelSave(p.opt.Db, p.header, "WorkOrderVoucherNo", trxs)

	if err != nil {
		return res, err
	}

	if _, err = scmlogic.NewInventBalanceCalc(p.opt.Db).Sync(p.inventTrxs); err != nil {
		return res, fmt.Errorf("update balance: %s", err.Error())
	}

	return res, err
}

func (p *WorkOrderPosting) Approved() error {
	return nil
}

func (p *WorkOrderPosting) Rejected() error {
	return nil
}

func (p *WorkOrderPosting) GetAccount() string {
	return ""
}

func (p *WorkOrderPosting) SubmitNotification(pa *ficomodel.PostingApproval) error {
	return nil
}

func (p *WorkOrderPosting) ApproveRejectNotification(pa *ficomodel.PostingApproval, op ficologic.PostOp) error {
	return nil
}
