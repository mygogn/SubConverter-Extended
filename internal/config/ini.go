package config

import (
	"bufio"
	"strings"

	"github.com/aethersailor/subconverter-extended/internal/utils"
)

type IniItem struct {
	Key   string
	Value string
}

type IniSection struct {
	Name   string
	Items  []IniItem
	Values map[string][]string
}

type IniFile struct {
	Sections map[string]*IniSection
	Order    []string

	AllowDuplicateSections bool
	StoreAnyLine           bool
	StoreIsolatedLine      bool
	IsolatedSection        string
}

func NewIni() *IniFile {
	return &IniFile{
		Sections: make(map[string]*IniSection),
	}
}

func (ini *IniFile) Parse(content string) {
	content = strings.TrimPrefix(content, "\xEF\xBB\xBF")
	scanner := bufio.NewScanner(strings.NewReader(content))
	var current *IniSection
	inIsolated := false

	if ini.StoreIsolatedLine && ini.IsolatedSection != "" {
		current = ini.getOrCreateSection(ini.IsolatedSection)
		inIsolated = true
	}

	for scanner.Scan() {
		line := utils.TrimWhitespace(scanner.Text(), true, true)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		line = utils.ProcessEscapeChar(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			current = ini.getOrCreateSection(sectionName)
			inIsolated = false
			continue
		}

		if current == nil {
			// Out of section and no isolated handling.
			continue
		}

		pos := strings.Index(line, "=")
		if pos == -1 {
			if ini.StoreAnyLine || inIsolated {
				current.addItem("{NONAME}", line)
			}
			continue
		}

		key := utils.Trim(line[:pos], true, true)
		value := ""
		if pos+1 < len(line) {
			value = utils.TrimWhitespace(line[pos+1:], true, true)
		}
		current.addItem(key, value)
	}
}

func (ini *IniFile) getOrCreateSection(name string) *IniSection {
	section, ok := ini.Sections[name]
	if !ok {
		section = &IniSection{
			Name:   name,
			Values: make(map[string][]string),
		}
		ini.Sections[name] = section
		ini.Order = append(ini.Order, name)
		return section
	}
	if ini.AllowDuplicateSections {
		return section
	}
	return section
}

func (s *IniSection) addItem(key, value string) {
	s.Items = append(s.Items, IniItem{Key: key, Value: value})
	s.Values[key] = append(s.Values[key], value)
}

func (ini *IniFile) Section(name string) *IniSection {
	return ini.Sections[name]
}

func (s *IniSection) Get(key string) string {
	if s == nil {
		return ""
	}
	values := s.Values[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (s *IniSection) GetAll(key string) []string {
	if s == nil {
		return nil
	}
	return append([]string(nil), s.Values[key]...)
}

func (s *IniSection) ItemPrefixExists(prefix string) bool {
	if s == nil {
		return false
	}
	for key := range s.Values {
		if key == prefix || strings.HasPrefix(key, prefix) {
			return true
		}
	}
	return false
}

func (s *IniSection) ItemsMap() map[string]string {
	if s == nil {
		return map[string]string{}
	}
	out := make(map[string]string)
	for _, item := range s.Items {
		if item.Key == "{NONAME}" {
			continue
		}
		out[item.Key] = item.Value
	}
	return out
}
