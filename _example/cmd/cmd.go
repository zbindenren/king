// Package cmd represents the command.
package cmd

import (
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-chi/chi"
	"github.com/postfinance/profiler"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zbindenren/king"
	"go.uber.org/zap"
)

// CLI represents the command line interface.
type CLI struct {
	Globals
	Server serverCmd `cmd:"" help:"daeira server"`
}

type serverCmd struct {
	Listen string `help:"server listen address" default:":3001"`
	Token  string `help:"the token" default:"very_secret"`
}

func (s serverCmd) Run(app *kong.Context, g *Globals, l *zap.SugaredLogger, reg *prometheus.Registry) error {
	l.Infow("starting server", king.FlagMap(app).Rm(
		"help", "version",
	).Register(
		app.Model.Name, reg,
	).Redact(
		regexp.MustCompile("token"),
		regexp.MustCompile("pw"),
	).List()...)

	r := chi.NewRouter()

	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.Method("GET", "/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    s.Listen,
		Handler: r,
	}

	l.Info("starting server")

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Globals are the gloabal flags.
type Globals struct {
	Debug    bool          `help:"enable debug output"`
	Profiler profilerFlags `embed:"" prefix:"profiler-"`
}

type profilerFlags struct {
	Enabled bool          `help:"Enable the profiler."`
	Listen  string        `help:"The profiles listen address." default:":6666"`
	Timeout time.Duration `help:"The timeout after the pprof handler will be shutdown." default:"5m"`
}

func (p profilerFlags) New(s os.Signal, h ...profiler.Hooker) *profiler.Profiler {
	return profiler.New(
		profiler.WithAddress(p.Listen),
		profiler.WithTimeout(p.Timeout),
		profiler.WithSignal(s),
		profiler.WithHooks(h...),
	)
}
