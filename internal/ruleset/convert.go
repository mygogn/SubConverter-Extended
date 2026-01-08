package ruleset

import (
	"regexp"
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

var payloadHeaderRe = regexp.MustCompile(`(?is)^payload:\r?\n`)

func ConvertRuleset(content string, typ config.RulesetType) string {
	if typ == config.RulesetSurge {
		return content
	}
	if payloadHeaderRe.MatchString(content) {
		return convertClashPayload(content, typ == config.RulesetClashClassical)
	}
	return convertQuanXRuleset(content)
}

func convertClashPayload(content string, classical bool) string {
	out := payloadHeaderRe.ReplaceAllString(content, "")
	out = regexp.MustCompile(`(?m)^\s*-\s+('|")?(.*?)\1\s*$`).ReplaceAllString(out, "\n$2")
	if classical {
		return out
	}
	var sb strings.Builder
	lines := utils.Split(out, string(utils.GetLineBreak(out)))
	for _, line := range lines {
		line = utils.TrimWhitespace(line, true, true)
		if line == "" {
			continue
		}
		if strings.Contains(line, "//") {
			line = line[:strings.Index(line, "//")]
			line = utils.TrimWhitespace(line, true, true)
		}
		if line == "" {
			continue
		}
		if strings.Contains(line, "/") {
			if utils.IsIPv4(line[:strings.Index(line, "/")]) {
				sb.WriteString("IP-CIDR,")
			} else {
				sb.WriteString("IP-CIDR6,")
			}
			sb.WriteString(line)
			sb.WriteByte('\n')
			continue
		}

		if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "+.") {
			keyword := false
			for strings.HasSuffix(line, ".*") {
				keyword = true
				line = line[:len(line)-2]
			}
			if keyword {
				sb.WriteString("DOMAIN-KEYWORD,")
			} else {
				sb.WriteString("DOMAIN-SUFFIX,")
			}
			if strings.HasPrefix(line, "+.") {
				line = line[2:]
			} else {
				line = line[1:]
			}
			sb.WriteString(line)
			sb.WriteByte('\n')
			continue
		}
		sb.WriteString("DOMAIN,")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func convertQuanXRuleset(content string) string {
	lines := utils.Split(content, string(utils.GetLineBreak(content)))
	var sb strings.Builder
	for _, line := range lines {
		line = utils.TrimWhitespace(line, true, true)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		line = replacePrefixFold(line, "host-keyword", "DOMAIN-KEYWORD")
		line = replacePrefixFold(line, "host-suffix", "DOMAIN-SUFFIX")
		line = replacePrefixFold(line, "host", "DOMAIN")
		line = replacePrefixFold(line, "ip6-cidr", "IP-CIDR6")

		fields := utils.Split(line, ",")
		if len(fields) < 2 {
			continue
		}
		typ := strings.ToUpper(strings.TrimSpace(fields[0]))
		value := strings.TrimSpace(fields[1])
		if value == "" {
			continue
		}
		sb.WriteString(typ)
		sb.WriteByte(',')
		sb.WriteString(value)
		if len(fields) > 2 {
			for _, field := range fields[2:] {
				if strings.TrimSpace(field) == "no-resolve" {
					sb.WriteString(",no-resolve")
					break
				}
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func replacePrefixFold(line, prefix, replacement string) string {
	if strings.HasPrefix(strings.ToLower(line), prefix) {
		return replacement + line[len(prefix):]
	}
	return line
}
