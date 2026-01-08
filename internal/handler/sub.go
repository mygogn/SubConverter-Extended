package handler

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/ruleset"
	"github.com/aethersailor/subconverter-extended/internal/template"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

var (
	regexBlacklist = map[string]struct{}{
		"(.*)*": {},
	}
	fallbackConfigURLs = []string{
		"https://gcore.jsdelivr.net/gh/Aethersailor/Custom_OpenClash_Rules@refs/heads/main/cfg/Custom_Clash.ini",
		"https://testingcf.jsdelivr.net/gh/Aethersailor/Custom_OpenClash_Rules@refs/heads/main/cfg/Custom_Clash.ini",
		"https://cdn.jsdelivr.net/gh/Aethersailor/Custom_OpenClash_Rules@refs/heads/main/cfg/Custom_Clash.ini",
		"https://raw.githubusercontent.com/Aethersailor/Custom_OpenClash_Rules/main/cfg/Custom_Clash.ini",
	}
)

type proxyProvider struct {
	Name          string
	URL           string
	Interval      int
	Filter        string
	ExcludeFilter string
	Path          string
	GroupID       int
}

type urlItem struct {
	URL     string
	GroupID int
	Group   string
}

func (h *Handler) Sub(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	args := queryArgs(r)
	target := args["target"]
	if target == "auto" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid target!"))
		return
	}
	if target == "clashr" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid target!"))
		return
	}
	if target != "clash" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid target!"))
		return
	}

	argURL := args["url"]
	if argURL == "" {
		if !shouldEnableInsert(args["insert"], h.Settings.EnableInsert) || h.Settings.InsertUrls == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Invalid request!"))
			return
		}
	}

	argInclude := args["include"]
	argExclude := args["exclude"]
	if _, banned := regexBlacklist[argInclude]; banned {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}
	if _, banned := regexBlacklist[argExclude]; banned {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	reqArgs := make(map[string]string, len(args))
	for k, v := range args {
		if k == "token" {
			continue
		}
		reqArgs[k] = v
	}
	reqArgs["target"] = target
	if _, ok := reqArgs["ver"]; !ok {
		reqArgs["ver"] = "3"
	}

	tplArgs := template.Args{
		GlobalVars:    h.Settings.TemplateVars,
		RequestParams: reqArgs,
		LocalVars: map[string]string{
			"clash.new_field_name": "true",
		},
	}

	renderer := &template.Renderer{
		IncludeScope:  h.Settings.TemplatePath,
		ManagedPrefix: h.Settings.ManagedPrefix,
		Fetch:         h.Fetcher.FetchConfig,
	}

	customGroups := h.Settings.CustomProxyGroups
	customRulesets := h.Settings.CustomRulesets
	enableRuleGen := h.Settings.EnableRuleGen
	clashBase := h.Settings.ClashBase
	includeRemarks := append([]string(nil), h.Settings.IncludeRemarks...)
	excludeRemarks := append([]string(nil), h.Settings.ExcludeRemarks...)
	renameRules := h.Settings.GetRenames()
	emojiRules := h.Settings.GetEmojis()
	addEmoji := h.Settings.AddEmoji
	removeEmoji := h.Settings.RemoveEmoji
	appendType := h.Settings.AppendType
	filterDeprecated := h.Settings.FilterDeprecated
	sortNodes := h.Settings.EnableSort

	argExternalConfig := args["config"]
	userProvidedConfig := argExternalConfig
	if argExternalConfig == "" {
		argExternalConfig = h.Settings.DefaultExtConfig
	}

	var extConf *config.ExternalConfig
	if argExternalConfig != "" {
		var err error
		extConf, err = config.LoadExternalConfig(argExternalConfig, tplArgs, renderer, h.Settings, h.Fetcher.FetchConfig)
		if err != nil && userProvidedConfig != "" && h.Settings.DefaultExtConfig != "" && argExternalConfig != h.Settings.DefaultExtConfig {
			for _, fallback := range fallbackConfigURLs {
				extConf, err = config.LoadExternalConfig(fallback, tplArgs, renderer, h.Settings, h.Fetcher.FetchConfig)
				if err == nil {
					break
				}
			}
		}
		if extConf != nil {
			if extConf.ClashRuleBase != "" {
				clashBase = pickExternalBase(extConf.ClashRuleBase, clashBase, h.Settings.BasePath)
			}
			if len(extConf.CustomProxyGroups) > 0 {
				customGroups = extConf.CustomProxyGroups
			}
			if len(extConf.SurgeRulesets) > 0 {
				customRulesets = extConf.SurgeRulesets
			}
			enableRuleGen = extConf.EnableRuleGenerator
			if len(extConf.Rename) > 0 {
				renameRules = extConf.Rename
			}
			if len(extConf.Emoji) > 0 {
				emojiRules = extConf.Emoji
			}
			if len(extConf.Include) > 0 {
				includeRemarks = extConf.Include
			}
			if len(extConf.Exclude) > 0 {
				excludeRemarks = extConf.Exclude
			}
			addEmoji = extConf.AddEmoji.Get(addEmoji)
			removeEmoji = extConf.RemoveOldEmoji.Get(removeEmoji)
			for key, value := range extConf.TemplateArgs {
				tplArgs.LocalVars[key] = value
			}
		}
	}

	baseContent, err := fetchBaseTemplate(clashBase, h.Fetcher.FetchConfig)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	rendered, err := renderer.Render(baseContent, tplArgs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(rendered))
		return
	}

	argEmoji := utils.TriBool{}
	argEmoji.SetFromString(args["emoji"])
	argAddEmoji := utils.TriBool{}
	argAddEmoji.SetFromString(args["add_emoji"])
	argRemoveEmoji := utils.TriBool{}
	argRemoveEmoji.SetFromString(args["remove_emoji"])
	if !argEmoji.IsUndef() {
		argAddEmoji.Set(argEmoji.Get(false))
		argRemoveEmoji.Set(true)
	}
	addEmoji = argAddEmoji.Get(addEmoji)
	removeEmoji = argRemoveEmoji.Get(removeEmoji)

	argAppendType := utils.TriBool{}
	argAppendType.SetFromString(args["append_type"])
	appendType = argAppendType.Get(appendType)

	argFilterDeprecated := utils.TriBool{}
	argFilterDeprecated.SetFromString(args["fdn"])
	filterDeprecated = argFilterDeprecated.Get(filterDeprecated)

	argSort := utils.TriBool{}
	argSort.SetFromString(args["sort"])
	sortNodes = argSort.Get(sortNodes)

	if argRename := args["rename"]; argRename != "" {
		renameRules = config.ParseRegexMatchFromINI(utils.Split(argRename, "`"), "@")
	}
	if argInclude != "" && utils.RegValid(argInclude) {
		includeRemarks = []string{argInclude}
	}
	if argExclude != "" && utils.RegValid(argExclude) {
		excludeRemarks = []string{argExclude}
	}

	urlItems := collectURLItems(argURL, h.Settings.InsertUrls, args["prepend"], h.Settings.PrependInsert, args["insert"], h.Settings.EnableInsert, h.Fetcher.FetchConfig)
	var providerItems []urlItem
	var nodeItems []urlItem
	for _, item := range urlItems {
		item.URL = strings.TrimSpace(item.URL)
		if item.URL == "" {
			continue
		}
		item.Group, item.URL = splitTaggedLink(item.URL)
		if isNodeLink(item.URL) {
			nodeItems = append(nodeItems, item)
			continue
		}
		if utils.IsLink(item.URL) {
			providerItems = append(providerItems, item)
			continue
		}
		nodeItems = append(nodeItems, item)
	}

	providers := buildProviders(providerItems, argInclude, argExclude)
	useProvider := len(providers) > 0

	var nodes []proxyNode
	for _, item := range nodeItems {
		parsed, err := parseNodeLink(item.URL, item.GroupID, item.Group)
		if err != nil {
			if h.Settings.SkipFailedLinks {
				continue
			}
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Invalid request!"))
			return
		}
		nodes = append(nodes, parsed...)
	}

	if len(nodes) == 0 && len(providers) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	nodes = filterNodes(nodes, includeRemarks, excludeRemarks)
	nodes = filterDeprecatedNodes(nodes, filterDeprecated)
	applyRenames(nodes, renameRules)
	applyEmojis(nodes, emojiRules, addEmoji, removeEmoji)
	sortNodesByName(nodes, sortNodes)
	appendTypePrefix(nodes, appendType)
	ensureUniqueNames(nodes)

	udpFlag := parseTriBool(args["udp"], h.Settings.UDPFlag)
	scvFlag := parseTriBool(args["scv"], h.Settings.SkipCertVerify)
	tfoFlag := parseTriBool(args["tfo"], h.Settings.TFOFlag)
	applyOverrides(nodes, udpFlag, scvFlag, tfoFlag, utils.TriBool{})

	providerNames := providerNameList(providers)
	groupDefs := buildProxyGroups(customGroups, nodes, providerNames, useProvider)
	groupBlock, err := marshalTopLevel("proxy-groups", groupDefs)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Invalid request!"))
		return
	}

	providerBlock := ""
	if useProvider {
		block, err := marshalTopLevel("proxy-providers", buildProviderMap(providers, udpFlag, scvFlag))
		if err == nil {
			providerBlock = block
		}
	}

	proxiesBlock := ""
	if len(nodes) > 0 {
		block, err := marshalTopLevel("proxies", buildProxyList(nodes))
		if err == nil {
			proxiesBlock = block
		}
	}

	rulesBlock := ""
	if enableRuleGen && len(customRulesets) > 0 {
		cacheable := extConf == nil || len(extConf.SurgeRulesets) == 0
		rulesets, err := h.resolveRulesets(customRulesets, cacheable)
		if err == nil {
			rulesList := ruleset.BuildClashRules(rulesets, h.Settings.MaxAllowedRules)
			rulesBlock, _ = marshalTopLevel("rules", rulesList)
		}
	}

	output := rendered
	output = replaceSection(output, "proxy-groups", groupBlock)
	if rulesBlock != "" {
		output = replaceSection(output, "rules", rulesBlock)
	}
	if providerBlock != "" {
		output = insertBeforeKey(output, "proxy-groups:", providerBlock)
	}
	if proxiesBlock != "" {
		output = insertBeforeKey(output, "proxy-groups:", proxiesBlock)
	}

	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(output))
}

func (h *Handler) resolveRulesets(list []config.RulesetConfig, cacheable bool) ([]config.RulesetContent, error) {
	if h.Settings.UpdateRulesetOnRequest {
		return ruleset.RefreshRulesets(list, h.Fetcher.FetchRuleset)
	}
	if cacheable {
		if cached, ok := h.CachedRulesets(); ok {
			return cached, nil
		}
	}
	contents, err := ruleset.RefreshRulesets(list, h.Fetcher.FetchRuleset)
	if err != nil {
		return nil, err
	}
	if cacheable {
		h.SetCachedRulesets(contents)
	}
	return contents, nil
}

func queryArgs(r *http.Request) map[string]string {
	values := r.URL.Query()
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

func collectURLItems(urlParam, insertURLs, prependParam string, defaultPrepend bool, insertParam string, enableInsert utils.TriBool, fetch func(string) (string, error)) []urlItem {
	urls := expandItems(utils.Split(urlParam, "|"), fetch)
	insertList := []string{}
	if shouldEnableInsert(insertParam, enableInsert) && insertURLs != "" {
		insertList = expandItems(utils.Split(insertURLs, "|"), fetch)
	}

	urlItems := make([]urlItem, 0, len(urls))
	groupID := 0
	for _, raw := range urls {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		urlItems = append(urlItems, urlItem{URL: item, GroupID: groupID})
		groupID++
	}

	insertItems := make([]urlItem, 0, len(insertList))
	insertID := -1
	for _, raw := range insertList {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		insertItems = append(insertItems, urlItem{URL: item, GroupID: insertID})
		insertID--
	}

	prepend := defaultPrepend
	if strings.TrimSpace(prependParam) != "" {
		tb := utils.TriBool{}
		tb.SetFromString(prependParam)
		prepend = tb.Get(defaultPrepend)
	}
	if prepend {
		return append(insertItems, urlItems...)
	}
	return append(urlItems, insertItems...)
}

func expandItems(items []string, fetch func(string) (string, error)) []string {
	if len(items) == 0 {
		return nil
	}
	expanded, _ := config.ImportItems(items, true, fetch)
	return expanded
}

func shouldEnableInsert(param string, defaultValue utils.TriBool) bool {
	tb := utils.TriBool{}
	tb.SetFromString(param)
	tb.Define(defaultValue)
	return tb.Get(false)
}

func buildProviders(items []urlItem, include, exclude string) []proxyProvider {
	var providers []proxyProvider
	for _, item := range items {
		link := strings.TrimSpace(item.URL)
		if link == "" {
			continue
		}
		hash := utils.GetMD5(utils.UrlDecode(link))
		if len(hash) > 10 {
			hash = hash[:10]
		}
		name := "provider_" + hash
		provider := proxyProvider{
			Name:     name,
			URL:      link,
			Interval: 3600,
			GroupID:  item.GroupID,
			Path:     "./providers/" + name + ".yaml",
		}
		if include != "" && utils.RegValid(include) {
			provider.Filter = include
		}
		if exclude != "" && utils.RegValid(exclude) {
			provider.ExcludeFilter = exclude
		}
		providers = append(providers, provider)
	}
	return providers
}

func providerNameList(providers []proxyProvider) []string {
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		if p.GroupID < 0 {
			continue
		}
		names = append(names, p.Name)
	}
	return names
}

func buildProviderMap(providers []proxyProvider, udp utils.TriBool, scv utils.TriBool) map[string]interface{} {
	result := make(map[string]interface{}, len(providers))
	for _, provider := range providers {
		entry := map[string]interface{}{
			"type":     "http",
			"url":      provider.URL,
			"interval": provider.Interval,
			"proxy":    "DIRECT",
			"path":     provider.Path,
			"health-check": map[string]interface{}{
				"enable":   true,
				"url":      "https://cp.cloudflare.com/generate_204",
				"interval": 300,
			},
		}
		if provider.Filter != "" {
			entry["filter"] = provider.Filter
		}
		if provider.ExcludeFilter != "" {
			entry["exclude-filter"] = provider.ExcludeFilter
		}
		override := map[string]interface{}{}
		if val, ok := scv.Value(); ok {
			override["skip-cert-verify"] = val
		}
		if val, ok := udp.Value(); ok {
			override["udp"] = val
		}
		if len(override) > 0 {
			entry["override"] = override
		}
		result[provider.Name] = entry
	}
	return result
}

func buildProxyGroups(groups []config.ProxyGroupConfig, nodes []proxyNode, providerNames []string, useProvider bool) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(groups))
	for _, group := range groups {
		entry := map[string]interface{}{
			"name": group.Name,
			"type": clashGroupType(group.Type),
		}
		switch group.Type {
		case config.GroupLoadBalance:
			if group.Strategy != "" {
				entry["strategy"] = string(group.Strategy)
			}
			fallthrough
		case config.GroupURLTest, config.GroupFallback, config.GroupSmart:
			if group.URL != "" {
				entry["url"] = group.URL
			}
			if group.Interval > 0 {
				entry["interval"] = group.Interval
			}
			if group.Tolerance > 0 {
				entry["tolerance"] = group.Tolerance
			}
			if val, ok := group.Lazy.Value(); ok {
				entry["lazy"] = val
			}
			if val, ok := group.EvaluateBefore.Value(); ok {
				entry["evaluate-before-use"] = val
			}
			if group.Timeout > 0 {
				entry["timeout"] = group.Timeout
			}
		}
		if val, ok := group.DisableUDP.Value(); ok {
			entry["disable-udp"] = val
		}
		if val, ok := group.Persistent.Value(); ok {
			entry["persistent"] = val
		}

		proxies, regex := collectGroupProxies(group.Proxies, nodes)
		if len(group.UsingProvider) > 0 {
			entry["use"] = group.UsingProvider
		} else if useProvider && regex != "" {
			if len(providerNames) > 0 {
				entry["use"] = providerNames
			}
			entry["filter"] = regex
		}
		if len(proxies) == 0 && !useProvider && len(group.UsingProvider) == 0 {
			proxies = append(proxies, "DIRECT")
		}
		if len(proxies) > 0 {
			entry["proxies"] = proxies
		}
		result = append(result, entry)
	}
	return result
}

func collectGroupProxies(rules []string, nodes []proxyNode) ([]string, string) {
	proxies := make([]string, 0)
	regex := ""
	seen := make(map[string]struct{})
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}
		if strings.HasPrefix(rule, "[]") {
			name := strings.TrimSpace(rule[2:])
			if name != "" {
				proxies = append(proxies, name)
			}
			continue
		}
		if rule == "DIRECT" || rule == "REJECT" {
			proxies = append(proxies, rule)
			continue
		}
		if strings.HasPrefix(rule, "script:") {
			continue
		}
		for _, node := range nodes {
			ok, real := applyMatcher(rule, node)
			if !ok {
				continue
			}
			if real != "" && !utils.RegFind(node.Name, real) {
				continue
			}
			if _, exists := seen[node.Name]; exists {
				continue
			}
			seen[node.Name] = struct{}{}
			proxies = append(proxies, node.Name)
		}
		if regex == "" {
			regex = rule
		}
	}
	return proxies, regex
}

func clashGroupType(t config.ProxyGroupType) string {
	if t == config.GroupSmart {
		return "url-test"
	}
	return string(t)
}

func parseTriBool(value string, fallback utils.TriBool) utils.TriBool {
	tb := utils.TriBool{}
	tb.SetFromString(value)
	tb.Define(fallback)
	return tb
}

func marshalTopLevel(key string, value interface{}) (string, error) {
	payload := map[string]interface{}{
		key: value,
	}
	data, err := yaml.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func replaceSection(content, key, replacement string) string {
	if replacement == "" {
		return content
	}
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s*~\s*$`)
	if !re.MatchString(content) {
		return content
	}
	replacement = strings.TrimRight(replacement, "\n")
	return re.ReplaceAllString(content, replacement)
}

func insertBeforeKey(content, key, insert string) string {
	if insert == "" {
		return content
	}
	insert = strings.TrimRight(insert, "\n") + "\n"
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key) {
			insertLines := strings.Split(strings.TrimRight(insert, "\n"), "\n")
			lines = append(lines[:i], append(insertLines, lines[i:]...)...)
			return strings.Join(lines, "\n")
		}
	}
	return content + "\n" + insert
}

func fetchBaseTemplate(path string, fetch func(string) (string, error)) (string, error) {
	if path == "" {
		return "", errors.New("empty base")
	}
	if utils.FileExists(path, true) {
		return utils.FileGet(path, true), nil
	}
	if utils.IsLink(path) {
		return fetch(path)
	}
	return "", errors.New("invalid base")
}

func pickExternalBase(value, fallback, basePath string) string {
	if value == "" {
		return fallback
	}
	if utils.IsLink(value) {
		return value
	}
	if basePath != "" && strings.HasPrefix(value, basePath) && utils.FileExists(value, true) {
		return value
	}
	return fallback
}

func splitTaggedLink(value string) (string, string) {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(trimmed), "tag:") {
		if idx := strings.Index(trimmed, ","); idx > 4 {
			return strings.TrimSpace(trimmed[4:idx]), strings.TrimSpace(trimmed[idx+1:])
		}
	}
	return "", value
}

func isNodeLink(link string) bool {
	lower := strings.ToLower(link)
	return strings.HasPrefix(lower, "vless://") ||
		strings.HasPrefix(lower, "vmess://") ||
		strings.HasPrefix(lower, "ss://") ||
		strings.HasPrefix(lower, "ssr://") ||
		strings.HasPrefix(lower, "trojan://") ||
		strings.HasPrefix(lower, "hysteria://") ||
		strings.HasPrefix(lower, "hysteria2://") ||
		strings.HasPrefix(lower, "hy2://") ||
		strings.HasPrefix(lower, "tuic://") ||
		strings.HasPrefix(lower, "snell://") ||
		strings.HasPrefix(lower, "socks5://") ||
		strings.HasPrefix(lower, "socks://")
}
