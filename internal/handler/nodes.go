package handler

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/mihomo"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type proxyNode struct {
	Name    string
	Type    string
	Server  string
	Port    int
	Group   string
	GroupID int
	Raw     map[string]any
}

var (
	groupIDRuleRe = regexp.MustCompile(`^!!(?:GROUPID|INSERT)=([\d\-+!,]+)(?:!!(.*))?$`)
	groupRuleRe   = regexp.MustCompile(`^!!(?:GROUP)=(.+?)(?:!!(.*))?$`)
	typeRuleRe    = regexp.MustCompile(`^!!(?:TYPE)=(.+?)(?:!!(.*))?$`)
	portRuleRe    = regexp.MustCompile(`^!!(?:PORT)=(.+?)(?:!!(.*))?$`)
	serverRuleRe  = regexp.MustCompile(`^!!(?:SERVER)=(.+?)(?:!!(.*))?$`)

	rangeNumRe      = regexp.MustCompile(`^-?\d+$`)
	rangeRe         = regexp.MustCompile(`^(\d+)-(\d+)$`)
	rangeNotRe      = regexp.MustCompile(`^!\-?\d+$`)
	rangeNotRangeRe = regexp.MustCompile(`^!(\d+)-(\d+)$`)
	rangeLessRe     = regexp.MustCompile(`^(\d+)-$`)
	rangeMoreRe     = regexp.MustCompile(`^(\d+)\+$`)
)

var (
	compatSupported = map[string]map[string]struct{}{
		"udp": {
			"anytls": {}, "socks": {}, "socks5": {}, "socks5h": {}, "ss": {}, "ssr": {}, "trojan": {}, "vless": {}, "vmess": {},
		},
		"skip-cert-verify": {
			"anytls": {}, "http": {}, "https": {}, "hy2": {}, "hysteria": {}, "hysteria2": {}, "socks": {}, "socks5": {}, "socks5h": {}, "trojan": {}, "tuic": {}, "vless": {}, "vmess": {},
		},
		"tfo": {
			"anytls": {}, "http": {}, "https": {}, "hy2": {}, "hysteria": {}, "hysteria2": {}, "socks": {}, "socks5": {}, "socks5h": {}, "ss": {}, "ssr": {}, "trojan": {}, "tuic": {}, "vless": {}, "vmess": {},
		},
		"xudp": {
			"vless": {}, "vmess": {},
		},
	}
	compatHardcoded = map[string]map[string]struct{}{
		"udp": {
			"anytls": {}, "ss": {}, "ssr": {}, "trojan": {}, "vmess": {},
		},
		"skip-cert-verify": {
			"http": {}, "vmess": {},
		},
		"xudp": {
			"vmess": {},
		},
	}
)

func parseNodeLink(link string, groupID int, group string) ([]proxyNode, error) {
	proxies, err := mihomo.ParseLinks([]string{link})
	if err != nil {
		return nil, err
	}
	nodes := make([]proxyNode, 0, len(proxies))
	for _, proxy := range proxies {
		node := newProxyNode(proxy, groupID, group)
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func newProxyNode(raw map[string]any, groupID int, group string) proxyNode {
	node := proxyNode{
		Group:   group,
		GroupID: groupID,
		Raw:     cloneMap(raw),
	}
	node.Name = stringValue(node.Raw["name"])
	node.Type = strings.ToLower(stringValue(node.Raw["type"]))
	node.Server = stringValue(node.Raw["server"])
	node.Port = intValue(node.Raw["port"])
	return node
}

func filterNodes(nodes []proxyNode, include, exclude []string) []proxyNode {
	out := nodes[:0]
	for _, node := range nodes {
		if shouldIgnoreNode(node, include, exclude) {
			continue
		}
		out = append(out, node)
	}
	return out
}

func shouldIgnoreNode(node proxyNode, include, exclude []string) bool {
	for _, rule := range exclude {
		if rule == "" {
			continue
		}
		ok, real := applyMatcher(rule, node)
		if ok && (real == "" || utils.RegFind(node.Name, real)) {
			return true
		}
	}
	if len(include) == 0 {
		return false
	}
	for _, rule := range include {
		if rule == "" {
			continue
		}
		ok, real := applyMatcher(rule, node)
		if ok && (real == "" || utils.RegFind(node.Name, real)) {
			return false
		}
	}
	return true
}

func applyMatcher(rule string, node proxyNode) (bool, string) {
	switch {
	case groupRuleRe.MatchString(rule):
		match := groupRuleRe.FindStringSubmatch(rule)
		target := match[1]
		real := match[2]
		return utils.RegFind(node.Group, target), real
	case groupIDRuleRe.MatchString(rule):
		match := groupIDRuleRe.FindStringSubmatch(rule)
		target := match[1]
		real := match[2]
		dir := 1
		if strings.HasPrefix(rule, "!!INSERT=") {
			dir = -1
		}
		return matchRange(target, dir*node.GroupID), real
	case typeRuleRe.MatchString(rule):
		match := typeRuleRe.FindStringSubmatch(rule)
		target := match[1]
		real := match[2]
		typ := nodeTypeForMatch(node.Type)
		if typ == "" {
			return false, real
		}
		return utils.RegMatch(typ, target), real
	case portRuleRe.MatchString(rule):
		match := portRuleRe.FindStringSubmatch(rule)
		target := match[1]
		real := match[2]
		return matchRange(target, node.Port), real
	case serverRuleRe.MatchString(rule):
		match := serverRuleRe.FindStringSubmatch(rule)
		target := match[1]
		real := match[2]
		return utils.RegFind(node.Server, target), real
	default:
		return true, rule
	}
}

func matchRange(expr string, target int) bool {
	parts := utils.Split(expr, ",")
	match := false
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case rangeNumRe.MatchString(part):
			if toInt(part, 0) == target {
				match = true
			}
		case rangeRe.MatchString(part):
			sub := rangeRe.FindStringSubmatch(part)
			begin := toInt(sub[1], 0)
			end := toInt(sub[2], 0)
			if target >= begin && target <= end {
				match = true
			}
		case rangeNotRe.MatchString(part):
			match = true
			value := toInt(strings.TrimPrefix(part, "!"), 0)
			if target == value {
				match = false
			}
		case rangeNotRangeRe.MatchString(part):
			match = true
			sub := rangeNotRangeRe.FindStringSubmatch(part)
			begin := toInt(sub[1], 0)
			end := toInt(sub[2], 0)
			if target >= begin && target <= end {
				match = false
			}
		case rangeLessRe.MatchString(part):
			sub := rangeLessRe.FindStringSubmatch(part)
			if toInt(sub[1], 0) >= target {
				match = true
			}
		case rangeMoreRe.MatchString(part):
			sub := rangeMoreRe.FindStringSubmatch(part)
			if toInt(sub[1], 0) <= target {
				match = true
			}
		}
	}
	return match
}

func applyRenames(nodes []proxyNode, rules []config.RegexMatchConfig) {
	for i := range nodes {
		name := nodes[i].Name
		original := name
		for _, rule := range rules {
			if rule.Script != "" || rule.Match == "" {
				continue
			}
			ok, real := applyMatcher(rule.Match, nodes[i])
			if ok && real != "" {
				name = utils.RegReplace(name, real, rule.Replace)
			}
		}
		if name == "" {
			name = original
		}
		nodes[i].Name = name
	}
}

func applyEmojis(nodes []proxyNode, rules []config.RegexMatchConfig, addEmoji bool, removeEmoji bool) {
	if !addEmoji && !removeEmoji {
		return
	}
	for i := range nodes {
		name := strings.TrimSpace(nodes[i].Name)
		if removeEmoji {
			name = strings.TrimSpace(removeLeadingEmoji(name))
		}
		if addEmoji {
			name = addEmojiPrefix(nodes[i], name, rules)
		}
		nodes[i].Name = name
	}
}

func addEmojiPrefix(node proxyNode, name string, rules []config.RegexMatchConfig) string {
	for _, rule := range rules {
		if rule.Script != "" || rule.Replace == "" || rule.Match == "" {
			continue
		}
		ok, real := applyMatcher(rule.Match, node)
		if ok && real != "" && utils.RegFind(name, real) {
			return rule.Replace + " " + name
		}
	}
	return name
}

func removeLeadingEmoji(value string) string {
	buf := []byte(value)
	for len(buf) >= 4 && buf[0] == 0xF0 && buf[1] == 0x9F {
		buf = buf[4:]
	}
	if len(buf) == 0 {
		return value
	}
	return string(buf)
}

func appendTypePrefix(nodes []proxyNode, enabled bool) {
	if !enabled {
		return
	}
	for i := range nodes {
		typ := displayType(nodes[i].Type)
		if typ == "" {
			continue
		}
		nodes[i].Name = "[" + typ + "] " + nodes[i].Name
	}
}

func sortNodesByName(nodes []proxyNode, enabled bool) {
	if !enabled {
		return
	}
	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})
}

func ensureUniqueNames(nodes []proxyNode) {
	used := make(map[string]struct{}, len(nodes))
	for i := range nodes {
		name := strings.ReplaceAll(strings.TrimSpace(nodes[i].Name), "=", "-")
		if name == "" {
			name = fmt.Sprintf("node_%d", i+1)
		}
		base := name
		for idx := 2; ; idx++ {
			if _, ok := used[name]; !ok {
				break
			}
			name = fmt.Sprintf("%s %d", base, idx)
		}
		used[name] = struct{}{}
		nodes[i].Name = name
	}
}

func applyOverrides(nodes []proxyNode, udp utils.TriBool, scv utils.TriBool, tfo utils.TriBool, xudp utils.TriBool) {
	for i := range nodes {
		typ := nodes[i].Type
		raw := nodes[i].Raw
		if raw == nil {
			raw = make(map[string]any)
			nodes[i].Raw = raw
		}
		if nodes[i].Name != "" {
			raw["name"] = nodes[i].Name
		}
		if nodes[i].Server != "" && raw["server"] == nil {
			raw["server"] = nodes[i].Server
		}
		if nodes[i].Port != 0 && raw["port"] == nil {
			raw["port"] = nodes[i].Port
		}

		if val, ok := udp.Value(); ok {
			applyOverride(raw, typ, "udp", val)
		}
		if val, ok := scv.Value(); ok {
			applyOverride(raw, typ, "skip-cert-verify", val)
		}
		if val, ok := tfo.Value(); ok {
			applyOverride(raw, typ, "tfo", val)
		}
		if val, ok := xudp.Value(); ok {
			applyOverride(raw, typ, "xudp", val)
		}
	}
}

func applyOverride(raw map[string]any, typ string, key string, value bool) {
	typ = strings.ToLower(typ)
	if typ == "" {
		return
	}
	supported := compatSupported[key]
	if supported != nil {
		if _, ok := supported[typ]; !ok {
			return
		}
	}
	if hardcoded := compatHardcoded[key]; hardcoded != nil {
		if _, ok := hardcoded[typ]; ok {
			return
		}
	}
	if _, ok := raw[key]; ok {
		return
	}
	raw[key] = value
}

func filterDeprecatedNodes(nodes []proxyNode, enabled bool) []proxyNode {
	if !enabled {
		return nodes
	}
	out := nodes[:0]
	for _, node := range nodes {
		if node.Type == "ss" {
			cipher := strings.ToLower(stringValue(node.Raw["cipher"]))
			if cipher == "chacha20" {
				continue
			}
		}
		out = append(out, node)
	}
	return out
}

func nodeTypeForMatch(typ string) string {
	switch strings.ToLower(typ) {
	case "ss":
		return "SS"
	case "ssr":
		return "SSR"
	case "vmess":
		return "VMESS"
	case "trojan":
		return "TROJAN"
	case "snell":
		return "SNELL"
	case "http":
		return "HTTP"
	case "https":
		return "HTTPS"
	case "socks", "socks5", "socks5h":
		return "SOCKS5"
	case "wireguard":
		return "WIREGUARD"
	case "vless":
		return "VLESS"
	case "hysteria":
		return "HYSTERIA"
	case "hysteria2", "hy2":
		return "HYSTERIA2"
	case "tuic":
		return "TUIC"
	}
	return strings.ToUpper(typ)
}

func displayType(typ string) string {
	switch strings.ToLower(typ) {
	case "ss":
		return "SS"
	case "ssr":
		return "SSR"
	case "vmess":
		return "VMESS"
	case "vless":
		return "VLESS"
	case "trojan":
		return "TROJAN"
	case "snell":
		return "SNELL"
	case "http":
		return "HTTP"
	case "https":
		return "HTTPS"
	case "socks", "socks5", "socks5h":
		return "SOCKS5"
	case "wireguard":
		return "WIREGUARD"
	case "hysteria":
		return "HYSTERIA"
	case "hysteria2", "hy2":
		return "HYSTERIA2"
	case "tuic":
		return "TUIC"
	case "anytls":
		return "ANYTLS"
	case "mieru":
		return "MIERU"
	}
	return strings.ToUpper(typ)
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case fmt.Stringer:
		return v.String()
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatInt(int64(v), 10)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprint(v)
	}
}

func intValue(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		return toInt(v, 0)
	default:
		return 0
	}
}

func toInt(value string, def int) int {
	if value == "" {
		return def
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return def
	}
	return parsed
}

func buildProxyList(nodes []proxyNode) []map[string]any {
	out := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		if node.Raw == nil {
			continue
		}
		out = append(out, node.Raw)
	}
	return out
}
