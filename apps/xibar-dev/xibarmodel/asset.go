package xibarmodel

import "time"

type AssetGrid struct {
	FileName    string    `form_read_only:"1"`
	Description string    ``
	UploadDate  time.Time `form_kind:"datetime" form_read_only:"1"`
	Content     string    `form_kind:"file" grid_label:"File"`
}
