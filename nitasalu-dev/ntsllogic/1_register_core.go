package ntsllogic

import (
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/nitasalu/ntslmodel"
	"github.com/sebarcode/dbmod"
)

func RegisterCore(s *kaos.Service) error {
	dbmod := dbmod.New()
	//uimod := suim.New()

	s.RegisterModel(new(ntslmodel.City), "city").SetMod(dbmod)
	s.RegisterModel(new(ntslmodel.Feature), "feature").SetMod(dbmod)
	s.RegisterModel(new(ntslmodel.Profile), "profile").SetMod(dbmod)

	return nil
}
