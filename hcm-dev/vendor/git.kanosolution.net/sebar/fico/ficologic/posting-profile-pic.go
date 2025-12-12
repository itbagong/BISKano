package ficologic

import (
	"errors"
	"fmt"
	"net/http"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
)

type PostingProfilePICHandler struct {
}

func (m *PostingProfilePICHandler) GetByPostingProfileID(ctx *kaos.Context, payload *dbflex.QueryParam) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	match := bson.M{}
	var query *dbflex.Filter
	if hr, ok := ctx.Data().Get("http_request", nil).(*http.Request); ok {
		queryValues := hr.URL.Query()
		for k, vs := range queryValues {
			match[k] = vs[0]
			query = dbflex.Eq(k, vs[0])
		}
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
		{
			"$sort": bson.M{
				"Priority": 1,
				"_id":      1,
			},
		},
		{
			"$skip": payload.Skip,
		},
		{
			"$limit": payload.Take,
		},
	}
	pics := []ficomodel.PostingProfilePIC{}
	cmd := dbflex.From(new(ficomodel.PostingProfilePIC).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &pics); err != nil {
		return nil, fmt.Errorf("err when get posting profile pic: %s", err.Error())
	}

	count, err := h.Count(new(ficomodel.PostingProfilePIC), dbflex.NewQueryParam().SetWhere(query))
	if err != nil {
		return nil, fmt.Errorf("err when get count: %s", err.Error())
	}

	return codekit.M{"data": pics, "count": count}, nil
}
