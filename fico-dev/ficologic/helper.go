package ficologic

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/ariefdarmawan/serde"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TernarySubledgerAccount(accounts ...ficomodel.SubledgerAccount) ficomodel.SubledgerAccount {
	account := ficomodel.SubledgerAccount{}

	for _, acc := range accounts {
		if acc.AccountType == "" {
			acc.AccountType = "LEDGERACCOUNT"
		}

		if account.AccountType == "" {
			acc.AccountType = account.AccountType
		}
		if account.AccountType == acc.AccountType && account.AccountID == "" && acc.AccountID != "" {
			account.AccountID = acc.AccountID
		}

		if account.AccountType != "" && account.AccountID != "" {
			return account
		}
	}

	return account
}

func validateQty(a, b float64) (float64, error) {
	if a < 0 && b > 0 {
		return 0, errors.New("different sign")
	}

	if a > 0 && b < 0 {
		return 0, errors.New("different sign")
	}

	if math.Abs(a) <= math.Abs(b) {
		return a, nil
	}

	return b, nil
}

func GetSnapshotRecord[T orm.DataModel](db *datahub.Hub, refModel T, origWhere *dbflex.Filter,
	dateFieldName string, balDate *time.Time, isPrevious bool) (*time.Time, error) {
	var (
		snapshotRecord T
		err            error
	)

	ssWhere := []*dbflex.Filter{origWhere}
	if balDate != nil {
		if isPrevious {
			ssWhere = append(ssWhere, dbflex.Lte(dateFieldName, balDate))
		} else {
			ssWhere = append(ssWhere, dbflex.Gte(dateFieldName, balDate))
		}
	}
	snapshotRecord, err = datahub.GetByParm(db, refModel, dbflex.NewQueryParam().SetWhere(dbflex.And(ssWhere...)).SetSort("-BalanceDate"))
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("get balance snapshot date: %s", err.Error())
	} else if err == io.EOF {
		return nil, nil
	}

	ssDate := new(time.Time)
	if err = reflector.From(snapshotRecord).GetTo(dateFieldName, ssDate); err != nil {
		return nil, fmt.Errorf("get balance date: %s", err.Error())
	}
	return ssDate, nil
}

func SaveModels(db *datahub.Hub, op string, models ...orm.DataModel) ([]orm.DataModel, []error) {
	failedModels := []orm.DataModel{}
	saveErrs := []error{}

forModel:
	for _, model := range models {
		var e error
		switch op {
		case dbflex.QueryInsert:
			e = db.Insert(model)

		case dbflex.QueryUpdate:
			e = db.Update(model)

		case dbflex.QuerySave:
			e = db.Save(model)

		default:
			e = errors.New("invalid op")
			saveErrs = append(saveErrs, e)
			failedModels = models
			break forModel
		}
		if e != nil {
			_, ids := model.GetID(nil)
			saveErrs = append(saveErrs, fmt.Errorf("%s: %s: %s", model.TableName(), ids[0].(string), e.Error()))
			failedModels = append(failedModels, model)
		}
	}

	return failedModels, saveErrs
}

func ExtractPattern(txt string) (string, int, error) {
	txtLen := len(txt)
	indexStr := ""
	for i := 0; i < txtLen; i++ {
		c := txt[txtLen-i-1]
		if !(c >= '0' && c <= '9') {
			break
		}
		indexStr = string(c) + indexStr
	}

	pattern := ""
	if len(indexStr) < txtLen {
		pattern = txt[:txtLen-len(indexStr)]
	}
	pattern += "%0" + fmt.Sprintf("%dd", len(indexStr))

	index, _ := strconv.Atoi(indexStr)
	return pattern, index, nil
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

func GetMatchDimension(dimIface []interface{}) bson.A {
	and := bson.A{}

	dim := tenantcoremodel.Dimension{}
	if err := serde.Serde(dimIface, &dim); err != nil {
		return and
	}

	dimGroup := lo.GroupBy(dim, func(item tenantcoremodel.DimensionItem) string {
		return item.Key
	})

	for _, val := range dimGroup {
		or := bson.A{}
		for _, di := range val {
			or = append(or, bson.M{
				"Dimension": bson.M{
					"$elemMatch": bson.M{
						"Key":   di.Key,
						"Value": di.Value,
					},
				},
			})
		}
		and = append(and, bson.M{
			"$or": or,
		})
	}

	return and
}

type TrxDate struct {
	Operator dbflex.FilterOp
	Value    time.Time
}

func GetMatchItems(Items []*dbflex.Filter, match bson.M) bson.M {
	if len(Items) > 0 {
		trxDate := []TrxDate{}
		for _, val := range Items {
			fieldVal := val.Field
			opVal := val.Op

			switch fieldVal {
			case "Text", "Name":
				aInterface := val.Value.([]interface{})
				if len(aInterface) > 0 {
					aString := make([]string, len(aInterface))
					for i, v := range aInterface {
						aString[i] = v.(string)
					}

					fieldVal = "Text"
					match[fieldVal] = bson.M{
						"$regex": primitive.Regex{Pattern: aString[0], Options: "i"},
					}
				}
			case "TrxDate":
				t, _ := time.Parse("2006-01-02T15:04:05.000Z", val.Value.(string))
				trxDate = append(trxDate, TrxDate{
					Operator: opVal,
					Value:    t,
				})
			case "Status":
				aInterface := val.Value.([]interface{})
				match[fieldVal] = bson.M{
					"$in": aInterface,
				}
			// Dimension
			case "":
				elemMatch := bson.M{}
				for _, it := range val.Items {
					field := strings.Split(it.Field, ".")[1]
					if field == "Key" {
						elemMatch[field] = it.Value
					} else {
						elemMatch[field] = bson.M{"$in": it.Value}
					}
				}
				match["Dimension"] = bson.M{
					"$elemMatch": elemMatch,
				}
			}
		}

		if len(trxDate) > 0 {
			filterDate := bson.M{}
			for _, valDate := range trxDate {
				filterDate[string(valDate.Operator)] = valDate.Value
			}
			match["TrxDate"] = filterDate
		}
	}

	return match
}

func GetFilterDimension(dimIface []interface{}) []*dbflex.Filter {
	and := []*dbflex.Filter{}

	dim := tenantcoremodel.Dimension{}
	if err := serde.Serde(dimIface, &dim); err != nil {
		return and
	}

	dimGroup := lo.GroupBy(dim, func(item tenantcoremodel.DimensionItem) string {
		return item.Key
	})

	for _, val := range dimGroup {
		df := make([]*dbflex.Filter, len(val))
		for idx, di := range val {
			df[idx] = dbflex.ElemMatch("Dimension", dbflex.Eq("Key", di.Key), dbflex.Eq("Value", di.Value))
		}
		and = append(and, dbflex.Or(df...))
	}

	return and
}

func GetFilterItems(Items []*dbflex.Filter) []*dbflex.Filter {
	filters2 := []*dbflex.Filter{}
	if len(Items) > 0 {
		for _, val := range Items {
			fieldVal := val.Field
			aInterface := val.Value.([]interface{})
			aString := make([]string, len(aInterface))
			for i, v := range aInterface {
				aString[i] = v.(string)
			}
			if len(aString) > 0 {
				filters2 = append(filters2, dbflex.Contains(fieldVal, aString[0]))
			}
		}
	}

	return filters2
}
