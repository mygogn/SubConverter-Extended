package handler

import (
	"sync"

	"github.com/aethersailor/subconverter-extended/internal/config"
)

type Handler struct {
	Settings *config.Settings
	Fetcher  *Fetcher

	mu             sync.RWMutex
	rulesetCache   []config.RulesetContent
	rulesetLoaded  bool
}

func New(settings *config.Settings) *Handler {
	return &Handler{
		Settings: settings,
		Fetcher:  &Fetcher{Settings: settings},
	}
}

func (h *Handler) CachedRulesets() ([]config.RulesetContent, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if !h.rulesetLoaded {
		return nil, false
	}
	out := make([]config.RulesetContent, len(h.rulesetCache))
	copy(out, h.rulesetCache)
	return out, true
}

func (h *Handler) SetCachedRulesets(contents []config.RulesetContent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rulesetCache = contents
	h.rulesetLoaded = true
}
