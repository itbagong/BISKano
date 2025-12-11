package bagonglogic

import (
	"fmt"
	"strconv"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
)

type AssetMovementPosting struct {
	header *bagongmodel.AssetMovement
	//journalType    *bagongmodel.InventJournalType
	lines          []bagongmodel.AssetMovementLine
	inventTrxs     []*bagongmodel.AssetMovement
	preview        *tenantcoremodel.PreviewReport
	postingProfile *ficomodel.PostingProfile

	db        *datahub.Hub
	ev        kaos.EventHub
	journalID string
	userID    string
	companyID string

	// items *sebar.MapRecord[*tenantcoremodel.Item]
}

func NewAssetMovementPosting(db *datahub.Hub, ev kaos.EventHub, journalID, userID, companyID string) *AssetMovementPosting {
	p := new(AssetMovementPosting)
	p.db = db
	p.ev = ev
	p.journalID = journalID
	p.userID = userID
	p.companyID = companyID

	// p.items = sebar.NewMapRecordWithORM(db, new(tenantcoremodel.Item))
	return p
}

func (p *AssetMovementPosting) ExtractHeader() error {
	var err error
	p.header, err = datahub.GetByID(p.db, new(bagongmodel.AssetMovement), p.journalID)
	if err != nil {
		return fmt.Errorf("invalid journal id")
	}

	jt, err := datahub.GetByID(p.db, new(ficomodel.AssetJournalType), p.header.JournalTypeID)
	if err != nil {
		return fmt.Errorf("missing: journal type: %s: %s", p.header.JournalTypeID, err.Error())
	}

	p.postingProfile, err = datahub.GetByID(p.db, new(ficomodel.PostingProfile), jt.PostingProfileID)
	if err != nil {
		return fmt.Errorf("invalid: posting profile: %s", p.header.PostingProfileID)
	}

	return nil
}

func (p *AssetMovementPosting) ExtractLines() error {
	p.lines = lo.Map(p.header.Lines, func(line bagongmodel.AssetMovementLine, index int) bagongmodel.AssetMovementLine {
		return line
	})
	return nil
}

func (p *AssetMovementPosting) Validate() error {
	// if len(p.lines) == 0 {
	// 	return fmt.Errorf("missing: lines")
	// }

	// groupedLines := lo.GroupBy(p.lines, func(l scmmodel.InventReceiveIssueLine) string {
	// 	return fmt.Sprintf("%s|%s", l.ItemID, l.InventDim.InventDimID)
	// })

	// for _, gls := range groupedLines {
	// 	bal, _ := NewInventBalanceCalc(p.db).Get(&InventBalanceCalcOpts{
	// 		CompanyID: p.header.CompanyID,
	// 		ItemID:    []string{gls[0].ItemID},
	// 		InventDim: scmmodel.InventDimension{InventDimID: gls[0].InventDim.InventDimID},
	// 	})

	// 	qtyPlanned := lo.SumBy(bal, func(b *scmmodel.ItemBalance) float64 {
	// 		return b.QtyPlanned
	// 	})

	// 	qtyConfirmed := lo.SumBy(gls, func(g scmmodel.InventReceiveIssueLine) float64 {
	// 		return g.Qty
	// 	})

	// 	if moreThan(qtyConfirmed, qtyPlanned, true) {
	// 		return fmt.Errorf("over qty: %s: receive %.2f, planned %.2f", gls[0].ItemID, qtyConfirmed, qtyPlanned)
	// 	}
	// }

	return nil
}

func (p *AssetMovementPosting) Calculate() error {
	// for index, line := range p.lines {
	// 	inventTrx, err := receiveIssueLineToTrx(p.db, p.header, line)
	// 	if err != nil {
	// 		return fmt.Errorf("create inventory transaction: line %d: %s", index, err.Error())
	// 	}
	// 	inventTrx.CompanyID = p.header.CompanyID
	// 	inventTrx.Status = scmmodel.ItemConfirmed
	// 	inventTrx.TrxDate = p.header.TrxDate
	// 	p.inventTrxs = append(p.inventTrxs, inventTrx)
	// }
	return nil
}

func (p *AssetMovementPosting) PostingProfile() *ficomodel.PostingProfile {
	return p.postingProfile
}

func (p *AssetMovementPosting) Status() string {
	return string(p.header.Status)
}

func (p *AssetMovementPosting) markAsReady(pa *ficomodel.PostingApproval) error {
	p.header.Status = "READY"
	p.db.Save(p.header)

	if p.postingProfile.DirectPosting {
		if postError := p.Post(); postError != nil {
			return postError
		}
	}

	return nil
}

func (p *AssetMovementPosting) Submit() (*ficomodel.PostingApproval, error) {
	pa, isNew, err := ficologic.GetOrCreatePostingApproval(p.db, p.userID, p.companyID, string(bagongmodel.ModuleAssetmovement), p.header.ID, *p.postingProfile, nil, true, true, receiveIssuelineToficoLines(p.lines), "", p.header.ID, p.header.TrxDate, 0)
	if p.postingProfile.NeedApproval {
		if err != nil {
			return nil, fmt.Errorf("create approval: %s", err.Error())
		}
		if !isNew {
			return nil, fmt.Errorf("duplicate: approval: %s, %s", ficomodel.SubledgerAccounting, p.header.ID)
		}
		p.header.Status = "SUBMITTED"
		p.db.Save(p.header)
	}

	return pa, nil
}

func (p *AssetMovementPosting) Approve(op string, txt string) (string, error) {
	pa, err := ficologic.GetPostingApprovalBySource(p.db, p.companyID, string(bagongmodel.ModuleAssetmovement), p.header.ID, true)
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
		p.header.Status = string(ficomodel.JournalStatus(pa.Status))
		p.db.Save(p.header)

	case "APPROVED":
		if err = p.markAsReady(pa); err != nil {
			return pa.Status, err
		}
	}
	return pa.Status, nil
}

func (p *AssetMovementPosting) Post() error {
	p.header.Status = string(ficomodel.JournalStatusPosted)
	p.db.Update(p.header, "Status")

	// add to asset and unit calendar
	err := PostAssets(p.db, p.journalID)
	if err != nil {
		return err
	}

	return nil
}

func (p *AssetMovementPosting) Preview() *tenantcoremodel.PreviewReport {
	return p.preview
}

func (p *AssetMovementPosting) Transactions(name string) []orm.DataModel {
	return ficologic.ToDataModels(p.inventTrxs)
}

func receiveIssuelineToficoLines(lineTrxs []bagongmodel.AssetMovementLine) []ficomodel.JournalLine {
	return lo.Map(lineTrxs, func(line bagongmodel.AssetMovementLine, index int) ficomodel.JournalLine {
		jl, _ := reflector.CopyAttributes(line, new(ficomodel.JournalLine))
		// TODO: get cost and assign yang lain-lain
		jl.Account = ficomodel.NewSubAccount(bagongmodel.ModuleAssetmovement, strconv.Itoa(line.LineNo))
		return *jl
	})
}
