package main

import (
	"flag"
	"os"
	"xibarCoreApp/xibarlogic"
	"xibarCoreApp/xibarmodel"

	"git.kanosolution.net/kano/appkit"
	"git.kanosolution.net/kano/kaos"
	"git.kanosolution.net/sebar/sebar"
	_ "github.com/ariefdarmawan/flexmgo"
	"github.com/ariefdarmawan/kasset"
	"github.com/ariefdarmawan/suim"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/kanoteknologi/hd"
	"github.com/kanoteknologi/knats"
)

var (
	config      = flag.String("config", "app.yml", "path to config file")
	serviceName = "v1/asset"
	logger      = appkit.LogWithPrefix(serviceName)
)

func main() {
	flag.Parse()
	sebar.StartApp(*config, "", serviceName, logger, registerModel)
}

func registerModel(s *kaos.Service, appConfig *sebar.AppConfig, ev kaos.EventHub) func() {
	//helper.EnsurePathExist(filepath.Join(cfg.Data.GetString("data_folder"), "asset"), 0644)
	// model
	storage_ep := appConfig.Data.GetString("storage_end_point")
	storage_region := appConfig.Data.GetString("storage_region")
	storage_key := appConfig.Data.GetString("storage_key")
	storage_secret := appConfig.Data.GetString("storage_secret")
	storage_bucket := appConfig.Data.GetString("storage_bucket")

	//-- get config
	cfg := aws.NewConfig().WithCredentials(credentials.NewStaticCredentials(storage_key, storage_secret, ""))
	cfg.WithRegion(storage_region)

	//-- end-point for S3-alike, ie:minio
	if storage_ep != "" {
		cfg.WithEndpoint(storage_ep)
	}
	cfg.DisableSSL = aws.Bool(true)
	cfg.S3ForcePathStyle = aws.Bool(true)
	s3, e := kasset.NewS3WithConfig(storage_bucket, cfg)

	if e != nil {
		s.Log().Errorf("fail to connect asset storage service. %s", e.Error())
		os.Exit(1)
	}

	modSuim := suim.New()

	kassetEngine := kasset.NewAssetEngine(s3, "/v1/asset")
	customAsset := xibarlogic.NewCustomAssetEngines(s3, kassetEngine)
	s.RegisterModel(kassetEngine, "").SetDeployer(knats.DeployerName).DisableRoute("view", "write-with-content")
	s.RegisterModel(kassetEngine, "").SetDeployer(hd.DeployerName).AllowOnlyRoute("view", "write-with-content", "delete")
	s.RegisterModel(customAsset, "").SetDeployer(hd.DeployerName).DisableRoute("update-tag-by-journal")
	s.RegisterModel(customAsset, "").SetDeployer(knats.DeployerName).AllowOnlyRoute("update-tag-by-journal")

	s.Group().SetMod(modSuim).SetDeployer(hd.DeployerName).Apply(
		s.RegisterModel(new(xibarmodel.AssetGrid), ""),
	)

	return nil
}
