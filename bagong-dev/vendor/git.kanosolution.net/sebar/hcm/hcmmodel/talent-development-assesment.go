package hcmmodel

import (
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"github.com/ariefdarmawan/suim"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssesmentItemDetail struct {
	Aspect        string
	MaxScore      int
	AchievedScore int
}

type AssesmentAttendanceDetail struct {
	Name  string
	Score int
}

type AssesmentAttendance struct {
	Presences []AssesmentAttendanceDetail
	Absents   []AssesmentAttendanceDetail
	Sicks     []AssesmentAttendanceDetail
	Leaves    []AssesmentAttendanceDetail
	Lates     []AssesmentAttendanceDetail
}

type AssesmentDetail struct {
	Attendace          AssesmentAttendance
	ItemTemplateID     string
	ItemDetails        []AssesmentItemDetail
	MaxScoreTotal      int     `form:"hide"`
	AchievedScoreTotal int     `form:"hide"`
	FinalScore         float64 `form:"hide"`
	IsProbationEnd     bool    `form_section:"Result" form_label:"Habis Probation/Kontrak" form_section_show_title:"1"`
	IsBecomeEmployee   bool    `form_section:"Result" form_label:"Diangkat menjadi karyawan tetap"`
	IsPromoted         bool    `form_section:"Result" form_label:"Dipromosikan"`
	PromoteTo          string  `form_label:"Dipromosikan menjadi" form_section:"Result" `
	Facility           string  `form_section:"Result"`
	Salary             float64 `form_section:"Result"`
	Status             string  `form_section:"Result"`
}

type AssesmentInterviewPointDetail struct {
	Subject string
	Weight  int
	Value   int
	Score   float32
	Note    string
}

type AssesmentInterviewDetail struct {
	Section string
	Details []AssesmentInterviewPointDetail
}

type AssesmentInterview struct {
	Interviewer     string                     `form_section:"General" form_lookup:"/tenant/employee/find|_id|Name"`
	Date            *time.Time                 `form_section:"General" form_kind:"date"`
	ItemTemplateID  string                     `form_section:"General"`
	Details         []AssesmentInterviewDetail `form_section:"Details"`
	TotalWeight     int                        `form_section:"D. Catatan / Feedback dari atasan :" form:"hide"`
	Score           float32                    `form_section:"D. Catatan / Feedback dari atasan :" form:"hide"`
	StrongSkill     string                     `form_section:"D. Catatan / Feedback dari atasan :" form_multi_row:"2" form_label:"Kekuatan yang bersangkutan yang perlu dipertahankan/ditingkatkan"`
	NeedImprovement string                     `form_section:"D. Catatan / Feedback dari atasan :" form_multi_row:"2" form_label:"Perbaikan yang diperlukan oleh yang bersangkutan"`
	Start           *time.Time                 `form_section:"Conclusion" form_kind:"date"`
	PeriodFrom      *time.Time                 `form_section:"Conclusion" form_kind:"date" form:"hide"`
	PeriodTo        *time.Time                 `form_section:"Conclusion" form_kind:"date" form:"hide"`
	Conclusion      string                     `form_section:"Conclusion" form:"hide"`
	ImportantNote   string                     `form_section:"Conclusion" form_multi_row:"2" form_label:"Catatan penting lain"`
}

func (o *AssesmentInterview) FormSections() []suim.FormSectionGroup {
	return []suim.FormSectionGroup{
		{Sections: []suim.FormSection{
			{Title: "General", ShowTitle: false, AutoCol: 3},
			{Title: "Details", ShowTitle: false, AutoCol: 1},
			{Title: "D. Catatan / Feedback dari atasan :", ShowTitle: true, AutoCol: 2},
			{Title: "Conclusion", ShowTitle: false, AutoCol: 2},
		}},
	}
}

type TalentDevelopmentAssesment struct {
	orm.DataModelBase   `bson:"-" json:"-"`
	ID                  string `bson:"_id" json:"_id" form:"hide"`
	CompanyID           string `grid:"hide" form:"hide"`
	JournalTypeID       string `grid:"hide" form:"hide" form_lookup:"/hcm/journaltype/find?TransactionType=Talent%20Development%20-%20Promotion%20-%20Tracking%20Assessment|_id|Name"`
	PostingProfileID    string `grid:"hide" form:"hide"`
	TalentDevelopmentID string `form:"hide"`
	JobTitle            string `form:"hide"`
	Assesment           AssesmentDetail
	Interview           AssesmentInterview        `form:"hide"`
	PsychoTests         []PsychologicalTestDetail `grid:"hide" form:"hide"`
	Status              ficomodel.JournalStatus   `form:"hide" form_read_only:"1"`
	Created             time.Time                 `grid:"hide" form:"hide"`
	LastUpdate          time.Time                 `grid:"hide" form:"hide"`
}

func (o *TalentDevelopmentAssesment) TableName() string {
	return "HCMTalentDevelopmentAssesments"
}

func (o *TalentDevelopmentAssesment) FK() []*orm.FKConfig {
	return orm.DefaultRelationManager().FKs(o)
}

func (o *TalentDevelopmentAssesment) ReverseFK() []*orm.ReverseFKConfig {
	return orm.DefaultRelationManager().ReverseFKs(o)
}

func (o *TalentDevelopmentAssesment) SetID(keys ...interface{}) {
	o.ID = keys[0].(string)
}

func (o *TalentDevelopmentAssesment) GetID(dbflex.IConnection) ([]string, []interface{}) {
	return []string{"_id"}, []interface{}{o.ID}
}

func (o *TalentDevelopmentAssesment) PreSave(dbflex.IConnection) error {
	if o.ID == "" {
		o.ID = primitive.NewObjectID().Hex()
	}
	if o.Created.IsZero() {
		o.Created = time.Now()
	}
	o.LastUpdate = time.Now()
	return nil
}

func (o *TalentDevelopmentAssesment) PostSave(dbflex.IConnection) error {
	return nil
}
