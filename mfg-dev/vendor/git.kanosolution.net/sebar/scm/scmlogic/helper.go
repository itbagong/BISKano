package scmlogic

import (
	"encoding/json"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/datahub"
	"github.com/leekchan/accounting"
	"github.com/sebarcode/codekit"
)

func GetCompanyIDFromContext(ctx *kaos.Context) (string, error) {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	if coID == "" {
		return "DEMO00", nil
		//return coID, errors.New("session: Company ID not found, please relogin")
	}
	return coID, nil
}

func GetCompanyAndUserIDFromContext(ctx *kaos.Context) (string, string, error) {
	coID := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M).GetString("CompanyID")
	userID := sebar.GetUserIDFromCtx(ctx)

	if coID == "" || userID == "" {
		return "", "", fmt.Errorf("session expired")
	}

	return coID, userID, nil
}

func SetIfEmpty(v *string, def string) {
	if *v == "" {
		*v = def
	}
}

// Deserialize decodes query string into a map
func Deserialize(str string, res interface{}) error {
	err := json.Unmarshal([]byte(str), &res)
	if err != nil {
		return fmt.Errorf("Deserialize Failed | %s:\n%s", err, str)
	}
	return err
}

type GeneralRequest struct {
	Select []string
	Sort   []string
	Skip   int
	Take   int
	Where  *dbflex.Filter
}

func (o *GeneralRequest) GetQueryParam() *dbflex.QueryParam {
	parm := dbflex.NewQueryParam()

	// TODO: do something with o.Select

	if len(o.Sort) > 0 {
		parm = parm.SetSort(o.Sort...)
	}

	parm = parm.SetSkip(o.Skip)
	parm = parm.SetTake(o.Take)

	if o.Where != nil {
		parm = parm.SetWhere(o.Where)
	}

	return parm
}

func (o *GeneralRequest) GetQueryParamWithURLQuery(ctx *kaos.Context) *dbflex.QueryParam {
	parm := dbflex.NewQueryParam()

	// TODO: do something with o.Select

	if len(o.Sort) > 0 {
		parm = parm.SetSort(o.Sort...)
	}

	parm = parm.SetSkip(o.Skip)
	parm = parm.SetTake(o.Take)

	fs := []*dbflex.Filter{}

	if qfs := GetURLQueryParamFilter(ctx); qfs != nil {
		fs = append(fs, qfs)
	}

	if o.Where != nil {
		fs = append(fs, o.Where)
	}

	if len(fs) > 0 {
		parm = parm.SetWhere(dbflex.And(fs...))
	}

	return parm
}

func GetURLQueryParams(ctx *kaos.Context) map[string]string {
	r, ok := sebar.GetHTTPRequest(ctx)
	if !ok {
		return map[string]string{}
	}

	res := map[string]string{}
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			res[key] = values[0]
		}
	}

	return res
}

func GetURLQueryParamFilter(ctx *kaos.Context, op ...dbflex.FilterOp) *dbflex.Filter {
	r, ok := sebar.GetHTTPRequest(ctx)
	if !ok {
		return nil
	}

	fs := []*dbflex.Filter{}

	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			fs = append(fs, dbflex.Eq(key, values[0]))
		} else if len(values) > 1 {
			fsor := []*dbflex.Filter{}
			for _, v := range values {
				fsor = append(fsor, dbflex.Eq(key, v))
			}
			fs = append(fs, dbflex.Or(fsor...))
		}
	}

	if len(fs) == 0 {
		return nil
	}

	if len(op) > 0 && op[0] == dbflex.OpOr {
		return dbflex.Or(fs...)
	}

	return dbflex.And(fs...)
}

// InventTrxSingleJournalSave hanya untuk trx / line yang berasal dari 1 journal saja, ketika trx / line bisa berasal dari beda journal Jangan Pakai Ini !!! pakai yang InventTrxMultiJournalSave
func InventTrxSingleJournalSave(h *datahub.Hub, trxs map[string][]orm.DataModel, companyID string, sourceType interface{}, sourceJournalID string, status scmmodel.ItemBalanceStatus) ([]*scmmodel.InventTrx, error) {
	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))
	if len(inventTrxs) == 0 {
		return nil, nil // prevent deleting without saving
	}

	h.DeleteByFilter(new(scmmodel.InventTrx),
		dbflex.Eqs("CompanyID", companyID, "SourceType", codekit.ToString(sourceType), "SourceJournalID", sourceJournalID),
	)

	for _, trx := range inventTrxs {
		trx.Status = status
		if e := h.Save(trx); e != nil {
			return nil, e
		}
	}

	return inventTrxs, nil
}

// InventTrxMultiJournalSave untuk trx / line yang bisa berasal dari journal yang berbeda
func InventTrxMultiJournalSave(h *datahub.Hub, trxs map[string][]orm.DataModel) ([]*scmmodel.InventTrx, error) {
	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))

	for _, trx := range inventTrxs {
		h.DeleteByFilter(new(scmmodel.InventTrx), dbflex.Eqs(
			"CompanyID", trx.CompanyID,
			"SourceType", trx.SourceType,
			"SourceJournalID", trx.SourceJournalID,
			"Status", trx.Status,
			"Item._id", trx.Item.ID,
			"SKU", trx.SKU,
			// TODO: butuh SourceLineNo ga ya? khawatir kalau ada 2 Line dengan Item._id dan SKU yang sama
		))

		if e := h.Save(trx); e != nil {
			return nil, e
		}
	}

	return inventTrxs, nil
}

func InventTrxSplitSave(h *datahub.Hub, trxs map[string][]orm.DataModel, sourceStatus scmmodel.ItemBalanceStatus) (map[string][]orm.DataModel, error) {
	inventTrxs := ficologic.FromDataModels(trxs[new(scmmodel.InventTrx).TableName()], new(scmmodel.InventTrx))

	splittedTrxs := []*scmmodel.InventTrx{}
	spliter := NewInventSplit(h)
	for _, trx := range inventTrxs {
		updatedSourceTrxs, newTrxs, err := spliter.SetOpts(&InventSplitOpts{
			SplitType:       SplitBySource,
			CompanyID:       trx.CompanyID,
			SourceType:      string(trx.SourceType),
			SourceJournalID: trx.SourceJournalID,
			SourceLineNo:    trx.SourceLineNo,
			SourceStatus:    string(sourceStatus),
		}).Split(trx.Qty, string(scmmodel.ItemConfirmed))
		if err != nil {
			return nil, err
		}

		splittedTrxs = append(splittedTrxs, append(updatedSourceTrxs, newTrxs...)...)
	}

	if len(splittedTrxs) == 0 {
		return nil, fmt.Errorf("fail split invent trx, return 0 trx")
	}

	delete(trxs, new(scmmodel.InventTrx).TableName())
	trxs[new(scmmodel.InventTrx).TableName()] = ficologic.ToDataModels(splittedTrxs)

	return trxs, nil
}

func FormatMoney(money float64) string {
	return accounting.FormatNumber(money, 2, ",", ".")
}

func FormatDate(paramDate *time.Time) string {
	if paramDate == nil {
		return ""
	}

	return paramDate.Format("02 January 2006")
}

func receiveIssueLineToTrx(db *datahub.Hub, header *scmmodel.InventReceiveIssueJournal, line scmmodel.InventReceiveIssueLine) (*scmmodel.InventTrx, error) {
	// var err error

	trx := new(scmmodel.InventTrx)
	trx.CompanyID = header.CompanyID
	trx.Text = line.Text
	trx.Item = line.Item
	trx.SKU = line.SKU
	trx.Dimension = line.Dimension
	trx.InventDim = line.InventDim

	trx.Qty = line.InventQty
	trx.TrxQty = line.Qty
	trx.TrxUnitID = line.UnitID

	trx.SourceType = line.SourceType
	trx.SourceJournalID = line.SourceJournalID
	trx.SourceTrxType = line.SourceTrxType
	trx.SourceLineNo = line.SourceLineNo

	trx.AmountPhysical = line.UnitCost * line.Qty
	trx.References = trx.References.Set(scmmodel.GetRefKey(line.SourceTrxType), line.SourceJournalID)

	trx.TrxDate = header.TrxDate
	return trx, nil
}
