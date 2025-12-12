package hcmmodel

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TrainingDevelopment struct {
	orm.DataModelBase `bson:"-" json:"-"`
	ID                string    `bson:"_id" json:"_id" form_read_only_edit:"1" form_read_only_new:"1" form_section:"General" form_section_size:"4"`
	RequestDate       time.Time `form_section:"General" form_kind:"date"`
	TrainingTitle     string    `form_section:"General"`
	TrainingType      string    `form_section:"General" form_items:"Recruitment|General"`
	TrainingRequestor string    `form_section:"General"`

	CompanyID        string `grid:"hide" form:"hide"`
	JournalTypeID    string `grid:"hide" form_required:"1" form_section:"General"`
	PostingProfileID string `grid:"hide" form:"hide"`

	RequestTrainingDateFrom time.Time `form_section:"Request Training Date" form_kind:"date" form_section_size:"4"`
	RequestTrainingDateTo   time.Time `form_section:"Request Training Date" form_kind:"date"`

	Status         ficomodel.JournalStatus `form:"hide"`
	TrainingStatus bool                    `form:"hide"`

	Dimension tenantcoremodel.Dimension `grid:"hide" form_section:"Dimension" form_section_direction:"row"`

	Created    time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info" form_section_auto_col:"2"`
	LastUpdate time.Time `form_kind:"datetime" form_read_only:"1" grid:"hide" form_section:"Time Info"`
}

func (o *TrainingDevelopment) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: true, AutoCol: 1},
			{Title: "Request Training Date", ShowTitle: true, AutoCol: 1},
			{Title: "Dimension", AutoCol: 1},
		}},
		{Sections: []suim.FormSection{
			{Title: "Time Info", ShowTitle: true, AutoCol: 2},
		}},
	}
}

func (o *TrainingDevelopment) TableName() string {
	return "HCMTrainingDevelopments"
}

func (o *TrainingDevelopment) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TrainingDevelopment) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TrainingDevelopment) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TrainingDevelopment) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TrainingDevelopment) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TrainingDevelopment) PostSave(dbflex.IConnection) error {
	return nil
}

func (o *TrainingDevelopment) KxPostSave(ctx *kaos.Context, mdl orm.DataModel) error {
	o = mdl.(*TrainingDevelopment)
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return errors.New("PostSave: missing connection")
	}

	if o.TrainingType == "Recruitment" {
		participants := []TrainingDevelopmentParticipant{}
		err := h.Gets(new(TrainingDevelopmentParticipant), dbflex.NewQueryParam().SetWhere(
			dbflex.Eq("TrainingCenterID", o.ID),
		), &participants)
		if err != nil {
			return fmt.Errorf("PostSave: error when get participant: %s", err.Error())
		}

		// only do when close
		if len(participants) > 0 && !o.TrainingStatus {
			for _, p := range participants {
				go func(participant TrainingDevelopmentParticipant) {
					training := new(Training)
					err = h.GetByParm(training, dbflex.NewQueryParam().SetWhere(
						dbflex.And(
							dbflex.Eq("JobVacancyID", participant.ManpowerRequestID),
							dbflex.Eq("CandidateID", participant.EmployeeID),
						),
					))
					if err != nil {
						ctx.Log().Errorf("error when get training with manpower id %s & participant id: %s", participant.ManpowerRequestID, participant.EmployeeID, err.Error())
					}

					training.TrainingStatus = "Close"
					err = h.Update(training, "TrainingStatus")
					if err != nil {
						ctx.Log().Errorf("error when update training with manpower id %s & participant id: %s", participant.ManpowerRequestID, participant.EmployeeID, err.Error())
					}
				}(p)
			}
		}
	}

	return nil
}
