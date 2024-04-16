package httpserver

import (
	"html"
	"log/slog"
	"net/http"

	"github.com/davidtannock/beanstalkd_exporter/v2/internal/beanstalkd"
	"github.com/davidtannock/beanstalkd_exporter/v2/internal/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsPath string
)

// Opts contains the options for configuring the http server.
type Opts struct {
	ListenAddress string
	MetricsPath   string

	BeanstalkdAddress         string
	BeanstalkdDialTimeout     uint
	BeanstalkdKeepAlivePeriod uint
	BeanstalkdSystemMetrics   []string
	BeanstalkdAllTubes        bool
	BeanstalkdTubes           []string
	BeanstalkdTubeMetrics     []string
}

// ListenAndServe initialises a http server and starts listening
// for http requests.
func ListenAndServe(opts Opts, logger *slog.Logger) error {
	metricsPath = opts.MetricsPath

	// Fetching all tubes overrides specific tubes.
	tubes := opts.BeanstalkdTubes
	if opts.BeanstalkdAllTubes {
		tubes = nil
	}

	beanstalkdServer, err := beanstalkd.NewServer(
		opts.BeanstalkdAddress,
		opts.BeanstalkdDialTimeout,
		opts.BeanstalkdKeepAlivePeriod,
	)
	if err != nil {
		return err
	}

	collector, err := exporter.NewBeanstalkdCollector(
		beanstalkdServer,
		exporter.CollectorOpts{
			SystemMetrics: opts.BeanstalkdSystemMetrics,
			AllTubes:      opts.BeanstalkdAllTubes,
			Tubes:         tubes,
			TubeMetrics:   opts.BeanstalkdTubeMetrics,
		},
		logger,
	)
	if err != nil {
		return err
	}

	prometheus.MustRegister(collector)

	http.HandleFunc("/", index)
	http.Handle(opts.MetricsPath, promhttp.Handler())

	logger.Info("started listening", "address", opts.ListenAddress)

	return http.ListenAndServe(opts.ListenAddress, nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(`<html>
	<head>
		<title>Beanstalkd Exporter</title>
	</head>
	<body>
		<h1>Beanstalkd Exporter</h1>
		<p><a href="` + html.EscapeString(metricsPath) + `">Metrics</a></p>
	</body>
</html>`))
}
