package scmlogic_test

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/fico/ficomodel"
	"git.kanosolution.net/sebar/scm/scmmodel"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/tenantcore/tenantcoremodel"
)

func injectSCMData(ctx *kaos.Context) error {
	var err error
	db := sebar.GetTenantDBFromContext(ctx)

	//-- journal type
	if err = InsertModel(db, []*scmmodel.InventJournalType{
		{ID: "MovementIn", PostingProfileID: "WADP2", TransactionType: scmmodel.JournalMovementIn},
		{ID: "Transfer", PostingProfileID: "WADP2", TransactionType: scmmodel.JournalTransfer},
		{ID: "MovementOut", PostingProfileID: "WADP2", TransactionType: scmmodel.JournalMovementOut},
		{ID: "StockOpname", PostingProfileID: "WADP2", TransactionType: scmmodel.JournalOpname},
	}); err != nil {
		return err
	}

	if err = InsertModel(db, []*scmmodel.PurchaseOrderJournalType{
		{ID: "PurchaseOrder", PostingProfileID: "WADP2", TrxType: scmmodel.PurchOrder},
	}); err != nil {
		return err
	}
	if err = InsertModel(db, []*scmmodel.PurchaseRequestJournalType{
		{ID: "PurchaseRequest", PostingProfileID: "WADP2", TrxType: scmmodel.PurchOrder},
	}); err != nil {
		return err
	}

	if err = InsertModel(db, []*scmmodel.ItemRequestJournalType{
		{ID: "ItemRequest", PostingProfileID: "WADP2", TrxType: scmmodel.ItemRequestType},
	}); err != nil {
		return err
	}

	if err = InsertModel(db, []*scmmodel.AssetAcquisitionJournalType{
		{ID: "AssetAcquisition", PostingProfileID: "WADP2", TrxType: scmmodel.AssetAcquisitionTrxType},
	}); err != nil {
		return err
	}

	//-- item group
	if err = InsertModel(db, []*tenantcoremodel.ItemGroup{
		{ID: "Consumable-NoSpec", LedgerAccountIDStock: "330011", CostUnitCalcMethod: tenantcoremodel.CostUnitCalcMethodFIFO,
			FinanceDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledLocationWarehouse: true,
			},
			DefaultUnitID: "Each", CostUnit: 150},
		{ID: "Consumable-WithSpec", LedgerAccountIDStock: "330011", CostUnitCalcMethod: tenantcoremodel.CostUnitCalcMethodFIFO,
			FinanceDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledSpecVariant:       true,
				IsEnabledSpecGrade:         true,
				IsEnabledLocationWarehouse: true,
			},
			DefaultUnitID: "Each", CostUnit: 150},
	}); err != nil {
		return err
	}

	//-- item
	if err := InsertModel(db, []*tenantcoremodel.Item{
		{ID: "Busi", DefaultUnitID: "Each", CostUnit: 1},
		{ID: "Mur", DefaultUnitID: "Each", CostUnit: 2},
		{ID: "Baut", DefaultUnitID: "Each", CostUnit: 10},
		{
			ID:                   "THRUST_WASHER",
			ItemType:             "STOCK",
			LedgerAccountIDStock: "110000",
			DefaultUnitID:        "EAC",
			CostUnit:             35000,
			PhysicalDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledLocationBox:       false,
				IsEnabledSpecGrade:         false,
				IsEnabledItemBatch:         false,
				IsEnabledLocationAisle:     false,
				IsEnabledLocationWarehouse: true,
				IsEnabledLocationSection:   false,
				IsEnabledSpecVariant:       true,
				IsEnabledSpecSize:          true,
				IsEnabledItemSerial:        false,
			},
			FinanceDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledLocationWarehouse: true,
				IsEnabledLocationSection:   false,
				IsEnabledLocationBox:       false,
				IsEnabledSpecVariant:       false,
				IsEnabledSpecGrade:         false,
				IsEnabledItemBatch:         false,
				IsEnabledItemSerial:        false,
				IsEnabledLocationAisle:     false,
				IsEnabledSpecSize:          false,
			},
			Name:               "THRUST WASHER",
			ItemGroupID:        "002",
			CostUnitCalcMethod: "MANUAL",
		},
		{
			ID:                   "SHOCKABSORBER",
			ItemType:             "STOCK",
			LedgerAccountIDStock: "110000",
			PhysicalDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledLocationAisle:     false,
				IsEnabledSpecVariant:       true,
				IsEnabledSpecGrade:         false,
				IsEnabledItemSerial:        false,
				IsEnabledLocationSection:   false,
				IsEnabledLocationBox:       false,
				IsEnabledSpecSize:          false,
				IsEnabledItemBatch:         false,
				IsEnabledLocationWarehouse: true,
			},
			Name:               "SHOCK ABSORBER",
			ItemGroupID:        "Consumable",
			DefaultUnitID:      "EAC",
			CostUnitCalcMethod: "FIFO",
			CostUnit:           400000,
			FinanceDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledSpecSize:          false,
				IsEnabledLocationSection:   false,
				IsEnabledSpecVariant:       true,
				IsEnabledSpecGrade:         false,
				IsEnabledItemBatch:         false,
				IsEnabledItemSerial:        false,
				IsEnabledLocationWarehouse: true,
				IsEnabledLocationAisle:     false,
				IsEnabledLocationBox:       false,
			},
		},
		{
			ID:                 "KAMPAS_REM",
			ItemGroupID:        "Consumable",
			DefaultUnitID:      "EAC",
			CostUnitCalcMethod: "FIFO",
			CostUnit:           350000,
			FinanceDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledSpecSize:          false,
				IsEnabledSpecGrade:         false,
				IsEnabledItemBatch:         false,
				IsEnabledItemSerial:        false,
				IsEnabledLocationWarehouse: true,
				IsEnabledLocationAisle:     false,
				IsEnabledSpecVariant:       true,
				IsEnabledLocationSection:   false,
				IsEnabledLocationBox:       false,
			},
			Name:                 "KAMPAS REM",
			ItemType:             "STOCK",
			LedgerAccountIDStock: "110000",
			PhysicalDimension: tenantcoremodel.ItemDimensionCheck{
				IsEnabledSpecVariant:       true,
				IsEnabledSpecSize:          false,
				IsEnabledSpecGrade:         false,
				IsEnabledLocationAisle:     false,
				IsEnabledItemBatch:         false,
				IsEnabledItemSerial:        false,
				IsEnabledLocationWarehouse: true,
				IsEnabledLocationSection:   false,
				IsEnabledLocationBox:       false,
			},
		},
	}); err != nil {
		return err
	}

	//-- item spec
	if err = InsertModel(db, []*tenantcoremodel.ItemSpec{
		{
			ID:            "65532d97342cac6d87908f03",
			SpecGradeID:   "",
			ItemID:        "THRUST_WASHER",
			SKU:           "A4003324562",
			SpecVariantID: "THK",
			SpecSizeID:    "1-5mm",
		},
		{
			ID:            "65532d1d342cac6d87908e1e",
			ItemID:        "THRUST_WASHER",
			SKU:           "A4003324462",
			SpecVariantID: "THK",
			SpecSizeID:    "1-4mm",
			SpecGradeID:   "",
		},
		{
			ID:            "6553097dbae8381cd665b8d4",
			SpecGradeID:   "",
			ItemID:        "SHOCKABSORBER",
			SKU:           "BA368 326 01 00/01",
			SpecVariantID: "RARE",
			SpecSizeID:    "",
		},
		{
			ID:            "6553097dbae8381cd665b8d2",
			ItemID:        "SHOCKABSORBER",
			SKU:           "BA368 323 01 00/01",
			SpecVariantID: "FRONT",
			SpecSizeID:    "",
			SpecGradeID:   "",
		},
		{
			ID:            "65531601bae8381cd665bdf8",
			SpecVariantID: "HYUNDAI_MIGHTY",
			SpecSizeID:    "",
			SpecGradeID:   "",
			ItemID:        "KAMPAS_REM",
			SKU:           "Y030-1",
		},
		{
			ID:            "65531601bae8381cd665bdf6",
			ItemID:        "KAMPAS_REM",
			SKU:           "MK321243",
			SpecVariantID: "COLTDIESEL-FE447",
			SpecSizeID:    "",
			SpecGradeID:   "",
		},
	}); err != nil {
		return err
	}

	//-- whs
	if err = InsertModel(db, []*tenantcoremodel.LocationWarehouse{
		{
			ID:       "W001",
			Name:     "Warehouse 1",
			IsActive: true,
		},
		{
			ID:       "W0001",
			Name:     "SURABAYA 1",
			IsActive: true,
		},
		{
			ID:       "W-MLG001",
			IsActive: true,
			Name:     "Malang 1",
		},
	}); err != nil {
		return err
	}

	//-- section
	if err = InsertModel(db, []*tenantcoremodel.LocationSection{
		{
			ID:       "S0001",
			Name:     "SURABAYA SECTION 1",
			IsActive: true,
		},
		{
			ID:       "S-MLG001",
			Name:     "Malang Section 1",
			IsActive: true,
		},
	}); err != nil {
		return err
	}

	//-- invent journal type
	if err = InsertModel(db, []*scmmodel.InventJournalType{
		{
			ID:                  "JT-0005",
			Name:                "Transaksi Inventory Adjusment",
			TransactionType:     "Stock Opname",
			Dimension:           nil,
			ChecklistTemplateID: "",
			NumberSequenceID:    "IVJ",
			ReferenceTemplateID: "",
			DefaultOffset: ficomodel.SubledgerAccount{
				AccountType: "",
				AccountID:   "",
			},
			Actions:          nil,
			Previews:         nil,
			PostingProfileID: "",
		},
		{
			ID:                  "JT-0004",
			TransactionType:     "Stock Opname",
			PostingProfileID:    "",
			Name:                "Transkasi Stock Opname",
			ReferenceTemplateID: "",
			Previews:            nil,
			Dimension:           nil,
			ChecklistTemplateID: "",
			Actions:             nil,
			NumberSequenceID:    "IVJ",
			DefaultOffset: ficomodel.SubledgerAccount{
				AccountType: "",
				AccountID:   "",
			},
		},
		{
			ID:                  "JT-0003",
			ReferenceTemplateID: "",
			Actions:             nil,
			Dimension:           nil,
			TransactionType:     "Movement Out",
			NumberSequenceID:    "IVJ",
			DefaultOffset: ficomodel.SubledgerAccount{
				AccountType: "",
				AccountID:   "",
			},
			ChecklistTemplateID: "",
			Name:                "Transaksi Movement Out",
			PostingProfileID:    "",
			Previews:            nil,
		},
		{
			ID:                  "JT-0002",
			ReferenceTemplateID: "",
			PostingProfileID:    "",
			Previews:            nil,
			Dimension:           nil,
			TransactionType:     "Movement In",
			DefaultOffset: ficomodel.SubledgerAccount{
				AccountType: "",
				AccountID:   "",
			},
			ChecklistTemplateID: "",
			Actions:             nil,
			Name:                "Transaksi Movement In",
			NumberSequenceID:    "IVJ",
		},
		{
			ID:              "JT-0001",
			TransactionType: "Transfer",
			DefaultOffset: ficomodel.SubledgerAccount{
				AccountType: "",
				AccountID:   "",
			},
			Previews:            nil,
			Name:                "Transaksi Inventory Transfer",
			NumberSequenceID:    "IVJ",
			PostingProfileID:    "",
			ReferenceTemplateID: "",
			ChecklistTemplateID: "",
			Actions:             nil,
			Dimension:           nil,
		},
	}); err != nil {
		return err
	}

	return nil
}
