package bagonglogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
)

type SiteEngine struct{}

type SiteByIDRequest struct {
	ID string
}

type SiteRequest struct {
	Name   string
	Filter *dbflex.Filter
}

type SiteResponse struct {
	ID   []string
	Site []SHESite
}

type SHESite struct {
	ID         string `bson:"_id" json:"_id" `
	Name       string
	Address    string
	Alias      string
	Dimension  tenantcoremodel.Dimension
	IsActive   bool
	Created    time.Time
	LastUpdate time.Time
}

func (o *SiteEngine) GetSiteById(ctx *kaos.Context, payload *SiteByIDRequest) (*bagongmodel.Site, error) {
	var (
		db  *datahub.Hub
		err error
		res bagongmodel.Site
	)

	if payload.ID == "" {
		return &res, fmt.Errorf("no id provided")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return &res, errors.New("missing: connection")
	}

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	//get data site by id
	site := new(bagongmodel.Site)
	if e := h.GetByID(site, payload.ID); e != nil {
		ctx.Log().Errorf("Failed populate data Site: %s", e.Error())
		return nil, e
	}

	return site, nil
}

func (o *SiteEngine) GetSiteIds(ctx *kaos.Context, payload *SiteRequest) (*SiteResponse, error) {
	var (
		db  *datahub.Hub
		err error
		res SiteResponse
	)

	if payload.Name == "" {
		return &res, fmt.Errorf("no name provided")
	}

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return &res, errors.New("missing: connection")
	}

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	//get data employee by id
	dataSite := []bagongmodel.Site{}
	if e := h.Gets(new(bagongmodel.Site), dbflex.NewQueryParam().SetWhere(dbflex.Contains("Name", payload.Name)), &dataSite); e != nil {
		return nil, e
	}

	if len(dataSite) > 0 {
		for _, val := range dataSite {
			res.ID = append(res.ID, val.ID)
		}
	}

	return &res, nil
}

func (o *SiteEngine) GetSites(ctx *kaos.Context, payload *SiteRequest) (*SiteResponse, error) {
	var (
		db  *datahub.Hub
		err error
		res SiteResponse
	)

	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return &res, errors.New("missing: connection")
	}

	db, _ = h.BeginTx()
	defer func() {
		if db.IsTx() {
			if err == nil {
				db.Commit()
			} else {
				db.Rollback()
			}
		}
	}()

	query := dbflex.NewQueryParam()
	filters := []*dbflex.Filter{}
	if payload.Filter != nil {
		if len(payload.Filter.Items) > 0 {
			vItems := payload.Filter.Items
			if len(vItems) > 0 {
				for _, val := range vItems {
					fieldVal := val.Field
					opVal := val.Op
					if opVal == dbflex.OpContains {
						aInterface := val.Value.([]interface{})
						aString := make([]string, len(aInterface))
						for i, v := range aInterface {
							aString[i] = v.(string)
						}
						if len(aString) > 0 {
							if aString[0] != "" {
								filters = append(filters, dbflex.Contains(fieldVal, aString[0]))
							}
						}
					}
				}
			}
		} else {
			fieldVal := payload.Filter.Field
			opVal := payload.Filter.Op
			if opVal == dbflex.OpEq {
				aInterface := payload.Filter.Value.(string)
				filters = append(filters, dbflex.Eq(fieldVal, aInterface))
			} else {
				filters = append(filters, payload.Filter)
			}
		}
	}
	if len(filters) > 0 {
		query = query.SetWhere(dbflex.And(filters...))
	}

	//get data employee by id
	dataSite := []bagongmodel.Site{}
	if e := h.Gets(new(bagongmodel.Site), query.SetSort("Name"), &dataSite); e != nil {
		return nil, e
	}

	if len(dataSite) > 0 {
		for _, val := range dataSite {
			siteTmp := SHESite{
				ID:         val.ID,
				Name:       val.Name,
				Address:    val.Address,
				Alias:      val.Alias,
				Dimension:  val.Dimension,
				IsActive:   val.IsActive,
				Created:    val.Created,
				LastUpdate: val.LastUpdate,
			}
			res.Site = append(res.Site, siteTmp)
		}
	}

	return &res, nil
}
