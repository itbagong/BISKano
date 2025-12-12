package ntsllogic_test

import (
	"fmt"
	"testing"
	"time"

	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/nitasalu/ntsllogic"
	"git.kanosolution.net/sebar/nitasalu/ntslmodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"github.com/smartystreets/goconvey/convey"
)

func TestPlan(t *testing.T) {
	convey.Convey("make 500 plan", t, func() {
		startDate := codekit.DateOnly(time.Now())
		db, _ := svc.GetDataHub("default", "")
		convey.So(db, convey.ShouldNotBeNil)

		plans := lo.RepeatBy(2000, func(index int) *ntslmodel.Plan {
			empID := fmt.Sprintf("user-%4d", codekit.RandInt(1000)-1)
			expStartDate := startDate.AddDate(0, 0, codekit.RandInt(200))
			expEndDate := expStartDate.AddDate(0, 0, 20)
			randCity := []string{"JKT", "SBY", "SAM"}[codekit.RandInt(3)]
			return &ntslmodel.Plan{
				UserID:           empID,
				MemberCount:      codekit.RandInt(5),
				ExpectedDateFrom: &expStartDate,
				ExpectedDateTo:   &expEndDate,
				ExpectedCities:   []string{randCity},
			}
		})

		var planErr error
		for _, plan := range plans {
			addr := "/v1/plan/insert"
			ctx := kaos.NewContextFromService(svc, svc.GetRoute(addr))
			planErr := svc.CallTo(addr, plan, ctx, plan)
			if planErr != nil {
				break
			}
		}
		convey.So(planErr, convey.ShouldBeNil)

		convey.Convey("grouping", func() {
			gids, err := ntsllogic.CalcPlan(db, lo.Map(plans, func(p *ntslmodel.Plan, i int) string {
				return p.ID
			})...)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(gids), convey.ShouldBeGreaterThan, 0)

			convey.Convey("validate", func() {
				var (
					pgHasErr error
				)
				groups := lo.Map(gids, func(id string, index int) *ntslmodel.PlanGroup {
					pg := new(ntslmodel.PlanGroup)
					addr := "/v1/plangroup/get"
					ctx := kaos.NewContextFromService(svc, svc.GetRoute(addr))
					if err := svc.CallTo(addr, &pg, ctx, []interface{}{id}); err != nil {
						pgHasErr = err
					}
					return pg
				})
				convey.So(pgHasErr, convey.ShouldBeNil)
				convey.So(len(groups), convey.ShouldBeGreaterThan, 0)
			})
		})
	})
}
