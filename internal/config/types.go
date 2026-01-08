package config

import "github.com/aethersailor/subconverter-extended/internal/utils"

type RegexMatchConfig struct {
	Match   string
	Replace string
	Script  string
}

type RulesetConfig struct {
	Group    string
	URL      string
	Interval int
}

type ProxyGroupType string

const (
	GroupSelect      ProxyGroupType = "select"
	GroupURLTest     ProxyGroupType = "url-test"
	GroupFallback    ProxyGroupType = "fallback"
	GroupLoadBalance ProxyGroupType = "load-balance"
	GroupRelay       ProxyGroupType = "relay"
	GroupSSID        ProxyGroupType = "ssid"
	GroupSmart       ProxyGroupType = "smart"
)

type BalanceStrategy string

const (
	BalanceConsistentHashing BalanceStrategy = "consistent-hashing"
	BalanceRoundRobin        BalanceStrategy = "round-robin"
)

type ProxyGroupConfig struct {
	Name            string
	Type            ProxyGroupType
	Proxies         []string
	UsingProvider   []string
	URL             string
	Interval        int
	Timeout         int
	Tolerance       int
	Strategy        BalanceStrategy
	Lazy            utils.TriBool
	DisableUDP      utils.TriBool
	Persistent      utils.TriBool
	EvaluateBefore  utils.TriBool
}

type CronTaskConfig struct {
	Name    string
	CronExp string
	Path    string
	Timeout int
}
