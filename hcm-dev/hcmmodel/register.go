package hcmmodel

type UserRegisterScreen struct {
	CompanyID       string `form_required:"1" form_label:"Company" form_lookup:"/tenant/company/find|_id|Name"`
	Email           string `form_required:"1"`
	Password        string `form_kind:"password" form_required:"1"`
	ConfirmPassword string `form_kind:"password" form_required:"1"`
}
