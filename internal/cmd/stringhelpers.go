package cmd

import (
	"strings"
)

func toStringArray(flag string) []string {
	flags := []string{}
	for _, part := range strings.Split(flag, ",") {
		if s := strings.Trim(part, " "); s != "" {
			flags = append(flags, s)
		}
	}
	return flags
}
