package sebar

import (
	"os"
	"os/signal"

	"git.kanosolution.net/kano/kaos"
	"github.com/sebarcode/logger"
)

func StartApp(config, portPrefix, serviceName string, logger *logger.LogEngine,
	registerModel func(s *kaos.Service, cfg *AppConfig, ev kaos.EventHub) func()) {
	if portPrefix == "" {
		portPrefix = os.Getenv("sebar_port_prefix")
	}
	if portPrefix == "" {
		portPrefix = "370"
	}
	appConfig, hm, ev, err := ConfigToApp(config, portPrefix)
	if err != nil {
		logger.Errorf("fail to prepare app. %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		ev.Close()
		hm.Close()
	}()

	// Service
	kaos.NamingType = kaos.NamingIsLower
	kaos.NamingJoiner = "-"
	s := kaos.NewService().SetBasePoint(serviceName).SetLogger(logger)
	CopyConfigDataToService(appConfig, s, "topic_asset_write", "addr_asset_view", "topic_send_message_template")
	s.SetHubManager(hm)
	s.RegisterEventHub(ev, "default", appConfig.EventServer.Group)

	// Model Relation
	/*
		orm.UseRelationManager = true
		rm := orm.DefaultRelationManager()
		if serviceName == "v1/admin" || serviceName == "v1/iam" {
			rm.AddRelation(&rbac.User{},
				orm.Relation{Children: &rbac.TenantUser{}, ParentField: "_id", ChildrenField: "UserID", MapFields: codekit.M{"LoginID": "LoginID"}})

			rm.AddRelation(&rbac.Tenant{},
				orm.Relation{Children: &rbac.TenantUser{}, ParentField: "_id", ChildrenField: "TenantID", MapFields: codekit.M{"TenantName": "Name"}},
				orm.Relation{Children: &rbac.Role{}, ParentField: "_id", ChildrenField: "TenantID", MapFields: codekit.M{"TenantName": "Name"}},
			)

		}
		orm.SetDefaultRelationManager(rm)
	*/

	if modelClose := registerModel(s, appConfig, ev); modelClose != nil {
		defer modelClose()
	}

	// deploy
	Deploy(appConfig, s, serviceName, logger, ev)

	// grace shutdown
	csign := make(chan os.Signal, 1)
	signal.Notify(csign, os.Interrupt)
	<-csign
	logger.Infof("stopping %v service", serviceName)
}
