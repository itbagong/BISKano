package scmlogic

import (
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

type InventIssuePosting struct {
	header         *scmmodel.InventReceiveIssueJournal
	lines          []scmmodel.InventReceiveIssueLine
	inventTrxs     []*scmmodel.InventTrx
	preview        *tenantcoremodel.PreviewReport
	postingProfile *ficomodel.PostingProfile

	db        *datahub.Hub
	ev        kaos.EventHub
	journalID string
	userID    string

	items *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewInventIssuancePosting(db *datahub.Hub, ev kaos.EventHub, journalID, userID string) *InventIssuePosting {
	p := new(InventIssuePosting)
	p.db = db
	p.ev = ev
	p.journalID = journalID
	p.userID = userID

	p.items = sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Item))
	return p
}

func (p *InventIssuePosting) ExtractHeader() error {
	var err error
	p.header, err = datahub.GetByID(p.db, new(scmmodel.InventReceiveIssueJournal), p.journalID)
	if err != nil {
		return fmt.Errorf("Invalid journal id")
	}

	p.postingProfile, err = datahub.GetByID(p.db, new(ficomodel.PostingProfile), p.header.PostingProfileID)
	if err != nil {
		return fmt.Errorf("Invalid: posting profile: %s", p.header.PostingProfileID)
	}

	return nil
}

func (p *InventIssuePosting) ExtractLines() error {
	p.lines = lo.Map(p.header.Lines, func(line scmmodel.InventReceiveIssueLine, index int) scmmodel.InventReceiveIssueLine {
		item, err := p.items.Get(line.ItemID)
		if err == nil {
			line.Item = *item
		}

		line.CostPerUnit = GetCostPerUnit(p.db, line.Item, line.InventDim, &p.header.TrxDate)
		return line
	})

	return nil
}

func (p *InventIssuePosting) Validate() error {
	if len(p.lines) == 0 {
		return fmt.Errorf("missing: lines")
	}

	groupedLines := lo.GroupBy(p.lines, func(l scmmodel.InventReceiveIssueLine) string {
		return fmt.Sprintf("%s|%s", l.ItemID, l.InventDim.InventDimID)
	})

	for _, gls := range groupedLines {
		bal, _ := NewInventBalanceCalc(p.db).Get(&InventBalanceCalcOpts{
			CompanyID: p.header.CompanyID,
			ItemID:    []string{gls[0].ItemID},
			InventDim: scmmodel.InventDimension{InventDimID: gls[0].InventDim.InventDimID},
		})

		qtyReserved := lo.SumBy(bal, func(b *scmmodel.ItemBalance) float64 {
			return b.QtyReserved
		})

		qtyConfirmed := lo.SumBy(gls, func(g scmmodel.InventReceiveIssueLine) float64 {
			return g.Qty
		})

		if moreThan(qtyConfirmed, qtyReserved, true) {
			return fmt.Errorf("over qty: %s: issue %2.f, reserved %2.f", gls[0].ItemID, qtyConfirmed, qtyReserved)
		}
	}

	return nil
}

func (p *InventIssuePosting) Calculate() error {
	for index, line := range p.lines {
		inventTrx, err := receiveIssueLineToTrx(p.db, p.header, line)
		if err != nil {
			return fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
		}

		inventTrx.CompanyID = p.header.CompanyID
		inventTrx.Status = scmmodel.ItemConfirmed
		inventTrx.TrxDate = p.header.TrxDate
		p.inventTrxs = append(p.inventTrxs, inventTrx)
	}

	return nil
}

func (p *InventIssuePosting) PostingProfile() *ficomodel.PostingProfile {
	return p.postingProfile
}

func (p *InventIssuePosting) Status() string {
	return string(p.header.Status)
}

func (p *InventIssuePosting) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *InventIssuePosting) Transactions(name string) []orm.DataModel {
	return ficologic.ToDataModels(p.inventTrxs)
}

func (p *InventIssuePosting) markAsReady(pa *ficomodel.PostingApproval) error {
	p.header.Status = "READY"
	p.db.Save(p.header)

	if p.postingProfile.DirectPosting {
		if postError := p.Post(); postError != nil {
			return postError
		}
	}

	return nil
}

func (p *InventIssuePosting) Submit() (*ficomodel.PostingApproval, error) {
	pa, isNew, err := ficologic.GetOrCreatePostingApproval(p.db, p.userID, p.header.CompanyID, string(scmmodel.ModuleInventory), p.header.ID, *p.postingProfile, p.header.Dimension, true, true, receiveIssuelineToficoLines(p.lines), "", p.header.Name, p.header.TrxDate, 0)
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
		if err = p.markAsReady(pa); err != nil {
			return pa, err
		}
	}

	return pa, nil
}

func (p *InventIssuePosting) Approve(op string, txt string) (string, error) {
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

func (p *InventIssuePosting) Post() error {
	p.header.Status = ficomodel.JournalStatusPosted
	p.db.Update(p.header, "Status")

	allTrxs := []*scmmodel.InventTrx{}
	splitter := NewInventSplit(p.db)

	for _, trx := range p.inventTrxs {
		_, splitTrxs, err := splitter.SetOpts(&InventSplitOpts{
			SplitType:       SplitBySource,
			CompanyID:       trx.CompanyID,
			SourceType:      string(trx.SourceType),
			SourceJournalID: trx.SourceJournalID,
			SourceLineNo:    trx.SourceLineNo,
			SourceStatus:    string(scmmodel.ItemReserved),
		}).Split(trx.Qty, string(scmmodel.ItemConfirmed))
		if err != nil {
			return err
		}

		allTrxs = append(allTrxs, splitTrxs...)
	}

	balEng := NewInventBalanceCalc(p.db)
	if _, err := balEng.Sync(allTrxs); err != nil {
		return err
	}

	//apa ada external logic setelah post?
	return nil
}

func (p *InventIssuePosting) Details() (*tenantcoremodel.PreviewReport, []scmmodel.InventTrxLine) {
	panic("not implemented") // TODO: Implement
}

func issueIssueLineToTrx(db *datahub.Hub, header *scmmodel.InventReceiveIssueJournal, line scmmodel.InventReceiveIssueLine) (*scmmodel.InventTrx, error) {
	var err error
	trx := new(scmmodel.InventTrx)
	trx.CompanyID = header.CompanyID
	trx.Item = line.Item
	trx.SKU = line.SKU
	trx.Dimension = line.Dimension
	trx.InventDim = line.InventDim
	trx.Qty, err = ConvertUnit(db, line.Qty, line.UnitID, line.Item.DefaultUnitID) // pada proses issue, apa perlu convert unit? sementara tak buat sama
	if err != nil {
		return nil, fmt.Errorf("convertunit: %s", err.Error())
	}
	trx.TrxQty = line.Qty
	trx.TrxUnitID = line.UnitID

	trx.SourceType = line.SourceType
	trx.SourceJournalID = line.SourceJournalID
	trx.SourceTrxType = string(header.TrxType)
	trx.SourceLineNo = line.SourceLineNo

	trx.TrxDate = header.TrxDate
	return trx, nil
}
