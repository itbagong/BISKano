package ficologic

import (
	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

func FindLegderTrxBySource(db *datahub.Hub, sourceModule tenantcoremodel.TrxModule, sourceID string, sourceLine int, getFirstOnly bool) []*ficomodel.LedgerTransaction {
	where := dbflex.Eqs("SourceType", sourceModule, "SourceJournalID", sourceID)
	if sourceLine > 0 {
		where = dbflex.And(where, dbflex.Eq("SourceLineNo", sourceLine))
	}

	parm := dbflex.NewQueryParam().SetWhere(where)
	if getFirstOnly {
		parm.SetTake(1)
	}

	res, _ := datahub.Find(db, new(ficomodel.LedgerTransaction), parm)
	return res
}
