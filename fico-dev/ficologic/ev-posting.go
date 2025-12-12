package ficologic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type EventPostingProfile struct {
}

type EnvelopePost struct {
	Data []*tenantcoremodel.PreviewReport
}

func (obj *EventPostingProfile) Post(ctx *kaos.Context, payloads EventPostRequest) (EnvelopePost, error) {
	res := EnvelopePost{}
	ctx.Data().Set("UserID", payloads.UserID)
	ctx.Data().Set("CompanyID", payloads.CompanyID)
	pph := new(PostingProfileHandler)
	post, err := pph.Post(ctx, payloads.PostRequest)
	if err != nil {
		return res, err
	}
	res.Data = post
	return res, nil
}
