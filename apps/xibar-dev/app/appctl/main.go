package main

import (
	"errors"
	"flag"
	"io"
	"os"

	"git.kanosolution.net/kano/dbflex"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	"git.kanosolution.net/sebar/sebarcore/rbaclogic"
	"git.kanosolution.net/sebar/sebarcore/rbacmodel"
	"github.com/ariefdarmawan/datahub"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/sebarcode/codekit"
)

var (
	config = flag.String("config", "app.yml", "path to config file")
	cmd    = flag.String("cmd", "", "command name")
	logger = sebar.LogWithPrefix("appctl")

	hm  *kaos.HubManager
	ev  kaos.EventHub
	err error

	// new user param
	userid   = flag.String("userid", "admin", "id of new admin")
	email    = flag.String("email", "admin@xibar.app", "email of new admin")
	secret   = flag.String("secret", "", "password for new admin")
	tenantid = flag.String("tenant", "demo", "tenant id")
)

func main() {
	flag.Parse()

	_, hm, ev, err = sebar.ConfigToApp(*config, "")
	if err != nil {
		logger.Errorf("fail to prepare app. %s", err.Error())
		os.Exit(1)
	}
	logger.Infof("tenant found: %d", len(hm.Keys())-2)

	defer func() {
		ev.Close()
		hm.Close()
	}()

	initTenant("Demo", false, false)

	switch *cmd {
	case "init-tenant":
		if _, err = initTenant(*tenantid, true, true); err != nil {
			logger.Errorf("init_tenant error: %s", err.Error())
		}

	case "save-admin":
		if err = saveadmin(); err != nil {
			logger.Errorf("save_admin error: %s", err.Error())
		}
	default:
		logger.Info("appctl - manage application data and configuration")
	}
}

func initTenant(tenantFID string, pushOtherData, rewrite bool) (string, error) {
	hadmin, _ := hm.Get("default", "")
	tenant := new(rbacmodel.Tenant)
	err := hadmin.GetByFilter(tenant, dbflex.Eq("FID", tenantFID))

	if err == io.EOF {
		tenant.ID = ""
		if tenantFID == "Demo" {
			tenant.ID = "Demo"
		}
		tenant.FID = tenantFID
		tenant.Name = tenantFID
		tenant.Enable = true
		panicIfError(hadmin.Save(tenant))
	} else {
		if pushOtherData && !rewrite {
			pushOtherData = false
		}
	}
	return tenant.ID, nil
}

func saveadmin() error {
	if *secret == "" {
		return errors.New("secret is mandatory")
	}

	registerAdminUser(hm.GetMust("iam", ""), ev, *userid, *secret)
	return nil
}

func registerAdminUser(hiam *datahub.Hub, ev kaos.EventHub, uid, password string) {
	fc := rbacmodel.FeatureCategory{
		ID:   "Administration",
		Name: "Administration",
	}
	panicIfError(hiam.Save(&fc))

	//-- feature and role
	feature := rbacmodel.Feature{
		ID:                "Administrator",
		Name:              "Administrator",
		NeedDimension:     false,
		FeatureCategoryID: fc.ID,
	}
	panicIfError(hiam.Save(&feature))

	//-- role
	role := rbacmodel.Role{
		ID:     "Administrators",
		Name:   "Administrators",
		Enable: true,
	}
	panicIfError(hiam.Save(&role))

	adme := rbaclogic.NewAdminEngine(hiam, ev)
	//-- role feature
	_, e := adme.AddFeatureToRole(&rbacmodel.RoleFeature{RoleID: role.ID, FeatureID: "Administrator", All: true})
	panicIfError(e)

	user := rbacmodel.User{
		ID:          "admin_" + uid,
		LoginID:     uid,
		DisplayName: uid,
		Email:       *email,
		Enable:      true,
		Status:      "Active",
	}
	panicIfError(hiam.Save(&user))

	userPass := rbacmodel.UserPassword{ID: user.ID, Password: codekit.ShaString(password, "")}
	panicIfError(hiam.Save(&userPass))
	adme.AddUserToRole(&rbacmodel.RoleMember{UserID: user.ID, RoleID: role.ID, Scope: rbacmodel.RoleScopeGlobal})
	logger.Infof("user %s password is reset to %s", user.ID, password)
}

func panicIfError(e error) {
	if e != nil {
		logger.Errorf("critical error, app will be stopped: %s", e.Error())
		os.Exit(1)
	}
}
