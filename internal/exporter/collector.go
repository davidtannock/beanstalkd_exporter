package exporter

import (
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/davidtannock/beanstalkd_exporter/v2/internal/beanstalkd"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "beanstalkd"
)

// BeanstalkdServer is the minimum interface required by a BeanstalkdCollector
type BeanstalkdServer interface {
	ListTubes() ([]string, error)
	FetchStats() (beanstalkd.ServerStats, error)
	FetchTubesStats(map[string]bool) (beanstalkd.ManyTubeStats, error)
}

// CollectorOpts contains the options for configuring the beanstalkd collector.
type CollectorOpts struct {
	SystemMetrics []string
	AllTubes      bool
	Tubes         []string
	TubeMetrics   []string
}

// BeanstalkdCollector collects metrics from a beanstalkd server
// for consumption by Prometheus
type BeanstalkdCollector struct {
	beanstalkd BeanstalkdServer
	mutex      sync.RWMutex

	opts   CollectorOpts
	logger *slog.Logger

	systemMetrics map[string]prometheus.Gauge
	tubesMetrics  map[string]*prometheus.GaugeVec

	totalScrapes prometheus.Counter
	up           prometheus.Gauge
}

func (opts *CollectorOpts) validate() (err error) {
	// Error on any invalid system metrics.
	for _, m := range opts.SystemMetrics {
		if _, ok := descSystemMetrics[m]; !ok {
			err = fmt.Errorf("unknown system metric: %v", m)
			return
		}
	}

	// Error on any invalid tube metrics.
	for _, m := range opts.TubeMetrics {
		if _, ok := descTubeMetrics[m]; !ok {
			err = fmt.Errorf("unknown tube metric: %v", m)
			return
		}
	}

	// If there are specific tube metrics, there
	// must be at least one tube.
	if len(opts.TubeMetrics) > 0 && len(opts.Tubes) == 0 && !opts.AllTubes {
		err = fmt.Errorf("tube metrics without tubes is not supported")
		return
	}

	// If there are no system metrics, fetch all of them.
	if len(opts.SystemMetrics) == 0 {
		for m := range descSystemMetrics {
			opts.SystemMetrics = append(opts.SystemMetrics, m)
		}
	}

	// If there are tubes but no metrics, fetch all of them.
	if (opts.AllTubes || len(opts.Tubes) > 0) && len(opts.TubeMetrics) == 0 {
		for m := range descTubeMetrics {
			opts.TubeMetrics = append(opts.TubeMetrics, m)
		}
	}

	err = nil
	return
}

// NewBeanstalkdCollector returns an initialised BeanstalkdCollector
func NewBeanstalkdCollector(beanstalkd BeanstalkdServer, opts CollectorOpts, logger *slog.Logger) (*BeanstalkdCollector, error) {
	err := opts.validate()
	if err != nil {
		return nil, err
	}

	// If we're describing all tubes in beanstalkd, we'll fetch them later.
	if opts.AllTubes {
		opts.Tubes = nil
	}

	systemMetrics := make(map[string]prometheus.Gauge, len(opts.SystemMetrics))
	for _, metric := range opts.SystemMetrics {
		stat := descSystemMetrics[metric].stat
		systemMetrics[stat] = prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metric,
			Help:      descSystemMetrics[metric].help,
		})
	}

	var tubesMetrics map[string]*prometheus.GaugeVec
	if opts.AllTubes || len(opts.Tubes) > 0 {
		tubeLabels := []string{"tube"}
		tubesMetrics = make(map[string]*prometheus.GaugeVec, len(opts.TubeMetrics))
		for _, metric := range opts.TubeMetrics {
			stat := descTubeMetrics[metric].stat
			tubesMetrics[stat] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      metric,
				Help:      descTubeMetrics[metric].help,
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
			b.logger.Error("error scraping beanstalkd", "err", err)
			b.up.Set(0)
		}
	}()

	// We've done another scrape.
	b.totalScrapes.Inc()

	// So far beanstalkd is up.
	b.up.Set(1)

	// Fetch the system stats from beanstalkd.
	err = b.scrapeSystemStats()
	if err != nil {
		return
	}

	// Fetch the tubes stats from beanstalkd.
	err = b.scrapeTubesStats()
	if err != nil {
		return
	}
}

func (b *BeanstalkdCollector) scrapeSystemStats() error {
	systemStats, err := b.beanstalkd.FetchStats()
	if err != nil {
		return err
	}
	for stat, value := range systemStats {
		if _, ok := b.systemMetrics[stat]; ok {
			v, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			b.systemMetrics[stat].Set(float64(v))
		}
	}
	return nil
}

func (b *BeanstalkdCollector) scrapeTubesStats() (err error) {
	var tubeNames []string
	tubeNames, err = b.getTubesToScrape()
	if err != nil {
		return
	}
	tubes := make(map[string]bool)
	for _, tube := range tubeNames {
		tubes[tube] = true
	}
	manyTubesStats, err := b.beanstalkd.FetchTubesStats(tubes)
	if err != nil {
		return
	}
	for tube, statsOrErr := range manyTubesStats {
		if statsOrErr.Err != nil {
			err = statsOrErr.Err
		}
		for stat, value := range statsOrErr.Stats {
			if _, ok := b.tubesMetrics[stat]; ok {
				var v int64
				v, err = strconv.ParseInt(value, 10, 64)
				if err == nil {
					b.tubesMetrics[stat].WithLabelValues(tube).Set(float64(v))
				}
			}
		}
	}
	return
}

func (b *BeanstalkdCollector) getTubesToScrape() ([]string, error) {
	var err error
	tubeNames := b.opts.Tubes
	if b.opts.AllTubes {
		tubeNames, err = b.beanstalkd.ListTubes()
		if err != nil {
			return nil, err
		}
	}
	return tubeNames, nil
}
