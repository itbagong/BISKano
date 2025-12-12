package flowmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FlowTemplate struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string `bson:"_id" json:"_id" key:"1" form_read_only_edit:"1" form_section:"General" form_section_auto_col:"2"`
	Name              string `form_required:"1" form_section:"General"`
	Tasks             Tasks
	Routes            Routes
	Created           time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate        time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *FlowTemplate) TableName() string {
	return "FlowTemplates"
}

func (o *FlowTemplate) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *FlowTemplate) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *FlowTemplate) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *FlowTemplate) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *FlowTemplate) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *FlowTemplate) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *FlowTemplate) GetGenesisTask() Tasks {
	destinationTaskIDs := lo.Map(o.Routes, func(r Route, i int) string {
		return r.ToID
	})
	genesisTasks := lo.Filter(o.Tasks, func(t Task, i int) bool {
		_, has := lo.Find(destinationTaskIDs, func(id string) bool {
			return id == t.ID
		})
		return !has
	})
	return genesisTasks
}

func (o *FlowTemplate) GetTaskFrom(fromID string) Tasks {
	routes := lo.Filter(o.Routes, func(r Route, i int) bool {
		return r.FromID == fromID
	})
	tasks := lo.Map(routes, func(r Route, i int) Task {
		ts := lo.Filter(o.Tasks, func(t Task, i int) bool {
			return t.ID == r.FromID
		})
		if len(ts) == 0 {
			return Task{ID: ""}
		}
		return ts[0]
	})
	return lo.Filter(tasks, func(ts Task, i int) bool {
		return ts.ID != ""
	})
}

func (o *FlowTemplate) GetTaskTo(toID string) Tasks {
	routes := lo.Filter(o.Routes, func(r Route, i int) bool {
		return r.ToID == toID
	})
	tasks := lo.Map(routes, func(r Route, i int) Task {
		ts := lo.Filter(o.Tasks, func(t Task, i int) bool {
			return t.ID == r.FromID
		})
		if len(ts) == 0 {
			return Task{ID: ""}
		}
		return ts[0]
	})
	return lo.Filter(tasks, func(ts Task, i int) bool {
		return ts.ID != ""
	})
}
