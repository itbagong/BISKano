package karamodel

import "git.kanosolution.net/kano/dbflex/orm"

func InitModelRelationAndIndex() {
	rm := orm.DefaultRelationManager()

	trx := new(AttendanceTrx)
	rm.AddParent(trx, new(WorkLocation), orm.Relation{ChildrenField: "WorkLocationID"})
	rm.AddParent(trx, new(UserProfile), orm.Relation{ChildrenField: "UserID"})
	rm.AddParent(trx, new(AttendanceRule), orm.Relation{ChildrenField: "RuleID"})

	ruleLine := new(RuleLine)
	rm.AddParent(ruleLine, new(AttendanceRule), orm.Relation{ChildrenField: "RuleID", AutoDeleteChildren: true})

	wlu := new(WorkLocationUser)
	rm.AddParent(wlu, new(UserProfile), orm.Relation{ChildrenField: "UserID"})
	rm.AddParent(wlu, new(AttendanceRule), orm.Relation{ChildrenField: "RuleID"})
}
