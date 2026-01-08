package config

import (
	"errors"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/aethersailor/subconverter-extended/internal/template"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type ExternalConfig struct {
	CustomProxyGroups []ProxyGroupConfig
	SurgeRulesets     []RulesetConfig
	ClashRuleBase     string
	SurgeRuleBase     string
	SurfboardRuleBase string
	MellowRuleBase    string
	QuanRuleBase      string
	QuanXRuleBase     string
	LoonRuleBase      string
	SSSubRuleBase     string
	SingBoxRuleBase   string
	Rename            []RegexMatchConfig
	Emoji             []RegexMatchConfig
	Include           []string
	Exclude           []string
	TemplateArgs      map[string]string
	OverwriteOriginalRules bool
	EnableRuleGenerator   bool
	AddEmoji         utils.TriBool
	RemoveOldEmoji   utils.TriBool
}

func LoadExternalConfig(path string, tplArgs template.Args, renderer *template.Renderer, settings *Settings, fetch FetchFunc) (*ExternalConfig, error) {
	if path == "" {
		return nil, errors.New("empty external config")
	}
	raw, err := fetch(path)
	if err != nil {
		return nil, err
	}
	content := raw
	if renderer != nil {
		renderer.IncludeScope = settings.TemplatePath
		renderer.ManagedPrefix = settings.ManagedPrefix
		rendered, renderErr := renderer.Render(raw, tplArgs)
		if renderErr == nil {
			content = rendered
		}
	}

	ext := &ExternalConfig{
		EnableRuleGenerator:   true,
		OverwriteOriginalRules: false,
		TemplateArgs:          make(map[string]string),
	}

	if err := loadExternalYAML(content, ext, settings, fetch); err == nil {
		return ext, nil
	}
	if err := loadExternalTOML(content, ext, settings, fetch); err == nil {
		return ext, nil
	}
	if err := loadExternalINI(content, ext, settings, fetch); err != nil {
		return nil, err
	}
	return ext, nil
}

func loadExternalYAML(content string, ext *ExternalConfig, settings *Settings, fetch FetchFunc) error {
	var root map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &root); err != nil {
		return err
	}
	custom, ok := root["custom"].(map[string]interface{})
	if !ok {
		return errors.New("no custom section")
	}

	ext.ClashRuleBase = toString(custom["clash_rule_base"])
	ext.SurgeRuleBase = toString(custom["surge_rule_base"])
	ext.SurfboardRuleBase = toString(custom["surfboard_rule_base"])
	ext.MellowRuleBase = toString(custom["mellow_rule_base"])
	ext.QuanRuleBase = toString(custom["quan_rule_base"])
	ext.QuanXRuleBase = toString(custom["quanx_rule_base"])
	ext.LoonRuleBase = toString(custom["loon_rule_base"])
	ext.SSSubRuleBase = toString(custom["sssub_rule_base"])
	ext.SingBoxRuleBase = toString(custom["singbox_rule_base"])

	if b, ok := toBool(custom["enable_rule_generator"]); ok {
		ext.EnableRuleGenerator = b
	}
	if b, ok := toBool(custom["overwrite_original_rules"]); ok {
		ext.OverwriteOriginalRules = b
	}

	groupKey := "custom_proxy_group"
	if _, ok := custom["proxy_groups"]; ok {
		groupKey = "proxy_groups"
	}
	if groups, ok := custom[groupKey].([]interface{}); ok {
		lines := readGroupsFromYAML(groups)
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.CustomProxyGroups = ParseProxyGroupsFromINI(lines)
	}

	rulesetKey := "surge_ruleset"
	if _, ok := custom["rulesets"]; ok {
		rulesetKey = "rulesets"
	}
	if rulesets, ok := custom[rulesetKey].([]interface{}); ok {
		lines := readRulesetListFromYAML(rulesets)
		if settings.MaxAllowedRulesets > 0 && len(lines) > settings.MaxAllowedRulesets {
			return errors.New("ruleset count exceeds limit")
		}
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.SurgeRulesets = ParseRulesetsFromINI(lines)
	}

	if rename, ok := custom["rename_node"].([]interface{}); ok {
		lines := readRegexMatchFromYAML(rename, "@")
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.Rename = ParseRegexMatchFromINI(lines, "@")
	}

	ext.AddEmoji.SetFromString(toString(custom["add_emoji"]))
	ext.RemoveOldEmoji.SetFromString(toString(custom["remove_old_emoji"]))

	if emoji, ok := custom["emojis"].([]interface{}); ok {
		lines := readEmojiFromYAML(emoji)
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.Emoji = ParseRegexMatchFromINI(lines, ",")
	}

	ext.Include = toStringSlice(custom["include_remarks"])
	ext.Exclude = toStringSlice(custom["exclude_remarks"])

	if args, ok := root["template_args"].([]interface{}); ok {
		for _, item := range args {
			m, _ := item.(map[string]interface{})
			if m == nil {
				continue
			}
			key := toString(m["key"])
			value := toString(m["value"])
			if key != "" {
				ext.TemplateArgs[key] = value
			}
		}
	}
	return nil
}

func loadExternalTOML(content string, ext *ExternalConfig, settings *Settings, fetch FetchFunc) error {
	var root map[string]interface{}
	if err := toml.Unmarshal([]byte(content), &root); err != nil {
		return err
	}
	if v, ok := toInt(root["version"]); !ok || v == 0 {
		return errors.New("missing version")
	}
	custom, ok := root["custom"].(map[string]interface{})
	if !ok {
		return errors.New("no custom section")
	}

	ext.ClashRuleBase = toString(custom["clash_rule_base"])
	ext.SurgeRuleBase = toString(custom["surge_rule_base"])
	ext.SurfboardRuleBase = toString(custom["surfboard_rule_base"])
	ext.MellowRuleBase = toString(custom["mellow_rule_base"])
	ext.QuanRuleBase = toString(custom["quan_rule_base"])
	ext.QuanXRuleBase = toString(custom["quanx_rule_base"])
	ext.LoonRuleBase = toString(custom["loon_rule_base"])
	ext.SSSubRuleBase = toString(custom["sssub_rule_base"])
	ext.SingBoxRuleBase = toString(custom["singbox_rule_base"])

	if b, ok := toBool(custom["enable_rule_generator"]); ok {
		ext.EnableRuleGenerator = b
	}
	if b, ok := toBool(custom["overwrite_original_rules"]); ok {
		ext.OverwriteOriginalRules = b
	}

	if groups, ok := root["custom_groups"].([]interface{}); ok {
		lines := readGroupsFromTOML(groups)
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.CustomProxyGroups = ParseProxyGroupsFromINI(lines)
	}
	if rulesets, ok := root["rulesets"].([]interface{}); ok {
		lines := readRulesetsFromTOML(rulesets)
		if settings.MaxAllowedRulesets > 0 && len(lines) > settings.MaxAllowedRulesets {
			return errors.New("ruleset count exceeds limit")
		}
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.SurgeRulesets = ParseRulesetsFromINI(lines)
	}
	if emoji, ok := root["emoji"].([]interface{}); ok {
		lines := readRegexMatchFromTOML(emoji, ",")
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.Emoji = ParseRegexMatchFromINI(lines, ",")
	}
	if rename, ok := root["rename_node"].([]interface{}); ok {
		lines := readRegexMatchFromTOML(rename, "@")
		lines, _ = ImportItems(lines, settings.APIMode, fetch)
		ext.Rename = ParseRegexMatchFromINI(lines, "@")
	}

	ext.AddEmoji.SetFromString(toString(custom["add_emoji"]))
	ext.RemoveOldEmoji.SetFromString(toString(custom["remove_old_emoji"]))
	ext.Include = toStringSlice(custom["include_remarks"])
	ext.Exclude = toStringSlice(custom["exclude_remarks"])

	if args, ok := root["template_args"].([]interface{}); ok {
		for _, item := range args {
			m, _ := item.(map[string]interface{})
			if m == nil {
				continue
			}
			key := toString(m["key"])
			value := toString(m["value"])
			if key != "" {
				ext.TemplateArgs[key] = value
			}
		}
	}
	return nil
}

func loadExternalINI(content string, ext *ExternalConfig, settings *Settings, fetch FetchFunc) error {
	ini := NewIni()
	ini.AllowDuplicateSections = true
	ini.StoreIsolatedLine = true
	ini.StoreAnyLine = true
	ini.IsolatedSection = "custom"
	ini.Parse(content)

	sec := ini.Section("custom")
	if sec == nil {
		return errors.New("custom section missing")
	}

	if sec.ItemPrefixExists("custom_proxy_group") {
		items := sec.GetAll("custom_proxy_group")
		items, _ = ImportItems(items, settings.APIMode, fetch)
		ext.CustomProxyGroups = ParseProxyGroupsFromINI(items)
	}

	key := "ruleset"
	if sec.ItemPrefixExists("surge_ruleset") {
		key = "surge_ruleset"
	}
	if sec.ItemPrefixExists(key) {
		items := sec.GetAll(key)
		items, _ = ImportItems(items, settings.APIMode, fetch)
		if settings.MaxAllowedRulesets > 0 && len(items) > settings.MaxAllowedRulesets {
			return errors.New("ruleset count exceeds limit")
		}
		ext.SurgeRulesets = ParseRulesetsFromINI(items)
	}

	if v := sec.Get("clash_rule_base"); v != "" {
		ext.ClashRuleBase = v
	}
	if v := sec.Get("surge_rule_base"); v != "" {
		ext.SurgeRuleBase = v
	}
	if v := sec.Get("surfboard_rule_base"); v != "" {
		ext.SurfboardRuleBase = v
	}
	if v := sec.Get("mellow_rule_base"); v != "" {
		ext.MellowRuleBase = v
	}
	if v := sec.Get("quan_rule_base"); v != "" {
		ext.QuanRuleBase = v
	}
	if v := sec.Get("quanx_rule_base"); v != "" {
		ext.QuanXRuleBase = v
	}
	if v := sec.Get("loon_rule_base"); v != "" {
		ext.LoonRuleBase = v
	}
	if v := sec.Get("sssub_rule_base"); v != "" {
		ext.SSSubRuleBase = v
	}
	if v := sec.Get("singbox_rule_base"); v != "" {
		ext.SingBoxRuleBase = v
	}
	if b, ok := toBool(sec.Get("overwrite_original_rules")); ok {
		ext.OverwriteOriginalRules = b
	}
	if b, ok := toBool(sec.Get("enable_rule_generator")); ok {
		ext.EnableRuleGenerator = b
	}

	if sec.ItemPrefixExists("rename") {
		items := sec.GetAll("rename")
		items, _ = ImportItems(items, settings.APIMode, fetch)
		ext.Rename = ParseRegexMatchFromINI(items, "@")
	}
	ext.AddEmoji.SetFromString(sec.Get("add_emoji"))
	ext.RemoveOldEmoji.SetFromString(sec.Get("remove_old_emoji"))
	if sec.ItemPrefixExists("emoji") {
		items := sec.GetAll("emoji")
		items, _ = ImportItems(items, settings.APIMode, fetch)
		ext.Emoji = ParseRegexMatchFromINI(items, ",")
	}
	if sec.ItemPrefixExists("include_remarks") {
		ext.Include = sec.GetAll("include_remarks")
	}
	if sec.ItemPrefixExists("exclude_remarks") {
		ext.Exclude = sec.GetAll("exclude_remarks")
	}

	if tmpl := ini.Section("template"); tmpl != nil {
		for _, item := range tmpl.Items {
			if item.Key == "{NONAME}" {
				continue
			}
			ext.TemplateArgs[item.Key] = item.Value
		}
	}
	return nil
}
