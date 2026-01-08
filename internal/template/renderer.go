package template

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nikolalohinski/gonja"
	"github.com/nikolalohinski/gonja/config"

	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type Args struct {
	GlobalVars    map[string]string
	RequestParams map[string]string
	LocalVars     map[string]string
}

type FetchFunc func(url string) (string, error)

type Renderer struct {
	IncludeScope  string
	ManagedPrefix string
	Fetch         FetchFunc
}

func (r *Renderer) Render(content string, args Args) (string, error) {
	data := make(map[string]interface{})
	data["global"] = buildMapFromFlat(args.GlobalVars)
	data["request"] = buildMapFromFlat(args.RequestParams)
	data["local"] = buildMapFromFlat(args.LocalVars)

	if req, ok := data["request"].(map[string]interface{}); ok {
		req["_args"] = buildArgs(args.RequestParams)
	}

	cfg := config.NewConfig()
	env := gonja.NewEnvironment(cfg, &templateLoader{
		scope: r.IncludeScope,
	})
	registerGlobals(env, data, r)

	normalized := applyLineStatements(content)
	tpl, err := env.FromString(normalized)
	if err != nil {
		return "Template render failed! Reason: " + err.Error(), err
	}
	out, err := tpl.Execute(data)
	if err != nil {
		return "Template render failed! Reason: " + err.Error(), err
	}
	return out, nil
}

func registerGlobals(env *gonja.Environment, data map[string]interface{}, r *Renderer) {
	env.Globals.Set("UrlEncode", func(val string) string {
		return utils.UrlEncode(val)
	})
	env.Globals.Set("UrlDecode", func(val string) string {
		return utils.UrlDecode(val)
	})
	env.Globals.Set("trim_of", func(val, target string) string {
		if target == "" {
			return val
		}
		return utils.TrimOf(val, rune(target[0]), true, true)
	})
	env.Globals.Set("trim", func(val string) string {
		return utils.Trim(val, true, true)
	})
	env.Globals.Set("find", func(src, target string) bool {
		return utils.RegFind(src, target)
	})
	env.Globals.Set("replace", func(src, target, rep string) string {
		return utils.RegReplace(src, target, rep)
	})
	env.Globals.Set("set", func(path, value string) string {
		setPath(data, path, value)
		return ""
	})
	env.Globals.Set("split", func(content, delim, dest string) string {
		list := utils.Split(content, delim)
		for i, item := range list {
			setPath(data, dest+"."+utils.ToString(i), item)
		}
		return ""
	})
	env.Globals.Set("append", func(path, value string) string {
		existing, ok := getPath(data, path).(string)
		if !ok {
			existing = ""
		}
		setPath(data, path, existing+value)
		return ""
	})
	env.Globals.Set("getLink", func(suffix string) string {
		return r.ManagedPrefix + suffix
	})
	env.Globals.Set("startsWith", func(hay, needle string) bool {
		return utils.StartsWith(hay, needle)
	})
	env.Globals.Set("endsWith", func(hay, needle string) bool {
		return utils.EndsWith(hay, needle)
	})
	env.Globals.Set("or", func(args ...interface{}) bool {
		for _, arg := range args {
			if truthy(arg) {
				return true
			}
		}
		return false
	})
	env.Globals.Set("and", func(args ...interface{}) bool {
		for _, arg := range args {
			if !truthy(arg) {
				return false
			}
		}
		return true
	})
	env.Globals.Set("bool", func(val string) int {
		val = strings.ToLower(val)
		if val == "true" || val == "1" {
			return 1
		}
		return 0
	})
	env.Globals.Set("string", func(val int) string {
		return utils.ToString(val)
	})
	env.Globals.Set("default", func(val, fallback interface{}) interface{} {
		if val == nil {
			return fallback
		}
		if s, ok := val.(string); ok {
			if s == "" {
				return fallback
			}
			return s
		}
		return val
	})
	env.Globals.Set("exists", func(path string) bool {
		return getPath(data, path) != nil
	})
	env.Globals.Set("fetch", func(url string) string {
		if r.Fetch == nil {
			return ""
		}
		content, err := r.Fetch(url)
		if err != nil {
			return ""
		}
		return content
	})
}

func buildMapFromFlat(input map[string]string) map[string]interface{} {
	out := make(map[string]interface{})
	for key, value := range input {
		setPath(out, key, value)
	}
	return out
}

func buildArgs(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	first := true
	for _, k := range keys {
		v := params[k]
		if !first {
			b.WriteByte('&')
		}
		first = false
		b.WriteString(k)
		if v != "" {
			b.WriteByte('=')
			b.WriteString(v)
		}
	}
	return b.String()
}

func setPath(root map[string]interface{}, path string, value interface{}) {
	if path == "" {
		return
	}
	parts := strings.Split(path, ".")
	cur := root
	for i := 0; i < len(parts)-1; i++ {
		key := parts[i]
		next, ok := cur[key].(map[string]interface{})
		if !ok {
			next = make(map[string]interface{})
			cur[key] = next
		}
		cur = next
	}
	cur[parts[len(parts)-1]] = value
}

func getPath(root map[string]interface{}, path string) interface{} {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	var cur interface{} = root
	for _, key := range parts {
		m, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur, ok = m[key]
		if !ok {
			return nil
		}
	}
	return cur
}

func truthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	default:
		return v != nil
	}
}

func applyLineStatements(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		idx := strings.Index(line, "#~#")
		if idx == -1 {
			continue
		}
		if strings.TrimSpace(line[:idx]) != "" {
			continue
		}
		stmt := strings.TrimSpace(line[idx+3:])
		lines[i] = "{% " + stmt + " %}"
	}
	return strings.Join(lines, "\n")
}

type templateLoader struct {
	scope string
}

func (l *templateLoader) Get(path string) (io.Reader, error) {
	resolved, err := l.Path(path)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(resolved)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(content), nil
}

func (l *templateLoader) Path(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty template path")
	}
	if l.scope == "" {
		return filepath.Abs(path)
	}
	scope, err := filepath.Abs(l.scope)
	if err != nil {
		return "", err
	}
	resolved := path
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(scope, resolved)
	}
	resolved, err = filepath.Abs(resolved)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(resolved, scope) {
		return "", errors.New("access denied when trying to include '" + path + "': out of scope")
	}
	return resolved, nil
}
