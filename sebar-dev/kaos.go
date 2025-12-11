package sebar

import (
	"net/http"

	"git.kanosolution.net/kano/kaos"
)

const (
	CtxJWTReferenceID = "jwt_reference_id"
	HTTP_REQUEST      = "http_request"
	HTTP_WRITER       = "http_writer"
)

// GetAccountID from given kaos context
func GetUserIDFromCtx(ctx *kaos.Context) string {
	return ctx.Data().Get(CtxJWTReferenceID, "").(string)
}

// GetHTTPRequest from given kaos context
func GetHTTPRequest(ctx *kaos.Context) (*http.Request, bool) {
	hr, ok := ctx.Data().Get(HTTP_REQUEST, nil).(*http.Request)
	return hr, ok
}
