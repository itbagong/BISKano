package mfglogic_test

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/mfg/mfgmodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

func injectMFGData(ctx *kaos.Context) error {
	var err error
	db := sebar.GetTenantDBFromContext(ctx)

	//-- insert work order journal type
	if err = InsertModel(db, []*mfgmodel.WorkOrderJournalType{
		{
			ID:                          "WorkOrder",
			Name:                        "WorkOrder",
			PostingProfileID:            "WADP2",
			PostingProfileIDConsumption: "WADP2",
			PostingProfileIDResource:    "WADP2",
			PostingProfileIDOutput:      "WADP2",
			// DefaultOffset:               ficomodel.SubledgerAccount{},
			TrxType: string(scmmodel.JournalWorkOrder),
		},
	}); err != nil {
		return err
	}

	return nil
}

func injectMFGDataBOM(ctx *kaos.Context) error {
	var err error
	db := sebar.GetTenantDBFromContext(ctx)

	//-- insert BOM
	if err = InsertModel(db, []*mfgmodel.BoM{
		{
			ID:          "BOM-001",
			Description: "Create new bus 1x1",
			BoMGroup:    "Production",
			OutputType:  "Item",
			SKU:         "65b590dc288776ac882da704",
			Title:       "New Bus 1x1",
			ItemID:      "BUS0001",
			LedgerID:    "",
			IsActive:    true,
		},
		{
			ID:          "BOM-004",
			Description: "Create new bus 4x4",
			BoMGroup:    "Production",
			OutputType:  "Item",
			SKU:         "65b590dc288776ac882da704",
			Title:       "New Bus 4X4",
			ItemID:      "BUS0004",
			LedgerID:    "",
			IsActive:    true,
		},
	}); err != nil {
		return err
	}

	//-- insert BOM Material
	if err = InsertModel(db, []*mfgmodel.BoMMaterial{
		{
			BoMID:       "BOM-001",
			ItemID:      "THRUST_WASHER",
			SKU:         "65532d97342cac6d87908f03",
			Description: "Thrust Washer",
			UoM:         "EAC",
			UnitPrice:   20000,
			Qty:         2,
			Total:       0,
		},
		{
			BoMID:       "BOM-004",
			ItemID:      "CAT",
			SKU:         "6552f890bae8381cd665b48b",
			Description: "CAT-DANAGLOSS CHROMEORANGE",
			UoM:         "EAC",
			UnitPrice:   20000,
			Qty:         2,
			Total:       0,
		},
		{
			BoMID:       "BOM-004",
			ItemID:      "BN001",
			SKU:         "6540adf0ea50d4de2033fcde",
			Description: "Ban-Micheline-R15-New",
			UoM:         "EAC",
			UnitPrice:   300000,
			Qty:         6,
			Total:       0,
		},
	}); err != nil {
		return err
	}

	//-- insert BOM ManPower
	if err = InsertModel(db, []*mfgmodel.BoMManpower{
		{
			BoMID: "BOM-001",
			ManPower: mfgmodel.ManPower{
				EmployeeQuantity: 3,
				StandartHour:     4,
				RatePerHour:      120000,
				ActivityName:     "Pemasangan Ban",
			},
		},
		{
			BoMID: "BOM-001",
			ManPower: mfgmodel.ManPower{
				EmployeeQuantity: 2,
				StandartHour:     5,
				RatePerHour:      200000,
				ActivityName:     "Pengecatan Rangka Bus",
			},
		},
	}); err != nil {
		return err
	}

	//-- insert BOM Machineries
	if err = InsertModel(db, []*mfgmodel.BoMMachinery{
		{
			BoMID: "BOM-001",
			Machinery: mfgmodel.Machinery{
				MachineCode:  "",
				MachineName:  "",
				StandartHour: 0,
				RatePerHour:  0,
			},
		},
		{
			BoMID: "BOM-001",
			Machinery: mfgmodel.Machinery{
				MachineCode:  "",
				MachineName:  "",
				StandartHour: 0,
				RatePerHour:  0,
			},
		},
		{
			BoMID: "BOM-001",
			Machinery: mfgmodel.Machinery{
				MachineCode:  "",
				MachineName:  "",
				StandartHour: 0,
				RatePerHour:  0,
			},
		},
	}); err != nil {
		return err
	}

	return nil
}
