package beanstalkd

// ServerStats is the map of beanstalkd server stats.
type ServerStats map[string]string

// TubeStats is the map of beanstalkd tube stats.
type TubeStats map[string]string

// TubeStatsOrError is used when fetching tube stats from
// beanstalkd succeeds or fails.
type TubeStatsOrError struct {
	Stats TubeStats
	Err   error
}

// ManyTubeStats is the collection of tube stats
// (or errors) for multiple tubes.
type ManyTubeStats map[string]TubeStatsOrError
