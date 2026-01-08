package mihomo

import (
	"net/url"
	"strings"

	"github.com/metacubex/mihomo/common/convert"
)

// ParseLinks converts a list of node links into Mihomo-compatible proxy maps.
func ParseLinks(links []string) ([]map[string]any, error) {
	var lines []string
	for _, link := range links {
		link = strings.TrimSpace(link)
		if link == "" {
			continue
		}
		if decoded, err := url.QueryUnescape(link); err == nil {
			link = decoded
		}
		lines = append(lines, link)
	}
	if len(lines) == 0 {
		return nil, nil
	}
	return convert.ConvertsV2Ray([]byte(strings.Join(lines, "\n")))
}
