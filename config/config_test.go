package config

import (
	"log"
	"testing"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/dnsoftware/mpmslib/pkg/configloader"
	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-miners-processor/internal/constants"
)

func TestConfigNew(t *testing.T) {
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config.yaml"
	envFile := basePath + "/.env"

	startConf, err := configloader.LoadStartConfig(basePath + constants.StartConfigFilename)
	if err != nil {
		log.Fatalf("start config load error: %s", err)
	}

	filePath, err := logger.GetLoggerMainLogPath()
	if err != nil {
		panic("Bad logger init: " + err.Error())
	}
	logger.InitLogger(logger.LogLevelDebug, filePath)

	err = LoadRemoteConfig(basePath, *startConf, logger.Log().Logger)
	if err != nil {
		logger.Log().Error("Remote config failed: " + err.Error())
	}

	cfg, err := New(configFile, envFile)

	cfg.AppID = startConf.AppID
	cfg.EtcdConfig.Endpoints = startConf.Etcd.Endpoints
	cfg.EtcdConfig.Username = startConf.Etcd.Auth.Username
	cfg.EtcdConfig.Password = startConf.Etcd.Auth.Password

	require.NoError(t, err)
	require.Equal(t, "Miners processor", cfg.AppName)
	require.Equal(t, "7878", cfg.GrpcPort)
	require.Equal(t, "minersprocessor", cfg.JWTServiceName)
}
