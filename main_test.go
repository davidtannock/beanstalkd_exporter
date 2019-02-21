package main

import (
	"reflect"
	"testing"

	"gopkg.in/alecthomas/kingpin.v2"
)

func TestInitApplication(t *testing.T) {
	expectedFlags := []string{
		"web.listen-address",
		"web.telemetry-path",
		"beanstalkd.address",
		"beanstalkd.systemMetrics",
		"beanstalkd.allTubes",
		"beanstalkd.tubes",
		"beanstalkd.tubeMetrics",
		"log.level",
		"log.format",
	}

	app := kingpin.New("Test", "")
	initApplication(app)

	for _, flag := range expectedFlags {
		actualFlag := app.GetFlag(flag)
		if actualFlag == nil {
			t.Errorf("expected flag %v, actual nil", flag)
		}
	}
}

func TestToStringArray(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{input: "", expected: []string{}},
		{input: "  ", expected: []string{}},
		{input: ",", expected: []string{}},
		{input: " , ", expected: []string{}},
		{input: "abc", expected: []string{"abc"}},
		{input: " abc ", expected: []string{"abc"}},
		{input: "abc,", expected: []string{"abc"}},
		{input: ",abc", expected: []string{"abc"}},
		{input: " abc, ", expected: []string{"abc"}},
		{input: " , abc , ", expected: []string{"abc"}},
		{input: "abc,def", expected: []string{"abc", "def"}},
		{input: " abc,def ", expected: []string{"abc", "def"}},
		{input: " abc , def , ", expected: []string{"abc", "def"}},
		{input: " , , abc , def,ghi ,", expected: []string{"abc", "def", "ghi"}},
	}

	for _, tt := range tests {
		actual := toStringArray(tt.input)
		if !reflect.DeepEqual(tt.expected, actual) {
			t.Errorf("expected %v for input %q, actual %v", tt.expected, tt.input, actual)
		}
	}
}
