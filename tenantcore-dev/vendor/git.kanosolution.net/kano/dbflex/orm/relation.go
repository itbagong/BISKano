package orm

import (
	"github.com/sebarcode/codekit"
)

var (
	UseRelationManager bool

	defaultRM RelationManager
)

type RelationManager map[string][]Relation

type Relation struct {
	Children            DataModel
	ParentField         string
	ChildrenField       string
	MapFields           codekit.M
	AutoDeleteChildren  bool
	AutoCreateParent    bool
	DefaultParentFields []string
}

func SetDefaultRelationManager(rm RelationManager) {
	defaultRM = rm
	UseRelationManager = true
}

func DefaultRelationManager() RelationManager {
	if defaultRM == nil {
		defaultRM = RelationManager{}
	}
	return defaultRM
}

func (rm RelationManager) AddRelation(parent DataModel, rels ...Relation) error {
	parentName := parent.TableName()
	_, ok := rm[parentName]
	if !ok {
		rm[parentName] = []Relation{}
	}
	rm[parentName] = append(rm[parentName], rels...)
	return nil
}

func (rm RelationManager) AddParent(child DataModel, parent DataModel, rel Relation) error {
	if rel.ParentField == "" {
		rel.ParentField = "_id"
	}
	rel.Children = child
	rm.AddRelation(parent, rel)
	return nil
}

func (rm RelationManager) FKs(obj DataModel) []*FKConfig {
	res := []*FKConfig{}
	objTableName := obj.TableName()
	for parentName, relations := range rm {
		for _, rel := range relations {
			if rel.Children.TableName() == objTableName {
				fk := &FKConfig{FieldID: rel.ChildrenField,
					RefTableName:     parentName,
					RefField:         rel.ParentField,
					AutoCreate:       rel.AutoCreateParent,
					RefDefaultFields: rel.DefaultParentFields,
					Map:              rel.MapFields,
				}

				res = append(res, fk)
			}
		}
	}
	return res
}

func (rm RelationManager) ReverseFKs(obj DataModel) []*ReverseFKConfig {
	res := []*ReverseFKConfig{}
	objTableName := obj.TableName()
	rels, ok := rm[objTableName]
	if !ok {
		return res
	}

	for _, rel := range rels {
		rfk := &ReverseFKConfig{
			FieldID:      rel.ParentField,
			RefField:     rel.ChildrenField,
			RefTableName: rel.Children.TableName(),
			AutoDelete:   rel.AutoDeleteChildren,
		}

		res = append(res, rfk)
	}

	return res
}
