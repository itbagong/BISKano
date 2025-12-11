package scmlogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type InventJournalPosting__ struct {
	header         *scmmodel.InventJournal
	journalType    *scmmodel.InventJournalType
	lines          []scmmodel.InventTrxLine
	inventTrxs     []*scmmodel.InventTrx
	preview        *tenantcoremodel.PreviewReport
	postingProfile *ficomodel.PostingProfile

	db        *datahub.Hub
	ev        kaos.EventHub
	journalID string
	userID    string

	items *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewInventJournalPosting__(db *datahub.Hub, ev kaos.EventHub, journalID, userID string) *InventJournalPosting__ {
	p := new(InventJournalPosting__)
	p.db = db
	p.ev = ev
	p.journalID = journalID
	p.userID = userID

	p.items = sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Item))
	return p
}

func (p *InventJournalPosting__) GetAccount() string {
	return p.header.Text
}

func (p *InventJournalPosting__) ExtractHeader() error {
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

func (p *InventJournalPosting__) ExtractLines() error {
	var err error
	p.lines = make([]scmmodel.InventTrxLine, len(p.header.Lines))

	for index, line := range p.header.Lines {
		lineTrx := scmmodel.InventTrxLine{}
		lineTrx.InventJournalLine = line
		lineTrx.JournalID = p.header.ID
		lineTrx.TrxDate = p.header.TrxDate
		lineTrx.TrxType = p.header.TrxType

		lineTrx.Item, err = p.items.Get(line.ItemID)
		if err != nil {
			return fmt.Errorf("invalid item: %s: %s", line.ItemID, err.Error())
		}

		lineTrx.InventDim = *NewInventDimHelper(InventDimHelperOpt{DB: p.db, SKU: line.SKU}).TernaryInventDimension(&lineTrx.InventDim, &p.header.InventDim)

		lineTrx.InventQty, err = ConvertUnit(p.db, line.Qty, line.UnitID, lineTrx.Item.DefaultUnitID)
		if err != nil {
			return fmt.Errorf("invalid: coversion: %s", err.Error())
		}
		lineTrx.CostPerUnit = GetCostPerUnit(p.db, *lineTrx.Item, lineTrx.InventDim, &p.header.TrxDate)
		p.lines[index] = lineTrx
	}

	return nil
}

func (p *InventJournalPosting__) Validate() error {
	if len(p.lines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	for _, line := range p.lines {
		switch p.header.TrxType {
		case scmmodel.JournalMovementOut:
			// validate against qty in item balance
			ibCalc := &InventBalanceCalcOpts{
				CompanyID: p.header.CompanyID,
				ItemID:    []string{line.ItemID},
				InventDim: line.InventDim,
				// BalanceDate: ???,
			}

			if ibs, err := NewInventBalanceCalc(p.db).Get(ibCalc); err != nil {
				return fmt.Errorf("no item balance found | param: %v", codekit.JsonStringIndent(ibCalc, "\t"))
			} else if len(ibs) > 0 && line.Qty > ibs[0].QtyAvail {
				return fmt.Errorf("Qty Available not enough, balance: %v", ibs[0].QtyAvail)
			}
		}
	}

	// check dimension

	// check inventory dimension

	return nil
}

func (p *InventJournalPosting__) Calculate() error {
	for index, line := range p.lines {
		inventTrx, err := trxLineToInventTrx(&line, p.db)
		if err != nil {
			return fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}
		inventTrx.CompanyID = p.header.CompanyID
		p.inventTrxs = append(p.inventTrxs, inventTrx)
	}
	return nil
}

func (p *InventJournalPosting__) PostingProfile() *ficomodel.PostingProfile {
	return p.postingProfile
}

func (p *InventJournalPosting__) Status() string {
	return string(p.header.Status)
}

func (p *InventJournalPosting__) Submit() (*ficomodel.PostingApproval, error) {
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

func (p *InventJournalPosting__) Approve(op string, txt string) (string, error) {
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

func (p *InventJournalPosting__) Post() error {
	for _, trx := range p.inventTrxs {
		trx.CompanyID = p.header.CompanyID
		trx.Status = lo.Ternary(trx.Qty > 0, scmmodel.ItemPlanned, scmmodel.ItemReserved)
		p.db.Save(trx)
	}

	if _, err := NewInventBalanceCalc(p.db).Sync(p.inventTrxs); err != nil {
		return fmt.Errorf("update balance: %s", err.Error())
	}

	p.header.Status = ficomodel.JournalStatusPosted
	p.db.Update(p.header, "Status")

	scmconfig.Config.EventHub().Publish(
		fmt.Sprintf(scmconfig.Config.PostingTopic, "inventjournal", "post"),
		p.header,
		nil, &kaos.PublishOpts{Headers: codekit.M{"CompanyID": p.header.CompanyID, sebar.CtxJWTReferenceID: p.userID}},
	)

	return nil
}

func (p *InventJournalPosting__) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *InventJournalPosting__) Transactions(name string) []orm.DataModel {
	return ficologic.ToDataModels(p.inventTrxs)
}

func (p *InventJournalPosting__) markAsReady(pa *ficomodel.PostingApproval) error {
	p.header.Status = "READY"
	p.db.Save(p.header)

	if p.postingProfile.DirectPosting {
		if postError := p.Post(); postError != nil {
			return postError
		}
	}

	return nil
}

func lineTrxsToFicoLines(lineTrxs []scmmodel.InventTrxLine) []ficomodel.JournalLine {
	return lo.Map(lineTrxs, func(line scmmodel.InventTrxLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}
