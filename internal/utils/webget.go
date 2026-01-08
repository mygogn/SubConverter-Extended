package utils

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultUserAgent = "clash.meta"

type FetchOptions struct {
	Proxy            string
	CacheTTL         int
	MaxSize          int64
	ServeCacheOnFail bool
	RequestHeaders   map[string]string
}

func FetchURL(rawURL string, opts FetchOptions) (string, error) {
	if rawURL == "" {
		return "", errors.New("empty url")
	}
	if StringsHasPrefixFold(rawURL, "data:") {
		return dataGet(rawURL), nil
	}

	cachePath := ""
	if opts.CacheTTL > 0 {
		cachePath = cacheFilePath(rawURL)
		if content, ok := readCache(cachePath, opts.CacheTTL); ok {
			return content, nil
		}
	}

	reqURL := rawURL
	headers := make(map[string]string)
	for k, v := range opts.RequestHeaders {
		headers[k] = v
	}

	proxy := strings.TrimSpace(opts.Proxy)
	if strings.HasPrefix(proxy, "cors:") {
		reqURL = proxy[len("cors:"):] + rawURL
		headers["X-Requested-With"] = "subconverter"
		proxy = ""
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			Proxy:           proxyFunc(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	if _, ok := headers["User-Agent"]; !ok {
		req.Header.Set("User-Agent", defaultUserAgent)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		if opts.ServeCacheOnFail && cachePath != "" {
			if content, ok := readCache(cachePath, 0); ok {
				return content, nil
			}
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if opts.ServeCacheOnFail && cachePath != "" {
			if content, ok := readCache(cachePath, 0); ok {
				return content, nil
			}
		}
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	limit := opts.MaxSize
	if limit <= 0 {
		limit = 1024 * 1024
	}
	body, err := readLimited(resp.Body, limit)
	if err != nil {
		return "", err
	}

	if cachePath != "" {
		_ = os.MkdirAll(filepath.Dir(cachePath), 0755)
		_ = os.WriteFile(cachePath, []byte(body), 0644)
	}

	return body, nil
}

func readLimited(reader io.Reader, limit int64) (string, error) {
	lr := &io.LimitedReader{R: reader, N: limit + 1}
	data, err := io.ReadAll(lr)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > limit {
		return "", errors.New("response exceeds size limit")
	}
	return string(data), nil
}

func cacheFilePath(rawURL string) string {
	hash := GetMD5(rawURL)
	return filepath.Join("cache", hash)
}

func readCache(path string, ttl int) (string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	if ttl > 0 {
		age := time.Since(info.ModTime())
		if age > time.Duration(ttl)*time.Second {
			return "", false
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func proxyFunc(proxy string) func(*http.Request) (*url.URL, error) {
	if proxy == "" {
		return nil
	}
	if proxy == "SYSTEM" {
		return http.ProxyFromEnvironment
	}
	if !strings.Contains(proxy, "://") {
		return nil
	}
	parsed, err := url.Parse(proxy)
	if err != nil {
		return nil
	}
	return http.ProxyURL(parsed)
}

func dataGet(rawURL string) string {
	if !StringsHasPrefixFold(rawURL, "data:") {
		return ""
	}
	comma := strings.IndexByte(rawURL, ',')
	if comma == -1 || comma == len(rawURL)-1 {
		return ""
	}
	meta := rawURL[:comma]
	data := UrlDecode(rawURL[comma+1:])
	if strings.HasSuffix(meta, ";base64") {
		return UrlSafeBase64Decode(data)
	}
	return data
}

func StringsHasPrefixFold(s, prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}
