package afycorelogic_test

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/afycore/afycoremodel"
	"git.kanosolution.net/sebar/sebar"
)

func injectCoreData(ctx *kaos.Context) error {
	prepareCtxData(ctx, "admin", testCoID1)
	db := sebar.GetTenantDBFromContext(ctx)

	err := insertModel(db, []*afycoremodel.MedicalLocation{
		{ID: "K001", Name: "Klinik 01"},
		{ID: "K002", Name: "Klinik 02"},
		{ID: "K003", Name: "Klinik 03"},
	})
	if err != nil {
		return err
	}

	if err = insertModel(db, []*afycoremodel.MedicalStaff{
		{ID: "D1", Name: "Dr 1", StaffRole: afycoremodel.StaffDoctor},
		{ID: "D2", Name: "Dr 2", StaffRole: afycoremodel.StaffDoctor},
		{ID: "D3", Name: "Dr 3", StaffRole: afycoremodel.StaffDoctor},
		{ID: "N1", Name: "N1", StaffRole: afycoremodel.StaffNurse},
		{ID: "N2", Name: "N2", StaffRole: afycoremodel.StaffNurse},
		{ID: "P1", Name: "P1", StaffRole: afycoremodel.StaffPharmacy},
	}); err != nil {
		return err
	}

	if err = insertModel(db, []*afycoremodel.Poli{
		{ID: "IGD", Name: "IGD"},
		{ID: "Umum", Name: "Umum"},
		{ID: "Lab", Name: "Lab"},
		{ID: "Farmasi", Name: "Farmasi"},
	}); err != nil {
		return err
	}

	if err = insertModel(db, []*afycoremodel.LocationPoli{
		{MedicalLocationID: "K001", PoliID: "IGD"},
		{MedicalLocationID: "K001", PoliID: "Umum"},
		{MedicalLocationID: "K001", PoliID: "Farmasi"},
		{MedicalLocationID: "K001", PoliID: "Lab"},
	}); err != nil {
		return err
	}

	return nil
}
