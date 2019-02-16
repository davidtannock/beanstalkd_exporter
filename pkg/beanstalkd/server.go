package beanstalkd

import (
	"net"

	"github.com/kr/beanstalk"
	"github.com/prometheus/common/log"
)

type beanstalkdConnection interface {
	Stats() (map[string]string, error)
	ListTubes() ([]string, error)
}

type beanstalkdTube interface {
	Stats() (map[string]string, error)
}

type beanstalkdDialer interface {
	Dial(network, address string) (net.Conn, error)
}

type logger interface {
	Debug(...interface{})
	Errorf(string, ...interface{})
}

// Server can be used to obtain stats from beanstalkd.
type Server struct {
	// Address is the address of the beanstalkd instance.
	Address string

	connection beanstalkdConnection
	dialer     beanstalkdDialer
	tubes      map[string]beanstalkdTube
	logger     logger
}

// NewServer returns an initialised Server.
func NewServer(address string) *Server {
	return &Server{
		Address:    address,
		connection: nil,
		dialer:     &net.Dialer{},
		tubes:      nil,
		logger:     log.Base(),
	}
}

// ListTubes returns the list of tubes from beanstalkd.
func (s *Server) ListTubes() ([]string, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	tubes, err := c.ListTubes()
	return tubes, err
}

// FetchStats returns the server stats from beanstalkd.
func (s *Server) FetchStats() (ServerStats, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	stats, err := c.Stats()
	if err != nil {
		s.logError("Failed to get server stats: %v", err)
		// Fetching stats failed, so maybe there's a connection problem.
		s.connection = nil
		s.tubes = make(map[string]beanstalkdTube)
	}
	return stats, err
}

// FetchTubesStats returns the tube stats from beanstalkd.
// The result is a map of stats per tube.
func (s *Server) FetchTubesStats(tubes map[string]bool) (ManyTubeStats, error) {
	if len(tubes) == 0 {
		return nil, nil
	}
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	allTubes, err := c.ListTubes()
	if err != nil {
		s.logError("Failed to list tubes: %v", err)
		return nil, err
	}
	if len(allTubes) == 0 {
		s.logDebug("There are no tubes")
		return nil, nil
	}
	tubesStats := make(ManyTubeStats)
	for _, tube := range allTubes {
		if _, ok := tubes[tube]; ok {
			tStats, err := s.tubeStats(tube)
			if err != nil {
				s.logError("Failed to fetch tube stats: %v", err)
			}
			tubesStats[tube] = TubeStatsOrError{
				Stats: tStats,
				Err:   err,
			}
		}
	}
	if len(tubesStats) == 0 {
		tubesStats = nil
	}
	return tubesStats, nil
}

func (s *Server) tubeStats(tubeName string) (TubeStats, error) {
	tube, err := s.initTube(tubeName)
	if err != nil {
		return nil, err
	}
	stats, err := tube.Stats()
	if err != nil {
		s.logError("Failed to get tube stats: %v", err)
		// Fetching stats failed, so maybe there's a connection problem.
		s.connection = nil
		s.tubes = make(map[string]beanstalkdTube)
	}
	return stats, err
}

func (s *Server) connect() (beanstalkdConnection, error) {
	if s.connection != nil {
		return s.connection, nil
	}
	c, err := s.dial()
	s.connection = c
	if err != nil {
		s.logError("Can't connect to beanstalkd: %v", err)
	}
	return s.connection, err
}

func (s *Server) dial() (beanstalkdConnection, error) {
	c, err := s.dialer.Dial("tcp", s.Address)
	if err != nil {
		return nil, err
	}
	return beanstalk.NewConn(c), nil
}

func (s *Server) initTube(tubeName string) (beanstalkdTube, error) {
	c, err := s.connect()
	if err != nil {
		return nil, err
	}
	if t, exists := s.tubes[tubeName]; exists {
		return t, nil
	}
	tube := &beanstalk.Tube{
		Conn: c.(*beanstalk.Conn),
		Name: tubeName,
	}
	if s.tubes == nil {
		s.tubes = make(map[string]beanstalkdTube)
	}
	s.tubes[tubeName] = tube
	return tube, nil
}

func (s *Server) logError(msg string, err error) {
	if s.logger != nil {
		s.logger.Errorf(msg, err)
	}
}

func (s *Server) logDebug(msg string) {
	if s.logger != nil {
		s.logger.Debug(msg)
	}
}
