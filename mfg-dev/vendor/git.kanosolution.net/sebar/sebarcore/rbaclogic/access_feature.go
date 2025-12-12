package rbaclogic

// . "git.kanosolution.net/sebar/sebarcore/rbacmodel"

/*
type RBACMap map[string]map[string][]RoleFeaturePure

func LoadMapToMemory(h *datahub.Hub, log *logger.LogEngine, loginIDs ...string) (RBACMap, error) {
	res := make(RBACMap)
	users := []User{}
	roleFeatures := []RoleFeature{}
	roleMembers := []RoleMember{}

	h.Gets(new(User), dbflex.NewQueryParam().SetSelect("_id", "LoginID"), &users)
	h.Gets(new(RoleMember), nil, &roleMembers)
	h.Gets(new(RoleFeature), nil, &roleFeatures)

	for _, user := range users {
		if !codekit.HasMember(loginIDs, user.LoginID) {
			continue
		}

		//-- get roles
		userRoleMember := lo.Filter(roleMembers, func(el RoleMember, index int) bool {
			return el.UserID == user.ID
		})
		if len(userRoleMember) == 0 {
			continue
		}

		//-- get features
		allUserRoleFeatures := []RoleFeaturePure{}
		mapFeatures := map[string][]RoleFeaturePure{}
		for _, role := range userRoleMember {
			useRoleFeatures := lo.Filter(roleFeatures, func(el RoleFeature, _ int) bool {
				return el.RoleID == role.RoleID
			})

			for _, f := range useRoleFeatures {
				allUserRoleFeatures = append(allUserRoleFeatures, RoleFeaturePure{
					FeatureID: f.FeatureID,
					Create:    f.Create || f.All,
					Read:      f.Read || f.All,
					Update:    f.Update || f.All,
					Delete:    f.Delete || f.All,
					Posting:   f.Posting || f.All,
					Special1:  f.Special1 || f.All,
					Special2:  f.Special2 || f.All,
					All:       f.All,
				})
			}
		}
		if len(allUserRoleFeatures) == 0 {
			continue
		}

		mapFeatures = lo.GroupBy(allUserRoleFeatures, func(el RoleFeaturePure) string {
			return el.FeatureID
		})
		res[user.ID] = mapFeatures
	}

	return res, nil
}

func HasFeatureAccess(ufm RBACMap, userid, featureID, access string, ignoreDimensionCheck bool, metrices DimensionItems) bool {
	userMap, ok := ufm[userid]
	if !ok {
		return false
	}

	userMapFeatures, ok := userMap[featureID]
	if !ok {
		return false
	}

	dimCheckCount := len(metrices)
	for _, mf := range userMapFeatures {
		hasAccess := (access == "Create" && mf.Create) ||
			(access == "Read" && mf.Read) ||
			(access == "Update" && mf.Update) ||
			(access == "Delete" && mf.Delete) ||
			(access == "Post" && mf.Posting) ||
			(access == "Special1" && mf.Special1) ||
			(access == "Special2" && mf.Special2)
		if !hasAccess {
			continue
		}

		if ignoreDimensionCheck {
			return true
		}

		//-- no dimension on setting's metrice, means it has access to all data
		if len(mf.Dimension.Items) == 0 {
			return true
		}

		validCheckCount := 0
		for _, dim := range metrices {
			comparatorMetrice := DimensionItem{
				Kind:  dim.Kind,
				Value: dim.Value,
			}

			for _, mfd := range mf.Dimension.Items {
				if mfd.Kind == comparatorMetrice.Kind &&
					mfd.Value == comparatorMetrice.Value {
					validCheckCount++
				}
			}
		}

		if validCheckCount > 0 && validCheckCount <= dimCheckCount {
			return true
		}
	}

	return false
}

func UserDimensionMetrices(ufm RBACMap, userid, featureID, access string) []Dimension {
	res := []Dimension{}

	userMap, ok := ufm[userid]
	if !ok {
		return res
	}

	userMapFeatures, ok := userMap[featureID]
	if !ok {
		return res
	}

	for _, mf := range userMapFeatures {
		hasAccess := (access == "Create" && mf.Create) ||
			(access == "Read" && mf.Read) ||
			(access == "Update" && mf.Update) ||
			(access == "Delete" && mf.Delete) ||
			(access == "Post" && mf.Posting) ||
			(access == "Special1" && mf.Special1) ||
			(access == "Special2" && mf.Special2)
		if !hasAccess {
			continue
		}

		res = append(res, mf.Dimension)
	}

	return res
}

func DimensionsToFilter(fieldName string, dims ...Dimension) *dbflex.Filter {
	var res *dbflex.Filter

	filters := make([]*dbflex.Filter, len(dims))
	for dimIdx, dim := range dims {
		dimFilters := make([]*dbflex.Filter, len(dim.Items))
		for dimFilterIdx, mt := range dim.Items {
			f := dbflex.ElemMatch(fieldName+".Items", dbflex.Eq("Kind", mt.Kind), dbflex.Eq("Value", mt.Value))
			dimFilters[dimFilterIdx] = f
		}
		dimF := dbflex.And(dimFilters...)
		filters[dimIdx] = dimF
	}
	if len(filters) > 0 {
		res = dbflex.Or(filters...)
	}

	return res
}

func UserDimensionMetricesFilter(ufm RBACMap, userid, featureID, access, fieldName string) *dbflex.Filter {
	dims := UserDimensionMetrices(ufm, userid, featureID, access)
	return DimensionsToFilter(fieldName, dims...)
}

func IsMemberOf(h *datahub.Hub, userid, roleid string, checkDim bool, dims DimensionItems) bool {
	rms := []RoleMember{}
	rm := new(RoleMember)
	if e := h.Gets(rm, dbflex.NewQueryParam().SetWhere(dbflex.Eqs("UserID", userid, "RoleID", roleid)), &rms); e != nil {
		return false
	}

	if len(rms) == 0 {
		return false
	}

	if !checkDim {
		return true
	}

	return false
}
*/
