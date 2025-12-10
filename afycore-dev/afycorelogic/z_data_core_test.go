package afycorelogic_test

import (
	"testing"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/afycore/afycoremodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/smartystreets/goconvey/convey"
)

func TestCoreData(t *testing.T) {
	convey.Convey("validate core data", t, func() {
		ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "user-01", testCoID1)
		db := sebar.GetTenantDBFromContext(ctx)

		locs, err := datahub.Find(db, new(afycoremodel.MedicalLocation), nil)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(locs), convey.ShouldBeGreaterThan, 0)
	})
}
