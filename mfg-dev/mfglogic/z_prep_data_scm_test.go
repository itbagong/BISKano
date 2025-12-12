package mfglogic_test

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
)

func injectSCMData(ctx *kaos.Context) error {
	var err error
	db := sebar.GetTenantDBFromContext(ctx)

	//-- insert Item balance
	bals := []*scmmodel.ItemBalance{
		{
			ItemID:           "THRUST_WASHER",
			SKU:              "65532d97342cac6d87908f03",
			BalanceDate:      nil,
			AmountFinancial:  0,
			AmountPhysical:   0,
			AmountAdjustment: 0,
			CompanyID:        testCoID1,
			Qty:              3000,
			QtyPlanned:       47000,
			QtyReserved:      0,
			QtyAvail:         50000,
			InventDim: scmmodel.InventDimension{
				WarehouseID:  WHSID,
				AisleID:      "",
				SectionID:    SECID,
				BoxID:        "",
				VariantID:    "THK",
				Size:         "1-5mm",
				Grade:        "",
				BatchID:      "",
				SerialNumber: "",
				SpecID:       "",
				InventDimID:  "",
			},
		},
		{
			ItemID:           "THRUST_WASHER",
			SKU:              "65532d1d342cac6d87908e1e",
			BalanceDate:      nil,
			AmountFinancial:  0,
			AmountPhysical:   0,
			AmountAdjustment: 0,
			CompanyID:        testCoID1,
			Qty:              3000,
			QtyPlanned:       47000,
			QtyReserved:      0,
			QtyAvail:         50000,
			InventDim: scmmodel.InventDimension{
				WarehouseID:  WHSID,
				AisleID:      "",
				SectionID:    SECID,
				BoxID:        "",
				VariantID:    "THK",
				Size:         "1-4mm",
				Grade:        "",
				BatchID:      "",
				SerialNumber: "",
				SpecID:       "",
				InventDimID:  "",
			},
		},
		{
			ItemID:           "SHOCKABSORBER",
			SKU:              "6553097dbae8381cd665b8d4",
			BalanceDate:      nil,
			AmountFinancial:  0,
			AmountPhysical:   0,
			AmountAdjustment: 0,
			CompanyID:        testCoID1,
			Qty:              3000,
			QtyPlanned:       47000,
			QtyReserved:      0,
			QtyAvail:         50000,
			InventDim: scmmodel.InventDimension{
				WarehouseID:  WHSID,
				AisleID:      "",
				SectionID:    SECID,
				BoxID:        "",
				VariantID:    "RARE",
				Size:         "",
				Grade:        "",
				BatchID:      "",
				SerialNumber: "",
				SpecID:       "",
				InventDimID:  "",
			},
		},
		{
			ItemID:           "CAT",
			SKU:              "6552f890bae8381cd665b48b",
			BalanceDate:      nil,
			AmountFinancial:  0,
			AmountPhysical:   0,
			AmountAdjustment: 0,
			CompanyID:        testCoID1,
			Qty:              3000,
			QtyPlanned:       47000,
			QtyReserved:      0,
			QtyAvail:         50000,
			InventDim: scmmodel.InventDimension{
				WarehouseID:  WHSID,
				AisleID:      "",
				SectionID:    SECID,
				BoxID:        "",
				VariantID:    "DANAGLOSS_CHROMEORANGE",
				Size:         "",
				Grade:        "",
				BatchID:      "",
				SerialNumber: "",
				SpecID:       "",
				InventDimID:  "",
			},
		},
		{
			ItemID:           "BN001",
			SKU:              "6540adf0ea50d4de2033fcde",
			BalanceDate:      nil,
			AmountFinancial:  0,
			AmountPhysical:   0,
			AmountAdjustment: 0,
			CompanyID:        testCoID1,
			Qty:              3000,
			QtyPlanned:       47000,
			QtyReserved:      0,
			QtyAvail:         50000,
			InventDim: scmmodel.InventDimension{
				WarehouseID:  WHSID,
				AisleID:      "",
				SectionID:    SECID,
				BoxID:        "",
				VariantID:    "V123",
				Size:         "S-001",
				Grade:        "",
				BatchID:      "",
				SerialNumber: "",
				SpecID:       "",
				InventDimID:  "",
			},
		},
	}

	for _, b := range bals {
		b.InventDim.Calc()
	}

	if err = InsertModel(db, bals); err != nil {
		return err
	}

	return nil
}
