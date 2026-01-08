package utils

import (
	"strconv"
)

func ToInt(value string, def int) int {
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}

func ToString(value int) string {
	return strconv.Itoa(value)
}
