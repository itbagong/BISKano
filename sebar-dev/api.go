package sebar

import (
	"net/http"
	"time"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/codekit"
)

func WrapApiError(ctx *kaos.Context, errTxt string) {
	w, ok := ctx.Data().Get("http_writer", nil).(http.ResponseWriter)
	if !ok {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(codekit.Jsonify(codekit.M{}.Set("error", errTxt).Set("timestamp", time.Now())))

	r, ok := ctx.Data().Get("http_request", nil).(*http.Request)
	if ok {
		ctx.Log().Errorf("%s | %s", r.URL.String(), errTxt)
	}
}
