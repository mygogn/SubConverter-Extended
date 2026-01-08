package config

import (
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/utils"
)

func ParseProxyGroupsFromINI(lines []string) []ProxyGroupConfig {
	var groups []ProxyGroupConfig
	for _, line := range lines {
		parts := utils.Split(line, "`")
		if len(parts) < 3 {
			continue
		}
		cfg := ProxyGroupConfig{
			Name:     parts[0],
			Strategy: BalanceConsistentHashing,
			Timeout:  0,
		}
		switch parts[1] {
		case string(GroupSelect):
			cfg.Type = GroupSelect
		case string(GroupRelay):
			cfg.Type = GroupRelay
		case string(GroupURLTest):
			cfg.Type = GroupURLTest
		case string(GroupFallback):
			cfg.Type = GroupFallback
		case string(GroupLoadBalance):
			cfg.Type = GroupLoadBalance
		case string(GroupSSID):
			cfg.Type = GroupSSID
		case string(GroupSmart):
			cfg.Type = GroupSmart
		default:
			continue
		}

		rulesUpperBound := len(parts)
		if cfg.Type == GroupURLTest || cfg.Type == GroupLoadBalance || cfg.Type == GroupFallback {
			if rulesUpperBound < 5 {
				continue
			}
			rulesUpperBound -= 2
			cfg.URL = parts[rulesUpperBound]
			parseGroupTimes(parts[rulesUpperBound+1], &cfg.Interval, &cfg.Timeout, &cfg.Tolerance)
		}

		for i := 2; i < rulesUpperBound; i++ {
			entry := parts[i]
			if strings.HasPrefix(entry, "!!PROVIDER=") {
				list := utils.Split(entry[len("!!PROVIDER="):], ",")
				cfg.UsingProvider = append(cfg.UsingProvider, list...)
			} else {
				cfg.Proxies = append(cfg.Proxies, entry)
			}
		}
		groups = append(groups, cfg)
	}
	return groups
}

func ParseRulesetsFromINI(lines []string) []RulesetConfig {
	var out []RulesetConfig
	for _, line := range lines {
		first := strings.Index(line, ",")
		if first == -1 {
			continue
		}
		cfg := RulesetConfig{
			Group:    line[:first],
			Interval: 86400,
		}
		rest := line[first+1:]
		if strings.HasPrefix(rest, "[]") {
			cfg.URL = rest
			out = append(out, cfg)
			continue
		}
		last := strings.LastIndex(rest, ",")
		if last != -1 {
			cfg.Interval = utils.ToInt(rest[last+1:], cfg.Interval)
			cfg.URL = rest[:last]
		} else {
			cfg.URL = rest
		}
		out = append(out, cfg)
	}
	return out
}

func ParseRegexMatchFromINI(lines []string, delimiter string) []RegexMatchConfig {
	var out []RegexMatchConfig
	for _, line := range lines {
		cfg := RegexMatchConfig{}
		if strings.HasPrefix(line, "script:") {
			cfg.Script = line[len("script:"):]
			out = append(out, cfg)
			continue
		}
		pos := strings.LastIndex(line, delimiter)
		if pos == -1 {
			cfg.Match = line
		} else {
			cfg.Match = line[:pos]
			if pos+1 < len(line) {
				cfg.Replace = line[pos+1:]
			}
		}
		out = append(out, cfg)
	}
	return out
}

func ParseCronTasksFromINI(lines []string) []CronTaskConfig {
	var out []CronTaskConfig
	for _, line := range lines {
		parts := utils.Split(line, "`")
		if len(parts) < 3 {
			continue
		}
		cfg := CronTaskConfig{
			Name:    parts[0],
			CronExp: parts[1],
			Path:    parts[2],
		}
		if len(parts) > 3 {
			cfg.Timeout = utils.ToInt(parts[3], 0)
		}
		out = append(out, cfg)
	}
	return out
}

func parseGroupTimes(src string, interval, timeout, tolerance *int) {
	parts := utils.Split(src, ",")
	if len(parts) > 0 && interval != nil {
		*interval = utils.ToInt(parts[0], 0)
	}
	if len(parts) > 1 && timeout != nil {
		*timeout = utils.ToInt(parts[1], 0)
	}
	if len(parts) > 2 && tolerance != nil {
		*tolerance = utils.ToInt(parts[2], 0)
	}
}
