package bagongmodel

import (
	"time"
)

type EmployeeAttachment struct {
	FileName    string
	Description string
	PIC         string    `grid_label:"PIC" form_label:"PIC"`
	UploadDate  time.Time `form_kind:"date"`
	URI         string    `grid_label:"Files"`
	ContentType string    `grid:"hide"`
	Size        int       `grid:"hide"`
}
