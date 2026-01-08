package utils

import "strings"

func IsLink(url string) bool {
	return StartsWith(url, "https://") || StartsWith(url, "http://") || StartsWith(url, "data:")
}

func IsIPv4(address string) bool {
	if address == "" {
		return false
	}
	parts := strings.Split(address, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		if len(part) > 1 && part[0] == '0' {
			// allow "0" but reject leading zeros for compatibility
			return false
		}
		val := 0
		for i := 0; i < len(part); i++ {
			ch := part[i]
			if ch < '0' || ch > '9' {
				return false
			}
			val = val*10 + int(ch-'0')
			if val > 255 {
				return false
			}
		}
	}
	return true
}

func IsIPv6(address string) bool {
	if address == "" {
		return false
	}
	// Lightweight check: must contain ':' and not contain invalid chars.
	if !strings.Contains(address, ":") {
		return false
	}
	for i := 0; i < len(address); i++ {
		ch := address[i]
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') || ch == ':' || ch == '.' {
			continue
		}
		return false
	}
	return true
}
