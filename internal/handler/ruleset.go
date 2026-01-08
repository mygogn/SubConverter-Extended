package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/ruleset"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

func (h *Handler) GetRuleset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	args := r.URL.Query()
	urlParam := utils.UrlSafeBase64Decode(args.Get("url"))
	typeStr := args.Get("type")
	group := utils.UrlSafeBase64Decode(args.Get("group"))
	typeInt, _ := strconv.Atoi(typeStr)

	if urlParam == "" || typeStr == "" || (typeInt == 2 && group == "") || typeInt < 1 || typeInt > 6 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	lines := make([]string, 0)
	for _, item := range utils.Split(urlParam, "|") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		lines = append(lines, "ruleset,"+item)
	}
	if len(lines) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	configs := config.ParseRulesetsFromINI(lines)
	contents, err := ruleset.RefreshRulesets(configs, h.Fetcher.FetchRuleset)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	var converted strings.Builder
	for _, item := range contents {
		converted.WriteString(ruleset.ConvertRuleset(item.Content, item.Type))
	}
	if converted.Len() == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	output := buildRulesetOutput(converted.String(), typeInt, group)
	if output == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}
	_, _ = w.Write([]byte(output))
}

func buildRulesetOutput(input string, typ int, group string) string {
	delimiter := utils.GetLineBreak(input)
	lines := utils.Split(input, string(delimiter))
	var out strings.Builder
	if typ == 3 || typ == 4 || typ == 6 {
		out.WriteString("payload:\n")
	}

	for _, line := range lines {
		if strings.Contains(line, "//") {
			line = line[:strings.Index(line, "//")]
			line = utils.TrimWhitespace(line, true, true)
		}
		switch typ {
		case 2:
			if !startsWithAny(line, ruleset.QuanXRuleTypes) {
				continue
			}
		case 1:
			if !startsWithAny(line, ruleset.SurgeRuleTypes) {
				continue
			}
		case 3:
			if !(strings.HasPrefix(line, "DOMAIN-SUFFIX,") || strings.HasPrefix(line, "DOMAIN,")) {
				continue
			}
			domain, ok := extractRuleValue(line)
			if !ok {
				continue
			}
			out.WriteString("  - '")
			if strings.HasSuffix(strings.SplitN(line, ",", 2)[0], "X") {
				out.WriteString("+.")
			}
			out.WriteString(domain)
			out.WriteString("'\n")
			continue
		case 4:
			if !(strings.HasPrefix(line, "IP-CIDR,") || strings.HasPrefix(line, "IP-CIDR6,")) {
				continue
			}
			value, ok := extractRuleValue(line)
			if !ok {
				continue
			}
			out.WriteString("  - '")
			out.WriteString(value)
			out.WriteString("'\n")
			continue
		case 5:
			if !(strings.HasPrefix(line, "DOMAIN-SUFFIX,") || strings.HasPrefix(line, "DOMAIN,")) {
				continue
			}
			domain, ok := extractRuleValue(line)
			if !ok {
				continue
			}
			if strings.HasSuffix(strings.SplitN(line, ",", 2)[0], "X") {
				out.WriteString(".")
			}
			out.WriteString(domain)
			out.WriteByte('\n')
			continue
		case 6:
			if !startsWithAny(line, ruleset.ClashRuleTypes) {
				continue
			}
			out.WriteString("  - ")
		}

		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if typ == 2 {
			if strings.HasPrefix(line, "IP-CIDR6") {
				line = "IP6-CIDR" + line[len("IP-CIDR6"):]
			}
			line = addQuanXGroup(line, group)
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}

	result := out.String()
	if result == "payload:\n" {
		switch typ {
		case 3:
			result += "  - '--placeholder--'"
		case 4:
			result += "  - '0.0.0.0/32'"
		case 6:
			result += "  - 'DOMAIN,--placeholder--'"
		}
	}
	return result
}

func extractRuleValue(line string) (string, bool) {
	first := strings.IndexByte(line, ',')
	if first == -1 {
		return "", false
	}
	second := strings.IndexByte(line[first+1:], ',')
	if second == -1 {
		return strings.TrimSpace(line[first+1:]), true
	}
	return strings.TrimSpace(line[first+1 : first+1+second]), true
}

func addQuanXGroup(line, group string) string {
	parts := utils.Split(line, ",")
	if len(parts) < 2 {
		return line
	}
	if len(parts) >= 3 && strings.TrimSpace(parts[2]) == "no-resolve" {
		return utils.Join([]string{parts[0], parts[1], group, "no-resolve"}, ",")
	}
	return utils.Join([]string{parts[0], parts[1], group}, ",")
}

func startsWithAny(value string, list []string) bool {
	for _, prefix := range list {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
