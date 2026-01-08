package ruleset

import (
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

func BuildClashRules(contents []config.RulesetContent, maxRules int) []string {
	var rules []string
	for _, item := range contents {
		if maxRules > 0 && len(rules) >= maxRules {
			break
		}
		group := item.Group
		if strings.HasPrefix(item.Content, "[]") {
			line := strings.TrimSpace(item.Content[2:])
			if strings.HasPrefix(line, "FINAL") {
				line = "MATCH" + strings.TrimPrefix(line, "FINAL")
			}
			if line != "" {
				rules = append(rules, transformRuleToCommon(line, group))
			}
			continue
		}
		converted := ConvertRuleset(item.Content, item.Type)
		if converted == "" {
			continue
		}
		lines := utils.Split(converted, string(utils.GetLineBreak(converted)))
		for _, line := range lines {
			if maxRules > 0 && len(rules) >= maxRules {
				break
			}
			line = utils.TrimWhitespace(line, true, true)
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
				continue
			}
			if strings.Contains(line, "//") {
				line = strings.TrimSpace(line[:strings.Index(line, "//")])
			}
			if line == "" {
				continue
			}
			if !startsWithAny(line, ClashRuleTypes) {
				continue
			}
			switch {
			case strings.HasPrefix(line, "AND") || strings.HasPrefix(line, "OR") || strings.HasPrefix(line, "NOT"):
				rules = append(rules, line+","+group)
			case strings.HasPrefix(line, "SUB-RULE") || strings.HasPrefix(line, "RULE-SET"):
				rules = append(rules, line)
			default:
				rules = append(rules, transformRuleToCommon(line, group))
			}
		}
	}
	return rules
}

func transformRuleToCommon(input, group string) string {
	parts := utils.Split(input, ",")
	if len(parts) < 2 {
		return input + "," + group
	}
	out := []string{parts[0], parts[1], group}
	if len(parts) > 2 {
		out = append(out, parts[2:]...)
	}
	return utils.Join(out, ",")
}

func startsWithAny(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
