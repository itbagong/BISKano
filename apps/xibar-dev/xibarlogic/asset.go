package xibarlogic

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"github.com/ariefdarmawan/kasset"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type CustomAssetEngines struct {
	fs          kasset.AssetFS
	assetengine *kasset.AssetEngine
}

func NewCustomAssetEngines(fs kasset.AssetFS, assetengine *kasset.AssetEngine) *CustomAssetEngines {
	return &CustomAssetEngines{
		fs:          fs,
		assetengine: assetengine,
	}
}

func (ae *CustomAssetEngines) ViewByJournal(ctx *kaos.Context, assetid string) ([]byte, error) {
	r, rOK := ctx.Data().Get("http_request", nil).(*http.Request)
	w, wOK := ctx.Data().Get("http_writer", nil).(http.ResponseWriter)

	if !rOK {
		return nil, errors.New("not a http compliant request")
	}

	if !wOK {
		return nil, errors.New("not a http compliant writer")
	}

	journalType := r.URL.Query().Get("journal_type")
	journalID := r.URL.Query().Get("journal_id")
	dl := r.URL.Query().Get("t") == "dl"

	ast := new(kasset.Asset)
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return []byte{}, fmt.Errorf("missing: db")
	}
	if e := h.GetByParm(ast, dbflex.NewQueryParam().SetWhere(&dbflex.Filter{Op: dbflex.OpAnd, Items: []*dbflex.Filter{{Op: dbflex.OpEq, Field: "Kind", Value: journalType}, {Op: dbflex.OpEq, Field: "RefID", Value: journalID}}})); e != nil {
		return nil, e
	}
	// if e := h.GetByID(ast, assetID); e != nil {
	// 	return nil, e
	// }

	content, e := ae.fs.Read(ast.URI)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(e.Error()))
	}

	if dl {
		if ast.OriginalFileName != "" {
			w.Header().Set("Content-Disposition", "attachment; filename=\""+ast.OriginalFileName+"\"")
		} else {
			newFileName := codekit.GenerateRandomString("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 16)
			uriParts := strings.Split(ast.URI, ".")
			newFileName += "." + uriParts[len(uriParts)-1]
			w.Header().Set("Content-Disposition", "attachment; filename=\""+newFileName+"\"")
		}
	}
	w.Header().Set("Content-Type", ast.ContentType)
	w.Write(content)

	ctx.Data().Set("kaos_command_1", "stop")
	return content, nil
}

type PayloadRead struct {
	JournalType string
	JournalID   string
	Tags        []string
	NewTags     []string
}

func (a *CustomAssetEngines) ReadByJournal(ctx *kaos.Context, read PayloadRead) ([]kasset.Asset, error) {
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	asts := []kasset.Asset{}
	filters := []*dbflex.Filter{}

	if read.JournalID != "" && read.JournalType != "" {
		filters = append(filters, dbflex.And(dbflex.Eq("Kind", read.JournalType), dbflex.Eq("RefID", read.JournalID)))
	}

	if len(read.Tags) > 0 {
		filters2 := []*dbflex.Filter{}
		for _, val := range read.Tags {
			filters2 = append(filters2, dbflex.Contains("Tags", val))
		}

		if len(filters2) > 0 {
			filters = append(filters, dbflex.Or(filters2...))
		}
	}

	if len(filters) > 0 {
		if e := h.GetsByFilter(new(kasset.Asset), dbflex.Or(filters...), &asts); e != nil {
			return nil, e
		}
	}

	// if e := h.GetsByFilter(new(kasset.Asset), &dbflex.Filter{Op: dbflex.OpAnd, Items: []*dbflex.Filter{{Op: dbflex.OpEq, Field: "Kind", Value: read.JournalType}, {Op: dbflex.OpEq, Field: "RefID", Value: read.JournalID}}}, &asts); e != nil {
	// 	return nil, e
	// }

	/*
		bs, e := a.fs.Read(ast.URI)
		if e != nil {
			return nil, fmt.Errorf("error reading file. %s", e.Error())
		}
	*/

	return asts, nil
}

type PayloadReads struct {
	JournalType string
	JournalIDs  []string
	Tags        []string
	NewTags     []string
}

func (a *CustomAssetEngines) ReadByJournals(ctx *kaos.Context, read PayloadReads) ([]kasset.Asset, error) {
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	asts := []kasset.Asset{}
	filters := []*dbflex.Filter{}

	if len(read.JournalIDs) > 0 && read.JournalType != "" {
		journalInterfaces := lo.Map(read.JournalIDs, func(journalID string, _ int) interface{} {
			return journalID
		})

		filters = append(filters, dbflex.And(dbflex.Eq("Kind", read.JournalType), dbflex.In("RefID", journalInterfaces...)))
	}

	if len(read.Tags) > 0 {
		filters2 := []*dbflex.Filter{}
		for _, val := range read.Tags {
			filters2 = append(filters2, dbflex.Eq("Tags", val))
		}

		if len(filters2) > 0 {
			filters = append(filters, dbflex.Or(filters2...))
		}
	}

	if len(filters) > 0 {
		if e := h.GetsByFilter(new(kasset.Asset), dbflex.Or(filters...), &asts); e != nil {
			return nil, e
		}
	}

	// if e := h.GetsByFilter(new(kasset.Asset), &dbflex.Filter{Op: dbflex.OpAnd, Items: []*dbflex.Filter{{Op: dbflex.OpEq, Field: "Kind", Value: read.JournalType}, {Op: dbflex.OpEq, Field: "RefID", Value: read.JournalID}}}, &asts); e != nil {
	// 	return nil, e
	// }

	/*
		bs, e := a.fs.Read(ast.URI)
		if e != nil {
			return nil, fmt.Errorf("error reading file. %s", e.Error())
		}
	*/

	return asts, nil
}

func (a *CustomAssetEngines) DeleteByJournal(ctx *kaos.Context, read PayloadRead) (string, error) {
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: db")
	}

	// delete asset
	asts := []kasset.Asset{}
	filters := []*dbflex.Filter{}

	if len(read.Tags) > 0 {
		if read.JournalID != "" && read.JournalType != "" {
			for _, val := range read.Tags {
				filters = append(filters, dbflex.And(dbflex.Eq("Kind", read.JournalType), dbflex.Eq("RefID", read.JournalID), dbflex.Eq("Tags", val)))
			}
		}
	}
	if len(filters) > 0 {
		if e := h.GetsByFilter(new(kasset.Asset), dbflex.Or(filters...), &asts); e != nil {
			return "", e
		}
	}

	// check Event hub
	ev, _ := ctx.DefaultEvent()

	if ev == nil {
		return "", errors.New("event cannot null")
	}

	if len(asts) > 0 {
		var vSize interface{}
		for _, valAsset := range asts {
			if e := ev.Publish("/v1/asset/delete", valAsset.ID, vSize, nil); e != nil {
				return "", e
			}
		}
	}

	// update asset
	asts2 := []kasset.Asset{}
	filters2 := []*dbflex.Filter{}
	if len(read.Tags) > 0 {
		for _, val := range read.Tags {
			filters2 = append(filters2, dbflex.Eq("Tags", val))
		}
	}
	if len(filters2) > 0 {
		if e := h.GetsByFilter(new(kasset.Asset), dbflex.Or(filters2...), &asts2); e != nil {
			return "", e
		}
	}

	if len(asts2) > 0 {
		for _, valTags := range read.Tags {
			for _, valAsset := range asts2 {
				isUpdate := false
				if len(valAsset.Tags) > 0 {
					for index, item := range valAsset.Tags {
						if item == valTags {
							valAsset.Tags = append(valAsset.Tags[:index], valAsset.Tags[index+1:]...)
							isUpdate = true
							break
						}
					}
				}
				if isUpdate {
					e := h.Save(&valAsset)
					if e != nil {
						return "", errors.New("error update asset tags")
					}
				}
			}
		}
	}

	return "ok", nil
}

func (a *CustomAssetEngines) UpdateTagByJournal(ctx *kaos.Context, read []PayloadRead) (string, error) {
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: db")
	}

	if len(read) > 0 {
		// update tags by journal id and journal type
		asts := []kasset.Asset{}
		filters := []*dbflex.Filter{}
		mapTag := map[string]PayloadRead{}
		for _, valTag := range read {
			if valTag.JournalID != "" && valTag.JournalType != "" {
				filters = append(filters, dbflex.And(dbflex.Eq("Kind", valTag.JournalType), dbflex.Eq("RefID", valTag.JournalID)))

				mapTag[valTag.JournalID] = valTag
			}
		}
		if len(filters) > 0 {
			if e := h.GetsByFilter(new(kasset.Asset), dbflex.Or(filters...), &asts); e != nil {
				return "", e
			}
		}
		if len(asts) > 0 {
			for _, valAsset := range asts {
				if v, ok := mapTag[valAsset.RefID]; ok {
					valAsset.Tags = append(valAsset.Tags, v.NewTags...)

					e := h.Save(&valAsset)
					if e != nil {
						return "", errors.New("error update asset tags")
					}
				}
			}
		}

		// update tags by tags
		asts2 := []kasset.Asset{}
		filters2 := []*dbflex.Filter{}
		mapTag2 := map[string]PayloadRead{}
		for _, valTag := range read {
			if len(valTag.Tags) > 0 {
				for _, val := range valTag.Tags {
					filters2 = append(filters2, dbflex.Eq("Tags", val))

					mapTag2[val] = valTag
				}
			}
		}

		if len(filters2) > 0 {
			if e := h.GetsByFilter(new(kasset.Asset), dbflex.Or(filters2...), &asts2); e != nil {
				return "", e
			}
		}

		if len(asts2) > 0 {
			for _, valAsset := range asts2 {
				isUpdate := false
				if len(valAsset.Tags) > 0 {
					for _, item := range valAsset.Tags {
						if v, ok := mapTag2[item]; ok {
							valAsset.Tags = append(valAsset.Tags, v.NewTags...)
							isUpdate = true
							break
						}
					}
				}
				if isUpdate {
					e := h.Save(&valAsset)
					if e != nil {
						return "", errors.New("error update asset tags")
					}
				}
			}
		}
	}

	return "ok", nil
}

type PayloadReadContent struct {
	JournalType string
	JournalID   string
}

func (a *CustomAssetEngines) ReadAssetContentByJournal(ctx *kaos.Context, read PayloadReadContent) ([]struct {
	Asset   kasset.Asset
	Content string
}, error) {
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, fmt.Errorf("missing: db")
	}

	assetContents := []struct {
		Asset   kasset.Asset
		Content string
	}{}

	assets := []kasset.Asset{}

	if e := h.GetsByFilter(new(kasset.Asset), &dbflex.Filter{Op: dbflex.OpAnd, Items: []*dbflex.Filter{{Op: dbflex.OpEq, Field: "Kind", Value: read.JournalType}, {Op: dbflex.OpEq, Field: "RefID", Value: read.JournalID}}}, &assets); e != nil {
		return nil, e
	}

	for _, asset := range assets {
		dataasset, e := a.fs.Read(asset.URI)
		if e != nil {
			return nil, fmt.Errorf("error reading file. %s", e.Error())
		}
		assetContents = append(assetContents, struct {
			Asset   kasset.Asset
			Content string
		}{
			Content: base64.StdEncoding.EncodeToString(dataasset),
			Asset:   asset,
		})
	}

	return assetContents, nil
}

func (a *CustomAssetEngines) WriteBatchWithContent(ctx *kaos.Context, attachReqs []kasset.AssetDataBase64) ([]kasset.Asset, error) {
	var wg sync.WaitGroup
	var mtx sync.Mutex
	var assets []kasset.Asset
	var err error

	for _, attachReq := range attachReqs {
		wg.Add(1)
		go func(attachReq kasset.AssetDataBase64) {
			mtx.Lock()
			defer mtx.Unlock()
			defer wg.Done()
			asset, errs := a.assetengine.WriteWithContent(ctx, &attachReq)
			if errs != nil {
				err = errs
				return
			}
			assets = append(assets, *asset)
		}(attachReq)
	}

	wg.Wait()

	if err != nil {
		return nil, err
	}

	return assets, nil
}

func (a *CustomAssetEngines) WriteAsset(ctx *kaos.Context, attachReq *kasset.AssetDataBase64) (*kasset.Asset, error) {
	req := new(kasset.AssetData)
	req.Asset = attachReq.Asset

	if attachReq.Asset.ID != "" && attachReq.Content == "" {
		h := kasset.GetTenantDBFromContext(ctx)
		if h == nil {
			return req.Asset, fmt.Errorf("missing: db")
		}

		e := h.Save(req.Asset)
		if e != nil {
			return req.Asset, errors.New("error update asset")
		}

		return req.Asset, nil
	}

	bs, e := base64.StdEncoding.DecodeString(attachReq.Content)
	if e != nil {
		return nil, fmt.Errorf("fail to decode content. %s", e.Error())
	}
	req.Content = bs

	return a.assetengine.Write(ctx, req)
}

func (a *CustomAssetEngines) WriteBatchAsset(ctx *kaos.Context, attachReqs []kasset.AssetDataBase64) ([]kasset.Asset, error) {
	var wg sync.WaitGroup
	var mtx sync.Mutex
	var assets []kasset.Asset
	var err error

	for _, attachReq := range attachReqs {
		wg.Add(1)
		go func(attachReq kasset.AssetDataBase64) {
			mtx.Lock()
			defer mtx.Unlock()
			defer wg.Done()
			asset, errs := a.WriteAsset(ctx, &attachReq)
			if errs != nil {
				err = errs
				return
			}
			assets = append(assets, *asset)
		}(attachReq)
	}

	wg.Wait()

	if err != nil {
		return nil, err
	}

	return assets, nil
}

type UpdateTagRequest struct {
	AddTags []string
	Tags    []string
}

func (a *CustomAssetEngines) UpdateTag(ctx *kaos.Context, payload UpdateTagRequest) (interface{}, error) {
	h := kasset.GetTenantDBFromContext(ctx)
	if h == nil {
		return "", fmt.Errorf("missing: db")
	}

	if len(payload.Tags) > 0 {
		filters := make([]*dbflex.Filter, len(payload.Tags))

		for i, t := range payload.Tags {
			filters[i] = dbflex.Eq("Tags", t)
		}

		assets := []kasset.Asset{}
		err := h.Gets(new(kasset.Asset), dbflex.NewQueryParam().SetWhere(
			dbflex.Or(filters...),
		), &assets)
		if err != nil {
			return nil, fmt.Errorf("error when get asset: %s", err.Error())
		}

		for _, asset := range assets {
			for _, tag := range payload.AddTags {
				isExist := false
				for _, t := range asset.Tags {
					if tag == t {
						isExist = true
						break
					}
				}

				if !isExist {
					asset.Tags = append(asset.Tags, tag)
				}
			}

			err = h.Save(&asset)
			if err != nil {
				return nil, fmt.Errorf("error when save asset: %s", err.Error())
			}
		}

		return "success", nil
	}

	return nil, fmt.Errorf("tags is required")
}
