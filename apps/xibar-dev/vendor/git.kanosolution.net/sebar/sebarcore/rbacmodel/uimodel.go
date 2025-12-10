package rbacmodel

type UserRegisterScreen struct {
	Email           string `form_required:"1"`
	Password        string `form_kind:"password" form_required:"1"`
	ConfirmPassword string `form_kind:"password" form_required:"1"`
}

type ChangePasswordScreen struct {
	CurrentPassword         string `form_kind:"password"`
	NewPassword             string `form_kind:"password"`
	NewPasswordConfirmation string `form_kind:"password"`
}

type LoginScreen struct {
	LoginID    string
	Password   string `form_kind:"password"`
	RememberMe bool
}
