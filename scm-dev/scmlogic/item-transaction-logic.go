package scmlogic

import (
	"fmt"
	"reflect"

	"git.kanosolution.net/sebar/scm/scmmodel"
	"github.com/ariefdarmawan/datahub"
)

type ItemTransactionLogic struct{}

type ItemTransactionUniqueParam struct {
	ReferenceID string `validate:"required"`
	ItemID      string `validate:"required"`
	SKU         string `validate:"required"`
}

func (o *ItemTransactionUniqueParam) Validate() error {
	if o == nil {
		return fmt.Errorf("param cannot be nil")
	}

	val := reflect.ValueOf(o).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		fieldName := fieldType.Name
		valTag := fieldType.Tag.Get("validate")

		if valTag == "required" {
			valStr := fmt.Sprintf("%v", field.Interface())
			if valStr == "" {
				return fmt.Errorf("Field '%s' is required and cannot be empty", fieldName)
			}
		}
	}

	return nil
}

func (o *ItemTransactionLogic) Add(h *datahub.Hub, param ItemTransactionUniqueParam, moveType MovementType, status scmmodel.TrxStatus, qty float64) error {
	if e := param.Validate(); e != nil {
		return e
	}

	mt := scmmodel.TrxType(moveType)
	if mt != scmmodel.ItemTransactionTypeMovementIn && mt != scmmodel.ItemTransactionTypeMovementOut {
		return fmt.Errorf("unknown movement type: %s", mt)
	}

	switch status {
	case scmmodel.ItemTransactionStatusPlanned:
		itp := scmmodel.ItemTransaction{
			ReferenceID:     param.ReferenceID,
			ItemID:          param.ItemID,
			SKU:             param.SKU,
			TransactionType: mt,
			Status:          scmmodel.ItemTransactionStatusPlanned,
			Qty:             qty, // always positive +
			CreatedBy:       "",  // TODO: set with user id
		}
		if e := h.Insert(&itp); e != nil {
			return e
		}

	case scmmodel.ItemTransactionStatusConfirmed:
		// plan reduction
		itp := scmmodel.ItemTransaction{
			ReferenceID:     param.ReferenceID,
			ItemID:          param.ItemID,
			SKU:             param.SKU,
			TransactionType: mt,
			Status:          scmmodel.ItemTransactionStatusPlanned,
			Qty:             (-1 * qty),
			CreatedBy:       "", // TODO: set with user id
		}
		if e := h.Insert(&itp); e != nil {
			return e
		}

		// confirm addition
		itc := scmmodel.ItemTransaction{
			ReferenceID:     param.ReferenceID,
			ItemID:          param.ItemID,
			SKU:             param.SKU,
			TransactionType: mt,
			Status:          scmmodel.ItemTransactionStatusConfirmed,
			Qty:             qty,
			CreatedBy:       "", // TODO: set with user id
		}
		if e := h.Insert(&itc); e != nil {
			return e
		}
	}

	return nil
}
