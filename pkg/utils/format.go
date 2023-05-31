package utils

import (
	"strings"
)

func FormatQuery(q string) string {
	return strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " ")
}

func HasNil(slice []interface{}) bool {
	for _, v := range slice {
		if v == nil {
			return true
		}
	}
	return false
}
