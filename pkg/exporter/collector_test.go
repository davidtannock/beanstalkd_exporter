package exporter

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davidtannock/beanstalkd_exporter/pkg/beanstalkd"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/log"
)

func readCounter(m prometheus.Counter) float64 {
	// TODO: Revisit this once client_golang offers better testing tools.
	pb := &dto.Metric{}
	m.Write(pb)
	return pb.GetCounter().GetValue()
}

func readGauge(m prometheus.Gauge) float64 {
	// TODO: Revisit this once client_golang offers better testing tools.
	pb := &dto.Metric{}
	m.Write(pb)
	return pb.GetGauge().GetValue()
}

func TestValidateErrors(t *testing.T) {
	tests := []struct {
		opts          CollectorOpts
		expectedError string
	}{
		// We expect an error when there's an unknown system metric.
		{
			opts:          CollectorOpts{SystemMetrics: []string{"does_not_exist"}},
			expectedError: "Unknown system metric: does_not_exist",
		},
		// We expect an error when there's an unknown tube metric.
		{
			opts:          CollectorOpts{TubeMetrics: []string{"tube_no_exist"}},
			expectedError: "Unknown tube metric: tube_no_exist",
		},
		// If specific tubes metrics are requested, we must have tubes.
		{
			opts:          CollectorOpts{Tubes: []string{}, TubeMetrics: []string{"tube_current_jobs_ready_count"}},
			expectedError: "Tube metrics without tubes is not supported",
		},
	}

	for _, tt := range tests {
		actualError := tt.opts.validate()
		if actualError.Error() != tt.expectedError {
			t.Errorf("expected error %v, actual %v", tt.expectedError, actualError.Error())
		}
	}
}

func TestValidateDefaultFetchAllSystemMetrics(t *testing.T) {
	opts := CollectorOpts{}
	err := opts.validate()
	if err != nil {
		t.Errorf("expected nil error, actual %v", err)
	}
	if len(opts.SystemMetrics) != len(systemMetricsToStats) {
		t.Errorf(
			"expected system metrics length to be %v, actual %v",
			len(systemMetricsToStats),
			len(opts.SystemMetrics),
		)
	}
	for _, m := range opts.SystemMetrics {
		if _, found := systemMetricsToStats[m]; !found {
			t.Errorf("unexpected system metric: %v", m)
		}
	}
}

func TestValidateDefaultFetchAllTubeMetrics(t *testing.T) {
	opts := CollectorOpts{Tubes: []string{"default"}}
	err := opts.validate()
	if err != nil {
		t.Errorf("expected nil error, actual %v", err)
	}
	if len(opts.TubeMetrics) != len(tubeMetricsToStats) {
		t.Errorf(
			"expected tube metrics length to be %v, actual %v",
			len(tubeMetricsToStats),
			len(opts.TubeMetrics),
		)
	}
	for _, m := range opts.TubeMetrics {
		if _, found := tubeMetricsToStats[m]; !found {
			t.Errorf("unexpected tube metric: %v", m)
		}
	}
}

func TestNewBeanstalkdCollector(t *testing.T) {
	tests := []struct {
		beanstalkd                  *beanstalkd.Server
		opts                        CollectorOpts
		expectedError               error
		expectedSystemMetricsLength int
		expectedTubeMetricsLength   int
	}{
		// We expect validation to return errors.
		{
			beanstalkd:                  beanstalkd.NewServer("localhost:11300"),
			opts:                        CollectorOpts{SystemMetrics: []string{"does_not_exist"}},
			expectedError:               fmt.Errorf("Unknown system metric: does_not_exist"),
			expectedSystemMetricsLength: 0,
			expectedTubeMetricsLength:   0,
		},
		// We expect an initialised collector when there are no errors.
		{
			beanstalkd:                  beanstalkd.NewServer("localhost:11300"),
			opts:                        CollectorOpts{},
			expectedError:               nil,
			expectedSystemMetricsLength: len(systemMetricsToStats),
			expectedTubeMetricsLength:   0,
		},
		{
			beanstalkd:                  beanstalkd.NewServer("localhost:11300"),
			opts:                        CollectorOpts{Tubes: []string{"default"}},
			expectedError:               nil,
			expectedSystemMetricsLength: len(systemMetricsToStats),
			expectedTubeMetricsLength:   len(tubeMetricsToStats),
		},
	}

	for _, tt := range tests {
		c, err := NewBeanstalkdCollector(tt.beanstalkd, tt.opts, log.NewNopLogger())
		if !reflect.DeepEqual(tt.expectedError, err) {
			t.Errorf("expected error %v, actual %v", tt.expectedError, err)
		}
		if err != nil && c != nil {
			t.Error("expected nil collector because of error")
		}
		if tt.expectedError == nil && tt.expectedSystemMetricsLength != len(c.systemMetrics) {
			t.Errorf(
				"expected system metrics length %v, actual %v",
				tt.expectedSystemMetricsLength,
				len(c.systemMetrics),
			)
		}
		if tt.expectedError == nil && tt.expectedTubeMetricsLength != len(c.tubesMetrics) {
			t.Errorf(
				"expected tube metrics length %v, actual %v",
				tt.expectedTubeMetricsLength,
				len(c.tubesMetrics),
			)
		}
	}
}

func TestHealtyBeanstalkdServer(t *testing.T) {
	collector, err := NewBeanstalkdCollector(
		mockHealthyBeanstalkd(),
		CollectorOpts{
			SystemMetrics: []string{"current_jobs_urgent_count", "current_jobs_ready_count"},
			Tubes:         []string{"default", "anotherTube"},
			TubeMetrics:   []string{"tube_current_jobs_urgent_count", "tube_current_jobs_ready_count"},
		},
		log.NewNopLogger(),
	)
	if err != nil {
		t.Errorf("expected nil error, actual %v", err)
	}

	ch := make(chan prometheus.Metric)

	go func() {
		defer close(ch)
		collector.Collect(ch)
	}()

	// "up" gauge
	if expected, actual := 1., readGauge((<-ch).(prometheus.Gauge)); expected != actual {
		t.Errorf("expected 'up' value %v, actual %v", expected, actual)
	}

	// "total scrapes" counter
	if expected, actual := 1., readCounter((<-ch).(prometheus.Counter)); expected != actual {
		t.Errorf("expected 'totalScrapes' value %v, actual %v", expected, actual)
	}

	// system metrics & tube metrics gauges
	expectedTotal := 6 // 2 system metrics, 4 tube metrics (2 + 2 labels)
	actualTotal := 0
	for range ch {
		actualTotal++
	}
	if expectedTotal != actualTotal {
		t.Errorf("expected %d metrics, actual %d", expectedTotal, actualTotal)
	}
}

/********************     MOCKS     ********************/

type mockBeanstalkdServer struct {
	stats           beanstalkd.ServerStats
	statsError      error
	tubesStats      beanstalkd.ManyTubeStats
	tubesStatsError error
}

func (m *mockBeanstalkdServer) FetchStats() (beanstalkd.ServerStats, error) {
	return m.stats, m.statsError
}

func (m *mockBeanstalkdServer) FetchTubesStats(tubes map[string]bool) (beanstalkd.ManyTubeStats, error) {
	return m.tubesStats, m.tubesStatsError
}

func mockHealthyBeanstalkd() *mockBeanstalkdServer {
	return &mockBeanstalkdServer{
		stats: beanstalkd.ServerStats{
			"current-jobs-urgent": "10",
			"current-jobs-ready":  "20",
		},
		statsError: nil,
		tubesStats: beanstalkd.ManyTubeStats{
			"default": beanstalkd.TubeStatsOrError{
				Stats: beanstalkd.TubeStats{
					"current-jobs-urgent": "5",
					"current-jobs-ready":  "10",
				},
				Err: nil,
			},
			"anotherTube": beanstalkd.TubeStatsOrError{
				Stats: beanstalkd.TubeStats{
					"current-jobs-urgent": "1",
					"current-jobs-ready":  "2",
				},
				Err: nil,
			},
		},
		tubesStatsError: nil,
	}
}
