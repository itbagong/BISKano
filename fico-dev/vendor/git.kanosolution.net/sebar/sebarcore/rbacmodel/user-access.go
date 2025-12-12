package rbacmodel

import (
	"github.com/samber/lo"
)

type AccessItem struct {
	FeatureID string
	Level     int
	Dimension Dimension
	Hash      string
}

type UserAccess []*AccessItem

type UserMatrix struct {
	RoleIDs []string
	Access  UserAccess
}

func (ua *UserAccess) Get(fid string, dim Dimension) (*AccessItem, bool) {
	fs := lo.Filter(*ua, func(ua *AccessItem, index int) bool {
		return ua.FeatureID == fid
	})

	var res *AccessItem

	matchDim := 0
	lowestLevel := 1000
	for _, f := range fs {
		match := f.Dimension.SetupCompare(dim)
		if match {
			matchDimF := len(f.Dimension)
			if matchDimF >= matchDim {
				matchDim = matchDimF

				if f.Level < lowestLevel {
					lowestLevel = f.Level
					res = f
				}
			}
		}
	}
	if lowestLevel == 1000 {
		lowestLevel = 0
	}

	return res, lowestLevel > 0
}

func (ua *UserAccess) Update(fid string, dim Dimension, level int) error {
	for _, a := range *ua {
		if a.FeatureID == fid && a.Dimension.SetupCompare(dim) && len(a.Dimension) == len(dim) {
			a.Level = level
			return nil
		}
	}

	*ua = append(*ua, &AccessItem{
		FeatureID: fid,
		Dimension: dim,
		Level:     level,
	})

	return nil
}

func (m *UserMatrix) MemberOf(roleID string) bool {
	memberOfRoleID := lo.IndexOf(m.RoleIDs, roleID) >= 0
	if memberOfRoleID {
		return true
	}

	memberOfAdmins := lo.IndexOf(m.RoleIDs, "Administrators") >= 0
	return memberOfAdmins
}

func (m *UserMatrix) AccessLevel(fid string, dim Dimension) int {
	return getAccessLevel(m.Access, fid, dim)
}

func getAccessLevel(access UserAccess, featureID string, dim Dimension) int {
	accesses := lo.Filter(access, func(d *AccessItem, index int) bool {
		if d.FeatureID == "Administrator" {
			return true
		}
		featureOK := d.FeatureID == featureID
		dimOK := false

		if len(d.Dimension) == 0 {
			dimOK = true
		} else {
			dimOK = d.Dimension.SetupCompare(dim)
		}
		return featureOK && dimOK
	})

	res := 0
	for _, ac := range accesses {
		if ac.Level > res {
			res = ac.Level
		}
	}

	return res
}
