package config

import (
	"log"
	"testing"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-miners-processor/internal/constants"
)

func TestConfigNew(t *testing.T) {
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config_example.yaml"
	envFile := basePath + "/.env"

	cfg, err := New(configFile, envFile)
	require.NoError(t, err)
	require.Equal(t, "Miners processor env", cfg.AppName)
	require.Equal(t, "7878", cfg.GrpcPort)
	require.Equal(t, "minersprocessor", cfg.JWTServiceName)
}
