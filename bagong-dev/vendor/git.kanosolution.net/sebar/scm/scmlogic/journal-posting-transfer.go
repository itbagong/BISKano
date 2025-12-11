package scmlogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type InventTransferJournalPosting struct {
	header         *scmmodel.InventJournal
	journalType    *scmmodel.InventJournalType
	lines          []scmmodel.InventTrxLine
	inventTrxFroms []*scmmodel.InventTrx
	inventTrxTos   []*scmmodel.InventTrx
	preview        *tenantcoremodel.PreviewReport
	postingProfile *ficomodel.PostingProfile

	db        *datahub.Hub
	ev        kaos.EventHub
	journalID string
	userID    string

	items *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewInventTransferJournalPosting(db *datahub.Hub, ev kaos.EventHub, journalID, userID string) *InventTransferJournalPosting {
	p := new(InventTransferJournalPosting)
	p.db = db
	p.ev = ev
	p.journalID = journalID
	p.userID = userID

	p.items = sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Item))
	return p
}

func (p *InventTransferJournalPosting) ExtractHeader() error {
	var err error

	p.header, err = datahub.GetByID(p.db, new(scmmodel.InventJournal), p.journalID)
	if err != nil {
		return fmt.Errorf("invalid: journal: %s", p.journalID)
	}

	if p.header.JournalTypeID == "" {
		return fmt.Errorf("invalid: journal type: %s: %s", p.header.JournalTypeID, "is empty")
	}

	p.journalType, err = datahub.GetByID(p.db, new(scmmodel.InventJournalType), p.header.JournalTypeID)
	if err != nil {
		return fmt.Errorf("invalid: journal type: %s: %s", p.header.JournalTypeID, err.Error())
	}

	p.header.PostingProfileID = tenantcorelogic.TernaryString(p.header.PostingProfileID, p.journalType.PostingProfileID)
	if p.header.PostingProfileID == "" {
		return fmt.Errorf("invalid: posting profile: %s: %s", p.header.JournalTypeID, "is empty")
	}

	p.postingProfile, err = datahub.GetByID(p.db, new(ficomodel.PostingProfile), p.header.PostingProfileID)

	if p.header.PostingProfileID == "" {
		return fmt.Errorf("invalid: posting profile: %s: %s", p.header.JournalTypeID, err.Error())
	}

	return nil
}

func (p *InventTransferJournalPosting) ExtractLines() error {
	var err error
	p.lines = make([]scmmodel.InventTrxLine, len(p.header.Lines))

	for index, line := range p.header.Lines {
		lineTrx := scmmodel.InventTrxLine{}
		lineTrx.InventJournalLine = line
		lineTrx.Qty = line.Qty
		lineTrx.JournalID = p.header.ID
		lineTrx.TrxDate = p.header.TrxDate
		lineTrx.TrxType = p.header.TrxType

		lineTrx.Item, err = p.items.Get(line.ItemID)
		if err != nil {
			return fmt.Errorf("invalid item: %s: %s", line.ItemID, err.Error())
		}

		lineTrx.InventQty, err = ConvertUnit(p.db, line.Qty, line.UnitID, lineTrx.Item.DefaultUnitID)
		if err != nil {
			return fmt.Errorf("invalid: coversion: %s", err.Error())
		}

		// lineTrx.CostPerUnit = GetCostPerUnit(&lineTrx)
		lineTrx.CostPerUnit = GetCostPerUnit(p.db, *lineTrx.Item, lineTrx.InventDim, &p.header.TrxDate)
		p.lines[index] = lineTrx
	}

	return nil
}

func (p *InventTransferJournalPosting) Validate() error {
	if len(p.lines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	// check dimension

	// check inventory dimension

	return nil
}

func (p *InventTransferJournalPosting) Calculate() error {
	for index, line := range p.lines {
		inventTrx, err := trxLineToInventTrx(&line, p.db)
		if err != nil {
			return fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}

		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.SourceTrxType = string(scmmodel.InventIssuance)
		p.inventTrxFroms = append(p.inventTrxFroms, inventTrx)

		inventTrxTo := *inventTrx
		inventTrxTo.InventDim = p.header.InventDimTo
		inventTrxTo.SourceTrxType = string(scmmodel.InventReceive)
		p.inventTrxTos = append(p.inventTrxTos, &inventTrxTo)
	}

	return nil
}

func (p *InventTransferJournalPosting) PostingProfile() *ficomodel.PostingProfile {
	return p.postingProfile
}

func (p *InventTransferJournalPosting) Status() string {
	return string(p.header.Status)
}

func (p *InventTransferJournalPosting) Submit() (*ficomodel.PostingApproval, error) {
	pa, isNew, err := ficologic.GetOrCreatePostingApproval(p.db, p.userID, p.header.CompanyID, string(scmmodel.ModuleInventory), p.header.ID, *p.postingProfile, p.header.Dimension, true, true, lineTrxsToFicoLines(p.lines), "", p.header.Text, p.header.TrxDate, 0)

	if p.postingProfile.NeedApproval {
		if err != nil {
			return nil, fmt.Errorf("create approval: %s", err.Error())
		}

		if !isNew {
			return nil, fmt.Errorf("duplicate: approval: %s, %s", ficomodel.SubledgerAccounting, p.header.ID)
		}

		p.header.Status = "SUBMITTED"
		p.db.Save(p.header)
	} else {
		p.markAsReady(pa)
	}

	return pa, nil
}

func (p *InventTransferJournalPosting) Approve(op string, txt string) (string, error) {
	pa, err := ficologic.GetPostingApprovalBySource(p.db, p.header.CompanyID, string(scmmodel.ModuleInventory), p.header.ID, true)
	if err != nil {
		return "", fmt.Errorf("posting approval: %s", err.Error())
	}

	if err = pa.UpdateApproval(p.db, p.userID, op, txt); err != nil {
		return pa.Status, fmt.Errorf("posting approval: %s", err.Error())
	}
	if err = p.db.Save(pa); err != nil {
		return pa.Status, fmt.Errorf("posting approval save: %s", err.Error())
	}
	switch pa.Status {
	case "REJECTED":
		p.header.Status = ficomodel.JournalStatus(pa.Status)
		p.db.Save(p.header)

	case "APPROVED":
		if err = p.markAsReady(pa); err != nil {
			return pa.Status, err
		}
	}
	return pa.Status, nil
}

func (p *InventTransferJournalPosting) Post() error {
	p.updateItemBalance(p.inventTrxTos)
	p.updateItemBalance(p.inventTrxFroms)

	p.header.Status = ficomodel.JournalStatusPosted
	p.db.Update(p.header, "Status")
	return nil
}

func (p *InventTransferJournalPosting) Details() (*tenantcoremodel.PreviewReport, []scmmodel.InventTrxLine) {
	return p.preview, p.lines
}

func (p *InventTransferJournalPosting) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *InventTransferJournalPosting) Transactions(name string) []orm.DataModel {
	return ficologic.ToDataModels(p.inventTrxFroms)
}

func (p *InventTransferJournalPosting) updateItemBalance(inventTrx []*scmmodel.InventTrx) error {
	for _, trx := range inventTrx {
		trx.Status = lo.Ternary(trx.Qty > 0, scmmodel.ItemPlanned, scmmodel.ItemReserved)
		p.db.Save(trx)

		balanceCalc := NewInventBalanceCalc(p.db)
		if _, err := balanceCalc.Update(&scmmodel.ItemBalance{
			CompanyID:   p.header.CompanyID,
			ItemID:      trx.Item.ID,
			InventDim:   trx.InventDim,
			Qty:         0,
			QtyReserved: lo.Ternary(trx.Qty < 0, trx.Qty, 0),
			QtyPlanned:  lo.Ternary(trx.Qty > 0, trx.Qty, 0),
		}); err != nil {
			return fmt.Errorf("update balance: %s", err.Error())
		}
	}

	return nil
}

func (p InventTransferJournalPosting) markAsReady(pa *ficomodel.PostingApproval) error {
	p.header.Status = "READY"
	p.db.Save(p.header)

	if p.postingProfile.DirectPosting {
		if postError := p.Post(); postError != nil {
			return postError
		}
	}

	return nil
}
