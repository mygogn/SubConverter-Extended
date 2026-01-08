package config

import (
	"errors"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/aethersailor/subconverter-extended/internal/utils"
)

const defaultExternalConfig = "https://gcore.jsdelivr.net/gh/Aethersailor/Custom_OpenClash_Rules@refs/heads/main/cfg/Custom_Clash.ini"

func LoadSettingsFromPath(path string, fetch FetchFunc) (*Settings, error) {
	content := utils.FileGet(path, false)
	if content == "" {
		return nil, errors.New("empty config")
	}
	settings := DefaultSettings()
	settings.PrefPath = path
	if fetch == nil {
		fetch = func(url string) (string, error) {
			return utils.FetchURL(url, utils.FetchOptions{
				Proxy:            settings.ProxyConfig,
				CacheTTL:         settings.CacheConfig,
				MaxSize:          settings.MaxAllowedDownloadSize,
				ServeCacheOnFail: settings.ServeCacheOnFetchFail,
			})
		}
	}
	if strings.Contains(content, "common:") {
		if err := settings.readYAML(content, fetch); err == nil {
			return settings, nil
		}
	}
	if err := settings.readTOML(content, fetch); err == nil {
		return settings, nil
	}
	if err := settings.readINI(content, fetch); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Settings) readINI(content string, fetch FetchFunc) error {
	ini := NewIni()
	ini.AllowDuplicateSections = true
	ini.Parse(content)

	common := ini.Section("common")
	if common != nil {
		if v := common.Get("default_url"); v != "" {
			s.DefaultUrls = v
		}
		s.EnableInsert.SetFromString(common.Get("enable_insert"))
		if v := common.Get("insert_url"); v != "" {
			s.InsertUrls = v
		}
		if v, ok := toBool(common.Get("prepend_insert_url")); ok {
			s.PrependInsert = v
		}
		if common.ItemPrefixExists("exclude_remarks") {
			s.ExcludeRemarks = common.GetAll("exclude_remarks")
		}
		if common.ItemPrefixExists("include_remarks") {
			s.IncludeRemarks = common.GetAll("include_remarks")
		}
		if v, ok := toBool(common.Get("enable_filter")); ok && v {
			s.FilterScript = common.Get("filter_script")
		} else {
			s.FilterScript = ""
		}
		if v := common.Get("base_path"); v != "" {
			s.BasePath = v
		}
		if v := common.Get("clash_rule_base"); v != "" {
			s.ClashBase = v
		}
		if v := common.Get("surge_rule_base"); v != "" {
			s.SurgeBase = v
		}
		if v := common.Get("surfboard_rule_base"); v != "" {
			s.SurfboardBase = v
		}
		if v := common.Get("mellow_rule_base"); v != "" {
			s.MellowBase = v
		}
		if v := common.Get("quan_rule_base"); v != "" {
			s.QuanBase = v
		}
		if v := common.Get("quanx_rule_base"); v != "" {
			s.QuanXBase = v
		}
		if v := common.Get("loon_rule_base"); v != "" {
			s.LoonBase = v
		}
		if v := common.Get("sssub_rule_base"); v != "" {
			s.SSSubBase = v
		}
		if v := common.Get("singbox_rule_base"); v != "" {
			s.SingBoxBase = v
		}
		if v := common.Get("default_external_config"); v != "" {
			s.DefaultExtConfig = v
		}
		if s.DefaultExtConfig == "" {
			s.DefaultExtConfig = defaultExternalConfig
		}
		if v, ok := toBool(common.Get("append_proxy_type")); ok {
			s.AppendType = v
		}
		if v := common.Get("proxy_config"); v != "" {
			s.ProxyConfig = v
		}
		if v := common.Get("proxy_ruleset"); v != "" {
			s.ProxyRuleset = v
		}
		if v := common.Get("proxy_subscription"); v != "" {
			s.ProxySubscription = v
		}
		if v, ok := toBool(common.Get("reload_conf_on_request")); ok {
			s.ReloadConfOnRequest = v
		}
	}

	if sec := ini.Section("surge_external_proxy"); sec != nil {
		if v := sec.Get("surge_ssr_path"); v != "" {
			s.SurgeSSRPath = v
		}
		if v, ok := toBool(sec.Get("resolve_hostname")); ok {
			s.SurgeResolveHostname = v
		}
	}

	if sec := ini.Section("node_pref"); sec != nil {
		s.UDPFlag.SetFromString(sec.Get("udp_flag"))
		s.TFOFlag.SetFromString(sec.Get("tcp_fast_open_flag"))
		s.SkipCertVerify.SetFromString(sec.Get("skip_cert_verify_flag"))
		s.TLS13Flag.SetFromString(sec.Get("tls13_flag"))
		if v, ok := toBool(sec.Get("sort_flag")); ok {
			s.EnableSort = v
		}
		if v := sec.Get("sort_script"); v != "" {
			s.SortScript = v
		}
		if v, ok := toBool(sec.Get("filter_deprecated_nodes")); ok {
			s.FilterDeprecated = v
		}
		if v, ok := toBool(sec.Get("append_sub_userinfo")); ok {
			s.AppendUserInfo = v
		}
		if v, ok := toBool(sec.Get("clash_use_new_field_name")); ok {
			s.ClashUseNewField = v
		}
		if v := sec.Get("clash_proxies_style"); v != "" {
			s.ClashProxiesStyle = v
		}
		if v, ok := toBool(sec.Get("singbox_add_clash_modes")); ok {
			s.SingBoxAddClashModes = v
		}
		if sec.ItemPrefixExists("rename_node") {
			items := sec.GetAll("rename_node")
			items, _ = ImportItems(items, false, fetch)
			s.SetRenames(ParseRegexMatchFromINI(items, "@"))
		}
	}

	if sec := ini.Section("userinfo"); sec != nil {
		if sec.ItemPrefixExists("stream_rule") {
			items := sec.GetAll("stream_rule")
			items, _ = ImportItems(items, false, fetch)
			s.SetStreamRules(ParseRegexMatchFromINI(items, "|"))
		}
		if sec.ItemPrefixExists("time_rule") {
			items := sec.GetAll("time_rule")
			items, _ = ImportItems(items, false, fetch)
			s.SetTimeRules(ParseRegexMatchFromINI(items, "|"))
		}
	}

	if sec := ini.Section("managed_config"); sec != nil {
		if v, ok := toBool(sec.Get("write_managed_config")); ok {
			s.WriteManaged = v
		}
		if v := sec.Get("managed_config_prefix"); v != "" {
			s.ManagedPrefix = v
		}
		if v, ok := toInt(sec.Get("config_update_interval")); ok {
			s.UpdateInterval = v
		}
		if v, ok := toBool(sec.Get("config_update_strict")); ok {
			s.UpdateStrict = v
		}
		if v := sec.Get("quanx_device_id"); v != "" {
			s.QuanXDevID = v
		}
	}

	if sec := ini.Section("emojis"); sec != nil {
		if v, ok := toBool(sec.Get("add_emoji")); ok {
			s.AddEmoji = v
		}
		if v, ok := toBool(sec.Get("remove_old_emoji")); ok {
			s.RemoveEmoji = v
		}
		if sec.ItemPrefixExists("rule") {
			items := sec.GetAll("rule")
			items, _ = ImportItems(items, false, fetch)
			s.SetEmojis(ParseRegexMatchFromINI(items, ","))
		}
	}

	sec := ini.Section("rulesets")
	if sec == nil {
		sec = ini.Section("ruleset")
	}
	if sec != nil {
		if v, ok := toBool(sec.Get("enabled")); ok {
			s.EnableRuleGen = v
		}
		if s.EnableRuleGen {
			if v, ok := toBool(sec.Get("overwrite_original_rules")); ok {
				s.OverwriteOriginalRules = v
			}
			if v, ok := toBool(sec.Get("update_ruleset_on_request")); ok {
				s.UpdateRulesetOnRequest = v
			}
			key := "ruleset"
			if sec.ItemPrefixExists("surge_ruleset") {
				key = "surge_ruleset"
			}
			if sec.ItemPrefixExists(key) {
				items := sec.GetAll(key)
				items, _ = ImportItems(items, false, fetch)
				s.CustomRulesets = ParseRulesetsFromINI(items)
			}
		} else {
			s.OverwriteOriginalRules = false
			s.UpdateRulesetOnRequest = false
		}
	}

	sec = ini.Section("proxy_groups")
	if sec == nil {
		sec = ini.Section("clash_proxy_group")
		if sec == nil {
			sec = ini.Section("proxy_group")
		}
	}
	if sec != nil && sec.ItemPrefixExists("custom_proxy_group") {
		items := sec.GetAll("custom_proxy_group")
		items, _ = ImportItems(items, false, fetch)
		s.CustomProxyGroups = ParseProxyGroupsFromINI(items)
	}

	if sec := ini.Section("template"); sec != nil {
		if v := sec.Get("template_path"); v != "" {
			s.TemplatePath = v
		}
		s.TemplateVars = make(map[string]string)
		for _, item := range sec.Items {
			if item.Key == "template_path" || item.Key == "{NONAME}" {
				continue
			}
			s.TemplateVars[item.Key] = item.Value
		}
		if s.ManagedPrefix != "" {
			s.TemplateVars["managed_config_prefix"] = s.ManagedPrefix
		}
	}

	if sec := ini.Section("aliases"); sec != nil {
		s.Aliases = make(map[string]string)
		for _, item := range sec.Items {
			if item.Key == "{NONAME}" {
				continue
			}
			s.Aliases[item.Key] = item.Value
		}
	}

	if sec := ini.Section("tasks"); sec != nil && sec.ItemPrefixExists("task") {
		items := sec.GetAll("task")
		items, _ = ImportItems(items, false, fetch)
		s.CronTasks = ParseCronTasksFromINI(items)
		s.EnableCron = len(s.CronTasks) > 0
	}

	if sec := ini.Section("server"); sec != nil {
		if v := sec.Get("listen"); v != "" {
			s.ListenAddress = v
		}
		if v, ok := toInt(sec.Get("port")); ok {
			s.ListenPort = v
		}
		if v := sec.Get("serve_file_root"); v != "" {
			s.ServeFileRoot = v
		}
	}

	if sec := ini.Section("advanced"); sec != nil {
		if v := sec.Get("log_level"); v != "" {
			s.LogLevel = ParseLogLevel(v)
		}
		if v, ok := toBool(sec.Get("print_debug_info")); ok {
			s.PrintDebugInfo = v
			if s.PrintDebugInfo {
				s.LogLevel = LogVerbose
			}
		}
		if v, ok := toInt(sec.Get("max_pending_connections")); ok {
			s.MaxPendingConns = v
		}
		if v, ok := toInt(sec.Get("max_concurrent_threads")); ok {
			s.MaxConcurThreads = v
		}
		if v, ok := toInt(sec.Get("max_allowed_rulesets")); ok {
			s.MaxAllowedRulesets = v
		}
		if v, ok := toInt(sec.Get("max_allowed_rules")); ok {
			s.MaxAllowedRules = v
		}
		if v, ok := toInt(sec.Get("max_allowed_download_size")); ok {
			s.MaxAllowedDownloadSize = int64(v)
		}
		if sec.Get("enable_cache") != "" {
			if v, ok := toBool(sec.Get("enable_cache")); ok && v {
				if v, ok := toInt(sec.Get("cache_subscription")); ok {
					s.CacheSubscription = v
				}
				if v, ok := toInt(sec.Get("cache_config")); ok {
					s.CacheConfig = v
				}
				if v, ok := toInt(sec.Get("cache_ruleset")); ok {
					s.CacheRuleset = v
				}
				if v, ok := toBool(sec.Get("serve_cache_on_fetch_fail")); ok {
					s.ServeCacheOnFetchFail = v
				}
			} else {
				s.CacheSubscription = 0
				s.CacheConfig = 0
				s.CacheRuleset = 0
				s.ServeCacheOnFetchFail = false
			}
		}
		if v, ok := toBool(sec.Get("script_clean_context")); ok {
			s.ScriptCleanContext = v
		}
		if v, ok := toBool(sec.Get("async_fetch_ruleset")); ok {
			s.AsyncFetchRuleset = v
		}
		if v, ok := toBool(sec.Get("skip_failed_links")); ok {
			s.SkipFailedLinks = v
		}
	}

	return nil
}

func (s *Settings) readYAML(content string, fetch FetchFunc) error {
	var root map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return err
	}
	common, _ := root["common"].(map[string]interface{})
	if common != nil {
		if v := toStringSlice(common["default_url"]); len(v) > 0 {
			s.DefaultUrls = utils.Join(v, "|")
		}
		s.EnableInsert.SetFromString(toString(common["enable_insert"]))
		if v := toStringSlice(common["insert_url"]); len(v) > 0 {
			s.InsertUrls = utils.Join(v, "|")
		}
		if b, ok := toBool(common["prepend_insert_url"]); ok {
			s.PrependInsert = b
		}
		if v := toStringSlice(common["exclude_remarks"]); len(v) > 0 {
			s.ExcludeRemarks = v
		}
		if v := toStringSlice(common["include_remarks"]); len(v) > 0 {
			s.IncludeRemarks = v
		}
		if b, ok := toBool(common["enable_filter"]); ok && b {
			s.FilterScript = toString(common["filter_script"])
		}
		if v := toString(common["base_path"]); v != "" {
			s.BasePath = v
		}
		if v := toString(common["clash_rule_base"]); v != "" {
			s.ClashBase = v
		}
		if v := toString(common["surge_rule_base"]); v != "" {
			s.SurgeBase = v
		}
		if v := toString(common["surfboard_rule_base"]); v != "" {
			s.SurfboardBase = v
		}
		if v := toString(common["mellow_rule_base"]); v != "" {
			s.MellowBase = v
		}
		if v := toString(common["quan_rule_base"]); v != "" {
			s.QuanBase = v
		}
		if v := toString(common["quanx_rule_base"]); v != "" {
			s.QuanXBase = v
		}
		if v := toString(common["loon_rule_base"]); v != "" {
			s.LoonBase = v
		}
		if v := toString(common["sssub_rule_base"]); v != "" {
			s.SSSubBase = v
		}
		if v := toString(common["singbox_rule_base"]); v != "" {
			s.SingBoxBase = v
		}
		if v := toString(common["default_external_config"]); v != "" {
			s.DefaultExtConfig = v
		}
		if s.DefaultExtConfig == "" {
			s.DefaultExtConfig = defaultExternalConfig
		}
		if b, ok := toBool(common["append_proxy_type"]); ok {
			s.AppendType = b
		}
		if v := toString(common["proxy_config"]); v != "" {
			s.ProxyConfig = v
		}
		if v := toString(common["proxy_ruleset"]); v != "" {
			s.ProxyRuleset = v
		}
		if v := toString(common["proxy_subscription"]); v != "" {
			s.ProxySubscription = v
		}
		if b, ok := toBool(common["reload_conf_on_request"]); ok {
			s.ReloadConfOnRequest = b
		}
	}

	if nodePref, ok := root["node_pref"].(map[string]interface{}); ok {
		s.UDPFlag.SetFromString(toString(nodePref["udp_flag"]))
		s.TFOFlag.SetFromString(toString(nodePref["tcp_fast_open_flag"]))
		s.SkipCertVerify.SetFromString(toString(nodePref["skip_cert_verify_flag"]))
		s.TLS13Flag.SetFromString(toString(nodePref["tls13_flag"]))
		if b, ok := toBool(nodePref["sort_flag"]); ok {
			s.EnableSort = b
		}
		if v := toString(nodePref["sort_script"]); v != "" {
			s.SortScript = v
		}
		if b, ok := toBool(nodePref["filter_deprecated_nodes"]); ok {
			s.FilterDeprecated = b
		}
		if b, ok := toBool(nodePref["append_sub_userinfo"]); ok {
			s.AppendUserInfo = b
		}
		if b, ok := toBool(nodePref["clash_use_new_field_name"]); ok {
			s.ClashUseNewField = b
		}
		if v := toString(nodePref["clash_proxies_style"]); v != "" {
			s.ClashProxiesStyle = v
		}
		if b, ok := toBool(nodePref["singbox_add_clash_modes"]); ok {
			s.SingBoxAddClashModes = b
		}
		if renameNode, ok := nodePref["rename_node"].([]interface{}); ok {
			lines := readRegexMatchFromYAML(renameNode, "@")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetRenames(ParseRegexMatchFromINI(lines, "@"))
		}
	}

	if userinfo, ok := root["userinfo"].(map[string]interface{}); ok {
		if streamRule, ok := userinfo["stream_rule"].([]interface{}); ok {
			lines := readRegexMatchFromYAML(streamRule, "|")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetStreamRules(ParseRegexMatchFromINI(lines, "|"))
		}
		if timeRule, ok := userinfo["time_rule"].([]interface{}); ok {
			lines := readRegexMatchFromYAML(timeRule, "|")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetTimeRules(ParseRegexMatchFromINI(lines, "|"))
		}
	}

	if managed, ok := root["managed_config"].(map[string]interface{}); ok {
		if b, ok := toBool(managed["write_managed_config"]); ok {
			s.WriteManaged = b
		}
		if v := toString(managed["managed_config_prefix"]); v != "" {
			s.ManagedPrefix = v
		}
		if v, ok := toInt(managed["config_update_interval"]); ok {
			s.UpdateInterval = v
		}
		if b, ok := toBool(managed["config_update_strict"]); ok {
			s.UpdateStrict = b
		}
		if v := toString(managed["quanx_device_id"]); v != "" {
			s.QuanXDevID = v
		}
	}

	if surgeProxy, ok := root["surge_external_proxy"].(map[string]interface{}); ok {
		if v := toString(surgeProxy["surge_ssr_path"]); v != "" {
			s.SurgeSSRPath = v
		}
		if b, ok := toBool(surgeProxy["resolve_hostname"]); ok {
			s.SurgeResolveHostname = b
		}
	}

	if emojis, ok := root["emojis"].(map[string]interface{}); ok {
		if b, ok := toBool(emojis["add_emoji"]); ok {
			s.AddEmoji = b
		}
		if b, ok := toBool(emojis["remove_old_emoji"]); ok {
			s.RemoveEmoji = b
		}
		if rules, ok := emojis["rules"].([]interface{}); ok {
			lines := readEmojiFromYAML(rules)
			lines, _ = ImportItems(lines, false, fetch)
			s.SetEmojis(ParseRegexMatchFromINI(lines, ","))
		}
	}

	if rulesets, ok := root["rulesets"].(map[string]interface{}); ok {
		readRulesetsFromYAML(rulesets, s, fetch)
	} else if ruleset, ok := root["ruleset"].(map[string]interface{}); ok {
		readRulesetsFromYAML(ruleset, s, fetch)
	}

	if groups, ok := root["proxy_groups"].(map[string]interface{}); ok {
		if custom, ok := groups["custom_proxy_group"].([]interface{}); ok {
			lines := readGroupsFromYAML(custom)
			lines, _ = ImportItems(lines, false, fetch)
			s.CustomProxyGroups = ParseProxyGroupsFromINI(lines)
		}
	} else if groups, ok := root["proxy_group"].(map[string]interface{}); ok {
		if custom, ok := groups["custom_proxy_group"].([]interface{}); ok {
			lines := readGroupsFromYAML(custom)
			lines, _ = ImportItems(lines, false, fetch)
			s.CustomProxyGroups = ParseProxyGroupsFromINI(lines)
		}
	}

	if template, ok := root["template"].(map[string]interface{}); ok {
		if v := toString(template["template_path"]); v != "" {
			s.TemplatePath = v
		}
		if globals, ok := template["globals"].([]interface{}); ok {
			s.TemplateVars = make(map[string]string)
			for _, item := range globals {
				m, _ := item.(map[string]interface{})
				if m == nil {
					continue
				}
				key := toString(m["key"])
				value := toString(m["value"])
				if key != "" {
					s.TemplateVars[key] = value
				}
			}
		}
		if s.ManagedPrefix != "" {
			if s.TemplateVars == nil {
				s.TemplateVars = make(map[string]string)
			}
			s.TemplateVars["managed_config_prefix"] = s.ManagedPrefix
		}
	}

	if aliases, ok := root["aliases"].([]interface{}); ok {
		s.Aliases = make(map[string]string)
		for _, item := range aliases {
			m, _ := item.(map[string]interface{})
			if m == nil {
				continue
			}
			uri := toString(m["uri"])
			target := toString(m["target"])
			if uri != "" && target != "" {
				s.Aliases[uri] = target
			}
		}
	}

	if tasks, ok := root["tasks"].([]interface{}); ok {
		s.CronTasks = readTasksFromTOML(tasks)
		s.EnableCron = len(s.CronTasks) > 0
	}

	if server, ok := root["server"].(map[string]interface{}); ok {
		if v := toString(server["listen"]); v != "" {
			s.ListenAddress = v
		}
		if v, ok := toInt(server["port"]); ok {
			s.ListenPort = v
		}
		if v := toString(server["serve_file_root"]); v != "" {
			s.ServeFileRoot = v
		}
	}

	if advanced, ok := root["advanced"].(map[string]interface{}); ok {
		if v := toString(advanced["log_level"]); v != "" {
			s.LogLevel = ParseLogLevel(v)
		}
		if b, ok := toBool(advanced["print_debug_info"]); ok {
			s.PrintDebugInfo = b
			if s.PrintDebugInfo {
				s.LogLevel = LogVerbose
			}
		}
		if v, ok := toInt(advanced["max_pending_connections"]); ok {
			s.MaxPendingConns = v
		}
		if v, ok := toInt(advanced["max_concurrent_threads"]); ok {
			s.MaxConcurThreads = v
		}
		if v, ok := toInt(advanced["max_allowed_rulesets"]); ok {
			s.MaxAllowedRulesets = v
		}
		if v, ok := toInt(advanced["max_allowed_rules"]); ok {
			s.MaxAllowedRules = v
		}
		if v, ok := toInt(advanced["max_allowed_download_size"]); ok {
			s.MaxAllowedDownloadSize = int64(v)
		}
		if v, ok := toBool(advanced["enable_cache"]); ok && v {
			if v, ok := toInt(advanced["cache_subscription"]); ok {
				s.CacheSubscription = v
			}
			if v, ok := toInt(advanced["cache_config"]); ok {
				s.CacheConfig = v
			}
			if v, ok := toInt(advanced["cache_ruleset"]); ok {
				s.CacheRuleset = v
			}
			if v, ok := toBool(advanced["serve_cache_on_fetch_fail"]); ok {
				s.ServeCacheOnFetchFail = v
			}
		} else if advanced["enable_cache"] != nil {
			s.CacheSubscription = 0
			s.CacheConfig = 0
			s.CacheRuleset = 0
			s.ServeCacheOnFetchFail = false
		}
		if b, ok := toBool(advanced["script_clean_context"]); ok {
			s.ScriptCleanContext = b
		}
		if b, ok := toBool(advanced["async_fetch_ruleset"]); ok {
			s.AsyncFetchRuleset = b
		}
		if b, ok := toBool(advanced["skip_failed_links"]); ok {
			s.SkipFailedLinks = b
		}
	}

	return nil
}

func (s *Settings) readTOML(content string, fetch FetchFunc) error {
	var root map[string]interface{}
	if err := toml.Unmarshal([]byte(content), &root); err != nil {
		return err
	}
	if v, ok := toInt(root["version"]); !ok || v == 0 {
		return errors.New("missing version")
	}
	common, _ := root["common"].(map[string]interface{})
	if common != nil {
		if v := toStringSlice(common["default_url"]); len(v) > 0 {
			s.DefaultUrls = utils.Join(v, "|")
		}
		s.EnableInsert.SetFromString(toString(common["enable_insert"]))
		if v := toStringSlice(common["insert_url"]); len(v) > 0 {
			s.InsertUrls = utils.Join(v, "|")
		}
		if b, ok := toBool(common["prepend_insert_url"]); ok {
			s.PrependInsert = b
		}
		if v := toStringSlice(common["exclude_remarks"]); len(v) > 0 {
			s.ExcludeRemarks = v
		}
		if v := toStringSlice(common["include_remarks"]); len(v) > 0 {
			s.IncludeRemarks = v
		}
		if b, ok := toBool(common["enable_filter"]); ok && b {
			s.FilterScript = toString(common["filter_script"])
		}
		if v := toString(common["base_path"]); v != "" {
			s.BasePath = v
		}
		if v := toString(common["clash_rule_base"]); v != "" {
			s.ClashBase = v
		}
		if v := toString(common["surge_rule_base"]); v != "" {
			s.SurgeBase = v
		}
		if v := toString(common["surfboard_rule_base"]); v != "" {
			s.SurfboardBase = v
		}
		if v := toString(common["mellow_rule_base"]); v != "" {
			s.MellowBase = v
		}
		if v := toString(common["quan_rule_base"]); v != "" {
			s.QuanBase = v
		}
		if v := toString(common["quanx_rule_base"]); v != "" {
			s.QuanXBase = v
		}
		if v := toString(common["loon_rule_base"]); v != "" {
			s.LoonBase = v
		}
		if v := toString(common["sssub_rule_base"]); v != "" {
			s.SSSubBase = v
		}
		if v := toString(common["singbox_rule_base"]); v != "" {
			s.SingBoxBase = v
		}
		if v := toString(common["default_external_config"]); v != "" {
			s.DefaultExtConfig = v
		}
		if s.DefaultExtConfig == "" {
			s.DefaultExtConfig = defaultExternalConfig
		}
		if b, ok := toBool(common["append_proxy_type"]); ok {
			s.AppendType = b
		}
		if v := toString(common["proxy_config"]); v != "" {
			s.ProxyConfig = v
		}
		if v := toString(common["proxy_ruleset"]); v != "" {
			s.ProxyRuleset = v
		}
		if v := toString(common["proxy_subscription"]); v != "" {
			s.ProxySubscription = v
		}
		if b, ok := toBool(common["reload_conf_on_request"]); ok {
			s.ReloadConfOnRequest = b
		}
	}

	if nodePref, ok := root["node_pref"].(map[string]interface{}); ok {
		s.UDPFlag.SetFromString(toString(nodePref["udp_flag"]))
		s.TFOFlag.SetFromString(toString(nodePref["tcp_fast_open_flag"]))
		s.SkipCertVerify.SetFromString(toString(nodePref["skip_cert_verify_flag"]))
		s.TLS13Flag.SetFromString(toString(nodePref["tls13_flag"]))
		if b, ok := toBool(nodePref["sort_flag"]); ok {
			s.EnableSort = b
		}
		if v := toString(nodePref["sort_script"]); v != "" {
			s.SortScript = v
		}
		if b, ok := toBool(nodePref["filter_deprecated_nodes"]); ok {
			s.FilterDeprecated = b
		}
		if b, ok := toBool(nodePref["append_sub_userinfo"]); ok {
			s.AppendUserInfo = b
		}
		if b, ok := toBool(nodePref["clash_use_new_field_name"]); ok {
			s.ClashUseNewField = b
		}
		if v := toString(nodePref["clash_proxies_style"]); v != "" {
			s.ClashProxiesStyle = v
		}
		if b, ok := toBool(nodePref["singbox_add_clash_modes"]); ok {
			s.SingBoxAddClashModes = b
		}
		if renameNode, ok := nodePref["rename_node"].([]interface{}); ok {
			lines := readRegexMatchFromTOML(renameNode, "@")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetRenames(ParseRegexMatchFromINI(lines, "@"))
		}
	}

	if userinfo, ok := root["userinfo"].(map[string]interface{}); ok {
		if streamRule, ok := userinfo["stream_rule"].([]interface{}); ok {
			lines := readRegexMatchFromTOML(streamRule, "|")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetStreamRules(ParseRegexMatchFromINI(lines, "|"))
		}
		if timeRule, ok := userinfo["time_rule"].([]interface{}); ok {
			lines := readRegexMatchFromTOML(timeRule, "|")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetTimeRules(ParseRegexMatchFromINI(lines, "|"))
		}
	}

	if managed, ok := root["managed_config"].(map[string]interface{}); ok {
		if b, ok := toBool(managed["write_managed_config"]); ok {
			s.WriteManaged = b
		}
		if v := toString(managed["managed_config_prefix"]); v != "" {
			s.ManagedPrefix = v
		}
		if v, ok := toInt(managed["config_update_interval"]); ok {
			s.UpdateInterval = v
		}
		if b, ok := toBool(managed["config_update_strict"]); ok {
			s.UpdateStrict = b
		}
		if v := toString(managed["quanx_device_id"]); v != "" {
			s.QuanXDevID = v
		}
	}

	if surgeProxy, ok := root["surge_external_proxy"].(map[string]interface{}); ok {
		if v := toString(surgeProxy["surge_ssr_path"]); v != "" {
			s.SurgeSSRPath = v
		}
		if b, ok := toBool(surgeProxy["resolve_hostname"]); ok {
			s.SurgeResolveHostname = b
		}
	}

	if emojis, ok := root["emojis"].(map[string]interface{}); ok {
		if b, ok := toBool(emojis["add_emoji"]); ok {
			s.AddEmoji = b
		}
		if b, ok := toBool(emojis["remove_old_emoji"]); ok {
			s.RemoveEmoji = b
		}
		if rules, ok := emojis["emoji"].([]interface{}); ok {
			lines := readRegexMatchFromTOML(rules, ",")
			lines, _ = ImportItems(lines, false, fetch)
			s.SetEmojis(ParseRegexMatchFromINI(lines, ","))
		}
	}

	if rulesets, ok := root["rulesets"].([]interface{}); ok {
		lines := readRulesetsFromTOML(rulesets)
		lines, _ = ImportItems(lines, false, fetch)
		s.CustomRulesets = ParseRulesetsFromINI(lines)
	}
	if groups, ok := root["custom_groups"].([]interface{}); ok {
		lines := readGroupsFromTOML(groups)
		lines, _ = ImportItems(lines, false, fetch)
		s.CustomProxyGroups = ParseProxyGroupsFromINI(lines)
	}

	if ruleset, ok := root["ruleset"].(map[string]interface{}); ok {
		if b, ok := toBool(ruleset["enabled"]); ok {
			s.EnableRuleGen = b
		}
		if b, ok := toBool(ruleset["overwrite_original_rules"]); ok {
			s.OverwriteOriginalRules = b
		}
		if b, ok := toBool(ruleset["update_ruleset_on_request"]); ok {
			s.UpdateRulesetOnRequest = b
		}
	}

	if template, ok := root["template"].(map[string]interface{}); ok {
		if v := toString(template["template_path"]); v != "" {
			s.TemplatePath = v
		}
		if globals, ok := template["globals"].([]interface{}); ok {
			s.TemplateVars = make(map[string]string)
			for _, item := range globals {
				m, _ := item.(map[string]interface{})
				if m == nil {
					continue
				}
				key := toString(m["key"])
				value := toString(m["value"])
				if key != "" {
					s.TemplateVars[key] = value
				}
			}
		}
		if s.ManagedPrefix != "" {
			if s.TemplateVars == nil {
				s.TemplateVars = make(map[string]string)
			}
			s.TemplateVars["managed_config_prefix"] = s.ManagedPrefix
		}
	}

	if aliases, ok := root["aliases"].([]interface{}); ok {
		s.Aliases = make(map[string]string)
		for _, item := range aliases {
			m, _ := item.(map[string]interface{})
			if m == nil {
				continue
			}
			uri := toString(m["uri"])
			target := toString(m["target"])
			if uri != "" && target != "" {
				s.Aliases[uri] = target
			}
		}
	}

	if tasks, ok := root["tasks"].([]interface{}); ok {
		s.CronTasks = readTasksFromTOML(tasks)
		s.EnableCron = len(s.CronTasks) > 0
	}

	if server, ok := root["server"].(map[string]interface{}); ok {
		if v := toString(server["listen"]); v != "" {
			s.ListenAddress = v
		}
		if v, ok := toInt(server["port"]); ok {
			s.ListenPort = v
		}
		if v := toString(server["serve_file_root"]); v != "" {
			s.ServeFileRoot = v
		}
	}

	if advanced, ok := root["advanced"].(map[string]interface{}); ok {
		if v := toString(advanced["log_level"]); v != "" {
			s.LogLevel = ParseLogLevel(v)
		}
		if b, ok := toBool(advanced["print_debug_info"]); ok {
			s.PrintDebugInfo = b
			if s.PrintDebugInfo {
				s.LogLevel = LogVerbose
			}
		}
		if v, ok := toInt(advanced["max_pending_connections"]); ok {
			s.MaxPendingConns = v
		}
		if v, ok := toInt(advanced["max_concurrent_threads"]); ok {
			s.MaxConcurThreads = v
		}
		if v, ok := toInt(advanced["max_allowed_rulesets"]); ok {
			s.MaxAllowedRulesets = v
		}
		if v, ok := toInt(advanced["max_allowed_rules"]); ok {
			s.MaxAllowedRules = v
		}
		if v, ok := toInt(advanced["max_allowed_download_size"]); ok {
			s.MaxAllowedDownloadSize = int64(v)
		}
		if b, ok := toBool(advanced["enable_cache"]); ok && b {
			if v, ok := toInt(advanced["cache_subscription"]); ok {
				s.CacheSubscription = v
			}
			if v, ok := toInt(advanced["cache_config"]); ok {
				s.CacheConfig = v
			}
			if v, ok := toInt(advanced["cache_ruleset"]); ok {
				s.CacheRuleset = v
			}
			if v, ok := toBool(advanced["serve_cache_on_fetch_fail"]); ok {
				s.ServeCacheOnFetchFail = v
			}
		} else if advanced["enable_cache"] != nil {
			s.CacheSubscription = 0
			s.CacheConfig = 0
			s.CacheRuleset = 0
			s.ServeCacheOnFetchFail = false
		}
		if b, ok := toBool(advanced["script_clean_context"]); ok {
			s.ScriptCleanContext = b
		}
		if b, ok := toBool(advanced["async_fetch_ruleset"]); ok {
			s.AsyncFetchRuleset = b
		}
		if b, ok := toBool(advanced["skip_failed_links"]); ok {
			s.SkipFailedLinks = b
		}
	}

	return nil
}

func readRegexMatchFromYAML(items []interface{}, delimiter string) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if script := toString(m["script"]); script != "" {
			out = append(out, "!!script:"+script)
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		match := toString(m["match"])
		replace := toString(m["replace"])
		if replace == "" {
			replace = toString(m["emoji"])
		}
		if match != "" && replace != "" {
			out = append(out, match+delimiter+replace)
		}
	}
	return out
}

func readEmojiFromYAML(items []interface{}) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if script := toString(m["script"]); script != "" {
			out = append(out, "!!script:"+script)
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		match := toString(m["match"])
		emoji := toString(m["emoji"])
		if match != "" && emoji != "" {
			out = append(out, match+","+emoji)
		}
	}
	return out
}

func readGroupsFromYAML(items []interface{}) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		name := toString(m["name"])
		typ := toString(m["type"])
		rules := toStringSlice(m["rule"])
		url := toString(m["url"])
		interval := toString(m["interval"])
		timeout := toString(m["timeout"])
		tolerance := toString(m["tolerance"])

		if name == "" || typ == "" {
			continue
		}
		parts := []string{name, typ}
		parts = append(parts, rules...)
		if typ == "url-test" || typ == "load-balance" || typ == "fallback" || typ == "smart" {
			if url == "" {
				url = "http://www.gstatic.com/generate_204"
			}
			if interval == "" {
				interval = "300"
			}
			parts = append(parts, url, interval+","+timeout+","+tolerance)
		}
		out = append(out, utils.Join(parts, "`"))
	}
	return out
}

func readRulesetsFromYAML(section map[string]interface{}, s *Settings, fetch FetchFunc) {
	if b, ok := toBool(section["enabled"]); ok {
		s.EnableRuleGen = b
	}
	if !s.EnableRuleGen {
		s.OverwriteOriginalRules = false
		s.UpdateRulesetOnRequest = false
		return
	}
	if b, ok := toBool(section["overwrite_original_rules"]); ok {
		s.OverwriteOriginalRules = b
	}
	if b, ok := toBool(section["update_ruleset_on_request"]); ok {
		s.UpdateRulesetOnRequest = b
	}
	var list []interface{}
	if v, ok := section["rulesets"].([]interface{}); ok {
		list = v
	} else if v, ok := section["surge_ruleset"].([]interface{}); ok {
		list = v
	}
	if list == nil {
		return
	}
	lines := readRulesetListFromYAML(list)
	lines, _ = ImportItems(lines, false, fetch)
	s.CustomRulesets = ParseRulesetsFromINI(lines)
}

func readRulesetListFromYAML(items []interface{}) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		group := toString(m["group"])
		url := toString(m["ruleset"])
		rule := toString(m["rule"])
		interval := toString(m["interval"])
		if url != "" {
			line := group + "," + url
			if interval != "" {
				line += "," + interval
			}
			out = append(out, line)
		} else if rule != "" {
			out = append(out, group+",[]"+rule)
		}
	}
	return out
}

func readRegexMatchFromTOML(items []interface{}, delimiter string) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		if script := toString(m["script"]); script != "" {
			out = append(out, "!!script:"+script)
			continue
		}
		match := toString(m["match"])
		replace := toString(m["replace"])
		if replace == "" {
			replace = toString(m["emoji"])
		}
		if match != "" && replace != "" {
			out = append(out, match+delimiter+replace)
		}
	}
	return out
}

func readGroupsFromTOML(items []interface{}) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		name := toString(m["name"])
		typ := toString(m["type"])
		if name == "" || typ == "" {
			continue
		}
		url := toString(m["url"])
		interval := toString(m["interval"])
		timeout := toString(m["timeout"])
		tolerance := toString(m["tolerance"])
		rules := toStringSlice(m["rule"])
		use := toStringSlice(m["use"])
		for _, u := range use {
			rules = append(rules, "!!PROVIDER="+u)
		}
		parts := []string{name, typ}
		parts = append(parts, rules...)
		if typ == "url-test" || typ == "load-balance" || typ == "fallback" || typ == "smart" {
			parts = append(parts, url, interval+","+timeout+","+tolerance)
		}
		out = append(out, utils.Join(parts, "`"))
	}
	return out
}

func readRulesetsFromTOML(items []interface{}) []string {
	var out []string
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			out = append(out, "!!import:"+imp)
			continue
		}
		group := toString(m["group"])
		ruleset := toString(m["ruleset"])
		typ := toString(m["type"])
		if typ != "" {
			switch typ {
			case "surge-ruleset":
				typ = "surge:"
			case "quantumultx":
				typ = "quanx:"
			case "clash-domain", "clash-ipcidr", "clash-classic":
				typ = typ + ":"
			default:
				typ = ""
			}
		}
		line := group + "," + typ + ruleset
		if interval := toString(m["interval"]); interval != "" {
			line += "," + interval
		}
		out = append(out, line)
	}
	return out
}

func readTasksFromTOML(items []interface{}) []CronTaskConfig {
	var out []CronTaskConfig
	for _, item := range items {
		m, _ := item.(map[string]interface{})
		if m == nil {
			continue
		}
		if imp := toString(m["import"]); imp != "" {
			continue
		}
		cfg := CronTaskConfig{
			Name:    toString(m["name"]),
			CronExp: toString(m["cronexp"]),
			Path:    toString(m["path"]),
		}
		if v, ok := toInt(m["timeout"]); ok {
			cfg.Timeout = v
		}
		if cfg.Name != "" && cfg.CronExp != "" && cfg.Path != "" {
			out = append(out, cfg)
		}
	}
	return out
}
