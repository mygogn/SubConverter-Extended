package config

import "strings"

const (
	LogFatal = iota
	LogError
	LogWarning
	LogInfo
	LogVerbose
	LogDebug
)

func ParseLogLevel(value string) int {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "fatal":
		return LogFatal
	case "error":
		return LogError
	case "warn", "warning":
		return LogWarning
	case "verbose":
		return LogVerbose
	case "debug":
		return LogDebug
	default:
		return LogInfo
	}
}
