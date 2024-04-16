package cmd

import (
	"reflect"
	"testing"
)

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
