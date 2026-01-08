package ruleset

import (
	"errors"
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

var rulesetPrefixes = map[string]config.RulesetType{
	"clash-domain:":  config.RulesetClashDomain,
	"clash-ipcidr:":  config.RulesetClashIPCIDR,
	"clash-classic:": config.RulesetClashClassical,
	"quanx:":         config.RulesetQuanX,
	"surge:":         config.RulesetSurge,
}

type FetchFunc func(url string) (string, error)

func RefreshRulesets(list []config.RulesetConfig, fetch FetchFunc) ([]config.RulesetContent, error) {
	var out []config.RulesetContent
	for _, item := range list {
		group := item.Group
		url := strings.TrimSpace(item.URL)
		if group == "" || url == "" {
			continue
		}
		if idx := strings.Index(url, "[]"); idx != -1 {
			out = append(out, config.RulesetContent{
				Group:          group,
				Type:           config.RulesetSurge,
				Content:        url[idx:],
				UpdateInterval: 0,
			})
			continue
		}

		typ := config.RulesetSurge
		typed := url
		for prefix, t := range rulesetPrefixes {
			if strings.HasPrefix(url, prefix) {
				typ = t
				url = strings.TrimPrefix(url, prefix)
				break
			}
		}

		content, _ := fetchRuleset(url, fetch)
		out = append(out, config.RulesetContent{
			Group:          group,
			Path:           url,
			PathTyped:      typed,
			Type:           typ,
			Content:        content,
			UpdateInterval: item.Interval,
		})
	}
	return out, nil
}

func fetchRuleset(path string, fetch FetchFunc) (string, error) {
	if path == "" {
		return "", errors.New("empty ruleset")
	}
	if utils.FileExists(path, true) {
		return utils.FileGet(path, true), nil
	}
	if utils.IsLink(path) {
		if fetch == nil {
			return "", errors.New("missing fetcher")
		}
		return fetch(path)
	}
	return "", errors.New("invalid ruleset path")
}
