package ntsllogic

import (
	"fmt"

	"git.kanosolution.net/sebar/nitasalu/ntslconfig"
	"git.kanosolution.net/sebar/nitasalu/ntslmodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
)

func CalcPlan(db *datahub.Hub, ids ...string) ([]string, error) {
	groupids := []string{}
	return groupids, nil
}

func AddPlanToGroup(db *datahub.Hub, groupID string, planIDs ...string) ([]string, []string, error) {
	group, err := datahub.GetByID(db, new(ntslmodel.PlanGroup), groupID)
	if err != nil {
		group = new(ntslmodel.PlanGroup)
	}

	memberCount := group.MemberCount
	oks := []string{}
	fails := []string{}
	for _, planID := range planIDs {
		if lo.IndexOf(group.PlanID, planID) >= 0 {
			fails = append(fails, planID)
			continue
		}

		if memberCount > ntslconfig.Config.MaxMemberPerGroup {
			fails = append(fails, planID)
			continue
		}

		plan, err := datahub.GetByID(db, &ntslmodel.Plan{}, planID)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid: plan: %s: %s", planID, err)
		}
		memberCount += plan.MemberCount
		if memberCount > ntslconfig.Config.MaxMemberPerGroup {
			fails = append(fails, planID)
			continue
		}

		group.PlanID = append(group.PlanID, planID)
		group.MemberCount = memberCount

		oks = append(oks, planID)
	}

	if err := db.Save(group); err != nil {
		return nil, nil, err
	}

	return oks, fails, nil
}

func RemovePlanFromGroup(db *datahub.Hub, groupID string, planIDs ...string) error {
	group, err := datahub.GetByID(db, new(ntslmodel.PlanGroup), groupID)
	if err != nil {
		group = new(ntslmodel.PlanGroup)
	}

	memberCount := group.MemberCount
	for _, planID := range planIDs {
		if lo.IndexOf(group.PlanID, planID) == -1 {
			continue
		}

		plan, err := datahub.GetByID(db, &ntslmodel.Plan{}, planID)
		if err != nil {
			return fmt.Errorf("invalid: plan: %s: %s", planID, err)
		}
		memberCount -= plan.MemberCount
		group.PlanID = lo.Filter(group.PlanID, func(pid string, index int) bool {
			return pid != planID
		})
		group.MemberCount = memberCount
		if group.MemberCount < 0 {
			group.MemberCount = 0
		}
	}
	if err := db.Save(group); err != nil {
		return err
	}
	return nil
}
