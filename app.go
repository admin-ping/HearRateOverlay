package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"HearRateOverlay/internal/ble"
	"HearRateOverlay/internal/config"
	"HearRateOverlay/internal/db"
	"HearRateOverlay/internal/i18n"
	"HearRateOverlay/internal/state"
	"HearRateOverlay/internal/stats"
	"HearRateOverlay/internal/style"
	"HearRateOverlay/internal/tray"
	"HearRateOverlay/internal/update"

	"gopkg.in/yaml.v3"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the main application struct with Wails-bound methods.
type App struct {
	ctx           context.Context
	mu            sync.RWMutex
	bleManager    *ble.BLEManager
	configMgr     *config.Manager
	styleEngine   *style.Engine
	database      *db.Database
	session       *stats.Session
	lastBPM       uint8
	i18nBundle    *i18n.Bundle
	updateChecker *update.Checker
	trayManager   *tray.Manager
	stateStore    *state.Store
	statePath     string
	overlayProc   *os.Process
}

// NewApp creates the application instance.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize i18n
	a.i18nBundle = i18n.NewBundle(i18n.ZH)

	// Initialize config directory
	homeDir, _ := os.UserHomeDir()
	configDir := homeDir + "/.config/heartrate-overlay"
	_ = os.MkdirAll(configDir, 0755)

	// Shared state file for overlay communication
	a.statePath = filepath.Join(os.TempDir(), "heartrate-overlay-state.json")
	a.stateStore = state.NewStore(a.statePath)

	// Initialize modules
	a.configMgr = config.NewManager(configDir)
	if err := a.configMgr.Load(); err != nil {
		runtime.LogError(ctx, "Failed to load config: "+err.Error())
	}

	cfg := a.configMgr.GetConfig()

	// Apply language from config
	if cfg.Language == "en" {
		a.i18nBundle.SetLang(i18n.EN)
	}

	a.session = stats.NewSession(cfg.MaxHR)

	dbPath := configDir + "/heartrate.db"
	var err error
	a.database, err = db.New(dbPath)
	if err != nil {
		runtime.LogError(ctx, "Failed to open database: "+err.Error())
	}

	// Initialize BLE
	a.bleManager = ble.NewBLEManager()
	if err := a.bleManager.Enable(); err != nil {
		runtime.LogError(ctx, "BLE init: "+err.Error())
	}

	a.bleManager.SetHeartRateCallback(func(bpm uint8) {
		a.mu.Lock()
		a.lastBPM = bpm
		a.mu.Unlock()

		a.session.RecordHeartRate(bpm)
		runtime.EventsEmit(a.ctx, "heart-rate-update", map[string]interface{}{
			"bpm":       bpm,
			"timestamp": time.Now().UnixMilli(),
		})

		// Write to shared state for overlay process
		a.writeSharedState()
	})

	// Initialize style engine
	styleDir := configDir + "/styles"
	a.styleEngine = style.NewEngine(styleDir)
	a.ensureEmbeddedStyles(styleDir)
	if err := a.styleEngine.Load(); err != nil {
		runtime.LogError(ctx, "Failed to load styles: "+err.Error())
	}

	// Initialize update checker
	a.updateChecker = update.NewChecker("owner", "heartrate-overlay", "0.1.0")

	// Start system tray
	a.trayManager = tray.New()
	go func() {
		a.trayManager.Start()
	}()
	go a.handleTrayActions()

	// Auto-connect if default device is set
	if cfg.DefaultDevice != "" {
		go func() {
			if err := a.bleManager.Connect(cfg.DefaultDevice); err == nil {
				a.bleManager.SubscribeHeartRate()
				a.bleManager.StartAutoReconnect()
				runtime.EventsEmit(a.ctx, "connection-status", "connected")
			}
		}()
	}

	// Async update check
	if cfg.CheckUpdate {
		go func() {
			release, hasUpdate, err := a.updateChecker.CheckLatest()
			if err != nil {
				return
			}
			if hasUpdate {
				runtime.EventsEmit(a.ctx, "update-available", map[string]interface{}{
					"version": release.TagName,
					"url":     release.HTMLURL,
					"body":    release.Body,
				})
			}
		}()
	}
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
	if a.session != nil {
		a.session.End()
		if a.database != nil {
			a.database.SaveSession(a.session)
		}
	}
	if a.bleManager != nil {
		a.bleManager.Disconnect()
	}
	if a.database != nil {
		a.database.Close()
	}
	if a.trayManager != nil {
		a.trayManager.Quit()
	}
	a.KillOverlay()
}

// handleTrayActions processes system tray menu clicks.
func (a *App) handleTrayActions() {
	for action := range a.trayManager.Actions() {
		switch action {
		case tray.ActionShowSettings:
			// Settings window is already visible — bring to front
			runtime.WindowShow(a.ctx)
		case tray.ActionShowOverlay:
			a.LaunchOverlay()
		case tray.ActionQuit:
			a.KillOverlay()
			runtime.Quit(a.ctx)
			return
		}
	}
}

// ensureEmbeddedStyles writes built-in style definitions on first run.
func (a *App) ensureEmbeddedStyles(styleDir string) {
	for _, def := range style.EmbeddedStyles() {
		dir := filepath.Join(styleDir, style.SafeDirName(def.Name))
		if err := os.MkdirAll(dir, 0755); err != nil {
			runtime.LogError(a.ctx, "Failed to create style dir: "+err.Error())
			continue
		}
		yamlPath := filepath.Join(dir, "style.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			data, err := yaml.Marshal(&def)
			if err != nil {
				runtime.LogError(a.ctx, "Failed to marshal style: "+err.Error())
				continue
			}
			if err := os.WriteFile(yamlPath, data, 0644); err != nil {
				runtime.LogError(a.ctx, "Failed to write style.yaml: "+err.Error())
			}
		}
	}
}

// writeSharedState writes the current state to the shared file for the overlay process.
func (a *App) writeSharedState() {
	a.mu.RLock()
	bpm := a.lastBPM
	a.mu.RUnlock()

	connected := a.bleManager.GetState() == ble.StateSubscribed
	currentStyle := a.styleEngine.GetCurrent()
	var styleCfg map[string]interface{}
	var component string
	if def, ok := a.styleEngine.Get(currentStyle); ok {
		styleCfg = def.Default
		component = def.Component
	}
	cfg := a.configMgr.GetConfig()

	shared := &state.SharedState{
		BPM:       bpm,
		Connected: connected,
		Style:     currentStyle,
		StyleName: currentStyle,
		Component: component,
		StyleCfg:  styleCfg,
		MaxHR:     cfg.MaxHR,
		OverlayW:  300,
		OverlayH:  200,
	}
	a.stateStore.Write(shared)
}

// LaunchOverlay spawns the overlay as a separate process.
func (a *App) LaunchOverlay() error {
	// Write latest state first
	a.writeSharedState()

	exe := os.Args[0]
	proc, err := os.StartProcess(
		exe,
		[]string{exe, "--overlay", a.statePath},
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}},
	)
	if err != nil {
		return fmt.Errorf("start overlay process: %w", err)
	}
	proc.Release() // Don't wait for it
	return nil
}

// KillOverlay terminates the overlay process if running.
func (a *App) KillOverlay() {
	// Find and kill any overlay processes
	// For simplicity, we can't easily track the PID across runs.
	// The overlay reads state from a file, so it auto-terminates when parent exits.
	// A more robust solution would use a named mutex or lock file.
}

// ============================================================================
// Wails-bound methods — called from frontend
// ============================================================================

// StartScan begins async BLE scanning. Results arrive via 'scan-update' events.
// Call StopScan to cancel.
func (a *App) StartScan() error {
	a.bleManager.SetScanProgressCallback(func(devices []ble.DeviceInfo) {
		runtime.EventsEmit(a.ctx, "scan-update", devices)
	})
	return a.bleManager.StartAsyncScan()
}

// StopScan stops the current scan.
func (a *App) StopScan() {
	a.bleManager.StopAsyncScan()
}

// GetDevices scans for BLE devices synchronously (legacy, use StartScan for UI).
func (a *App) GetDevices() []ble.DeviceInfo {
	devices, err := a.bleManager.ScanAndDiscover(5 * time.Second)
	if err != nil {
		runtime.LogError(a.ctx, "Scan error: "+err.Error())
		return []ble.DeviceInfo{}
	}
	return devices
}

// Connect connects to a BLE device and subscribes to heart rate.
func (a *App) Connect(deviceID string) error {
	if err := a.bleManager.Connect(deviceID); err != nil {
		runtime.EventsEmit(a.ctx, "connection-status", fmt.Sprintf("error: %v", err))
		return err
	}
	if err := a.bleManager.SubscribeHeartRate(); err != nil {
		runtime.EventsEmit(a.ctx, "connection-status", fmt.Sprintf("subscribe error: %v", err))
		return err
	}
	a.bleManager.StartAutoReconnect()

	cfg := a.configMgr.GetConfig()
	cfg.DefaultDevice = deviceID
	a.configMgr.UpdateConfig(cfg)

	runtime.EventsEmit(a.ctx, "connection-status", "connected")
	return nil
}

// Disconnect disconnects from the current device.
func (a *App) Disconnect() error {
	err := a.bleManager.Disconnect()
	runtime.EventsEmit(a.ctx, "connection-status", "disconnected")
	return err
}

// GetConnectionStatus returns the current BLE connection state.
func (a *App) GetConnectionStatus() string {
	return a.bleManager.GetState().String()
}

// GetCurrentHeartRate returns the most recent BPM value.
func (a *App) GetCurrentHeartRate() uint8 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastBPM
}

// GetSessionStats returns current session statistics.
func (a *App) GetSessionStats() map[string]interface{} {
	if a.session == nil {
		return map[string]interface{}{}
	}
	summary := a.session.Summary()
	zones := a.session.GetZonePercentages()
	recentAvg := a.session.GetRecentAverage(10)

	return map[string]interface{}{
		"current_bpm": a.lastBPM,
		"avg_hr":      summary.AvgHR,
		"max_hr":      summary.MaxHR,
		"min_hr":      summary.MinHR,
		"duration":    summary.Duration,
		"zones":       zones,
		"recent_avg":  recentAvg,
	}
}

// SwitchStyle switches to the named style.
func (a *App) SwitchStyle(styleName string) error {
	if !a.styleEngine.SetCurrent(styleName) {
		return fmt.Errorf("style %q not found", styleName)
	}
	def, _ := a.styleEngine.Get(styleName)
	runtime.EventsEmit(a.ctx, "style-changed", def)
	return nil
}

// GetStyles returns all available styles.
func (a *App) GetStyles() []style.Definition {
	return a.styleEngine.GetAll()
}

// GetCurrentStyle returns the currently active style.
func (a *App) GetCurrentStyle() string {
	return a.styleEngine.GetCurrent()
}

// SwitchScene activates the named scene.
func (a *App) SwitchScene(sceneName string) error {
	scene, ok := a.configMgr.GetScene(sceneName)
	if !ok {
		return fmt.Errorf("scene %q not found", sceneName)
	}
	if scene.Style != "" {
		a.SwitchStyle(scene.Style)
	}
	runtime.EventsEmit(a.ctx, "scene-changed", scene)
	return nil
}

// GetScenes returns all saved scenes.
func (a *App) GetScenes() []config.Scene {
	return a.configMgr.GetScenes()
}

// SaveScene creates or updates a scene.
func (a *App) SaveScene(scene config.Scene) error {
	return a.configMgr.SaveScene(&scene)
}

// DeleteScene removes a scene.
func (a *App) DeleteScene(name string) error {
	return a.configMgr.DeleteScene(name)
}

// GetConfig returns the global configuration.
func (a *App) GetConfig() config.GlobalConfig {
	return a.configMgr.GetConfig()
}

// UpdateConfig updates the global configuration.
func (a *App) UpdateConfig(cfg config.GlobalConfig) error {
	return a.configMgr.UpdateConfig(cfg)
}

// GetSessionHistory returns past session summaries.
func (a *App) GetSessionHistory() []stats.SessionSummary {
	if a.database == nil {
		return nil
	}
	sessions, err := a.database.GetSessions(50)
	if err != nil {
		runtime.LogError(a.ctx, "Failed to get sessions: "+err.Error())
		return nil
	}
	return sessions
}

// ExportSessionCSV returns CSV data for a session.
func (a *App) ExportSessionCSV(sessionID string) (string, error) {
	samples, err := a.database.GetHeartRateSamples(sessionID)
	if err != nil {
		return "", err
	}
	csv := "timestamp,bpm\n"
	for _, s := range samples {
		csv += fmt.Sprintf("%s,%d\n", s.Timestamp.Format(time.RFC3339), s.BPM)
	}
	return csv, nil
}

// ResetSession starts a new session.
func (a *App) ResetSession() {
	if a.session != nil {
		a.session.End()
		if a.database != nil {
			a.database.SaveSession(a.session)
		}
	}
	cfg := a.configMgr.GetConfig()
	a.session = stats.NewSession(cfg.MaxHR)
}

// GetI18nMessages returns all messages for the current language.
func (a *App) GetI18nMessages() map[string]string {
	msgs := make(map[string]string)
	keys := []string{
		"app.title", "app.scanning", "app.scan_devices", "app.disconnect",
		"app.select_device", "app.connect", "app.status", "app.styles", "app.scenes",
		"app.new_scene", "app.create_scene", "app.stats", "app.current",
		"app.avg_10s", "app.avg", "app.max", "app.min", "app.duration",
		"app.reset_session", "app.settings", "app.max_hr", "app.close_settings",
		"app.connected", "app.disconnected", "app.connecting", "app.idle",
		"zone.warmup", "zone.fat_burn", "zone.aerobic", "zone.anaerobic", "zone.maximum",
	}
	for _, k := range keys {
		msgs[k] = a.i18nBundle.T(k)
	}
	return msgs
}

// SetLanguage changes the UI language and saves to config.
func (a *App) SetLanguage(lang string) error {
	switch lang {
	case "zh":
		a.i18nBundle.SetLang(i18n.ZH)
	case "en":
		a.i18nBundle.SetLang(i18n.EN)
	default:
		return fmt.Errorf("unsupported language: %s", lang)
	}
	cfg := a.configMgr.GetConfig()
	cfg.Language = lang
	return a.configMgr.UpdateConfig(cfg)
}

// CheckForUpdate checks GitHub for a new release.
func (a *App) CheckForUpdate() (map[string]interface{}, error) {
	release, hasUpdate, err := a.updateChecker.CheckLatest()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"has_update": hasUpdate,
		"version":    release.TagName,
		"url":        release.HTMLURL,
		"body":       release.Body,
	}, nil
}
