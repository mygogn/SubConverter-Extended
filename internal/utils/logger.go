package utils

import (
	"log"
)

const (
	LogFatal = iota
	LogError
	LogWarning
	LogInfo
	LogVerbose
	LogDebug
)

var logLevel = LogInfo

func SetLogLevel(level int) {
	logLevel = level
}

func Logf(level int, format string, args ...interface{}) {
	if level > logLevel {
		return
	}
	log.Printf(format, args...)
}
