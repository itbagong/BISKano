package hcmlogic

import (
	"errors"
	"fmt"
	"io"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/hcm/hcmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

type TalentDevelopmentHandler struct {
}

type TalentDevelopmentGetDetailRequest struct {
	ID         string
	EmployeeID string
}

type TalentDevelopmentGetDetailResponse struct {
	Existing   hcmmodel.TalentDevelopmentDetail
	NewPropose hcmmodel.TalentDevelopmentDetail
}

func (m *TalentDevelopmentHandler) GetDetail(ctx *kaos.Context, payload *TalentDevelopmentGetDetailRequest) (*TalentDevelopmentGetDetailResponse, error) {
	h := sebar.GetTenantDBFromContext(ctx)
	if h == nil {
		return nil, errors.New("missing: connection")
	}

	result := new(TalentDevelopmentGetDetailResponse)
	// get current detail
	current := new(hcmmodel.TalentDevelopmentDetail)
	err := h.GetByID(current, payload.ID)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error when get current talent development detail: %s", err.Error())
	}
	result.NewPropose = *current

	// get latest
	latest := []hcmmodel.TalentDevelopment{}
	err = h.Gets(new(hcmmodel.TalentDevelopment), dbflex.NewQueryParam().SetWhere(
		dbflex.And(
			dbflex.Eq("EmployeeID", payload.EmployeeID),
			dbflex.Eq("Status", "APPROVED"),
		),
	).SetSort("-Created").SetTake(1), &latest)
	if err != nil {
		return nil, fmt.Errorf("error when get latest talent development: %s", err.Error())
	}

	// if doesn't exists get from employee
	if len(latest) == 0 {
		// get employee
		employee := new(tenantcoremodel.Employee)
		err = h.GetByParm(employee, dbflex.NewQueryParam().SetWhere(dbflex.Eq("_id", payload.EmployeeID)))
		if err != nil {
			return nil, fmt.Errorf("error when get employee: %s", err.Error())
		}

		// get employee detail
		employeeDetail := new(bagongmodel.EmployeeDetail)
		err = h.GetByParm(employeeDetail, dbflex.NewQueryParam().SetWhere(dbflex.Eq("EmployeeID", payload.EmployeeID)))
		if err != nil {
			return nil, fmt.Errorf("error when get employee detail: %s", err.Error())
		}

		detail := hcmmodel.TalentDevelopmentDetail{
			Department:  employeeDetail.Department,
			Position:    employeeDetail.Position,
			Grade:       employeeDetail.Grade,
			Group:       employeeDetail.Group,
			SubGroup:    employeeDetail.SubGroup,
			Site:        employee.Dimension.Get("Site"),
			PointOfHire: employeeDetail.POH,
			BasicSalary: employeeDetail.BasicSalary,
			Allowance:   "",
		}

		// set existing
		result.Existing = detail
	} else {
		// get latest talent detail
		talentDetail := new(hcmmodel.TalentDevelopmentDetail)
		err = h.GetByID(talentDetail, latest[0].ID)
		if err != nil {
			return nil, fmt.Errorf("error when get latest talent development detail: %s", err.Error())
		}

		result.Existing = *talentDetail
	}

	return result, nil
}
