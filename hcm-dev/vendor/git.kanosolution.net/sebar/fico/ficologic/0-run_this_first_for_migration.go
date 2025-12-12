package ficologic

import (
	"fmt"
	"reflect"
	"strings"

	"git.kanosolution.net/kano/dbflex/orm"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
)

type MigrationOpt struct {
}

type MigrationFn func(db *datahub.Hub) error

func RunMigration(db *datahub.Hub, opt MigrationOpt) error {
	fns := []MigrationFn{configureRBAC, configureTables}

	for _, fn := range fns {
		if e := fn(db); e != nil {
			rt := reflect.TypeOf(fn)
			return fmt.Errorf("%s: %s", rt.Name(), e.Error())
		}
	}

	return nil
}

func configureRBAC(h *datahub.Hub) error {
	featureCategory := &rbacmodel.FeatureCategory{
		ID: "Fico", Name: "Finance and Accounting",
	}
	h.Insert(featureCategory)

	features := []rbacmodel.Feature{
		{ID: "FicoMaster", Name: "Master Data"},
		{ID: "FicoLedger", Name: "Ledger Journal", NeedDimension: true},
		{ID: "FicoCustomer", Name: "Customer Transaction", NeedDimension: true},
		{ID: "FicoVendor", Name: "Vendor Transaction", NeedDimension: true},
		{ID: "FicoAsset", Name: "Asset Transaction", NeedDimension: true},
		{ID: "FicoWHS", Name: "Item & Warehouse Transaction", NeedDimension: true},
	}

	for _, feat := range features {
		feat.FeatureCategoryID = "Fico"
		h.Save(&feat)
	}

	// dibawah ini bisa diautomate
	roles := []rbacmodel.Role{
		{ID: "FicoUsers", Name: "Fico Users"},
	}

	// -- start automate here --
	roles = append(roles,
		rbacmodel.Role{ID: "FicoLedgerEntry", Name: "Fico Ledger Entry"},
		rbacmodel.Role{ID: "FicoLedgerPost", Name: "Fico Ledger Posting"},
	)
	for _, role := range roles {
		role.Enable = true
		h.Save(&role)
	}

	roleFeatures := []rbacmodel.RoleFeature{
		{RoleID: "FicoLedgerEntry", FeatureID: "FicoLedger", Create: true, Read: true, Update: true, Delete: true},
		{RoleID: "FicoLedgerPost", FeatureID: "FicoLedger", Read: true, Posting: true},
	}

	for _, rf := range roleFeatures {
		rf.ID = rf.RoleID
		h.Save(&rf)
	}
	//-- end automate --

	return nil
}

func configureTables(h *datahub.Hub) error {
	// TODO: masukkan obj disini, dan pastikan mereka punya index yang benar
	objs := []orm.DataModel{}

	sliceErrors := []string{}
	for _, obj := range objs {
		if err := h.EnsureDb(obj); err != nil {
			sliceErrors = append(sliceErrors, fmt.Sprintf("%s: %s", obj.TableName(), err.Error()))
		}
	}

	if len(sliceErrors) > 0 {
		return fmt.Errorf("error init table: %s", strings.Join(sliceErrors, ", "))
	}

	return nil
}
