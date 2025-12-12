package tenantcoremodel

import "github.com/sebarcode/codekit"

type PreviewSectionType string

const (
	PreviewAsGrid PreviewSectionType = "Grid"
	PreviewAsText PreviewSectionType = "Text"
	PreviewAsHTML PreviewSectionType = "HTML"
)

type PreviewSection struct {
	Title       string
	SectionType PreviewSectionType
	Content     string
	Items       [][]string
}

type PreviewReport struct {
	Header   codekit.M
	Sections []PreviewSection
}
