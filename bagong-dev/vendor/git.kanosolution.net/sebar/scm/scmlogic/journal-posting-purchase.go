package scmlogic

import (
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type PurchasePosting struct {
	this ficologic.JournalPosting

	db *datahub.Hub
	ev kaos.EventHub
}

func NewPurchasePosting(db *datahub.Hub, ev kaos.EventHub, id string) *PurchasePosting {
	p := new(PurchasePosting)
	p.db = db
	p.ev = ev
	return p
}

func (p *PurchasePosting) SetThis(jp ficologic.JournalPosting) ficologic.JournalPosting {
	p.this = jp
	return p
}

func (p *PurchasePosting) This() ficologic.JournalPosting {
	if p.this == nil {
		return p
	}
	return p.this
}

func (p *PurchasePosting) ExtractHeader() error {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) ExtractLines() error {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Validate() error {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Calculate() error {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) PostingProfile() *ficomodel.PostingProfile {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Status() string {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Submit() (*ficomodel.PostingApproval, error) {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Approve(_ string, _ string) (string, error) {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Post() error {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Preview() *tenantcoremodel.PreviewReport {
	panic("not implemented") // TODO: Implement
}

func (p *PurchasePosting) Transactions(name string) []orm.DataModel {
	return []orm.DataModel{}
}
