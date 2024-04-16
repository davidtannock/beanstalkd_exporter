package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/davidtannock/beanstalkd_exporter/v2/internal/httpserver"
	"github.com/urfave/cli/v2"
)

const version = "2.0.0"

var logger = slog.Default()

var (
	flagBeanstalkdAddress = &cli.StringFlag{
		Name:  "beanstalkd.address",
		Value: "localhost:11300",
		Usage: "address of beanstalkd process",
	}
	flagBeanstalkdDialTimeout = &cli.UintFlag{
		Name:  "beanstalkd.dialTimeout",
		Value: 10,
		Usage: "seconds (between 1 and 30) to wait for the connection to beanstalkd",
		Action: func(ctx *cli.Context, v uint) error {
			if v < 1 || v > 30 {
				return fmt.Errorf("flag beanstalkd.dialTimeout value %v out of range[1-30]", v)
			}
			return nil
		},
	}
	flagBeanstalkdKeepAlivePeriod = &cli.UintFlag{
		Name:  "beanstalkd.keepAlivePeriod",
		Value: 10,
		Usage: "seconds (> 0) between TCP keepalive messages to beanstalkd",
		Action: func(ctx *cli.Context, v uint) error {
			if v < 1 {
				return fmt.Errorf("flag beanstalkd.keepAlivePeriod value < 1")
			}
			return nil
		},
	}
	flagBeanstalkdSystemMetrics = &cli.StringFlag{
		Name:  "beanstalkd.systemMetrics",
		Value: "",
		Usage: "comma separated beanstalkd system metrics to collect (all metrics will be collected if this flag is not set)",
	}
	flagBeanstalkdAllTubes = &cli.BoolFlag{
		Name:  "beanstalkd.allTubes",
		Value: false,
		Usage: "collect metrics for all tubes",
	}
	flagBeanstalkdTubes = &cli.StringFlag{
		Name:  "beanstalkd.tubes",
		Value: "",
		Usage: "comma separated beanstalkd tubes for which to collect metrics (ignored when 'beanstalkd.allTubes' is true)",
	}
	flagBeanstalkdTubeMetrics = &cli.StringFlag{
		Name:  "beanstalkd.tubeMetrics",
		Value: "",
		Usage: "comma separated beanstalkd tube metrics to collect for the targeted tubes (all metrics are collected when this is not set)",
	}
	flagListenAddress = &cli.StringFlag{
		Name:  "web.listen-address",
		Value: ":8080",
		Usage: "address to listen on for web interface and telemetry",
	}
	flagMetricsPath = &cli.StringFlag{
		Name:  "web.telemetry-path",
		Value: "/metrics",
		Usage: "path under which to expose metrics",
	}
)

func newApp() *cli.App {
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Printf("%s\n", ctx.App.Version)
	}
	return &cli.App{
		Name:    filepath.Base(os.Args[0]),
		Version: version,
		Usage:   "a simple server that scrapes beanstalkd stats and exports them via http for prometheus consumption",
		Flags: []cli.Flag{
			flagBeanstalkdAddress,
			flagBeanstalkdDialTimeout,
			flagBeanstalkdKeepAlivePeriod,
			flagBeanstalkdSystemMetrics,
			flagBeanstalkdAllTubes,
			flagBeanstalkdTubes,
			flagBeanstalkdTubeMetrics,
			flagListenAddress,
			flagMetricsPath,
		},
		Action: runCmd,
	}
}

func runCmd(ctx *cli.Context) error {
	if ctx.NArg() != 0 {
		return cli.ShowAppHelp(ctx)
	}

	// Fetching all tubes overrides specific tubes.
	beanstalkdAllTubes := ctx.Bool(flagBeanstalkdAllTubes.Name)
	beanstalkdTubes := ctx.String(flagBeanstalkdTubes.Name)
	if beanstalkdAllTubes {
		beanstalkdTubes = ""
	}

	serverOptions := httpserver.Opts{
		BeanstalkdAddress:         ctx.String(flagBeanstalkdAddress.Name),
		BeanstalkdDialTimeout:     ctx.Uint(flagBeanstalkdDialTimeout.Name),
		BeanstalkdKeepAlivePeriod: ctx.Uint(flagBeanstalkdKeepAlivePeriod.Name),
		BeanstalkdSystemMetrics:   toStringArray(ctx.String(flagBeanstalkdSystemMetrics.Name)),
		BeanstalkdAllTubes:        beanstalkdAllTubes,
		BeanstalkdTubes:           toStringArray(beanstalkdTubes),
		BeanstalkdTubeMetrics:     toStringArray(ctx.String(flagBeanstalkdTubeMetrics.Name)),
		ListenAddress:             ctx.String(flagListenAddress.Name),
		MetricsPath:               ctx.String(flagMetricsPath.Name),
	}

	return httpserver.ListenAndServe(serverOptions, logger)
}

func RunAndExit() {
	app := newApp()
	err := app.Run(os.Args)
	if err != nil {
		logger.Error("fatal error", "err", err)
		os.Exit(1)
	}
}
