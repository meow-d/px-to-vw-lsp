package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// config
type Config struct {
	ViewportWidth float64 `json:"viewportWidth"`
	UnitPrecision int     `json:"unitPrecision"`
}

// TODO clean up vibe coded code
// GlobalConfig holds the global configuration and file monitoring
type GlobalConfig struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
	watcher    *fileWatcher
}

// fileWatcher monitors file changes
type fileWatcher struct {
	path    string
	watcher chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewGlobalConfig creates a new global config manager
func NewGlobalConfig(ctx context.Context, logger *zap.Logger) (*GlobalConfig, error) {
	sugar := logger.Sugar()

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		sugar.Warnf("Failed to get user config dir: %v", err)
		return &GlobalConfig{
			config:     &Config{ViewportWidth: 1440, UnitPrecision: 3},
			configPath: "",
		}, nil
	}

	configPath := filepath.Join(userConfigDir, "px-to-vw-lsp", "config.json")

	globalConfig := &GlobalConfig{
		config:     &Config{ViewportWidth: 1440, UnitPrecision: 3},
		configPath: configPath,
	}

	// Load global config if it exists
	if err := globalConfig.load(logger); err != nil {
		sugar.Warnf("Failed to load global config: %v", err)
	}

	// Start file monitoring
	if err := globalConfig.startWatcher(ctx, logger); err != nil {
		sugar.Warnf("Failed to start global config watcher: %v", err)
	}

	return globalConfig, nil
}

// Get returns the current global config
func (g *GlobalConfig) Get() *Config {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config
}

// load reads and parses the global config file
func (g *GlobalConfig) load(logger *zap.Logger) error {
	if g.configPath == "" {
		return nil
	}

	file, err := os.ReadFile(g.configPath)
	if err != nil {
		return err
	}

	cssremConfig, err := parseCssremConfig(file)
	if err != nil {
		return err
	}

	config := convertToConfig(*cssremConfig)

	g.mu.Lock()
	g.config = &config
	g.mu.Unlock()

	logger.Sugar().Infof("Loaded global config from %s: viewport=%.0f, precision=%d",
		g.configPath, config.ViewportWidth, config.UnitPrecision)

	return nil
}

// startWatcher monitors the global config file for changes
func (g *GlobalConfig) startWatcher(ctx context.Context, logger *zap.Logger) error {
	if g.configPath == "" {
		return nil
	}

	// Create a dedicated context for file watching with cancellation
	watcherCtx, cancel := context.WithCancel(ctx)

	g.watcher = &fileWatcher{
		path:   g.configPath,
		ctx:    watcherCtx,
		cancel: cancel,
	}

	// Ensure config directory exists
	if err := os.MkdirAll(filepath.Dir(g.configPath), 0755); err != nil {
		logger.Sugar().Warnf("Failed to create global config directory: %v", err)
		return nil
	}

	// Start file monitoring in a goroutine
	go g.monitorFile(logger)
	return nil
}

// monitorFile watches for file changes using stat polling
func (g *GlobalConfig) monitorFile(logger *zap.Logger) {
	if g.watcher == nil {
		return
	}

	var lastModTime time.Time
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-g.watcher.ctx.Done():
			return
		case <-ticker.C:
			if g.watchFile(&lastModTime) {
				if err := g.load(logger); err != nil {
					logger.Sugar().Warnf("Failed to reload global config: %v", err)
				}
			}
		}
	}
}

// watchFile checks if the config file has been modified
func (g *GlobalConfig) watchFile(lastModTime *time.Time) bool {
	info, err := os.Stat(g.configPath)
	if err != nil {
		return false
	}

	if info.ModTime().After(*lastModTime) {
		*lastModTime = info.ModTime()
		return true
	}

	return false
}

// Close stops the file watcher
func (g *GlobalConfig) Close() {
	if g.watcher != nil {
		// Cancel the watcher's context to stop the goroutine
		if g.watcher.cancel != nil {
			g.watcher.cancel()
		}
		// Clear the watcher
		g.watcher = nil
	}
}

func loadDefaultConfig() Config {
	return Config{
		ViewportWidth: 1440,
		UnitPrecision: 3,
	}
}

func loadConfig(root string, logger *zap.Logger) Config {
	sugar := logger.Sugar()
	defaultConfig := loadDefaultConfig()

	cssremPath := filepath.Join(root, ".cssrem")
	file, err := os.ReadFile(cssremPath)
	if err != nil {
		sugar.Warnf("Failed to open config file %s: %v", cssremPath, err)
		return defaultConfig
	}

	cssremConfig, err := parseCssremConfig(file)
	if err != nil {
		sugar.Warnf("Failed to parse config file %s: %v", cssremPath, err)
		return defaultConfig
	}

	config := convertToConfig(*cssremConfig)

	sugar.Infof("Loaded config from %s: viewport=%.0f, precision=%d",
		cssremPath, config.ViewportWidth, config.UnitPrecision)

	return config
}

func parseCssremConfig(data []byte) (*SchemaJson, error) {
	cssremConfig := SchemaJson{}
	if err := cssremConfig.UnmarshalJSON(data); err != nil {
		return nil, err
	}
	return &cssremConfig, nil
}

// mergeConfigs implements priority: default < global < project
// Returns a new Config with values from the highest priority source
func mergeConfigs(defaultConfig, globalConfig, projectConfig Config) Config {
	result := defaultConfig

	// Global config overrides defaults
	if globalConfig.ViewportWidth != 0 {
		result.ViewportWidth = globalConfig.ViewportWidth
	}
	if globalConfig.UnitPrecision != 0 {
		result.UnitPrecision = globalConfig.UnitPrecision
	}

	// Project config overrides global and defaults
	if projectConfig.ViewportWidth != 0 {
		result.ViewportWidth = projectConfig.ViewportWidth
	}
	if projectConfig.UnitPrecision != 0 {
		result.UnitPrecision = projectConfig.UnitPrecision
	}

	return result
}

// loadEffectiveConfig loads the final config with priority: default < global < project
func (h *Handler) loadEffectiveConfig(globalConfig *GlobalConfig, root string, logger *zap.Logger) Config {
	defaultConfig := loadDefaultConfig()

	var globalConfigValues Config
	if globalConfig != nil {
		globalConfigValues = *globalConfig.Get()
	}

	projectConfig := loadConfig(root, logger)

	// Merge configs with priority: default < global < project
	effectiveConfig := mergeConfigs(defaultConfig, globalConfigValues, projectConfig)

	logger.Sugar().Debugf("Effective config for %s: viewport=%.0f, precision=%d (default: %.0f/%d, global: %.0f/%d, project: %.0f/%d)",
		root, effectiveConfig.ViewportWidth, effectiveConfig.UnitPrecision,
		defaultConfig.ViewportWidth, defaultConfig.UnitPrecision,
		globalConfigValues.ViewportWidth, globalConfigValues.UnitPrecision,
		projectConfig.ViewportWidth, projectConfig.UnitPrecision)

	return effectiveConfig
}

func convertToConfig(schema SchemaJson) Config {
	return Config{
		ViewportWidth: schema.VwDesign,
		UnitPrecision: int(schema.FixedDigits),
	}
}
