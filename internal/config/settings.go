package config

import (
	"sync"

	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type Settings struct {
	PrefPath         string
	DefaultExtConfig string
	ExcludeRemarks   []string
	IncludeRemarks   []string
	CustomRulesets   []RulesetConfig
	StreamNodeRules  []RegexMatchConfig
	TimeNodeRules    []RegexMatchConfig
	RulesetsContent  []RulesetContent
	ListenAddress    string
	ListenPort       int
	MaxPendingConns  int
	MaxConcurThreads int
	DefaultUrls      string
	InsertUrls       string
	ManagedPrefix    string
	PrependInsert    bool
	SkipFailedLinks  bool
	APIMode          bool
	WriteManaged     bool
	EnableRuleGen    bool
	UpdateRulesetOnRequest bool
	OverwriteOriginalRules bool
	PrintDebugInfo   bool
	CFWChildProcess  bool
	AppendUserInfo   bool
	AsyncFetchRuleset bool
	SurgeResolveHostname bool
	BasePath         string
	CustomGroup      string
	LogLevel         int
	MaxAllowedDownloadSize int64

	TemplatePath string
	TemplateVars map[string]string

	GeneratorMode   bool
	GenerateProfiles string

	ReloadConfOnRequest bool
	Renames            []RegexMatchConfig
	Emojis             []RegexMatchConfig
	AddEmoji           bool
	RemoveEmoji        bool
	AppendType         bool
	FilterDeprecated   bool
	UDPFlag            utils.TriBool
	TFOFlag            utils.TriBool
	SkipCertVerify     utils.TriBool
	TLS13Flag          utils.TriBool
	EnableInsert       utils.TriBool
	EnableSort         bool
	UpdateStrict       bool
	ClashUseNewField   bool
	SingBoxAddClashModes bool
	ClashProxiesStyle  string
	ClashProxyGroupsStyle string
	ProxyConfig        string
	ProxyRuleset       string
	ProxySubscription  string
	UpdateInterval     int
	SortScript         string
	FilterScript       string

	ClashBase   string
	CustomProxyGroups []ProxyGroupConfig
	SurgeBase   string
	SurfboardBase string
	MellowBase  string
	QuanBase    string
	QuanXBase   string
	LoonBase    string
	SSSubBase   string
	SingBoxBase string
	SurgeSSRPath string
	QuanXDevID   string

	ServeCacheOnFetchFail bool
	CacheSubscription int
	CacheConfig int
	CacheRuleset int

	MaxAllowedRulesets int
	MaxAllowedRules    int
	ScriptCleanContext bool

	EnableCron bool
	CronTasks  []CronTaskConfig

	Aliases       map[string]string
	ServeFileRoot string

	mu sync.RWMutex
}

func DefaultSettings() *Settings {
	return &Settings{
		PrefPath:         "pref.ini",
		ListenAddress:    "127.0.0.1",
		ListenPort:       25500,
		MaxPendingConns:  10,
		MaxConcurThreads: 4,
		APIMode:          true,
		BasePath:         "base",
		LogLevel:         0,
		MaxAllowedDownloadSize: 1048576,
		TemplatePath:     "templates",
		ClashProxiesStyle: "flow",
		ClashProxyGroupsStyle: "block",
		CacheSubscription: 60,
		CacheConfig:       300,
		CacheRuleset:      21600,
		MaxAllowedRulesets: 64,
		MaxAllowedRules:    32768,
		AppendUserInfo:     true,
		EnableRuleGen:      true,
		OverwriteOriginalRules: true,
		TemplateVars:       make(map[string]string),
	}
}

func (s *Settings) SetEmojis(items []RegexMatchConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Emojis = items
}

func (s *Settings) SetRenames(items []RegexMatchConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Renames = items
}

func (s *Settings) SetStreamRules(items []RegexMatchConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.StreamNodeRules = items
}

func (s *Settings) SetTimeRules(items []RegexMatchConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TimeNodeRules = items
}

func (s *Settings) GetEmojis() []RegexMatchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]RegexMatchConfig(nil), s.Emojis...)
}

func (s *Settings) GetRenames() []RegexMatchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]RegexMatchConfig(nil), s.Renames...)
}

func (s *Settings) GetStreamRules() []RegexMatchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]RegexMatchConfig(nil), s.StreamNodeRules...)
}

func (s *Settings) GetTimeRules() []RegexMatchConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]RegexMatchConfig(nil), s.TimeNodeRules...)
}
