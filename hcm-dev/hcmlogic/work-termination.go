package hcmlogic

import (
	"errors"
	"fmt"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/she/shemodel"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/samber/lo"
)

type WorkTerminationHandler struct {
}

func (m *WorkTerminationHandler) GetDetail(ctx *kaos.Context, ids []interface{}) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	wt := new(hcmmodel.WorkTermination)
	if err := h.GetByID(wt, ids...); err != nil {
		return nil, fmt.Errorf("error when get work termination: %s", err.Error())
	}

	site := new(bagongmodel.Site)
	if err := h.GetByID(site, wt.Dimension.Get("Site")); err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}

	employee := new(tenantcoremodel.Employee)
	if err := h.GetByID(employee, wt.EmployeeID); err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}

	empDetail := new(bagongmodel.EmployeeDetail)
	if err := h.GetByFilter(empDetail, dbflex.Eq("EmployeeID", wt.EmployeeID)); err != nil {
		return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
	}

	// check is last salary calculated
	isLastSalaryCalculated := false
	lastSalaryIndex := -1
	for index, i := range wt.NonTaxableIncome {
		if i.Name == "Upah Terakhir" {
			lastSalaryIndex = index
			if i.Amount != 0 {
				isLastSalaryCalculated = true
				break
			}
		}
	}

	// calculate and fill last salary
	if !isLastSalaryCalculated && lastSalaryIndex > -1 {
		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			return nil, fmt.Errorf("error when get convert timezone: %s", err.Error())
		}

		cutOffSalaryDate := time.Date(wt.ResignDate.Year(), wt.ResignDate.Month()-1, 24, 0, 0, 0, 0, loc)
		dayDifference := int(wt.ResignDate.Sub(cutOffSalaryDate).Hours()/24) + 1
		lastSalary := site.Configuration.UMK * float64(dayDifference) / 30

		wt.NonTaxableIncome[lastSalaryIndex].Amount = lastSalary
	}

	// check severance pay, award
	severanceIndex := -1
	isSeverancePayCalculated := false
	awardIndex := -1
	isAwardCalculated := false
	for index, i := range wt.TaxableIncome {
		if i.Name == "Uang Pesangon" {
			severanceIndex = index
			if i.Amount != 0 {
				isSeverancePayCalculated = true
			}
		} else if i.Name == "Uang Penghargaan Masa Kerja" {
			awardIndex = index
			if i.Amount != 0 {
				isAwardCalculated = true
			}
		}
	}

	// calculate and fill severance pay
	yearDifference := wt.ResignDate.Sub(employee.JoinDate).Hours() / 24 / 365.25
	if !isSeverancePayCalculated && severanceIndex > -1 {
		times := 0.0
		if yearDifference < 1 {
			times = 1
		} else if yearDifference < 2 {
			times = 2
		} else if yearDifference < 3 {
			times = 3
		} else if yearDifference < 4 {
			times = 4
		} else if yearDifference < 5 {
			times = 5
		} else if yearDifference < 6 {
			times = 6
		} else if yearDifference < 7 {
			times = 7
		} else if yearDifference < 8 {
			times = 8
		} else {
			times = 9
		}

		reasonConvert := 0.0
		switch wt.Type {
		case "PHK":
			reasonConvert = 0.5
		case "Pensiun":
			reasonConvert = 1.75
		case "SakitBerkepanjangan":
			reasonConvert = 2
		case "Meninggal":
			reasonConvert = 2
		}
		wt.TaxableIncome[severanceIndex].Calculation = fmt.Sprintf("%.0f x %.0f x %.2f", site.Configuration.UMK, times, reasonConvert)
		wt.TaxableIncome[severanceIndex].Amount = site.Configuration.UMK * times * reasonConvert
	}

	// calculate and fill award
	if !isAwardCalculated && awardIndex > -1 {
		times := 0.0
		if yearDifference >= 21 {
			times = 8
		} else if yearDifference >= 18 {
			times = 7
		} else if yearDifference >= 15 {
			times = 6
		} else if yearDifference >= 12 {
			times = 5
		} else if yearDifference >= 9 {
			times = 4
		} else if yearDifference >= 6 {
			times = 3
		} else if yearDifference >= 3 {
			times = 2
		}

		wt.TaxableIncome[awardIndex].Calculation = fmt.Sprintf("%.0f x %.0f x 1", site.Configuration.UMK, times)
		wt.TaxableIncome[awardIndex].Amount = site.Configuration.UMK * times
	}

	masterIDs := []string{empDetail.Grade, empDetail.Position, empDetail.Department}
	masters := make([]tenantcoremodel.MasterData, 0)
	if err := h.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(dbflex.In("_id", masterIDs...)), &masters); err != nil {
		return nil, fmt.Errorf("error when get master: %s", err.Error())
	}
	mapMaster := lo.Associate(masters, func(detail tenantcoremodel.MasterData) (string, string) {
		return detail.ID, detail.Name
	})
	result := struct {
		*hcmmodel.WorkTermination
		Name         string
		NIK          string
		Grade        string
		Position     string
		Department   string
		Address      string
		Education    string
		JoinDate     time.Time
		ResignDate   time.Time
		WorkDuration float64
		Salary       float64
	}{
		WorkTermination: wt,
		Name:            employee.Name,
		NIK:             empDetail.EmployeeNo,
		Grade:           mapMaster[empDetail.Grade],
		Position:        mapMaster[empDetail.Position],
		Department:      mapMaster[empDetail.Department],
		Address:         empDetail.Address,
		Education:       empDetail.LastEducation,
		JoinDate:        employee.JoinDate,
		ResignDate:      wt.ResignDate,
		WorkDuration:    yearDifference,
		Salary:          site.Configuration.UMK,
	}

	return result, nil
}

func (m *WorkTerminationHandler) Insert(ctx *kaos.Context, model *hcmmodel.WorkTermination) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	templates := []shemodel.MCUItemTemplate{}
	err := h.Gets(new(shemodel.MCUItemTemplate), dbflex.NewQueryParam().SetWhere(
		dbflex.In("_id", []string{"67502f706f28e3712e4818cb", "67502d946f28e3712e4818c0", "67501b6c6f28e3712e4818a9"}...),
	), &templates)
	if err != nil {
		return nil, fmt.Errorf("error when get item template: %s", err.Error())
	}

	journalType := new(hcmmodel.JournalType)
	err = h.GetByID(journalType, model.JournalTypeID)
	if err != nil {
		return false, fmt.Errorf("error when get journal type: %s", err.Error())
	}

	kind := ""
	switch journalType.TransactionType {
	case "Work Termination - Resign":
		kind = "WorkTerminationResign"
	case "Work Termination - PHK":
		kind = "WorkTerminationPHK"
	case "Work Termination - Sakit Berkepanjangan":
		kind = "WorkTerminationSakit"
	case "Work Termination - Meninggal":
		kind = "WorkTerminationMeninggal"
	case "Work Termination - Pensiun":
		kind = "WorkTerminationPensiun"
	}

	for _, t := range templates {
		switch t.Name {
		case "Work Termination - Severance - Kewajiban Pekerja":
			model.MandatoryWorkTemplateID = t.ID
			model.MandatoryWork = make([]hcmmodel.MandatoryWorkDetail, len(t.Lines))
			for i, l := range t.Lines {
				model.MandatoryWork[i] = hcmmodel.MandatoryWorkDetail{
					Number: l.Number,
					Name:   l.Description,
				}
			}
		case "Work Termination - Severance - Hak Pekerja Part II":
			model.TaxableIncomeTemplateID = t.ID
			model.TaxableIncome = make([]hcmmodel.TaxableIncomeDetail, len(t.Lines))
			for i, l := range t.Lines {
				model.TaxableIncome[i] = hcmmodel.TaxableIncomeDetail{
					Number: l.Number,
					Name:   l.Description,
				}
			}
		case "Work Termination - Severance - Hak Pekerja Part I":
			model.NonTaxableIncomeTemplateID = t.ID
			model.NonTaxableIncome = make([]hcmmodel.NonTaxableIncomeDetail, len(t.Lines))
			for i, l := range t.Lines {
				model.NonTaxableIncome[i] = hcmmodel.NonTaxableIncomeDetail{
					Number: l.Number,
					Name:   l.Description,
				}
			}
		}
	}

	tenantcorelogic.MWPreAssignSequenceNo(kind, false, "")(ctx, model)
	err = h.Insert(model)
	if err != nil {
		return nil, fmt.Errorf("error when insert: %s", err.Error())
	}

	return model, nil
}
