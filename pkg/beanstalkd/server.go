package beanstalkd

import (
	"github.com/kr/beanstalk"
	"github.com/prometheus/common/log"
)

// Server can be used to obtain stats from beanstalkd.
type Server struct {
	Address    string
	connection *beanstalk.Conn
}

// NewServer returns an initialised Server.
func NewServer(address string) *Server {
	return &Server{
		Address:    address,
		connection: nil,
	}
}

// FetchStats returns the server stats from beanstalkd.
func (s *Server) FetchStats() (ServerStats, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	stats, err := c.Stats()
	if err != nil {
		log.Errorf("Failed to get server stats: %v", err)
		// Fetching stats failed, so maybe there's a connection problem.
		s.connection = nil
	}
	return stats, err
}

// FetchTubesStats returns the tube stats from beanstalkd.
// The result is a map of stats per tube.
func (s *Server) FetchTubesStats(tubes map[string]bool) (ManyTubeStats, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	allTubes, err := c.ListTubes()
	if err != nil {
		log.Errorf("Failed to list tubes: %v", err)
		return nil, err
	}
	tubesStats := make(ManyTubeStats)
	if len(allTubes) == 0 {
		log.Debug("There are no tubes")
		return tubesStats, nil
	}
	for _, tube := range allTubes {
		if _, ok := tubes[tube]; ok {
			tStats, err := s.tubeStats(tube)
			if err != nil {
				log.Errorf("Failed to fetch tube stats: %v", err)
			}
			tubesStats[tube] = TubeStatsOrError{
				Stats: tStats,
				Err:   err,
			}
		}
	}
	return tubesStats, nil
}

func (s *Server) tubeStats(tubeName string) (TubeStats, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	tube := beanstalk.Tube{
		Conn: c,
		Name: tubeName,
	}
	stats, err := tube.Stats()
	if err != nil {
		log.Errorf("Failed to get tube stats: %v", err)
		// Fetching stats failed, so maybe there's a connection problem.
		s.connection = nil
	}
	return stats, err
}

func (s *Server) connect() (*beanstalk.Conn, error) {
	if s.connection != nil {
		return s.connection, nil
	}
	c, err := beanstalk.Dial("tcp", s.Address)
	s.connection = c
	if err != nil {
		log.Errorf("Can't connect to beanstalkd: %v", err)
	}
	return s.connection, err
}
