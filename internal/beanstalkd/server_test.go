package beanstalkd

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		address         string
		dialTimeout     uint
		keepAlivePeriod uint
		expectedAddress string
		expectedError   error
	}{
		// We expect the address to be set correctly.
		{
			address:         "localhost:11300",
			dialTimeout:     10,
			keepAlivePeriod: 10,
			expectedAddress: "localhost:11300",
			expectedError:   nil,
		},
		// We expect dialTimeout to be validated.
		{
			address:         "localhost:12345",
			dialTimeout:     0,
			keepAlivePeriod: 10,
			expectedAddress: "localhost:12345",
			expectedError:   fmt.Errorf("dialTimeout 0 out of range[1-30]"),
		},
		{
			address:         "localhost:12345",
			dialTimeout:     31,
			keepAlivePeriod: 10,
			expectedAddress: "localhost:12345",
			expectedError:   fmt.Errorf("dialTimeout 31 out of range[1-30]"),
		},
		// We expect keepAlivePeriod to be validated.
		{
			address:         "localhost:54321",
			dialTimeout:     10,
			keepAlivePeriod: 0,
			expectedAddress: "localhost:54321",
			expectedError:   fmt.Errorf("keepAlivePeriod < 1"),
		},
	}

	for _, tt := range tests {
		server, err := NewServer(tt.address, tt.dialTimeout, tt.keepAlivePeriod)
		if server != nil && !reflect.DeepEqual(tt.expectedAddress, server.Address) {
			t.Errorf("expected address %v, actual %v", tt.expectedAddress, server.Address)
		}
		if err != nil && !reflect.DeepEqual(tt.expectedError, err) {
			t.Errorf("expected error %v, actual %v", tt.expectedError, err)
		}
	}
}

func TestListTubes(t *testing.T) {
	conn := &mockConnection{
		tubes:              []string{"default", "one", "two", "three"},
		listTubesCallCount: 0,
	}
	server := &Server{
		Address:    "localhost:11300",
		connection: conn,
	}
	actualTubes, err := server.ListTubes()
	if err != nil {
		t.Error(err)
	}
	expectedTubes := []string{"default", "one", "two", "three"}
	if !reflect.DeepEqual(expectedTubes, actualTubes) {
		t.Errorf("expected tubes %v, actual %v", expectedTubes, actualTubes)
	}
}

func TestFetchStats(t *testing.T) {
	conn := &mockConnection{
		stats: map[string]string{
			"current-jobs-urgent": "10",
			"current-jobs-ready":  "20",
		},
		statsCallCount: 0,
	}
	server := &Server{
		Address:    "localhost:11300",
		connection: conn,
	}
	actualStats, err := server.FetchStats()
	if err != nil {
		t.Error(err)
	}
	expectedStats := ServerStats{
		"current-jobs-urgent": "10",
		"current-jobs-ready":  "20",
	}
	if !reflect.DeepEqual(expectedStats, actualStats) {
		t.Errorf("expected stats %v, actual %v", expectedStats, actualStats)
	}
	if conn.statsCallCount != 1 {
		t.Errorf("expected Stats() to be called 1 time, actual %v", conn.statsCallCount)
	}
}

func TestFetchStatsConnectError(t *testing.T) {
	dialer := &mockDialer{
		conn:      nil,
		connError: fmt.Errorf("Bad network"),
	}
	server := &Server{
		Address: "localhost:11300",
		dialer:  dialer,
	}
	_, err := server.FetchStats()
	if err == nil {
		t.Errorf("expected a connection error, but got nil")
	}
}

func TestFetchStatsError(t *testing.T) {
	errorMessage := "Something bad happened"
	conn := &mockConnection{
		statsError: fmt.Errorf(errorMessage),
	}
	server := &Server{
		Address:    "localhost:11300",
		connection: conn,
	}
	if server.connection == nil {
		t.Errorf("not expecting connection to be nil")
	}
	_, err := server.FetchStats()
	if err == nil {
		t.Error("expected an error, but got nil")
	}
	if err.Error() != errorMessage {
		t.Errorf("expected error %v, actual %v", errorMessage, err.Error())
	}
	if server.connection != nil {
		t.Error("expected connection to be nil")
	}
}

func TestFetchTubesStats(t *testing.T) {
	conn := &mockConnection{
		tubes:              []string{"default", "anotherTube", "errorTube"},
		listTubesCallCount: 0,
	}
	server := &Server{
		Address:    "localhost:11300",
		connection: conn,
		tubes: map[string]beanstalkdTube{
			"default": &mockTube{
				stats: map[string]string{
					"current-jobs-urgent": "10",
					"current-jobs-ready":  "20",
				},
			},
			"anotherTube": &mockTube{
				stats: map[string]string{
					"current-jobs-urgent": "0",
					"current-jobs-ready":  "0",
				},
			},
			"errorTube": &mockTube{
				statsError: fmt.Errorf("Oops"),
			},
		},
	}

	tests := []struct {
		num                        string
		tubes                      map[string]bool
		listTubesError             error
		expectedListTubesCallCount int
		expectedTubesStats         ManyTubeStats
	}{
		// We expect empty tubes stats when we don't specify the tube names.
		{
			num:                        "1) ",
			tubes:                      nil,
			listTubesError:             nil,
			expectedListTubesCallCount: 1,
			expectedTubesStats:         nil,
		},
		// We expect empty tubes stats when the tubes don't exist.
		{
			num:                        "2) ",
			tubes:                      map[string]bool{"doesNotExist": true},
			listTubesError:             nil,
			expectedListTubesCallCount: 1,
			expectedTubesStats:         nil,
		},
		// We expect empty tubes stats when there are errors.
		{
			num:                        "3) ",
			tubes:                      map[string]bool{"default": true},
			listTubesError:             fmt.Errorf("Something went wrong"),
			expectedListTubesCallCount: 1,
			expectedTubesStats:         nil,
		},
		// We expect the stats for the tubes we ask for.
		{
			num:                        "4) ",
			tubes:                      map[string]bool{"anotherTube": true},
			listTubesError:             nil,
			expectedListTubesCallCount: 1,
			expectedTubesStats: ManyTubeStats{
				"anotherTube": TubeStatsOrError{
					Stats: map[string]string{
						"current-jobs-urgent": "0",
						"current-jobs-ready":  "0",
					},
					Err: nil,
				},
			},
		},
		// We expect the stats for the tubes we ask for, even
		// if there are errors for only some tubes.
		{
			num:                        "5) ",
			tubes:                      map[string]bool{"default": true, "errorTube": true},
			listTubesError:             nil,
			expectedListTubesCallCount: 1,
			expectedTubesStats: ManyTubeStats{
				"default": TubeStatsOrError{
					Stats: map[string]string{
						"current-jobs-urgent": "10",
						"current-jobs-ready":  "20",
					},
					Err: nil,
				},
				"errorTube": TubeStatsOrError{
					Stats: nil,
					Err:   fmt.Errorf("Oops"),
				},
			},
		},
	}

	for _, tt := range tests {
		conn.listTubesCallCount = 0
		conn.listTubesError = tt.listTubesError
		actualTubesStats, err := server.FetchTubesStats(tt.tubes)
		if tt.listTubesError == nil && err != nil {
			t.Error(err)
		}
		if tt.expectedListTubesCallCount != conn.listTubesCallCount {
			t.Errorf(
				tt.num+"expected ListTubes() to be called %v times, actual %v",
				tt.expectedListTubesCallCount,
				conn.listTubesCallCount,
			)
		}
		if !reflect.DeepEqual(tt.expectedTubesStats, actualTubesStats) {
			t.Errorf(
				tt.num+"expected tube stats %v, actual %v",
				tt.expectedTubesStats,
				actualTubesStats,
			)
		}
		if tt.listTubesError != nil {
			if err == nil {
				t.Error("expected tubes stats error, but got nil")
			}
			expectedError := "Something went wrong"
			if err.Error() != expectedError {
				t.Errorf("expected error %v, actual %v", expectedError, err.Error())
			}
		}
	}
}

func TestFetchTubesStatsConnectError(t *testing.T) {
	dialer := &mockDialer{
		conn:      nil,
		connError: fmt.Errorf("Bad network"),
	}
	server := &Server{
		Address: "localhost:11300",
		dialer:  dialer,
	}
	_, err := server.FetchTubesStats(map[string]bool{"default": true})
	if err == nil {
		t.Errorf("expected a connection error, but got nil")
	}
}

func TestFetchTubesStatsWithNoListedTubes(t *testing.T) {
	conn := &mockConnection{
		tubes:              []string{},
		listTubesCallCount: 0,
	}
	server := &Server{
		Address:    "localhost:11300",
		connection: conn,
	}
	actualTubesStats, err := server.FetchTubesStats(map[string]bool{"default": true})
	if actualTubesStats != nil {
		t.Errorf("expected nil tubes stats, actual %v", actualTubesStats)
	}
	if err != nil {
		t.Errorf("expected nil tubes stats error, actual %v", err)
	}
}

func TestConnect(t *testing.T) {
	server := &Server{
		Address: "localhost:11300",
		dialer: &mockDialer{
			conn: &mockNetConn{},
		},
	}
	connection, err := server.connect()
	if err != nil {
		t.Errorf("expecting no error, actual %v", err)
	}
	if connection == nil {
		t.Errorf("expecting a connection, but got nil")
	}
}

/********************     MOCKS     ********************/

type mockConnection struct {
	stats              map[string]string
	statsError         error
	statsCallCount     int
	tubes              []string
	listTubesError     error
	listTubesCallCount int
}

func (m *mockConnection) Stats() (map[string]string, error) {
	m.statsCallCount++
	return m.stats, m.statsError
}

func (m *mockConnection) ListTubes() ([]string, error) {
	m.listTubesCallCount++
	return m.tubes, m.listTubesError
}

type mockTube struct {
	stats          map[string]string
	statsError     error
	statsCallCount int
}

func (m *mockTube) Stats() (map[string]string, error) {
	m.statsCallCount++
	return m.stats, m.statsError
}

type mockDialer struct {
	conn      net.Conn
	connError error
}

func (m *mockDialer) Dial(network, address string) (net.Conn, error) {
	return m.conn, m.connError
}

type mockNetConn struct {
	bytes.Buffer
}

func (m *mockNetConn) Close() error {
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr {
	return nil
}

func (m *mockNetConn) RemoteAddr() net.Addr {
	return nil
}

func (m *mockNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *mockNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *mockNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}
