package shelogic

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
)

type MCUTransactionLogic struct {
}

type MCURequest struct {
	MCUTransaction shemodel.MCUTransaction
	Site           string
	IsResult       bool
}

func (obj *MCUTransactionLogic) GetMcuItemLastChild(ctx *kaos.Context, payload *dbflex.QueryParam) ([]shemodel.MCUItemTemplateLine, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	query := dbflex.NewQueryParam()
	filters := []*dbflex.Filter{}
	var valueVal interface{}

	if payload != nil {
		if payload.Where != nil {
			filters2 := []*dbflex.Filter{}
			if len(payload.Where.Items) > 0 {
				vItems := payload.Where.Items
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
									valueVal = aString[0]
									filters2 = append(filters2, dbflex.ElemMatch("Lines", dbflex.Contains(fieldVal, aString[0])))
								}
							}
						}
					}
				}
			} else {
				fieldVal := payload.Where.Field
				opVal := payload.Where.Op
				valueValType := reflect.TypeOf(payload.Where.Value).Kind()
				if opVal == dbflex.OpEq && valueValType == reflect.String {
					valueVal = payload.Where.Value.(string)
					filters2 = append(filters2, dbflex.ElemMatch("Lines", dbflex.Eq(fieldVal, valueVal)))
					// fmt.Println("val", codekit.JsonString(filters2))
				} else if opVal == dbflex.OpIn && valueValType == reflect.Slice {
					valueVal = payload.Where.Value.([]interface{})
					aString := []string{}
					for _, v := range payload.Where.Value.([]interface{}) {
						aString = append(aString, v.(string))
					}
					filters2 = append(filters2, dbflex.ElemMatch("Lines", dbflex.In(fieldVal, aString...)))
				}
			}
			if len(filters2) > 0 {
				filters = append(filters, dbflex.Or(filters2...))
			}
		}
	}

	if len(filters) > 0 {
		query = query.SetWhere(dbflex.And(filters...))
	}

	items := []shemodel.MCUItemTemplate{}
	itemLines := []shemodel.MCUItemTemplateLine{}
	err := hub.Gets(new(shemodel.MCUItemTemplate), query, &items)
	if err != nil {
		return nil, err
	}

	if len(items) > 0 {
		for _, val := range items {
			if len(val.Lines) > 0 {
				for _, line := range val.Lines {
					if line.Type != "" {
						if len(filters) > 0 {
							valType := reflect.TypeOf(valueVal).Kind()
							if valType == reflect.String {
								if strings.Contains(line.Description, valueVal.(string)) || strings.Contains(line.ID, valueVal.(string)) {
									itemLines = append(itemLines, line)
								}
							} else if valType == reflect.Slice {
								for _, valval := range valueVal.([]interface{}) {
									if strings.Contains(line.Description, valval.(string)) || strings.Contains(line.ID, valval.(string)) {
										itemLines = append(itemLines, line)
									}
								}
							}
						} else {
							itemLines = append(itemLines, line)
						}
					}
				}
			}
		}
	}

	return itemLines, nil
}

func (obj *MCUTransactionLogic) SavePreview(ctx *kaos.Context, payload *MCURequest) (*tenantcoremodel.Preview, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	// get mcu item template
	mapMcuitemtemplates := map[string]string{}
	mcuitemtemplates := []shemodel.MCUItemTemplate{}
	e := hub.Gets(new(shemodel.MCUItemTemplate), nil, &mcuitemtemplates)
	if e != nil {
		return nil, errors.New("Failed populate data master: " + e.Error())
	}

	if len(mcuitemtemplates) > 0 {
		for _, val := range mcuitemtemplates {
			for _, valLine := range val.Lines {
				mapMcuitemtemplates[valLine.ID] = valLine.Description
			}
		}
	}

	res := tenantcoremodel.Preview{}
	res.PreviewReport = &tenantcoremodel.PreviewReport{
		Header: make(codekit.M),
	}
	vSourceType := "MCU"
	additionalItemArr := []string{}
	// get preview
	tenantPreview := []*tenantcoremodel.Preview{}
	if payload.IsResult {
		// save preview result with key SourceType = "MCU"
		e := hub.Gets(new(tenantcoremodel.Preview),
			dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.Eq("SourceType", vSourceType),
					dbflex.Eq("SourceJournalID", payload.MCUTransaction.ID),
				)),
			&tenantPreview)
		if e != nil {
			return nil, errors.New("Failed populate data preview: " + e.Error())
		}

		for _, val := range payload.MCUTransaction.AdditionalItem {
			if v, ok := mapMcuitemtemplates[val]; ok {
				additionalItemArr = append(additionalItemArr, v)
			}
		}
	} else {
		vSourceType = "MCU_FOLLOWUP"
		// save preview follow up with key SourceType = "MCU_FOLLOWUP"
		e := hub.Gets(new(tenantcoremodel.Preview),
			dbflex.NewQueryParam().SetWhere(
				dbflex.And(
					dbflex.Eq("SourceType", vSourceType),
					dbflex.Eq("SourceJournalID", payload.MCUTransaction.ID),
				)),
			&tenantPreview)
		if e != nil {
			return nil, errors.New("Failed populate data preview: " + e.Error())
		}

		// set last follow up additional item
		for i, val := range payload.MCUTransaction.FollowUp {
			if i == len(payload.MCUTransaction.FollowUp) {
				for _, valAdd := range val.AdditionalItem {
					if v, ok := mapMcuitemtemplates[valAdd]; ok {
						additionalItemArr = append(additionalItemArr, v)
					}
				}
			}
		}
	}

	if len(tenantPreview) > 0 {
		res.ID = tenantPreview[0].ID
	}

	// get master data
	masterDatas := []tenantcoremodel.MasterData{}
	e = hub.Gets(new(tenantcoremodel.MasterData), nil, &masterDatas)
	if e != nil {
		return nil, errors.New("Failed populate data master: " + e.Error())
	}

	mapMasterDatas := lo.Associate(masterDatas, func(masterData tenantcoremodel.MasterData) (string, tenantcoremodel.MasterData) {
		return masterData.ID, masterData
	})

	Provider := ""
	if val, ok := mapMasterDatas[payload.MCUTransaction.Provider]; ok {
		Provider = val.Name
	}
	Gender := ""
	if val, ok := mapMasterDatas[payload.MCUTransaction.Gender]; ok {
		Gender = val.Name
	}
	Position := ""
	if val, ok := mapMasterDatas[payload.MCUTransaction.Position]; ok {
		Position = val.Name
	}

	emp := new(tenantcoremodel.Employee)
	if e := hub.GetByID(emp, payload.MCUTransaction.Name); e != nil {
		return nil, errors.New("Failed populate data employee: " + e.Error())
	}

	mcuPakage := new(shemodel.MCUMasterPackage)
	if e := hub.GetByID(mcuPakage, payload.MCUTransaction.MCUPackage); e != nil {
		return nil, errors.New("Failed populate data mcu master package: " + e.Error())
	}

	additionalItem := strings.Join(additionalItemArr, ", ")
	res.Name = "Default"
	res.SourceType = vSourceType
	res.SourceJournalID = payload.MCUTransaction.ID

	// set header
	var datas [][]string
	datas = append(datas, []string{"", "", fmt.Sprintf("Kepanjen, %s :R", payload.MCUTransaction.Date.Format("02 January 2006"))})
	datas = append(datas, []string{"Nomor ", fmt.Sprintf(": %s", payload.MCUTransaction.ID), ""})
	datas = append(datas, []string{"Perihal ", ": Surat Pengantar MCU", ""})
	datas = append(datas, []string{"Kepada ", fmt.Sprintf(": Lab. %s", Provider), "", "", "", ""})

	var footer [][]string
	footer = append(footer, []string{payload.MCUTransaction.DoctorParamedic, "", ""})

	var section [][]string
	section = append(section, []string{"Nama ", fmt.Sprintf(": %s", emp.Name), ""})
	section = append(section, []string{"Jenis Kelamin ", fmt.Sprintf(": %s", Gender), ""})
	section = append(section, []string{"Usia ", fmt.Sprintf(": %s", strconv.Itoa(payload.MCUTransaction.Age)), ""})
	section = append(section, []string{"Jabatan ", fmt.Sprintf(": %s", Position), ""})
	section = append(section, []string{"Site ", fmt.Sprintf(": %s", payload.Site), ""})
	section = append(section, []string{"Paket MCU ", fmt.Sprintf(": %s", mcuPakage.PackageName), ""})
	section = append(section, []string{additionalItem, "", ""})

	sectionLine := tenantcoremodel.PreviewSection{
		SectionType: tenantcoremodel.PreviewAsGrid,
		Title:       "MCU",
		Items:       section,
	}

	res.PreviewReport.Header["Data"] = datas
	res.PreviewReport.Header["Footer"] = footer
	res.PreviewReport.Sections = append(res.PreviewReport.Sections, sectionLine)

	if e := hub.Save(&res); e != nil {
		return nil, errors.New("error save mcu preview: " + e.Error())
	}

	return &res, nil
}
