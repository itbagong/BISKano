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
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type InventReceivePosting struct {
	header *scmmodel.InventReceiveIssueJournal
	//journalType    *scmmodel.InventJournalType
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

func NewInventReceivePosting(db *datahub.Hub, ev kaos.EventHub, journalID, userID string) *InventReceivePosting {
	p := new(InventReceivePosting)
	p.db = db
	p.ev = ev
	p.journalID = journalID
	p.userID = userID

	p.items = sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Item))
	return p
}

func (p *InventReceivePosting) ExtractHeader() error {
	var err error
	p.header, err = datahub.GetByID(p.db, new(scmmodel.InventReceiveIssueJournal), p.journalID)
	if err != nil {
		return fmt.Errorf("invalid journal id")
	}

	p.postingProfile, err = datahub.GetByID(p.db, new(ficomodel.PostingProfile), p.header.PostingProfileID)
	if err != nil {
		return fmt.Errorf("invalid: posting profile: %s", p.header.PostingProfileID)
	}

	return nil
}

func (p *InventReceivePosting) ExtractLines() error {
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

func (p *InventReceivePosting) Validate() error {
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

		qtyPlanned := lo.SumBy(bal, func(b *scmmodel.ItemBalance) float64 {
			return b.QtyPlanned
		})

		qtyConfirmed := lo.SumBy(gls, func(g scmmodel.InventReceiveIssueLine) float64 {
			return g.Qty
		})

		if moreThan(qtyConfirmed, qtyPlanned, true) {
			return fmt.Errorf("over qty: %s: receive %.2f, planned %.2f", gls[0].ItemID, qtyConfirmed, qtyPlanned)
		}
	}

	return nil
}

func (p *InventReceivePosting) Calculate() error {
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

func (p *InventReceivePosting) PostingProfile() *ficomodel.PostingProfile {
	return p.postingProfile
}

func (p *InventReceivePosting) Status() string {
	return string(p.header.Status)
}

func (p *InventReceivePosting) markAsReady(pa *ficomodel.PostingApproval) error {
	p.header.Status = "READY"
	p.db.Save(p.header)

	if p.postingProfile.DirectPosting {
		if postError := p.Post(); postError != nil {
			return postError
		}
	}

	return nil
}

func (p *InventReceivePosting) Submit() (*ficomodel.PostingApproval, error) {
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

func (p *InventReceivePosting) Approve(op string, txt string) (string, error) {
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

func (p *InventReceivePosting) Post() error {
	p.header.Status = ficomodel.JournalStatusPosted
	p.db.Update(p.header, "Status")

	allTrxs := []*scmmodel.InventTrx{}
	spliter := NewInventSplit(p.db)
	for _, trx := range p.inventTrxs {
		_, splitTrxs, err := spliter.SetOpts(&InventSplitOpts{
			SplitType:       SplitBySource,
			CompanyID:       trx.CompanyID,
			SourceType:      string(trx.SourceType),
			SourceJournalID: trx.SourceJournalID,
			SourceLineNo:    trx.SourceLineNo,
			SourceStatus:    string(scmmodel.ItemPlanned),
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

	scmconfig.Config.EventHub().Publish(
		fmt.Sprintf(scmconfig.Config.PostingTopic, "inventreceive", "post"),
		p.header,
		nil, &kaos.PublishOpts{Headers: codekit.M{"CompanyID": p.header.CompanyID, sebar.CtxJWTReferenceID: p.userID}},
	)

	return nil
}

func (p *InventReceivePosting) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *InventReceivePosting) Transactions(name string) []orm.DataModel {
	return ficologic.ToDataModels(p.inventTrxs)
}

func receiveIssuelineToficoLines(lineTrxs []scmmodel.InventReceiveIssueLine) []ficomodel.JournalLine {
	return lo.Map(lineTrxs, func(line scmmodel.InventReceiveIssueLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(scmmodel.ModuleInventory, line.ItemID)
		jl.Amount = line.CostPerUnit * line.InventQty
		return *jl
	})
}
