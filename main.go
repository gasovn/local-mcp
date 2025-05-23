package main

import (
	"log"

	"local-mcp/tools"

	"github.com/strowk/foxy-contexts/pkg/app"
	"github.com/strowk/foxy-contexts/pkg/mcp"
	"github.com/strowk/foxy-contexts/pkg/stdio"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

const (
	appName    = "local-mcp"
	appVersion = "1.0.0"
)

func main() {
	logger := createLogger()

	app.
		NewBuilder().
		WithTool(tools.NewSearchTool).
		WithTool(tools.NewClickHouseQueryTool).
		WithTool(tools.NewClickHouseSchemasTool).
		WithTool(tools.NewClickHouseTablesTool).
		WithName(appName).
		WithVersion(appVersion).
		WithServerCapabilities(&mcp.ServerCapabilities{
			Tools: &mcp.ServerCapabilitiesTools{},
		}).
		WithTransport(stdio.NewTransport()).
		WithFxOptions(
			fx.Provide(func() *zap.Logger { return logger }),
			fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
				return &fxevent.ZapLogger{Logger: logger}
			}),
		).
		Run()
}

func createLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zap.ErrorLevel)

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	return logger
}
