package ruleset

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type ClashRuleProviderOptions struct {
	ManagedPrefix string
	Script        bool
	Classic       bool
	MaxRules      int
}

type ClashRuleProviderResult struct {
	Providers map[string]any
	Rules     []string
	Script    string
}

type ruleMeta struct {
	Name         string
	Group        string
	URL          string
	RuleType     config.RulesetType
	Interval     int
	HasDomain    bool
	HasIPCIDR    bool
	HasNoResolve bool
	Original     bool
	Keywords     []string
}

type scriptRule struct {
	Name       string
	Group      string
	HasDomain  bool
	HasIPCIDR  bool
	Original   bool
	Keywords   []string
}

func BuildClashRuleProviders(contents []config.RulesetContent, opts ClashRuleProviderOptions) ClashRuleProviderResult {
	result := ClashRuleProviderResult{
		Providers: make(map[string]any),
	}
	if len(contents) == 0 {
		return result
	}

	usedNames := make(map[string]int)
	metaOrder := make([]string, 0, len(contents))
	metaMap := make(map[string]*ruleMeta, len(contents))
	matchGroup := ""
	geoips := make(map[string]string)

	for _, item := range contents {
		group := item.Group
		if group == "" {
			continue
		}

		if item.Path == "" && strings.HasPrefix(strings.TrimSpace(item.Content), "[]") {
			line := strings.TrimSpace(item.Content)
			if strings.HasPrefix(line, "[]") {
				line = strings.TrimSpace(line[2:])
			}
			if line == "" {
				continue
			}
			if opts.Script {
				if strings.HasPrefix(line, "MATCH") || strings.HasPrefix(line, "FINAL") {
					matchGroup = group
					continue
				}
				if strings.HasPrefix(line, "GEOIP") {
					parts := utils.Split(line, ",")
					if len(parts) >= 2 {
						geo := strings.TrimSpace(parts[1])
						if geo != "" {
							geoips[geo] = group
						}
					}
				}
				continue
			}
			if strings.HasPrefix(line, "FINAL") {
				line = "MATCH" + strings.TrimPrefix(line, "FINAL")
			}
			appendRule(&result.Rules, opts.MaxRules, transformRuleToCommon(line, group))
			continue
		}

		ruleName := uniqueRuleName(ruleNameFromPath(item.Path), usedNames)
		meta := &ruleMeta{
			Name:     ruleName,
			Group:    group,
			URL:      item.PathTyped,
			RuleType: item.Type,
			Interval: item.UpdateInterval,
		}

		switch item.Type {
		case config.RulesetClashDomain:
			meta.URL = "*" + item.Path
			meta.HasDomain = true
			meta.Original = true
		case config.RulesetClashIPCIDR:
			meta.URL = "*" + item.Path
			meta.HasIPCIDR = true
			meta.Original = true
		case config.RulesetClashClassical:
			meta.URL = "*" + item.Path
		}

		if item.Type == config.RulesetClashDomain || item.Type == config.RulesetClashIPCIDR || item.Type == config.RulesetClashClassical {
			metaOrder = append(metaOrder, ruleName)
			metaMap[ruleName] = meta
			if !opts.Script {
				appendRule(&result.Rules, opts.MaxRules, "RULE-SET,"+ruleName+","+group)
			}
			continue
		}

		if opts.Classic {
			metaOrder = append(metaOrder, ruleName)
			metaMap[ruleName] = meta
			if !opts.Script {
				appendRule(&result.Rules, opts.MaxRules, "RULE-SET,"+ruleName+","+group)
			}
			continue
		}

		if item.Content == "" {
			continue
		}
		converted := ConvertRuleset(item.Content, item.Type)
		if converted == "" {
			continue
		}
		delimiter := utils.GetLineBreak(converted)
		lines := utils.Split(converted, string(delimiter))
		for _, line := range lines {
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
			if strings.HasPrefix(line, "DOMAIN-KEYWORD,") {
				parts := utils.Split(line, ",")
				if len(parts) < 2 {
					continue
				}
				keyword := strings.TrimSpace(parts[1])
				if keyword == "" {
					continue
				}
				if opts.Script {
					meta.Keywords = append(meta.Keywords, keyword)
				} else {
					rule := parts[0] + "," + keyword + "," + group
					if len(parts) > 2 {
						rule = rule + "," + strings.TrimSpace(parts[2])
					}
					appendRule(&result.Rules, opts.MaxRules, rule)
				}
				continue
			}
			if !meta.HasDomain && (strings.HasPrefix(line, "DOMAIN,") || strings.HasPrefix(line, "DOMAIN-SUFFIX,")) {
				meta.HasDomain = true
				continue
			}
			if !meta.HasIPCIDR && (strings.HasPrefix(line, "IP-CIDR,") || strings.HasPrefix(line, "IP-CIDR6,")) {
				meta.HasIPCIDR = true
				if strings.Contains(line, ",no-resolve") {
					meta.HasNoResolve = true
				}
			}
		}

		metaOrder = append(metaOrder, ruleName)
		metaMap[ruleName] = meta
		if !opts.Script {
			if meta.HasDomain {
				appendRule(&result.Rules, opts.MaxRules, "RULE-SET,"+ruleName+" (Domain),"+group)
			}
			if meta.HasIPCIDR {
				rule := "RULE-SET," + ruleName + " (IP-CIDR)," + group
				if meta.HasNoResolve {
					rule += ",no-resolve"
				}
				appendRule(&result.Rules, opts.MaxRules, rule)
			}
			if !meta.HasDomain && !meta.HasIPCIDR {
				appendRule(&result.Rules, opts.MaxRules, "RULE-SET,"+ruleName+","+group)
			}
		}
	}

	for _, name := range metaOrder {
		meta := metaMap[name]
		if meta == nil || meta.URL == "" {
			continue
		}
		if meta.HasDomain {
			key := meta.Name
			if meta.RuleType != config.RulesetClashDomain {
				key = meta.Name + " (Domain)"
			}
			result.Providers[key] = buildProviderEntry(meta, "domain", 3, "_domain.yaml")
		}
		if meta.HasIPCIDR {
			key := meta.Name
			if meta.RuleType != config.RulesetClashIPCIDR {
				key = meta.Name + " (IP-CIDR)"
			}
			result.Providers[key] = buildProviderEntry(meta, "ipcidr", 4, "_ipcidr.yaml")
		}
		if !meta.HasDomain && !meta.HasIPCIDR {
			result.Providers[meta.Name] = buildProviderEntry(meta, "classical", 6, ".yaml")
		}
	}

	if opts.Script {
		scriptRules := make([]scriptRule, 0, len(metaOrder))
		for _, name := range metaOrder {
			meta := metaMap[name]
			if meta == nil {
				continue
			}
			scriptRules = append(scriptRules, scriptRule{
				Name:      meta.Name,
				Group:     meta.Group,
				HasDomain: meta.HasDomain,
				HasIPCIDR: meta.HasIPCIDR,
				Original:  meta.Original,
				Keywords:  meta.Keywords,
			})
		}
		result.Script = buildClashScript(scriptRules, matchGroup, geoips)
	}

	return result
}

func buildProviderEntry(meta *ruleMeta, behavior string, typ int, suffix string) map[string]any {
	url := meta.URL
	pathHash := hashRuleURL(url)
	entry := map[string]any{
		"type":     "http",
		"behavior": behavior,
		"path":     "./providers/" + pathHash + suffix,
	}
	entry["url"] = rulesetURL(meta.URL, typ, meta.URL, meta.Interval, meta.ManagedPrefix())
	if meta.Interval > 0 {
		entry["interval"] = meta.Interval
	}
	return entry
}

func rulesetURL(raw string, typ int, url string, _ int, prefix string) string {
	target := raw
	if strings.HasPrefix(target, "*") {
		return target[1:]
	}
	base := strings.TrimRight(prefix, "/")
	encoded := utils.UrlSafeBase64Encode(target)
	if base == "" {
		return "/getruleset?type=" + strconv.Itoa(typ) + "&url=" + encoded
	}
	return base + "/getruleset?type=" + strconv.Itoa(typ) + "&url=" + encoded
}

func (m *ruleMeta) ManagedPrefix() string {
	return ""
}

func buildClashScript(rules []scriptRule, matchGroup string, geoips map[string]string) string {
	var b strings.Builder
	b.WriteString("def main(ctx, md):\n")
	b.WriteString("  host = md[\"host\"]\n")
	for _, rule := range rules {
		if (!rule.HasDomain && !rule.HasIPCIDR) || rule.Original {
			b.WriteString(fmt.Sprintf("  if ctx.rule_providers[%q].match(md):\n", rule.Name))
			b.WriteString(fmt.Sprintf("    ctx.log('[Script] matched %s rule')\n", rule.Group))
			b.WriteString(fmt.Sprintf("    return %q\n", rule.Group))
			continue
		}
		if rule.HasDomain {
			b.WriteString(fmt.Sprintf("  if ctx.rule_providers[%q].match(md):\n", rule.Name+"_domain"))
			b.WriteString(fmt.Sprintf("    ctx.log('[Script] matched %s DOMAIN rule')\n", rule.Group))
			b.WriteString(fmt.Sprintf("    return %q\n", rule.Group))
		}
		if len(rule.Keywords) > 0 {
			b.WriteString("  keywords = [" + joinQuoted(rule.Keywords) + "]\n")
			b.WriteString("  for keyword in keywords:\n")
			b.WriteString("    if keyword in host:\n")
			b.WriteString(fmt.Sprintf("      ctx.log('[Script] matched %s DOMAIN-KEYWORD rule')\n", rule.Group))
			b.WriteString(fmt.Sprintf("      return %q\n", rule.Group))
		}
		if rule.HasIPCIDR {
			b.WriteString(fmt.Sprintf("  if ctx.rule_providers[%q].match(md):\n", rule.Name+"_ipcidr"))
			b.WriteString(fmt.Sprintf("    ctx.log('[Script] matched %s IP rule')\n", rule.Group))
			b.WriteString(fmt.Sprintf("    return %q\n", rule.Group))
		}
	}

	if len(geoips) > 0 {
		keys := make([]string, 0, len(geoips))
		for key := range geoips {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		b.WriteString("  geoips = { ")
		for i, key := range keys {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(fmt.Sprintf("%q: %q", key, geoips[key]))
		}
		b.WriteString(" }\n")
		b.WriteString("  ip = md[\"dst_ip\"]\n")
		b.WriteString("  if ip == \"\":\n")
		b.WriteString("    ip = ctx.resolve_ip(host)\n")
		b.WriteString("    if ip == \"\":\n")
		b.WriteString(fmt.Sprintf("      ctx.log('[Script] dns lookup error use %s')\n", matchGroup))
		b.WriteString(fmt.Sprintf("      return %q\n", matchGroup))
		b.WriteString("  for key in geoips:\n")
		b.WriteString("    if ctx.geoip(ip) == key:\n")
		b.WriteString("      return geoips[key]\n")
	}

	b.WriteString(fmt.Sprintf("  return %q", matchGroup))
	return b.String()
}

func joinQuoted(values []string) string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, strconv.Quote(value))
	}
	return strings.Join(out, ", ")
}

func uniqueRuleName(base string, used map[string]int) string {
	if base == "" {
		base = "ruleset"
	}
	if _, ok := used[base]; !ok {
		used[base] = 1
		return base
	}
	idx := used[base] + 1
	for {
		candidate := fmt.Sprintf("%s %d", base, idx)
		if _, ok := used[candidate]; !ok {
			used[base] = idx
			used[candidate] = 1
			return candidate
		}
		idx++
	}
}

func ruleNameFromPath(path string) string {
	if path == "" {
		return ""
	}
	pos := strings.LastIndexAny(path, "/\\")
	start := 0
	if pos != -1 {
		start = pos + 1
	}
	end := strings.LastIndex(path, ".")
	if end == -1 || end < start {
		end = len(path)
	}
	return utils.UrlDecode(path[start:end])
}

func hashRuleURL(value string) string {
	const prime uint64 = 0x100000001B3
	const basis uint64 = 0xCBF29CE484222325
	hash := basis
	for i := 0; i < len(value); i++ {
		hash ^= uint64(value[i])
		hash *= prime
	}
	return strconv.FormatUint(hash, 10)
}

func appendRule(list *[]string, max int, rule string) {
	if rule == "" {
		return
	}
	if max > 0 && len(*list) >= max {
		return
	}
	*list = append(*list, rule)
}

