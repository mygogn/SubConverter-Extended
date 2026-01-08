package utils

import "strings"

// TriBool represents a tri-state boolean with an undefined state.
type TriBool struct {
	set   bool
	value bool
}

func (t *TriBool) Set(value bool) {
	t.set = true
	t.value = value
}

func (t *TriBool) SetFromString(value string) {
	if value == "" {
		return
	}
	switch strings.ToLower(value) {
	case "true", "1":
		t.Set(true)
	case "false", "0":
		t.Set(false)
	}
}

func (t TriBool) IsUndef() bool {
	return !t.set
}

func (t TriBool) Get(defaultValue bool) bool {
	if t.set {
		return t.value
	}
	return defaultValue
}

func (t *TriBool) Define(other TriBool) *TriBool {
	if t.set {
		return t
	}
	if other.set {
		t.set = true
		t.value = other.value
	}
	return t
}

func (t TriBool) Value() (bool, bool) {
	return t.value, t.set
}
