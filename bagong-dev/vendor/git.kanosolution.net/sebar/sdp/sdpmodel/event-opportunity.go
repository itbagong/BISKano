package sdpmodel

import "time"

type EventOpportunity struct {
	Event    string    `form_kind:"text"`
	DueDate  time.Time `form_kind:"date"`
	PIC      string    `form_kind:"text" form_lookup:"/tenant/employee/find|_id|Name"`
	Attendee []string  `form_lookup:"/tenant/employee/find|_id|Name"`
	Notes    string    `form_kind:"text"`
}
