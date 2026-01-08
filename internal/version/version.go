package version

import "strings"

var (
	Version   = "dev"
	BuildID   = ""
	BuildDate = ""
)

func FormatBuildDate(value string) string {
	if value == "" {
		return "unknown"
	}
	splitPos := strings.IndexByte(value, 'T')
	if splitPos == -1 {
		splitPos = strings.IndexByte(value, ' ')
	}
	candidate := value
	if splitPos != -1 {
		candidate = value[:splitPos]
	}
	if len(candidate) >= 10 && isDate(candidate) {
		return candidate[:10]
	}
	return candidate
}

func isDate(value string) bool {
	if len(value) < 10 {
		return false
	}
	return value[4] == '-' && value[7] == '-' &&
		isDigit(value[0]) && isDigit(value[1]) && isDigit(value[2]) && isDigit(value[3]) &&
		isDigit(value[5]) && isDigit(value[6]) && isDigit(value[8]) && isDigit(value[9])
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
