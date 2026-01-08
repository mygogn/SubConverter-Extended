package utils

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func IsInScope(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	if runtime.GOOS == "windows" {
		if strings.Contains(path, ":\\") || strings.Contains(path, ":/") {
			return false
		}
	} else {
		if strings.HasPrefix(path, "/") {
			return false
		}
	}
	return true
}

func FileExists(path string, scopeLimit bool) bool {
	if scopeLimit && !IsInScope(path) {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

func FileGet(path string, scopeLimit bool) string {
	if scopeLimit && !IsInScope(path) {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func FileWrite(path string, content string, overwrite bool) error {
	flag := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flag |= os.O_TRUNC
	} else {
		flag |= os.O_APPEND
	}
	f, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.WriteString(f, content)
	return err
}

func FileCopy(source, dest string) bool {
	in, err := os.Open(source)
	if err != nil {
		return false
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return false
	}
	out, err := os.Create(dest)
	if err != nil {
		return false
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return false
	}
	return true
}
