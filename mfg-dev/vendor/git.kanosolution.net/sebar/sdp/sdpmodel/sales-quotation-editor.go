package sdpmodel

type SalesQuotationEditorForm struct {
	LetterHeadAsset string `form_kind:"file" form_section:"Head" form_section_show_title:"1"`
	LetterHeadFirst bool   `form_section:"Head"`
	FooterAsset     string `form_kind:"file" form_section:"Footer" form_section_show_title:"1"`
	FooterLastPage  bool   `form_section:"Footer"`
	Editor          string `form_kind:"html" form_section:"Editor" form_section_show_title:"1"`
}
