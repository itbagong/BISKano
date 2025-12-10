package afycorelogic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/afycore/afycoremodel"
	"git.kanosolution.net/sebar/sebar"
	"github.com/ariefdarmawan/suim"
)

func RegisterCore(s *kaos.Service) error {
	dbmod := sebar.NewDBModFromContext()
	uimod := suim.New()

	s.RegisterModel(new(afycoremodel.MedicalLocation), "location").SetMod(dbmod, uimod)
	s.RegisterModel(new(afycoremodel.Poli), "poli").SetMod(dbmod, uimod)
	s.RegisterModel(new(afycoremodel.MedicalStaff), "staff").SetMod(dbmod, uimod)
	s.RegisterModel(new(afycoremodel.Patient), "patient").SetMod(dbmod, uimod)
	s.RegisterModel(new(afycoremodel.CaseAction), "emr").SetMod(dbmod, uimod)

	s.RegisterModel(new(CaseLogic), "case")

	return nil
}
