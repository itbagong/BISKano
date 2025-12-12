package tenantcoremodel

import (
	"testing"
	"time"
)

func TestNumberSequence(t *testing.T) {
	seq := NumberSequence{}
	seq.OutFormat = "O%04d"
	seq.UseDate = ""
	target := "O0001"
	if target != seq.Format(1, nil) {
		t.FailNow()
	}
	seq.UseDate = "2006"
	seq.OutFormat = "O%s%04d"
	inpTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)
	target = "O20230001"
	if target != seq.Format(1, &inpTime) {
		t.FailNow()
	}

}
