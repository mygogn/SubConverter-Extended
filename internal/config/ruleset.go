package config

type RulesetType int

const (
	RulesetSurge RulesetType = iota
	RulesetQuanX
	RulesetClashDomain
	RulesetClashIPCIDR
	RulesetClashClassical
)

type RulesetContent struct {
	Group          string
	Path           string
	PathTyped      string
	Type           RulesetType
	Content        string
	UpdateInterval int
}
