package sebar

import (
	"time"

	"github.com/ariefdarmawan/datahub"
	"github.com/kanoteknologi/kns"
)

func GetSequenceNo(h *datahub.Hub, seqid string, dt *time.Time, reserve bool) (string, error) {
	mgrSN := kns.NewManager(h)
	num, err := mgrSN.GetNo(seqid, dt, reserve)
	if err != nil {
		return "", err
	}
	return mgrSN.Format(num), nil
}
