package ntsllogic_test

import (
	"testing"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/nitasalu/ntslmodel"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestAPICity(t *testing.T) {
	convey.Convey("test api city", t, func() {
		convey.Convey("kediri", func() {
			addr := "/v1/city/insert"
			kdr := &ntslmodel.City{ID: "KDR", Name: "Kediri"}
			ctx := kaos.NewContextFromService(svc, svc.GetRoute(addr))
			err := svc.CallTo(addr, kdr, ctx, kdr)
			convey.So(err, convey.ShouldBeNil)
			convey.So(kdr.Created.Year(), convey.ShouldBeGreaterThan, 2023)
		})

		convey.Convey("sby - negative test", func() {
			addr := "/v1/city/insert"
			sby := &ntslmodel.City{ID: "SBY", Name: "Surabaya"}
			ctx := kaos.NewContextFromService(svc, svc.GetRoute(addr))
			err := svc.CallTo(addr, sby, ctx, sby)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(sby.Created.Year(), convey.ShouldBeGreaterThan, 2023)
		})
	})
}

func TestUser(t *testing.T) {
	convey.Convey("get users", t, func() {
		addr := "/v1/profile/gets"
		qp := dbflex.NewQueryParam().SetSort("-_id").SetTake(1000)
		ctx := kaos.NewContextFromService(svc, svc.GetRoute(addr))
		res := codekit.M{}
		err := svc.CallTo(addr, &res, ctx, qp)
		convey.So(err, convey.ShouldBeNil)

		profiles := *(res.Get("data").(*[]ntslmodel.Profile))
		convey.So(len(profiles), convey.ShouldEqual, 1000)
		convey.So(profiles[0].Name, convey.ShouldEqual, "Name 0999")
	})
}
