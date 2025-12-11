package bagonglogic

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/ariefdarmawan/reflector"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"github.com/xuri/excelize/v2"
)

type MasterUploadEngine struct{}

type MURequest struct {
	TableName string
	Content   string
}

type MUResult struct {
	Result string
}

type ListModelResult struct {
	Key      string
	Name     string
	Service  string
	Download string
	Model    orm.DataModel
}

var listDownloadFunc = map[string](func(h *datahub.Hub) ([]codekit.M, error)){
	// new(tenantcoremodel.CustomItemDownload).TableName(): new(tenantcorelogic.ItemEngine).Download(),
	new(tenantcoremodel.CustomItemDownload).TableName(): new(tenantcorelogic.ItemSpecEngine).Download(),
	new(tenantcoremodel.ItemGroup).TableName(): func(h *datahub.Hub) ([]codekit.M, error) {
		res := []codekit.M{}
		h.Gets(new(tenantcoremodel.ItemGroup), dbflex.NewQueryParam().SetSort("Name"), &res)
		return res, nil
	},
	new(tenantcoremodel.SpecGrade).TableName(): func(h *datahub.Hub) ([]codekit.M, error) {
		res := []codekit.M{}
		h.Gets(new(tenantcoremodel.SpecGrade), dbflex.NewQueryParam().SetSort("Name"), &res)
		return res, nil
	},
	new(tenantcoremodel.SpecSize).TableName(): func(h *datahub.Hub) ([]codekit.M, error) {
		res := []codekit.M{}
		h.Gets(new(tenantcoremodel.SpecSize), dbflex.NewQueryParam().SetSort("Name"), &res)
		return res, nil
	},
	new(tenantcoremodel.SpecVariant).TableName(): func(h *datahub.Hub) ([]codekit.M, error) {
		res := []codekit.M{}
		h.Gets(new(tenantcoremodel.SpecVariant), dbflex.NewQueryParam().SetSort("Name"), &res)
		return res, nil
	},
	new(tenantcoremodel.Item).TableName(): func(h *datahub.Hub) ([]codekit.M, error) {
		res := []codekit.M{}
		h.Gets(new(tenantcoremodel.Item), dbflex.NewQueryParam().SetSort("Name"), &res)
		return res, nil
	},
}

func getListModels() []ListModelResult {
	res := []ListModelResult{
		{
			Key:     new(tenantcoremodel.ExpenseType).TableName(),
			Name:    "Expense Type",
			Service: "/tenant/expensetype",
			Model:   new(tenantcoremodel.ExpenseType),
		},
		{
			Key:     new(tenantcoremodel.ExpenseTypeGroup).TableName(),
			Name:    "Expense Type Group",
			Service: "/tenant/expensetypegroup",
			Model:   new(tenantcoremodel.ExpenseTypeGroup),
		},
		{
			Key:      new(tenantcoremodel.ItemGroup).TableName(),
			Name:     "Item Group",
			Service:  "/tenant/itemgroup",
			Download: "true",
			Model:    new(tenantcoremodel.ItemGroup),
		},
		{
			Key:      new(tenantcoremodel.SpecGrade).TableName(),
			Name:     "Spec Grade",
			Service:  "/tenant/specgrade",
			Download: "true",
			Model:    new(tenantcoremodel.SpecGrade),
		},
		{
			Key:      new(tenantcoremodel.SpecSize).TableName(),
			Name:     "Spec Size",
			Service:  "/tenant/specsize",
			Download: "true",
			Model:    new(tenantcoremodel.SpecSize),
		},
		{
			Key:      new(tenantcoremodel.SpecVariant).TableName(),
			Name:     "Spec Variant",
			Service:  "/tenant/specvariant",
			Download: "true",
			Model:    new(tenantcoremodel.SpecVariant),
		},
		// {
		// 	Key:     new(ficomodel.TaxSetup).TableName(),
		// 	Name:    "Tax Code",
		// 	Service: "/fico/taxsetup",
		// 	Model:   new(ficomodel.TaxSetup),
		// },
		{
			Key:      new(tenantcoremodel.Item).TableName(),
			Name:     "Item",
			Service:  "/tenant/item",
			Download: "true",
			Model:    new(tenantcoremodel.Item),
		},
		{
			Key:      new(tenantcoremodel.CustomItemDownload).TableName(),
			Name:     "Item with SKU",
			Service:  "/tenant/custom-item-download",
			Download: "true",
			// Model:    new(tenantcoremodel.CustomItemDownload),
			Model: new(tenantcorelogic.ItemSpecSearchResult),
		},
	}

	return res
}

func getHeaderNameAndType(vModel orm.DataModel) ([]string, []string) {
	v := reflect.ValueOf(vModel)
	headerNames := []string{}
	headerTypes := []string{}
	for i := 0; i < v.Elem().NumField(); i++ {
		fieldName := v.Elem().Type().Field(i).Name
		fieldType := v.Elem().Type().Field(i).Type

		if i > 0 {
			if !strings.Contains(fieldName, "Dimension") && (fieldName != "Created" || fieldType.String() != "time.Time") && (fieldName != "LastUpdate" || fieldType.String() != "time.Time") {
				headerNames = append(headerNames, lo.Ternary(fieldName == "ID", "_id", fieldName))
				headerTypes = append(headerTypes, fieldType.String())
			}
		}
	}

	return headerNames, headerTypes
}

func convertValue(typeDate string, valueData string) (interface{}, error) {
	var res interface{}
	var err error
	if typeDate == "bool" {
		if strings.ToLower(valueData) == "true" {
			res = true
		} else {
			res = false
		}
	} else if typeDate == "int" {
		res, err = strconv.Atoi(valueData)
		if err != nil {
			return res, fmt.Errorf("error : convert string to int : %s", valueData)
		}
	} else if typeDate == "float64" {
		res, err = strconv.ParseFloat(valueData, 64)
		if err != nil {
			return res, fmt.Errorf("error : convert string to float64 : %s", valueData)
		}
	} else if typeDate == "tenantcoremodel.ItemType" {
		res = tenantcoremodel.ItemType(valueData)
	} else if typeDate == "tenantcoremodel.CostUnitCalcMethod" {
		res = tenantcoremodel.CostUnitCalcMethod(valueData)
	} else {
		res = valueData
	}

	return res, err
}

func (o *MasterUploadEngine) GetListModels(ctx *kaos.Context, payload *MURequest) ([]ListModelResult, error) {
	res := getListModels()

	return res, nil
}

func (o *MasterUploadEngine) DownloadTemplateExcel(ctx *kaos.Context, payload *MURequest) (*ResponseHttp, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	var vModel orm.DataModel

	listModels := getListModels()

	mapListModels := lo.Associate(listModels, func(list ListModelResult) (string, ListModelResult) {
		return list.Key, list
	})

	if v, ok := mapListModels[payload.TableName]; ok {
		vModel = v.Model
	} else {
		return nil, errors.New("nothing matches with list model ")
	}

	headerNames, headerTypes := getHeaderNameAndType(vModel)

	t := time.Now()
	strdate := t.Format("20060102")
	filename := fmt.Sprintf("%s_%s.xlsx", payload.TableName, strdate)

	truckPerformancesData := []codekit.M{}

	// Craete new response
	var err error
	kr := NewResponseHttp()
	kr.Header.Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	kr.Header.Add("Content-Type", GetMimeByExtention(".xlsx"))
	kr.Body, err = CreateExcelWithStyleBackgroundColor(headerNames, headerTypes, truckPerformancesData)
	if err != nil {
		return nil, fmt.Errorf("error when create response : %s", err)
	}

	return kr, nil
}

func (o *MasterUploadEngine) Download(ctx *kaos.Context, payload *MURequest) (*ResponseHttp, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	listModels := getListModels()

	mapListModels := lo.Associate(listModels, func(list ListModelResult) (string, ListModelResult) {
		return list.Key, list
	})

	if v, ok := mapListModels[payload.TableName]; ok {
		headerNames, headerTypes := getHeaderNameAndType(v.Model)
		t := time.Now()
		strdate := t.Format("20060102")
		filename := fmt.Sprintf("%s_%s.xlsx", payload.TableName, strdate)

		if downloadFunc, ok := listDownloadFunc[payload.TableName]; !ok {
			return nil, fmt.Errorf("%s not implemented yet", payload.TableName)
		} else {
			res, e := downloadFunc(h)
			if e != nil {
				return nil, e
			}

			if len(res) > 0 {
				kr := NewResponseHttp()
				kr.Header.Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
				kr.Header.Add("Content-Type", GetMimeByExtention(".xlsx"))
				kr.Body, e = CreateExcelWithStyleBackgroundColor(headerNames, headerTypes, res)
				if e != nil {
					return nil, fmt.Errorf("error when create response : %s", e)
				}

				return kr, nil
			}
		}
	} else {
		return nil, errors.New("nothing matches with list model ")
	}

	return nil, errors.New("Fail generate excel download")
}

func (o *MasterUploadEngine) Upload(ctx *kaos.Context, payload *MURequest) (*MUResult, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	res := MUResult{}

	var vModel orm.DataModel
	listModels := getListModels()

	mapListModels := lo.Associate(listModels, func(list ListModelResult) (string, ListModelResult) {
		return list.Key, list
	})

	if v, ok := mapListModels[payload.TableName]; ok {
		vModel = v.Model
	} else {
		return nil, errors.New("nothing matches with list model ")
	}

	headerNames, headerTypes := getHeaderNameAndType(vModel)

	ex, err := ExcelFromBase64(payload.Content)
	if err != nil {
		return nil, err
	}

	exFile, err := excelize.OpenReader(bytes.NewReader(ex))
	if err != nil {
		return nil, fmt.Errorf("cannot open excel file : %s", err)
	}

	now := time.Now()
	rows, err := exFile.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	for k, row := range rows {
		// index k=0 is type
		// index k=1 is header
		if k == 1 {
			for i, cell := range row {
				if strings.TrimSpace(cell) != headerNames[i] {
					return nil, fmt.Errorf("column header did not match. between excel: %s and database: %s", strings.TrimSpace(cell), headerNames[i])
				}
			}
		}
		if k > 1 {
			reflector.From(vModel).Set("Created", now).Flush()
			reflector.From(vModel).Set("LastUpdate", now).Flush()
			for i, colCell := range row {
				if headerNames[i] == "ID" || headerNames[i] == "Name" {
					if colCell == "" {
						return nil, fmt.Errorf("column can't be empty")
					}
				}

				newVal, err := convertValue(headerTypes[i], colCell)
				if err != nil {
					return nil, fmt.Errorf("error : convert data : %s", newVal)
				}
				reflector.From(vModel).Set(headerNames[i], newVal).Flush()
			}
			if e := h.Save(vModel); e != nil {
				return nil, fmt.Errorf("error : save : %s", payload.TableName)
			}
		}
	}

	res.Result = "Success"
	return &res, nil
}
