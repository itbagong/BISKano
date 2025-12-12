package ficologic

import (
	"time"

	"github.com/sebarcode/codekit"
)

type VendorClosing struct {
}

func (obj *VendorClosing) Close(companyID string, dt time.Time) (codekit.M, error) {
	res := codekit.M{}

	return res, nil
}
