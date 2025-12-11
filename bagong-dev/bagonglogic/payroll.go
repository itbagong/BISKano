package bagonglogic

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficologic"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/kara/karamodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcorelogic"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
	"github.com/ariefdarmawan/datahub"
	"github.com/samber/lo"
	"github.com/sebarcode/codekit"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayrollHandler struct{}
type PayrollDetailRequest struct {
	SiteID    string
	DateStart *time.Time
	DateEnd   *time.Time
	Take      int
	Skip      int
}
type PayrollSummaryRespItem struct {
	SiteID        string
	Period        string
	Name          string
	BaseSalary    float64
	Benefit       []bagongmodel.DetailBenDeduct
	Deduction     []bagongmodel.DetailBenDeduct
	AttendanceNum int
}
type PayrollSummaryDetailRespItem struct {
	Details []PayrollSummaryRespItem
}
type PayrollDetailResponse struct {
	Details []bagongmodel.BGPayrollDetail
}
type PostSubmissionRequest struct {
	bagongmodel.BGPayrollSubmission
	Lines         []ficomodel.JournalLine
	CompanyID     string
	PayrollPeriod string
}

func (obj *PayrollHandler) PostSubmission(ctx *kaos.Context, payload PostSubmissionRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "NOK", errors.New("missing: connection")
	}

	if payload.JournalType == "" {
		return nil, fmt.Errorf("please choose journal type")
	}

	journalType := new(ficomodel.LedgerJournalType)
	err := hub.GetByID(journalType, payload.JournalType)
	if err != nil {
		return nil, fmt.Errorf("error when get journal type: %s", err.Error())
	}

	journal := new(ficomodel.LedgerJournal)
	if payload.ID != "" {
		err = hub.GetByID(journal, payload.ID)
		if err != nil {
			return nil, fmt.Errorf("error when get journal: %s", err.Error())
		}
	}

	reference := new(tenantcoremodel.ReferenceTemplate)
	hub.GetByID(reference, journalType.ReferenceTemplateID)
	// if err != nil {
	// 	return nil, fmt.Errorf("error when get reference type: %s", err.Error())
	// }

	for i := range payload.Lines {
		payload.Lines[i].OffsetAccount.AccountID = "211004"
		payload.Lines[i].OffsetAccount.AccountType = ficomodel.SubledgerAccounting
	}

	journal.JournalTypeID = payload.JournalType
	journal.TrxDate = payload.SubmissionDate
	journal.Dimension = payload.Dimension
	journal.PostingProfileID = journalType.PostingProfileID
	journal.Text = payload.Text
	journal.DefaultOffset = journalType.DefaultOffset
	journal.Status = ficomodel.JournalStatus(payload.Status)
	journal.Lines = payload.Lines
	journal.ReferenceTemplateID = journalType.ReferenceTemplateID

	if payload.CompanyID != "" {
		journal.CompanyID = payload.CompanyID
	} else {
		coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
		journal.CompanyID = coID
	}

	if len(reference.Items) > 0 {
		journal.References = journal.References.Set(reference.Items[0].Label, "Payroll")
	}

	if payload.ID == "" {
		tenantcorelogic.MWPreAssignCustomSequenceNo("PayrollSubmission")(ctx, journal)
	} else {
		journal.ID = payload.ID
	}

	if payload.PayrollPeriod != "" {
		journal.References = journal.References.Set("PayrollPeriod", payload.PayrollPeriod)
	}

	err = hub.Save(journal)
	if err != nil {
		return nil, fmt.Errorf("error when save payroll journal: %s", err.Error())
	}
	return journal, nil
}

func (obj *PayrollHandler) Post(ctx *kaos.Context, payload []ficologic.PostRequest) ([]*tenantcoremodel.PreviewReport, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	result, err := new(ficologic.PostingProfileHandler).Post(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("error when post journal: %s", err.Error())
	}

	journal := new(ficomodel.LedgerJournal)
	err = hub.GetByID(journal, payload[0].JournalID)
	if err != nil {
		return nil, fmt.Errorf("error when get journal: %s", err.Error())
	}

	if journal.Status == ficomodel.JournalStatusPosted {
		if len(journal.Lines) > 0 {
			ids := make([]string, len(journal.Lines))
			for i, d := range journal.Lines {
				ids[i] = strings.Split(d.Text, " - ")[0]
			}

			loans := make([]hcmmodel.Loan, 0)
			err = hub.Gets(new(hcmmodel.Loan), dbflex.NewQueryParam().SetWhere(
				dbflex.In("EmployeeID", ids...),
			), &loans)
			if err != nil {
				return nil, fmt.Errorf("error when get loan: %s", err.Error())
			}

			payrollPeriod := journal.References.Get("PayrollPeriod", "").(string)
			if payrollPeriod == "" {
				return nil, errors.New("payroll period is empty")
			}

			coID := tenantcorelogic.GetCompanyIDFromContext(ctx)
			if ctx.Data().Get("CompanyID", "").(string) != "" {
				coID = ctx.Data().Get("CompanyID", "").(string)
			}

			// get company
			company, err := datahub.GetByParm(hub, new(tenantcoremodel.Company), dbflex.NewQueryParam().
				SetWhere(dbflex.Eqs("_id", coID)))
			if err != nil {
				return nil, fmt.Errorf("error when get company: %s", err)
			}

			loc, err := time.LoadLocation(company.LocationCode)
			if err != nil {
				return nil, fmt.Errorf("error convert company location: %s", err)
			}

			period, err := time.Parse("2006-01", payrollPeriod)
			if err != nil {
				return nil, fmt.Errorf("error when get loan: %s", err.Error())
			}

			oneMonthBefore := period.AddDate(0, -1, 0).In(loc)
			twoMonthBefore := period.AddDate(0, -2, 0).In(loc)
			startDate := time.Date(twoMonthBefore.Year(), twoMonthBefore.Month(), 24, 0, 0, 0, 0, period.Location())
			endDate := time.Date(oneMonthBefore.Year(), oneMonthBefore.Month(), 23, 0, 0, 0, 0, period.Location())

			// update status
			for _, l := range loans {
				for i, d := range l.Lines {
					if (d.Date.After(startDate) || d.Date.Equal(startDate)) && (d.Date.Before(endDate) || d.Date.Equal(endDate)) {
						l.Lines[i].Status = hcmmodel.LoanPaid
					}
				}

				hub.Update(&l, "Lines")
			}
		}
	}

	return result, nil
}

type TotalResult struct {
	Total float64
}

func (obj *PayrollHandler) GetSubmissionTotal(ctx *kaos.Context, payload PayrollDetailRequest) (TotalResult, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return TotalResult{Total: 0}, errors.New("missing: connection")
	}
	period := payload.DateStart.Format("200601")
	filterDetail := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{
		dbflex.NewFilter("SiteID", dbflex.OpEq, payload.SiteID, nil),
		dbflex.NewFilter("Period", dbflex.OpEq, period, nil),
	})

	detail := []bagongmodel.BGPayrollDetail{}
	err := hub.GetsByFilter(&bagongmodel.BGPayrollDetail{}, filterDetail, &detail)
	if err != nil {
		return TotalResult{Total: 0}, err
	}
	total := float64(0)
	for _, det := range detail {
		total += det.GetTakeHome()
	}
	return TotalResult{Total: total}, nil
}
func (obj *PayrollHandler) GetSubmission(ctx *kaos.Context, payload dbflex.QueryParam) ([]bagongmodel.BGPayrollSubmission, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}
	submission := bagongmodel.BGPayrollSubmission{}
	dest := []bagongmodel.BGPayrollSubmission{}
	err := hub.Gets(&submission, &payload, &dest)
	//err := hub.GetsByFilter(&submission, &payload, &dest)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return dest, nil
}
func GetEmployeeAttendance(event kaos.EventHub, userid string, dateStart, dateEnd time.Time) (int, error) {
	res := []map[string]interface{}{}
	filterDateStart := dbflex.NewFilter("TrxDate", dbflex.OpGte, dateStart, nil)
	filterDateEnd := dbflex.NewFilter("TrxDate", dbflex.OpLte, dateEnd, nil)
	filterUserId := dbflex.NewFilter("UserID", dbflex.OpEq, userid, nil)
	// fmt.Println(date1.Local(), date2.Local(), val.UserID)
	allFilter := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{filterUserId, filterDateEnd, filterDateStart})
	// karaAttendanceTrxQuery := dbflex.NewQueryParam()
	// karaAttendanceTrxQuery.Where = allFilter
	err := event.Publish("/v1/kara/attendancetrx/find", allFilter, &res, nil)
	if err != nil {
		return 0, err
	}
	return len(res), nil
}
func GetEmployee(event kaos.EventHub, payload PayrollDetailRequest) ([]tenantcoremodel.Employee, error) {
	respEvt := []tenantcoremodel.Employee{}
	siteFilter := dbflex.NewFilter("Dimension.Key", dbflex.OpEq, "Site", nil)
	siteIdFilter := dbflex.NewFilter("Dimension.Value", dbflex.OpEq, payload.SiteID, nil)
	allFilter := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{siteFilter, siteIdFilter})

	tenantEmpQuery := dbflex.NewQueryParam()
	// filterSite := dbflex.NewFilter("Dimension.Key", dbflex.OpEq, "Site", nil)
	// siteIdFilter = dbflex.NewFilter("Dimension.Value", dbflex.OpEq, payload.SiteID, nil)
	// allFilter = dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{filterSite, siteIdFilter})
	tenantEmpQuery.Where = allFilter
	tenantEmpQuery.Take = payload.Take
	tenantEmpQuery.Skip = payload.Skip
	err := event.Publish("/v1/tenant/employee/find", tenantEmpQuery, &respEvt, nil)
	if err != nil {
		return nil, err
	}
	return respEvt, nil
}
func (obj *PayrollHandler) Save(ctx *kaos.Context, payload PayrollDetailResponse) (string, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return "", errors.New("missing: connection")
	}
	for _, detail := range payload.Details {
		xx := bagongmodel.BGPayrollDetail{}
		filterDetail := dbflex.NewFilter("", dbflex.OpAnd, nil, []*dbflex.Filter{
			dbflex.NewFilter("EmployeeID", dbflex.OpEq, detail.EmployeeID, nil),
			dbflex.NewFilter("Period", dbflex.OpEq, detail.Period, nil),
		})
		err := hub.GetByFilter(&xx, filterDetail)
		if err != nil && err != io.EOF {
			return "NOK", err
		}
		detail.ID = xx.ID
		err = hub.Save(&detail)
		if err != nil {
			return "NOK", err
		}
	}
	return "OK", nil
}

type SiteDetail struct {
	SalaryUsed string
	UMK        float64
}

func (obj *PayrollHandler) GetSite(ctx *kaos.Context, payload *PayrollDetailRequest) (*PayrollSummaryDetailRespItem, error) {
	response := &PayrollSummaryDetailRespItem{Details: []PayrollSummaryRespItem{}}
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	jwtData := ctx.Data().Get("jwt_data", codekit.M{}).(codekit.M)
	dimIface := jwtData.Get("Dimension", []interface{}{}).([]interface{})

	filter := []*dbflex.Filter{
		dbflex.Eq("IsActive", true),
	}

	if len(dimIface) > 0 {
		filter = append(filter, GetFilterDimension(dimIface)...)
	}

	if payload.SiteID != "" {
		filter = append(filter, dbflex.Eq("_id", payload.SiteID))
	}

	// get sites
	sites := []bagongmodel.Site{}
	err := hub.Gets(new(bagongmodel.Site), dbflex.NewQueryParam().SetWhere(
		dbflex.And(filter...),
	), &sites)
	if err != nil {
		return nil, fmt.Errorf("error when get sites: %s", err.Error())
	}

	mapSiteUMK := map[string]*SiteDetail{}
	mapOvertimeConfig := map[string]bagongmodel.Overtime{}
	siteIDs := make([]string, len(sites))
	for i, v := range sites {
		siteIDs[i] = v.ID
		mapSiteUMK[v.ID] = &SiteDetail{
			SalaryUsed: v.SalaryUsed,
			UMK:        v.Configuration.UMK,
		}
		for _, l := range v.Overtime {
			mapOvertimeConfig[v.ID+l.Position] = l
		}
	}

	hoSite := new(bagongmodel.Site)
	err = hub.GetByID(hoSite, "SITE020")
	if err != nil {
		return nil, fmt.Errorf("error when get HO Site: %s", err.Error())
	}

	// get tenant employee
	tenantEmployees := []tenantcoremodel.Employee{}
	err = hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.In("Value", siteIDs...)),
		),
	), &tenantEmployees)
	if err != nil {
		return nil, fmt.Errorf("error when tenant employee: %s", err.Error())
	}
	employeeIDs := lo.Map(tenantEmployees, func(e tenantcoremodel.Employee, index int) string {
		return e.ID
	})

	// build map employee per site
	mapTenantEmployeePerSite := map[string][]string{}
	for _, te := range tenantEmployees {
		site := te.Dimension.Get("Site")
		mapTenantEmployeePerSite[site] = append(mapTenantEmployeePerSite[site], te.ID)
	}

	// get employee detail
	employees := []bagongmodel.EmployeeDetail{}
	err = hub.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("EmployeeID", employeeIDs...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}
	mapEmployeeDetail := lo.Associate(employees, func(detail bagongmodel.EmployeeDetail) (string, bagongmodel.EmployeeDetail) {
		return detail.EmployeeID, detail
	})

	// get master data
	masters := []tenantcoremodel.MasterData{}
	err = hub.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("MasterDataTypeID", []interface{}{"Relation", "GDE", "PBJTK"}),
		),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}
	mapGrade := map[string]string{}
	mapRelation := map[string]string{}
	mapPBJTK := map[string]string{}
	for _, d := range masters {
		switch d.MasterDataTypeID {
		case "Relation":
			mapRelation[d.ID] = d.Name
		case "GDE":
			mapGrade[d.ID] = d.Name
		case "PBJTK":
			mapPBJTK[d.ID] = d.Name
		}
	}

	date1 := payload.DateEnd
	period := date1.Format("200601")
	// get payrol detail
	payrollDetails := []bagongmodel.BGPayrollDetail{}
	err = hub.Gets(new(bagongmodel.BGPayrollDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("EmployeeID", employeeIDs...),
			dbflex.Eq("Period", period),
		),
	), &payrollDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get payroll detail: %s", err.Error())
	}

	mapPayrollDetailPerEmployee := lo.Associate(payrollDetails, func(detail bagongmodel.BGPayrollDetail) (string, bagongmodel.BGPayrollDetail) {
		return detail.EmployeeID, detail
	})

	mapWorkDay, err := obj.getWorkDays(hub, siteIDs, payload)
	if err != nil {
		return nil, err
	}

	monthBefore := codekit.FirstOfMonth(*payload.DateEnd).AddDate(0, -1, 0)
	twoMonthBefore := codekit.FirstOfMonth(*payload.DateEnd).AddDate(0, -2, 0)
	start := time.Date(twoMonthBefore.Year(), twoMonthBefore.Month(), 24, 0, 0, 0, 0, twoMonthBefore.Location())
	end := time.Date(monthBefore.Year(), monthBefore.Month(), 23, 0, 0, 0, 0, monthBefore.Location())

	mapOvertime, err := obj.getOvertimes(hub, mapOvertimeConfig, mapEmployeeDetail, mapSiteUMK, siteIDs, &start, &end)
	if err != nil {
		return nil, err
	}

	mapMealAllowance, err := obj.getMealAllowances(hub, siteIDs, &start, &end)
	if err != nil {
		return nil, err
	}

	mapLoan, err := obj.getLoans(hub, siteIDs, employeeIDs, &start, &end)
	if err != nil {
		return nil, err
	}

	for _, site := range sites {
		benMap := map[string]float64{}
		baseTotal := float64(0)
		deductMap := map[string]float64{}
		allowance := 0.0
		for _, b := range site.Benefits {
			if b.ID == "UangMakan" {
				allowance = b.Value
				break
			}
		}

		if employees, ok := mapTenantEmployeePerSite[site.ID]; ok {
			pp := PayrollSummaryRespItem{}
			for _, id := range employees {
				empDetail := mapEmployeeDetail[id]
				workDay := float64(mapWorkDay[site.ID+id])
				if pr, ok := mapPayrollDetailPerEmployee[id]; ok {
					baseTotal += pr.BaseSalary

					for _, ben := range site.Benefits {
						if ben.IsCash {
							if pr.CheckBenefitExists(ben.ID) {
								for _, val := range pr.Benefits {
									if val.ID == ben.ID {
										benMap[ben.ID] += val.Amount
										break
									}
								}
							} else {
								if _, ok := benMap[ben.ID]; !ok {
									benMap[ben.ID] = 0
								}

								if ben.ID == "Overtime" {
									benMap[ben.ID] += mapOvertime[site.ID+id]
								} else if ben.ID == "UangMakan" {
									benMap[ben.ID] += float64(mapMealAllowance[site.ID+id]) * allowance
								} else {
									if ben.CalcType == ficomodel.PayrollComponentPercentage {
										benMap[ben.ID] += ben.Value / 100 * pr.BaseSalary
									} else if ben.CalcType == ficomodel.PayrollComponentDaily {
										benMap[ben.ID] += ben.Value * workDay
									} else {
										benMap[ben.ID] += ben.Value
									}
								}
							}
						}
					}

					for _, ded := range site.Deductions {
						if ded.IsCash {
							if pr.CheckDeductionExists(ded.ID) {
								for _, val := range pr.Benefits {
									if val.ID == ded.ID {
										deductMap[ded.ID] += val.Amount
										break
									}
								}
							} else {
								if _, ok := deductMap[ded.ID]; !ok {
									deductMap[ded.ID] = 0
								}

								switch ded.ID {
								case "BPJSKS":
									deductMap[ded.ID] += obj.calculateHealthBPJS(&empDetail, &site, mapGrade[empDetail.Grade], mapRelation)
								case "BPJSTK":
									deductMap[ded.ID] += obj.calculateEmploymentBPJS(&empDetail, mapGrade[empDetail.Grade], mapPBJTK[empDetail.BPJSTKProgram], hoSite.Configuration.UMK)
								case "Loan":
									deductMap[ded.ID] += mapLoan[site.ID+id]
								default:
									if ded.CalcType == ficomodel.PayrollComponentPercentage {
										deductMap[ded.ID] += ded.Value / 100 * pr.BaseSalary
									} else if ded.CalcType == ficomodel.PayrollComponentDaily {
										deductMap[ded.ID] += ded.Value * workDay
									} else {
										deductMap[ded.ID] += ded.Value
									}
								}
							}
						}
					}
				} else {
					if empDetail, ok := mapEmployeeDetail[id]; ok {
						baseTotal += empDetail.BasicSalary

						for _, ben := range site.Benefits {
							if ben.IsCash {
								if _, ok := benMap[ben.ID]; !ok {
									benMap[ben.ID] = 0
								}

								if ben.ID == "Overtime" {
									benMap[ben.ID] += mapOvertime[site.ID+id]
								} else if ben.ID == "UangMakan" {
									benMap[ben.ID] += float64(mapMealAllowance[site.ID+id]) * allowance
								} else {
									if ben.CalcType == ficomodel.PayrollComponentPercentage {
										benMap[ben.ID] += ben.Value / 100 * empDetail.BasicSalary
									} else if ben.CalcType == ficomodel.PayrollComponentDaily {
										benMap[ben.ID] += ben.Value * workDay
									} else {
										benMap[ben.ID] += ben.Value
									}
								}
							}
						}

						for _, ded := range site.Deductions {
							if ded.IsCash {
								if _, ok := deductMap[ded.ID]; !ok {
									deductMap[ded.ID] = 0
								}

								switch ded.ID {
								case "BPJSKS":
									deductMap[ded.ID] += obj.calculateHealthBPJS(&empDetail, &site, mapGrade[empDetail.Grade], mapRelation)
								case "BPJSTK":
									deductMap[ded.ID] += obj.calculateEmploymentBPJS(&empDetail, mapGrade[empDetail.Grade], mapPBJTK[empDetail.BPJSTKProgram], hoSite.Configuration.UMK)
								case "Loan":
									deductMap[ded.ID] += mapLoan[site.ID+id]
								default:
									if ded.CalcType == ficomodel.PayrollComponentPercentage {
										deductMap[ded.ID] += ded.Value / 100 * empDetail.BasicSalary
									} else if ded.CalcType == ficomodel.PayrollComponentDaily {
										deductMap[ded.ID] += ded.Value * workDay
									} else {
										deductMap[ded.ID] += ded.Value
									}
								}
							}
						}
					}
				}
			}

			pp.SiteID = site.ID
			pp.Name = site.Name
			pp.BaseSalary = baseTotal
			pp.Period = period
			pp.Benefit = []bagongmodel.DetailBenDeduct{}
			pp.Deduction = []bagongmodel.DetailBenDeduct{}
			for key, benefitSum := range benMap {
				newDetail := bagongmodel.DetailBenDeduct{Name: key, Amount: benefitSum}
				pp.Benefit = append(pp.Benefit, newDetail)
			}
			for key, deductSum := range deductMap {
				newDetail := bagongmodel.DetailBenDeduct{Name: key, Amount: deductSum}
				pp.Deduction = append(pp.Deduction, newDetail)
			}
			response.Details = append(response.Details, pp)
		}
	}
	return response, nil
}

func (obj *PayrollHandler) Get(ctx *kaos.Context, payload *PayrollDetailRequest) (*PayrollDetailResponse, error) {
	response := &PayrollDetailResponse{Details: []bagongmodel.BGPayrollDetail{}}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	mapSiteUMK := map[string]*SiteDetail{}
	// get site
	site := new(bagongmodel.Site)
	err := hub.GetByID(site, payload.SiteID)
	if err != nil {
		return nil, fmt.Errorf("error when get site: %s", err.Error())
	}
	mapOvertimeConfig := map[string]bagongmodel.Overtime{}
	for _, l := range site.Overtime {
		mapOvertimeConfig[site.ID+l.Position] = l
	}

	mapSiteUMK[site.ID] = &SiteDetail{
		SalaryUsed: site.SalaryUsed,
		UMK:        site.Configuration.UMK,
	}

	// get tenant employee
	tenantEmployees := []tenantcoremodel.Employee{}
	err = hub.Gets(new(tenantcoremodel.Employee), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.Eq("Value", site.ID)),
		),
	), &tenantEmployees)
	if err != nil {
		return nil, fmt.Errorf("error when get tenant employee: %s", err.Error())
	}
	employeeIDs := lo.Map(tenantEmployees, func(e tenantcoremodel.Employee, index int) string {
		return e.ID
	})

	// get employee detail
	employees := []bagongmodel.EmployeeDetail{}
	err = hub.Gets(new(bagongmodel.EmployeeDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.In("EmployeeID", employeeIDs...),
	), &employees)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}
	mapEmpDetail := lo.Associate(employees, func(detail bagongmodel.EmployeeDetail) (string, bagongmodel.EmployeeDetail) {
		return detail.EmployeeID, detail
	})

	date1 := payload.DateEnd
	period := date1.Format("200601")
	// get payrol detail
	payrollDetails := []bagongmodel.BGPayrollDetail{}
	err = hub.Gets(new(bagongmodel.BGPayrollDetail), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("EmployeeID", employeeIDs...),
			dbflex.Eq("Period", period),
		),
	), &payrollDetails)
	if err != nil {
		return nil, fmt.Errorf("error when get payroll detail: %s", err.Error())
	}

	mapPayrollDetailPerEmployee := lo.Associate(payrollDetails, func(detail bagongmodel.BGPayrollDetail) (string, bagongmodel.BGPayrollDetail) {
		return detail.EmployeeID, detail
	})

	mapWorkDay, err := obj.getWorkDays(hub, []string{site.ID}, payload)
	if err != nil {
		return nil, err
	}

	monthBefore := codekit.FirstOfMonth(*payload.DateEnd).AddDate(0, -1, 0)
	twoMonthBefore := codekit.FirstOfMonth(*payload.DateEnd).AddDate(0, -2, 0)
	start := time.Date(twoMonthBefore.Year(), twoMonthBefore.Month(), 24, 0, 0, 0, 0, twoMonthBefore.Location())
	end := time.Date(monthBefore.Year(), monthBefore.Month(), 23, 0, 0, 0, 0, monthBefore.Location())

	mapOvertime, err := obj.getOvertimes(hub, mapOvertimeConfig, mapEmpDetail, mapSiteUMK, []string{site.ID}, &start, &end)
	if err != nil {
		return nil, err
	}

	mapMealAllowance, err := obj.getMealAllowances(hub, []string{site.ID}, &start, &end)
	if err != nil {
		return nil, err
	}

	mapLoan, err := obj.getLoans(hub, []string{site.ID}, employeeIDs, &start, &end)
	if err != nil {
		return nil, err
	}

	// get master data
	masters := []tenantcoremodel.MasterData{}
	err = hub.Gets(new(tenantcoremodel.MasterData), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.In("MasterDataTypeID", []interface{}{"Relation", "GDE", "PBJTK"}...),
		),
	), &masters)
	if err != nil {
		return nil, fmt.Errorf("error when get employee: %s", err.Error())
	}
	mapGrade := map[string]string{}
	mapRelation := map[string]string{}
	mapPBJTK := map[string]string{}
	for _, d := range masters {
		switch d.MasterDataTypeID {
		case "Relation":
			mapRelation[d.ID] = d.Name
		case "GDE":
			mapGrade[d.ID] = d.Name
		case "PBJTK":
			mapPBJTK[d.ID] = d.Name
		}
	}

	hoSite := new(bagongmodel.Site)
	err = hub.GetByID(hoSite, "SITE020")
	if err != nil {
		return nil, fmt.Errorf("error when get HO Site: %s", err.Error())
	}

	allowance := 0.0
	for _, b := range site.Benefits {
		if b.ID == "UangMakan" {
			allowance = b.Value
			break
		}
	}
	for _, employee := range tenantEmployees {
		payroll := bagongmodel.BGPayrollDetail{}
		if pr, ok := mapPayrollDetailPerEmployee[employee.ID]; ok {
			payroll = pr
		}

		empDetail := mapEmpDetail[employee.ID]
		payroll.Period = period
		payroll.BaseSalary = empDetail.BasicSalary
		payroll.AttendanceNum = mapWorkDay[site.ID+employee.ID]
		payroll.EmployeeID = employee.ID
		payroll.Name = employee.Name
		payroll.SiteID = payload.SiteID
		payroll.Dimension = employee.Dimension
		for _, ben := range site.Benefits {
			if ben.IsCash {
				if payroll.CheckBenefitExists(ben.ID) {
					continue
				}

				amount := float64(0.0)
				if ben.ID == "Overtime" {
					amount = mapOvertime[site.ID+employee.ID]
				} else if ben.ID == "UangMakan" {
					amount = float64(mapMealAllowance[site.ID+employee.ID]) * allowance
				} else {
					if ben.CalcType == ficomodel.PayrollComponentPercentage {
						amount = ben.Value / 100 * payroll.BaseSalary
					} else if ben.CalcType == ficomodel.PayrollComponentDaily {
						amount = ben.Value * float64(payroll.AttendanceNum)
					} else {
						amount = ben.Value
					}
				}

				payroll.SetBenefitValue(ben.ID, ben.Name, amount, ben.CalcType)
			}
		}
		for _, ded := range site.Deductions {
			if ded.IsCash {
				if payroll.CheckBenefitExists(ded.ID) {
					continue
				}

				amount := float64(0.0)
				switch ded.ID {
				case "BPJSKS":
					amount = obj.calculateHealthBPJS(&empDetail, site, mapGrade[empDetail.Grade], mapRelation)
				case "BPJSTK":
					amount = obj.calculateEmploymentBPJS(&empDetail, mapGrade[empDetail.Grade], mapPBJTK[empDetail.BPJSTKProgram], hoSite.Configuration.UMK)
				case "Loan":
					amount = mapLoan[site.ID+employee.ID]
				default:
					if ded.CalcType == ficomodel.PayrollComponentPercentage {
						amount = ded.Value / 100 * payroll.BaseSalary
					} else if ded.CalcType == ficomodel.PayrollComponentDaily {
						amount = ded.Value * float64(payroll.AttendanceNum)
					} else {
						amount = ded.Value
					}
				}

				payroll.SetDeductionValue(ded.ID, ded.Name, amount, ded.CalcType)
			}
		}

		response.Details = append(response.Details, payroll)
	}

	return response, nil
}

func (obj *PayrollHandler) getWorkDays(h *datahub.Hub, siteIds []string, payload *PayrollDetailRequest) (map[string]int, error) {
	pipe := []bson.M{
		{
			"$match": bson.M{
				"TrxDate": bson.M{
					"$gte": payload.DateStart,
					"$lte": payload.DateEnd,
				},
				"Dimension": bson.M{
					"$elemMatch": bson.M{
						"Key":   "Site",
						"Value": bson.M{"$in": siteIds},
					},
				},
			},
		},
		{
			"$project": bson.M{
				"Dimension": 1,
				"UserID":    1,
			},
		},
		{
			"$unwind": "$Dimension",
		},
		{
			"$match": bson.M{
				"Dimension.Key": "Site",
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"UserID": "$UserID",
					"Site":   "$Dimension.Value",
				},
				"Count": bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"UserID": "$_id.UserID",
				"Site":   "$_id.Site",
				"Count":  1,
			},
		},
	}

	type workday struct {
		UserID string
		Site   string
		Count  int
	}

	// get work day count
	workdDays := []workday{}
	cmd := dbflex.From(new(karamodel.AttendanceTrx).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &workdDays); err != nil {
		return nil, fmt.Errorf("err when get work day count: %s", err.Error())
	}

	result := map[string]int{}
	for _, d := range workdDays {
		result[d.Site+d.UserID] = d.Count
	}

	return result, nil
}

func (obj *PayrollHandler) getOvertimes(h *datahub.Hub, mapOvertimeConfig map[string]bagongmodel.Overtime,
	mapEmployeeDetail map[string]bagongmodel.EmployeeDetail, mapSiteUMK map[string]*SiteDetail,
	siteIds []string, start, end *time.Time) (map[string]float64, error) {
	// get overtime hcm
	overtimes := []hcmmodel.Overtime{}
	err := h.Gets(new(hcmmodel.Overtime), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Gte("OvertimeDate", start),
			dbflex.Lte("OvertimeDate", end),
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.In("Value", siteIds...)),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted),
		),
	), &overtimes)
	if err != nil {
		return nil, fmt.Errorf("error when get overtime: %s", err.Error())
	}

	// calculate overtime
	mapOvertime := map[string]float64{}
	for _, overtime := range overtimes {
		for _, detail := range overtime.Lines {
			site := overtime.Dimension.Get("Site")
			// check position configuration in site
			if v, ok := mapOvertimeConfig[site+detail.Position]; ok {
				index := site + detail.EmployeeID
				// check calculation method
				if v.Method == "Flat" {
					if detail.OffDay {
						mapOvertime[index] += detail.ActualOvertime * float64(v.TULHoliday)
					} else {
						mapOvertime[index] += detail.ActualOvertime * float64(v.TUL)
					}
				} else {
					if v.Divider != 0 {
						tul := 0.0
						if s, ok := mapSiteUMK[site]; ok {
							if s.SalaryUsed == "UMK Site" {
								tul = s.UMK / float64(v.Divider)
							} else if s.SalaryUsed == "Basic Salary" {
								tul = mapEmployeeDetail[detail.EmployeeID].BasicSalary / float64(v.Divider)
							}
						}

						// if off day
						if detail.OffDay {
							if detail.ActualOvertime <= 7 {
								mapOvertime[index] += float64(detail.ActualOvertime) * 2 * tul
							} else {
								// 1-7 hour
								mapOvertime[index] += 7 * 2 * tul
								// 8 hour
								mapOvertime[index] += 1 * 3 * tul
								// 9-11 hour
								if detail.ActualOvertime > 8 {
									mapOvertime[index] += float64(detail.ActualOvertime-8) * 4 * tul
								}
							}
						} else {
							// 1 hour
							firstHour := 1.5 * tul
							// 2 - 11 hour
							remainingHour := float64(detail.ActualOvertime-1) * 2 * tul
							mapOvertime[index] += firstHour + remainingHour
						}
					}
				}
			}
		}
	}

	return mapOvertime, nil
}

func (obj *PayrollHandler) getLoans(h *datahub.Hub, siteIds, employeeIds []string, start, end *time.Time) (map[string]float64, error) {
	// get loan hcm
	loans := []hcmmodel.Loan{}
	err := h.Gets(new(hcmmodel.Loan), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.ElemMatch("Lines", dbflex.And(
				dbflex.Gte("Date", start),
				dbflex.Lte("Date", end),
			)),
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.In("Value", siteIds...)),
			dbflex.Eq("Status", ficomodel.JournalStatusPosted),
			dbflex.In("EmployeeID", employeeIds...),
		),
	), &loans)
	if err != nil {
		return nil, fmt.Errorf("error when get loan: %s", err.Error())
	}

	// calculate overtime
	mapLoan := map[string]float64{}
	for _, loan := range loans {
		for _, detail := range loan.Lines {
			site := loan.Dimension.Get("Site")
			if (detail.Date.Equal(*start) || detail.Date.After(*start)) && (detail.Date.Before(*end) || detail.Date.Equal(*end)) {
				mapLoan[site+loan.EmployeeID] += loan.ApprovedInstallment
			}
		}
	}

	return mapLoan, nil
}

func (obj *PayrollHandler) getMealAllowances(h *datahub.Hub, siteIds []string, start, end *time.Time) (map[string]int, error) {
	type attendance struct {
		UserID    string
		Op        karamodel.OpCode
		Dimension tenantcoremodel.Dimension
		TrxDate   time.Time
	}

	// get attendance
	attendances := []attendance{}
	err := h.Gets(new(karamodel.AttendanceTrx), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Gte("TrxDate", start),
			dbflex.Lte("TrxDate", end),
			dbflex.ElemMatch("Dimension", dbflex.Eq("Key", "Site"), dbflex.In("Value", siteIds...)),
		),
	).SetSelect("UserID", "Op", "Dimension", "TrxDate").SetSort("TrxDate"), &attendances)
	if err != nil {
		return nil, fmt.Errorf("error when get attendances: %s", err.Error())
	}

	type detail struct {
		Op      karamodel.OpCode
		TrxDate time.Time
	}
	result := map[string]int{}
	mapAttendance := map[string]*detail{}
	for _, a := range attendances {
		index := fmt.Sprintf("%s%s", a.Dimension.Get("Site"), a.UserID)
		det := &detail{
			Op:      a.Op,
			TrxDate: a.TrxDate,
		}
		// for first time only record check in
		// if first record is check out, skip it
		// because only count when check in and check out is normal
		if v, ok := mapAttendance[index]; !ok {
			if a.Op == karamodel.Checkin {
				mapAttendance[index] = det
			}
		} else {
			// record only when after check in, there is check out
			if v.Op == karamodel.Checkin && a.Op == karamodel.Checkout {
				if a.TrxDate.Sub(v.TrxDate).Hours() >= 4 {
					result[index] += 1
				}
			}
		}

		mapAttendance[index] = det
	}

	return result, nil
}

func (obj *PayrollHandler) calculateHealthBPJS(employeeDetail *bagongmodel.EmployeeDetail, site *bagongmodel.Site, grade string, mapRelation map[string]string) float64 {
	value := 0.01
	childCount := 0
	for _, m := range employeeDetail.FamilyMembers {
		if mapRelation[m.Relation] == "Child" {
			childCount++
		} else if mapRelation[m.Relation] != "Spouse" {
			value += 0.01
		}
	}

	if childCount > 3 {
		value += 0.01
	}

	if grade == "Supervisor" || grade == "Manager" || grade == "General Manager" || grade == "Director" {
		return value * employeeDetail.BasicSalary
	} else if grade == "Foreman" || grade == "Non Staff" || grade == "Officer" {
		return value * float64(site.Configuration.UMK)
	}

	return 0
}

func (obj *PayrollHandler) calculateEmploymentBPJS(employeeDetail *bagongmodel.EmployeeDetail, grade, employementBPJS string, umkHO float64) float64 {
	if grade == "Manager" || grade == "General Manager" || grade == "Director" {
		return 0.03 * employeeDetail.BasicSalary
	} else if grade == "Supervisor" || grade == "Foreman" || grade == "Non Staff" || grade == "Officer" {
		if employementBPJS == "3" {
			return 0.02 * umkHO
		} else if employementBPJS == "4" {
			return 0.03 * umkHO
		}
	}

	return 0
}

func (obj *PayrollHandler) GetSubmissionDetail(ctx *kaos.Context, payload *PayrollDetailRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	param := JournalPostReq{
		SiteID: payload.SiteID,
		Type:   "Payroll",
	}
	journalType, err := new(SiteEntryEngine).GetDetailLedgerJournalPost(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("error when get mapping journal type: %s", err.Error())
	}

	detail, err := obj.Get(ctx, payload)
	if err != nil {
		return nil, err
	}

	jt := journalType.(*ficomodel.LedgerJournalType)
	lines := make([]ficomodel.JournalLine, len(detail.Details))
	for i, det := range detail.Details {
		line := ficomodel.JournalLine{
			LineNo:        i,
			Qty:           1,
			PriceEach:     det.GetTakeHome(),
			Amount:        det.GetTakeHome(),
			UnitID:        "Each",
			Taxable:       true,
			Dimension:     det.Dimension,
			OffsetAccount: jt.DefaultOffset,
			Text:          fmt.Sprintf("%s - %s", det.EmployeeID, det.Name),
		}

		lines[i] = line
	}

	return lines, nil
}

type GetJournalPayrollRequest struct {
	CompanyID string
	SiteID    []string
	Text      string
	Period    string
	Skip      int
	Take      int
}

func (obj *PayrollHandler) GetJournalPayroll(ctx *kaos.Context, payload *GetJournalPayrollRequest) (interface{}, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	coID := tenantcorelogic.GetCompanyIDFromContext(ctx)

	start, err := time.Parse("2006-01", payload.Period)
	if err != nil {
		return nil, fmt.Errorf("error parsing perion: %s", err.Error())
	}
	end := start.AddDate(0, 1, 0)
	match := bson.M{
		"References": bson.M{
			"$elemMatch": bson.M{
				"Key":   "Submission Type",
				"Value": "Payroll",
			},
		},
		"TrxDate": bson.M{
			"$gte": start,
			"$lt":  end,
		},
	}

	if len(payload.SiteID) > 0 {
		match["Dimension"] = bson.M{
			"$elemMatch": bson.M{
				"Key":   "Site",
				"Value": bson.M{"$in": payload.SiteID},
			},
		}
	}

	if payload.CompanyID != "" {
		match["CompanyID"] = payload.CompanyID
	} else {
		match["CompanyID"] = coID
	}

	if payload.Text != "" {
		match["Text"] = bson.M{
			"$regex": primitive.Regex{Pattern: payload.Text, Options: "i"},
		}
	}

	pipe := []bson.M{
		{
			"$match": match,
		},
		{
			"$group": bson.M{
				"_id":   nil,
				"Count": bson.M{"$sum": 1},
			},
		},
	}

	type CountJournal struct {
		Count int
	}

	// get count
	counts := []CountJournal{}
	cmd := dbflex.From(new(ficomodel.LedgerJournal).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &counts); err != nil {
		return nil, fmt.Errorf("err when get payroll journal count: %s", err.Error())
	}

	count := 0
	if len(counts) > 0 {
		count = counts[0].Count
	}

	pipe = []bson.M{
		{
			"$match": match,
		},
		{
			"$sort": bson.M{"TrxDate": -1},
		},
		{
			"$skip": payload.Skip,
		},
		{
			"$limit": payload.Take,
		},
	}

	// get payroll cash journal
	journals := []codekit.M{}
	cmd = dbflex.From(new(ficomodel.LedgerJournal).TableName()).Command("pipe", pipe)
	if _, err := h.Populate(cmd, &journals); err != nil {
		return nil, fmt.Errorf("err when get petty cash journal: %s", err.Error())
	}

	ids := make([]interface{}, len(journals))
	for i := range journals {
		journals[i]["SiteID"] = ""

		dimension := journals[i]["Dimension"].(primitive.A)
		for _, d := range dimension {
			dim := d.(codekit.M)
			if dim["Key"].(string) == "Site" {
				ids[i] = dim["Value"]

				journals[i]["SiteID"] = dim["Value"]
				break
			}
		}
	}

	sites := []bagongmodel.Site{}
	h.GetsByFilter(new(bagongmodel.Site), dbflex.In("_id", ids...), &sites)
	mapSite := lo.Associate(sites, func(site bagongmodel.Site) (string, string) {
		return site.ID, site.Name
	})

	for i := range journals {
		journals[i]["SiteName"] = mapSite[journals[i]["SiteID"].(string)]
	}

	result := codekit.M{}.Set("data", journals).Set("count", count)
	return result, nil
}
