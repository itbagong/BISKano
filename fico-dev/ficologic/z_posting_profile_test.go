package ficologic_test

import (
	"testing"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/samber/lo"
	"github.com/smartystreets/goconvey/convey"
)

func TestFindPostingProfile(t *testing.T) {

	convey.Convey("find posting profile", t, func() {
		convey.Convey("find for public", func() {
			ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "random-user", testCoID1)
			pps, err := new(ficologic.PostingProfileHandler).FindPostingProfile(ctx, "")
			ids := lo.Map(pps, func(pp *ficomodel.PostingProfile, index int) string {
				return pp.ID
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(pps), convey.ShouldBeGreaterThan, 0)
			convey.So(lo.Contains(ids, "DEF"), convey.ShouldBeTrue)
			convey.So(lo.Contains(ids, "WADP"), convey.ShouldBeFalse)
		})

		convey.Convey("find for limited user", func() {
			ctx := prepareCtxData(kaos.NewContextFromService(svc, nil), "finance-exec", testCoID1)
			pps, err := new(ficologic.PostingProfileHandler).FindPostingProfile(ctx, "")
			ids := lo.Map(pps, func(pp *ficomodel.PostingProfile, index int) string {
				return pp.ID
			})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(pps), convey.ShouldBeGreaterThan, 0)
			convey.So(lo.Contains(ids, "WADP"), convey.ShouldBeTrue)
		})
	})
}
