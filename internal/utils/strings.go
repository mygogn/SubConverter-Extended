package utils

import (
	"strings"
)

func Split(s, separator string) []string {
	if separator == "" {
		return []string{s}
	}
	var result []string
	bpos := 0
	epos := strings.Index(s, separator)
	for bpos < len(s) {
		if epos == -1 {
			epos = len(s)
		}
		result = append(result, s[bpos:epos])
		bpos = epos + len(separator)
		if bpos > len(s) {
			break
		}
		epos = strings.Index(s[bpos:], separator)
		if epos != -1 {
			epos += bpos
		}
	}
	return result
}

func Join(arr []string, delimiter string) string {
	if len(arr) == 0 {
		return ""
	}
	if len(arr) == 1 {
		return arr[0]
	}
	var b strings.Builder
	b.WriteString(arr[0])
	for i := 1; i < len(arr); i++ {
		b.WriteString(delimiter)
		b.WriteString(arr[i])
	}
	return b.String()
}

func TrimOf(s string, target rune, before bool, after bool) string {
	if !before && !after {
		return s
	}
	if before && after {
		return strings.Trim(s, string(target))
	}
	if before {
		return strings.TrimLeft(s, string(target))
	}
	return strings.TrimRight(s, string(target))
}

func Trim(s string, before bool, after bool) string {
	return TrimOf(s, ' ', before, after)
}

func TrimWhitespace(s string, before bool, after bool) string {
	if !before && !after {
		return s
	}
	if before && after {
		return strings.TrimSpace(s)
	}
	if before {
		return strings.TrimLeftFunc(s, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\f' || r == '\v' || r == '\n' || r == '\r'
		})
	}
	return strings.TrimRightFunc(s, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\f' || r == '\v' || r == '\n' || r == '\r'
	})
}

func StartsWith(hay, needle string) bool {
	return strings.HasPrefix(hay, needle)
}

func EndsWith(hay, needle string) bool {
	return strings.HasSuffix(hay, needle)
}

func ReplaceAllDistinct(str, oldValue, newValue string) string {
	if oldValue == "" {
		return str
	}
	for {
		pos := strings.Index(str, oldValue)
		if pos == -1 {
			break
		}
		str = str[:pos] + newValue + str[pos+len(oldValue):]
	}
	return str
}

func CountLeast(hay string, needle rune, cnt int) bool {
	if cnt <= 0 {
		return true
	}
	for _, ch := range hay {
		if ch == needle {
			cnt--
			if cnt == 0 {
				return true
			}
		}
	}
	return false
}

func GetLineBreak(s string) byte {
	if CountLeast(s, '\n', 1) {
		return '\n'
	}
	return '\r'
}

func ProcessEscapeChar(s string) string {
	if !strings.Contains(s, "\\") {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
				continue
			case 'r':
				b.WriteByte('\r')
				i++
				continue
			case 't':
				b.WriteByte('\t')
				i++
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func ProcessEscapeCharReverse(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\n':
			b.WriteString("\\n")
		case '\r':
			b.WriteString("\\r")
		case '\t':
			b.WriteString("\\t")
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}
