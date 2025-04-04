// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package utils

import (
	"slices"
	"strings"
)

// AggregateCols Take in a default set of columns, then modify it based on the second parameter
// eg: status,startdate,name and "+tags" show status,startdate,name,tags
// eg: status,startdate,name and "-startdate" show status,name
// eg: status,startdate,name and "name" show name
func AggregateCols(def []string, modify string) []string {
	if modify[0] == '+' {
		for _, c := range strings.Split(modify[1:], ",") {
			if !slices.Contains(def, c) {
				def = append(def, c)
			}
		}

		return def
	} else {
		return strings.Split(modify, ",")
	}
}
