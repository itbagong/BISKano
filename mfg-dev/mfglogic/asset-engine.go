package mfglogic

import (
	"encoding/base64"
	"errors"
	"fmt"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmconfig"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/kasset"
	"github.com/ariefdarmawan/reflector"
)

type AssetEngine struct{}

type AssetUploadRequest struct {
	JournalType string
	JournalID   string
	Assets      []AssetField
}

type AssetField struct {
	AssetData kasset.AssetDataBase64
	Field     string // "PhotoID"
}

func (o *AssetEngine) Upload(ctx *kaos.Context, req *AssetUploadRequest) (interface{}, error) {
	db := sebar.GetTenantDBFromContext(ctx)
	if db == nil {
		return nil, errors.New("missing: db")
	}

	var m orm.DataModel

	switch req.JournalType {
	case string(scmmodel.ModuleWorkorder):
		m = new(mfgmodel.WorkOrderJournal)
	default:
		return nil, fmt.Errorf("invalid module: %s", req.JournalType)
	}

	if e := db.GetByID(m, req.JournalID); e != nil {
		return nil, e
	}

	jf := reflector.From(m)

	for _, a := range req.Assets {
		bs, e := base64.StdEncoding.DecodeString(a.AssetData.Content)
		if e != nil {
			return nil, e
		}

		resAss := new(kasset.Asset)
		if e := scmconfig.Config.EventHub().Publish("/v1/asset/write", &kasset.AssetData{
			Asset:   a.AssetData.Asset,
			Content: bs,
		}, resAss, nil); e != nil {
			return nil, e
		}

		jf.Set(a.Field, resAss.ID) // TODO: consider when Field is hierarchical such as "Attachment.DocumentID"
	}

	if e := db.Save(m); e != nil {
		return nil, e
	}

	return m, nil
}

func (o *AssetEngine) getJournal(J orm.DataModel) {

}
