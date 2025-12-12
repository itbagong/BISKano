package ficologic

import (
	"fmt"
	"strings"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/reflector"
)

// GetSubledgerFromMapRecord get subledger and its ledger account from a MapRecord
func GetSubledgerFromMapRecord[M orm.DataModel](
	id, ledgerFieldID string,
	mapSubledger *sebar.MapRecord[M],
	mapLedger *sebar.MapRecord[*tenantcoremodel.LedgerAccount]) (M, *tenantcoremodel.LedgerAccount, error) {
	res, err := mapSubledger.Get(id)
	if err != nil {
		return res, nil, err
	}
	if ledgerFieldID == "" {
		return res, nil, nil
	}

	var ledgerID string
	err = reflector.From(res).GetTo(ledgerFieldID, &ledgerID)
	if ledgerID == "" {
		return res, nil, fmt.Errorf("missing: ledger id value")
	}
	if err != nil {
		return res, nil, err
	}

	ledger, err := mapLedger.Get(ledgerID)
	if err != nil {
		return res, nil, err
	}
	return res, ledger, nil
}

func GetSubledgerFromMapRecordAndGroup[M orm.DataModel, G orm.DataModel](
	id, ledgerFieldID, groupFieldID string,
	mapSubledger *sebar.MapRecord[M],
	mapGroup *sebar.MapRecord[G],
	mapLedger *sebar.MapRecord[*tenantcoremodel.LedgerAccount]) (M, *tenantcoremodel.LedgerAccount, error) {
	res, err := mapSubledger.Get(id)
	if err != nil {
		return res, nil, err
	}
	if ledgerFieldID == "" {
		return res, nil, nil
	}

	var (
		ledgerID string
		groupID  string
	)
	reflector.From(res).GetTo(ledgerFieldID, &ledgerID)
	if ledgerID == "" {
		// get from group
		var group G
		reflector.From(res).GetTo(groupFieldID, &groupID)
		group, err = mapGroup.Get(groupID)
		if err != nil {
			return res, nil, fmt.Errorf("invalid: group: %s, %s", groupFieldID, err.Error())
		}

		err = reflector.From(group).GetTo(ledgerFieldID, &ledgerID)
	}
	if err != nil {
		return res, nil, err
	}

	ledger, err := mapLedger.Get(ledgerID)
	if err != nil {
		return res, nil, err
	}
	return res, ledger, nil
}

func getDimensionHash(dim tenantcoremodel.Dimension, dimNames ...string) string {
	values := make([]string, len(dimNames))
	for idx, name := range dimNames {
		values[idx] = dim.Get(name)
	}
	return strings.Join(values, "|")
}
