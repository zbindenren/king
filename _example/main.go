package main

import (
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zbindenren/king"
	"github.com/zbindenren/king/_example/cmd"
	"go.uber.org/zap"
)

// set variables with goreleaser
var (
	version = "1.0.0"
	date    = "2020-09-22T11:11:10+02:00"
	commit  = "123456789"
)

const appName = "example"

func main() {
	cli := cmd.CLI{}
	registry := prometheus.NewRegistry()
	logger, _ := zap.NewProduction()

	defer logger.Sync() // nolint: errcheck

	l := logger.Sugar()

	b, err := king.NewBuildInfo(version,
		king.WithDateString(date),
		king.WithRevision(commit),
		king.WithLocation("Europe/Zurich"),
	)
	if err != nil {
		l.Fatal(err)
	}

	app := kong.Parse(&cli, king.DefaultOptions(
		king.Config{
			Name:        appName,
			Description: "Daeira cli and server.",
			BuildInfo:   b,
		},
	)...)

	if cli.Profiler.Enabled {
		cli.Profiler.New(syscall.SIGUSR2).Start()
	}

	if err := app.Run(&cli.Globals, l, registry); err != nil {
		l.Fatal(err)
	}
}
