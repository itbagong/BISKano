package karamodel

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"github.com/ariefdarmawan/suim"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WorkLocation struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" key:"1" form_read_only:"1" grid:"hide" form_section:"General" form_section_show_title:"1"`
	Name              string    `form_required:"1" form_section_size:"3"`
	Enable            bool      `form_section:"Setting" form_section_show_title:"1"`
	Virtual           bool      `form_section:"Setting"`
	AcceptNonRule     bool      `form_section:"Setting"`
	TimeLoc           string    `form_section:"Setting"`
	DistanceTolerance float32   `form_section:"Setting"`
	Address           string    `form_section:"Map" form_multi_row:"3"`
	Location          *Location `form_section:"Map"`
	// Longitude         float64   `form_section:"Map"`
	// Latitude          float64   `form_section:"Map" form:"hide"`
	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Timestamp" form_section_show_title:"1"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Timestamp"`
}

type Location struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

func (o *WorkLocation) TableName() string {
	return "KaraWorkLocations"
}

func (o *WorkLocation) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *WorkLocation) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *WorkLocation) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *WorkLocation) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *WorkLocation) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	if o.TimeLoc == "" {
		return errors.New("INVALID_TIMELOC")
	}
	_, err := time.LoadLocation(o.TimeLoc)
	if err != nil {
		return fmt.Errorf("INVALID_TIMEZONE, %s", err.Error())
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *WorkLocation) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *WorkLocation) PreDelete(conn dbflex.IConnection) error {
	ms := []codekit.M{}
	conn.Cursor(dbflex.From(new(WorkLocationUser).TableName()).Select("_id").Where(dbflex.Eq("WorkLocationID", o.ID)).Take(1), nil).Fetchs(&ms, 0).Close()
	if len(ms) > 1 {
		return errors.New("reference: work location user")
	}

	ms = []codekit.M{}
	conn.Cursor(dbflex.From(new(ProfileRole).TableName()).Select("_id").Where(dbflex.Eq("WorkLocationID", o.ID)).Take(1), nil).Fetchs(&ms, 0).Close()
	if len(ms) > 1 {
		return errors.New("reference: profile role user")
	}

	return nil
}

func (o *WorkLocation) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: true, AutoCol: 1},
			{Title: "Setting", ShowTitle: true, AutoCol: 1},
		}},
		{Sections: []suim.FormSection{{Title: "Map", ShowTitle: true, AutoCol: 1}}},
		{Sections: []suim.FormSection{{Title: "Timestamp", ShowTitle: true, AutoCol: 1}}},
	}
}
