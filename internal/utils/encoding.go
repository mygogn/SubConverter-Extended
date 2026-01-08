package utils

import (
	"encoding/base64"
	"net/url"
	"strings"
)

func UrlEncode(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' || ch == '~' {
			b.WriteByte(ch)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(toHex(ch >> 4))
		b.WriteByte(toHex(ch & 0x0f))
	}
	return b.String()
}

func UrlDecode(value string) string {
	var b strings.Builder
	for i := 0; i < len(value); i++ {
		ch := value[i]
		switch ch {
		case '+':
			b.WriteByte(' ')
		case '%':
			if i+2 >= len(value) {
				return b.String()
			}
			hi := fromHex(value[i+1])
			lo := fromHex(value[i+2])
			if hi >= 0 && lo >= 0 {
				b.WriteByte(byte(hi<<4 | lo))
				i += 2
				continue
			}
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}

func UrlSafeBase64Encode(value string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(value))
	encoded = strings.TrimRight(encoded, "=")
	encoded = strings.ReplaceAll(encoded, "+", "-")
	encoded = strings.ReplaceAll(encoded, "/", "_")
	return encoded
}

func UrlSafeBase64Decode(value string) string {
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "-", "+")
	value = strings.ReplaceAll(value, "_", "/")
	if pad := len(value) % 4; pad != 0 {
		value += strings.Repeat("=", 4-pad)
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func Base64Encode(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func Base64Decode(value string) string {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return ""
	}
	return string(decoded)
}

func JoinArguments(args map[string]string) string {
	if len(args) == 0 {
		return ""
	}
	var b strings.Builder
	first := true
	for k, v := range args {
		if !first {
			b.WriteByte('&')
		}
		first = false
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(UrlEncode(v))
	}
	return b.String()
}

func ParseQueryArg(rawURL, key string) string {
	if rawURL == "" || key == "" {
		return ""
	}
	idx := strings.LastIndex(rawURL, key+"=")
	for idx != -1 {
		if idx == 0 || rawURL[idx-1] == '&' || rawURL[idx-1] == '?' {
			start := idx + len(key) + 1
			end := strings.IndexByte(rawURL[start:], '&')
			if end == -1 {
				return rawURL[start:]
			}
			return rawURL[start : start+end]
		}
		if idx == 0 {
			break
		}
		idx = strings.LastIndex(rawURL[:idx], key+"=")
	}
	return ""
}

func ParseQueryParams(rawQuery string) map[string]string {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return map[string]string{}
	}
	result := make(map[string]string, len(values))
	for key, vals := range values {
		if len(vals) > 0 {
			result[key] = vals[0]
		} else {
			result[key] = ""
		}
	}
	return result
}

func toHex(x byte) byte {
	if x > 9 {
		return x + 55
	}
	return x + 48
}

func fromHex(x byte) int {
	switch {
	case x >= 'A' && x <= 'Z':
		return int(x - 'A' + 10)
	case x >= 'a' && x <= 'z':
		return int(x - 'a' + 10)
	case x >= '0' && x <= '9':
		return int(x - '0')
	}
	return -1
}
