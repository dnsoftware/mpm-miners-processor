package main

import (
	"context"
	"log"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"

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

	cfg, err := config.New(configFile, envFile)
	if err != nil {

	}

	app.Run(ctx, cfg)
}
