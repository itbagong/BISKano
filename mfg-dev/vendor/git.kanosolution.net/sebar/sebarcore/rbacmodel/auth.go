package rbacmodel

import "github.com/sebarcode/codekit"

type AuthChangeDataScope string

const (
	AuthScopeIsJWT     AuthChangeDataScope = "JWT"
	AuthScopeIsSession AuthChangeDataScope = "SESSION"
	AuthScopeIsBoth    AuthChangeDataScope = "BOTH"
)

type AuthChangeDataRequest struct {
	Scope         AuthChangeDataScope
	Token         string
	Data          codekit.M
	ImpersonateAs string
}
