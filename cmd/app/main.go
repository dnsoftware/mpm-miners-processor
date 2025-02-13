package main

import (
	"context"
	"log"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/dnsoftware/mpmslib/pkg/configloader"

	"github.com/dnsoftware/mpm-miners-processor/config"
	"github.com/dnsoftware/mpm-miners-processor/internal/app"
	"github.com/dnsoftware/mpm-miners-processor/internal/constants"
)

func main() {
	ctx := context.Background()

	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config.yaml"
	envFile := basePath + "/.env"

	filePath, err := logger.GetLoggerMainLogPath()
	if err != nil {
		panic("Bad logger init: " + err.Error())
	}
	logger.InitLogger(logger.LogLevelDebug, filePath)

	startConf, err := configloader.LoadStartConfig(basePath + constants.StartConfigFilename)
	if err != nil {
		log.Fatalf("start config load error: %w", err)
	}

	err = config.LoadRemoteConfig(basePath, *startConf, logger.Log().Logger)
	if err != nil {
		logger.Log().Error("Remote config failed: " + err.Error())
	}

	cfg, err := config.New(configFile, envFile)
	if err != nil {
		log.Fatalf("config load error: %w", err)
	}

	cfg.AppID = startConf.AppID
	cfg.EtcdConfig.Endpoints = startConf.Etcd.Endpoints
	cfg.EtcdConfig.Username = startConf.Etcd.Auth.Username
	cfg.EtcdConfig.Password = startConf.Etcd.Auth.Password

	app.Run(ctx, cfg)
}
