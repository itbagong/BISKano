package tenantcorelogic

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

func GetPreviewBySource(db *datahub.Hub, sourceType, sourceID, voucherNo, name string) (*tenantcoremodel.Preview, error) {
	res, err := datahub.GetByFilter(db, new(tenantcoremodel.Preview), dbflex.Eqs("SourceType", sourceType, "SourceJournalID", sourceID,
		"VoucherNo", voucherNo, "Name", name))
	if err != nil {
		res.SourceType = sourceType
		res.SourceJournalID = sourceID
		res.VoucherNo = voucherNo
		res.Name = name
	}
	return res, err
}

func FindPreviewBySource(db *datahub.Hub, sourceType, sourceID, name string) ([]*tenantcoremodel.Preview, error) {
	prevs, err := datahub.Find(db, new(tenantcoremodel.Preview),
		dbflex.NewQueryParam().SetWhere(dbflex.Eqs("SourceType", sourceType, "SourceJournalID", sourceID,
			"Name", name)).SetSort("-Created"))
	return prevs, err
}
