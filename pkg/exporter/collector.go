package exporter

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/davidtannock/beanstalkd_exporter/pkg/beanstalkd"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "beanstalkd"
)

type beanstalkdServer interface {
	FetchStats() (beanstalkd.ServerStats, error)
	FetchTubesStats(map[string]bool) (beanstalkd.ManyTubeStats, error)
}

// CollectorOpts contains the options for configuring the beanstalkd collector.
type CollectorOpts struct {
	SystemMetrics []string
	Tubes         []string
	TubeMetrics   []string
}

// BeanstalkdCollector collects metrics from a beanstalkd server
// for consumption by Prometheus
type BeanstalkdCollector struct {
	beanstalkd beanstalkdServer
	mutex      sync.RWMutex

	opts   CollectorOpts
	logger log.Logger

	systemMetrics map[string]prometheus.Gauge
	tubesMetrics  map[string]*prometheus.GaugeVec

	totalScrapes prometheus.Counter
	up           prometheus.Gauge
}

func (opts *CollectorOpts) validate() (err error) {
	// Panic on any invalid system metrics.
	for _, m := range opts.SystemMetrics {
		if _, ok := systemMetricsToStats[m]; !ok {
			err = fmt.Errorf("Unknown system metric: %v", m)
			return
		}
	}

	// Panic on any invalid tube metrics.
	for _, m := range opts.TubeMetrics {
		if _, ok := tubeMetricsToStats[m]; !ok {
			err = fmt.Errorf("Unknown tube metric: %v", m)
			return
		}
	}

	// If there are specific tube metrics, there
	// must be at least one tube.
	if len(opts.TubeMetrics) > 0 && len(opts.Tubes) == 0 {
		err = fmt.Errorf("Tube metrics without tubes is not supported")
		return
	}

	// If there are no system metrics, fetch all of them.
	if len(opts.SystemMetrics) == 0 {
		var systemMetrics []string
		for m := range systemMetricsToStats {
			systemMetrics = append(systemMetrics, m)
		}
		opts.SystemMetrics = systemMetrics
	}

	// If there are specific tubes but no metrics, fetch all of them.
	if len(opts.Tubes) > 0 && len(opts.TubeMetrics) == 0 {
		var tubeMetrics []string
		for m := range tubeMetricsToStats {
			tubeMetrics = append(tubeMetrics, m)
		}
		opts.TubeMetrics = tubeMetrics
	}

	err = nil
	return
}

// NewBeanstalkdCollector returns an initialised BeanstalkdCollector
func NewBeanstalkdCollector(beanstalkd beanstalkdServer, opts CollectorOpts, logger log.Logger) (*BeanstalkdCollector, error) {
	err := opts.validate()
	if err != nil {
		return nil, err
	}

	systemMetrics := make(map[string]prometheus.Gauge, len(opts.SystemMetrics))
	for _, metric := range opts.SystemMetrics {
		stat := systemMetricsToStats[metric]
		systemMetrics[stat] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metric,
			Help:      systemStatsHelp[stat],
		})
	}

	var tubesMetrics map[string]*prometheus.GaugeVec
	if len(opts.Tubes) > 0 {
		tubeLabels := []string{"tube"}
		tubesMetrics = make(map[string]*prometheus.GaugeVec, len(opts.TubeMetrics))
		for _, metric := range opts.TubeMetrics {
			stat := tubeMetricsToStats[metric]
			tubesMetrics[stat] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      metric,
				Help:      tubeStatsHelp[stat],
			}, tubeLabels)
		}
	}

	return &BeanstalkdCollector{
		beanstalkd:    beanstalkd,
		opts:          opts,
		logger:        logger,
		systemMetrics: systemMetrics,
		tubesMetrics:  tubesMetrics,
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_scrapes_total",
			Help:      "Current total number of beanstalkd scrapes.",
		}),
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Current health status of the backend (1 = UP, 0 = DOWN).",
		}),
	}, nil
}

// Describe implements the prometheus.Collector interface
// to describe the collected metrics.
func (b *BeanstalkdCollector) Describe(ch chan<- *prometheus.Desc) {
	b.up.Describe(ch)
	b.totalScrapes.Describe(ch)
	for _, m := range b.systemMetrics {
		m.Describe(ch)
	}
	for _, m := range b.tubesMetrics {
		m.Describe(ch)
	}
}

// Collect implements the prometheus.Collector interface
// to collect the beanstalkd metrics.
func (b *BeanstalkdCollector) Collect(ch chan<- prometheus.Metric) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.resetMetrics()
	b.scrape()

	b.up.Collect(ch)
	b.totalScrapes.Collect(ch)
	for _, m := range b.systemMetrics {
		m.Collect(ch)
	}
	for _, m := range b.tubesMetrics {
		m.Collect(ch)
	}
}

func (b *BeanstalkdCollector) resetMetrics() {
	for _, m := range b.tubesMetrics {
		m.Reset()
	}
}

func (b *BeanstalkdCollector) scrape() {
	// If there are any errors at the end of this func
	// then mark the backend "down".
	var err error
	defer func() {
		if err != nil {
			b.logger.Errorf("Error scraping beanstalkd: %v", err)
			b.up.Set(0)
		}
	}()

	// We've done another scrape.
	b.totalScrapes.Inc()
	b.up.Set(1)

	// Fetch the system stats from beanstalkd.
	systemStats, err := b.beanstalkd.FetchStats()
	if err != nil {
		return
	}
	for stat, value := range systemStats {
		if _, ok := b.systemMetrics[stat]; ok {
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return
			}
			b.systemMetrics[stat].Set(float64(v))
		}
	}

	// Fetch the tubes stats from beanstalkd.
	if len(b.opts.Tubes) == 0 {
		return
	}
	tubes := make(map[string]bool)
	for _, tube := range b.opts.Tubes {
		tubes[tube] = true
	}
	manyTubesStats, err := b.beanstalkd.FetchTubesStats(tubes)
	if err != nil {
		return
	}
	for tube, statsOrErr := range manyTubesStats {
		if statsOrErr.Err != nil {
			err = statsOrErr.Err
			return
		}
		for stat, value := range statsOrErr.Stats {
			if _, ok := b.tubesMetrics[stat]; ok {
				v, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return
				}
				b.tubesMetrics[stat].WithLabelValues(tube).Set(float64(v))
			}
		}
	}
}
