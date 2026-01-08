package handler

import (
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type Fetcher struct {
	Settings *config.Settings
}

func (f *Fetcher) FetchConfig(url string) (string, error) {
	return utils.FetchURL(url, utils.FetchOptions{
		Proxy:            parseProxy(f.Settings.ProxyConfig),
		CacheTTL:         f.Settings.CacheConfig,
		MaxSize:          f.Settings.MaxAllowedDownloadSize,
		ServeCacheOnFail: f.Settings.ServeCacheOnFetchFail,
	})
}

func (f *Fetcher) FetchRuleset(url string) (string, error) {
	return utils.FetchURL(url, utils.FetchOptions{
		Proxy:            parseProxy(f.Settings.ProxyRuleset),
		CacheTTL:         f.Settings.CacheRuleset,
		MaxSize:          f.Settings.MaxAllowedDownloadSize,
		ServeCacheOnFail: f.Settings.ServeCacheOnFetchFail,
	})
}

func (f *Fetcher) FetchSubscription(url string) (string, error) {
	return utils.FetchURL(url, utils.FetchOptions{
		Proxy:            parseProxy(f.Settings.ProxySubscription),
		CacheTTL:         f.Settings.CacheSubscription,
		MaxSize:          f.Settings.MaxAllowedDownloadSize,
		ServeCacheOnFail: f.Settings.ServeCacheOnFetchFail,
	})
}

func parseProxy(source string) string {
	switch strings.ToUpper(strings.TrimSpace(source)) {
	case "SYSTEM":
		return "SYSTEM"
	case "NONE":
		return ""
	default:
		return strings.TrimSpace(source)
	}
}
