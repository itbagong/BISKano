package ficologic

import (
	"errors"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
)

type VendorTransactionAPI struct {
}

func (obj *VendorTransactionAPI) Create(ctx *kaos.Context, payload *ficomodel.VendorJournal) (*ficomodel.VendorJournal, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missingDBConn")
	}

	h.Save(payload)

	return payload, nil
}
