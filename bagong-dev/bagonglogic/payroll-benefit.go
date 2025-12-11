package bagonglogic

import (
	"errors"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/bagong/bagongmodel"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/sebar"
)

type PayrollBenefitEngine struct{}

type PayrollBenefitGetResponse struct {
	ficomodel.PayrollBenefit
	Detail bagongmodel.BGPayrollBenefitDetail
}

func (engine *PayrollBenefitEngine) Get(ctx *kaos.Context, req []interface{}) (*PayrollBenefitGetResponse, error) {
	if len(req) == 0 {
		return nil, errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	res := new(PayrollBenefitGetResponse)

	benefitID := req[0]
	ficoPayrollBenefit := new(ficomodel.PayrollBenefit)
	e := hub.GetByID(ficoPayrollBenefit, benefitID)
	if e == nil {
		res.PayrollBenefit = *ficoPayrollBenefit

		//get detail of vendor
		bagongBenefitDetail := new(bagongmodel.BGPayrollBenefitDetail)
		e = hub.GetByFilter(bagongBenefitDetail, dbflex.Eq("PayrollBenefitID", ficoPayrollBenefit.ID))
		if e == nil {
			res.Detail = *bagongBenefitDetail
		}
	}

	return res, nil
}

type PayrollBenefitSaveRequest struct {
	ficomodel.PayrollBenefit
	Detail bagongmodel.BGPayrollBenefitDetail
}

func (engine *PayrollBenefitEngine) Save(ctx *kaos.Context, req *PayrollBenefitSaveRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	//save to payroll benefit fico
	ficoBenefit := new(ficomodel.PayrollBenefit)
	e := hub.GetByID(ficoBenefit, req.ID)
	if e != nil {
		//insert fico payroll benefit from payload
		e = hub.Insert(&req.PayrollBenefit)
		if e != nil {
			return nil, errors.New("error insert payroll benefit: " + e.Error())
		}
	} else {
		//upsert fico payroll benefit from  payload
		e = hub.Save(&req.PayrollBenefit)
		if e != nil {
			return nil, errors.New("error update payroll benefit: " + e.Error())
		}
	}

	req.Detail.PayrollBenefitID = req.ID

	//save to bagong
	bagongDetailBenefit := new(bagongmodel.BGPayrollBenefitDetail)
	e = hub.GetByID(bagongDetailBenefit, req.Detail.ID)
	if e != nil {
		//insert bagong payroll benefit detail from payload
		e = hub.Insert(&req.Detail)
		if e != nil {
			return nil, errors.New("error insert bagong payroll benefit detail: " + e.Error())
		}
	} else {
		//upsert bagong payroll benefit detail from  payload
		e = hub.Save(&req.Detail)
		if e != nil {
			return nil, errors.New("error update bagong payroll benefit detail: " + e.Error())
		}
	}

	return req, nil
}

type PayrollDeductionEngine struct{}

type PayrollDeductionGetResponse struct {
	ficomodel.PayrollDeduction
	Detail bagongmodel.BGPayrollDeductionDetail
}

func (engine *PayrollDeductionEngine) Get(ctx *kaos.Context, req []interface{}) (*PayrollDeductionGetResponse, error) {
	if len(req) == 0 {
		return nil, errors.New("missing: invalid request, please check your payload")
	}

	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	res := new(PayrollDeductionGetResponse)

	deductionID := req[0]
	ficoPayrollDeduction := new(ficomodel.PayrollDeduction)
	e := hub.GetByID(ficoPayrollDeduction, deductionID)
	if e == nil {
		res.PayrollDeduction = *ficoPayrollDeduction

		//get detail of vendor
		bagongDeductionDetail := new(bagongmodel.BGPayrollDeductionDetail)
		e = hub.GetByFilter(bagongDeductionDetail, dbflex.Eq("PayrollDeductionID", ficoPayrollDeduction.ID))
		if e == nil {
			res.Detail = *bagongDeductionDetail
		}
	}

	return res, nil
}

type PayrollDeductionSaveRequest struct {
	ficomodel.PayrollDeduction
	Detail bagongmodel.BGPayrollDeductionDetail
}

func (engine *PayrollDeductionEngine) Save(ctx *kaos.Context, req *PayrollDeductionSaveRequest) (interface{}, error) {
	hub := sebar.GetTenantDBFromContext(ctx)
	if hub == nil {
		return nil, errors.New("missing: connection")
	}

	//save to payroll deduction fico
	ficoDeduction := new(ficomodel.PayrollDeduction)
	e := hub.GetByID(ficoDeduction, req.ID)
	if e != nil {
		//insert fico payroll deduction from payload
		e = hub.Insert(&req.PayrollDeduction)
		if e != nil {
			return nil, errors.New("error insert payroll deduction: " + e.Error())
		}
	} else {
		//upsert fico payroll deduction from  payload
		e = hub.Save(&req.PayrollDeduction)
		if e != nil {
			return nil, errors.New("error update payroll deduction: " + e.Error())
		}
	}

	req.Detail.PayrollDeductionID = req.ID

	//save to bagong
	bagongDetailDeduction := new(bagongmodel.BGPayrollDeductionDetail)
	e = hub.GetByID(bagongDetailDeduction, req.Detail.ID)
	if e != nil {
		//insert bagong payroll deduction detail from payload
		e = hub.Insert(&req.Detail)
		if e != nil {
			return nil, errors.New("error insert bagong payroll deduction detail: " + e.Error())
		}
	} else {
		//upsert bagong payroll deduction detail from  payload
		e = hub.Save(&req.Detail)
		if e != nil {
			return nil, errors.New("error update bagong payroll deduction detail: " + e.Error())
		}
	}

	return req, nil
}
