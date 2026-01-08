package app

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aethersailor/subconverter-extended/internal/config"
	"github.com/aethersailor/subconverter-extended/internal/handler"
	"github.com/aethersailor/subconverter-extended/internal/ruleset"
	"github.com/aethersailor/subconverter-extended/internal/utils"
	"github.com/aethersailor/subconverter-extended/internal/version"
)

// Run starts the service entrypoint.
func Run() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Chdir(filepath.Dir(exe))
	}

	options := parseArgs(os.Args[1:])
	prefPath, err := ensurePrefPath(options.PrefPath)
	if err != nil {
		utils.Logf(config.LogFatal, "config: %v", err)
		return
	}
	if abs, err := filepath.Abs(prefPath); err == nil {
		prefPath = abs
	}
	_ = os.Chdir(filepath.Dir(prefPath))

	settings, err := config.LoadSettingsFromPath(prefPath, nil)
	if err != nil {
		utils.Logf(config.LogFatal, "load config failed: %v", err)
		return
	}

	if managedPrefix := os.Getenv("MANAGED_PREFIX"); managedPrefix != "" {
		settings.ManagedPrefix = managedPrefix
		if settings.TemplateVars == nil {
			settings.TemplateVars = make(map[string]string)
		}
		settings.TemplateVars["managed_config_prefix"] = managedPrefix
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil {
			settings.ListenPort = port
		}
	}

	utils.SetLogLevel(settings.LogLevel)
	utils.Logf(config.LogInfo, "SubConverter %s starting up..", version.Version)

	h := handler.New(settings)
	if settings.EnableRuleGen && !settings.UpdateRulesetOnRequest && len(settings.CustomRulesets) > 0 {
		if contents, err := ruleset.RefreshRulesets(settings.CustomRulesets, h.Fetcher.FetchRuleset); err == nil {
			h.SetCachedRulesets(contents)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/version", h.Version)
	mux.HandleFunc("/sub", h.Sub)
	mux.HandleFunc("/getruleset", h.GetRuleset)

	addr := fmt.Sprintf("%s:%d", settings.ListenAddress, settings.ListenPort)
	utils.Logf(config.LogInfo, "Startup completed. Serving HTTP @ http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		utils.Logf(config.LogFatal, "server error: %v", err)
	}
}

type runOptions struct {
	PrefPath string
}

func parseArgs(args []string) runOptions {
	var opts runOptions
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-f", "--file":
			if i+1 < len(args) {
				opts.PrefPath = args[i+1]
				i++
			}
		}
	}
	return opts
}

func ensurePrefPath(path string) (string, error) {
	if path != "" {
		if utils.FileExists(path, false) {
			return path, nil
		}
		return "", fmt.Errorf("pref file not found: %s", path)
	}
	if utils.FileExists("pref.toml", false) {
		return "pref.toml", nil
	}
	if utils.FileExists("pref.yml", false) {
		return "pref.yml", nil
	}
	if utils.FileExists("pref.ini", false) {
		return "pref.ini", nil
	}
	if utils.FileExists("pref.example.toml", false) {
		return copyExample("pref.example.toml", "pref.toml")
	}
	if utils.FileExists("pref.example.yml", false) {
		return copyExample("pref.example.yml", "pref.yml")
	}
	if utils.FileExists("pref.example.ini", false) {
		return copyExample("pref.example.ini", "pref.ini")
	}
	return "", errors.New("no pref file found")
}

func copyExample(src, dst string) (string, error) {
	content := utils.FileGet(src, false)
	if content == "" {
		return "", fmt.Errorf("empty example config: %s", src)
	}
	if err := os.WriteFile(dst, []byte(content), 0644); err != nil {
		return "", err
	}
	return dst, nil
}
