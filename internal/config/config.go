package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// GlobalConfig holds application-wide settings.
type GlobalConfig struct {
	DefaultDevice string `yaml:"default_device" json:"default_device"`
	MaxHR         int    `yaml:"max_hr" json:"max_hr"`
	Language      string `yaml:"language" json:"language"`
	AutoStart     bool   `yaml:"auto_start" json:"auto_start"`
	CheckUpdate   bool   `yaml:"check_update" json:"check_update"`
}

// Scene holds window position, size, and style settings for a scene.
type Scene struct {
	Name        string       `yaml:"name" json:"name"`
	Window      WindowConfig `yaml:"window" json:"window"`
	Style       string       `yaml:"style" json:"style"`
	FontSize    int          `yaml:"font_size" json:"font_size"`
}

// WindowConfig holds window geometry and display settings.
type WindowConfig struct {
	X           int     `yaml:"x" json:"x"`
	Y           int     `yaml:"y" json:"y"`
	Width       int     `yaml:"width" json:"width"`
	Height      int     `yaml:"height" json:"height"`
	AlwaysOnTop bool    `yaml:"always_on_top" json:"always_on_top"`
	Opacity     float64 `yaml:"opacity" json:"opacity"`
}

// Manager handles config and scene persistence.
type Manager struct {
	mu         sync.RWMutex
	configDir  string
	configPath string
	scenesDir  string
	config     GlobalConfig
	scenes     map[string]*Scene
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() GlobalConfig {
	return GlobalConfig{
		MaxHR:       185,
		Language:    "zh",
		AutoStart:   false,
		CheckUpdate: true,
	}
}

// DefaultScene returns a default scene configuration.
func DefaultScene() *Scene {
	return &Scene{
		Name:   "默认",
		Style:  "minimalist",
		FontSize: 80,
		Window: WindowConfig{
			Width:       300,
			Height:      200,
			AlwaysOnTop: true,
			Opacity:     1.0,
		},
	}
}

// NewManager creates a config manager. configDir is typically ~/.config/heartrate-overlay/.
func NewManager(configDir string) *Manager {
	return &Manager{
		configDir: configDir,
		configPath: filepath.Join(configDir, "config.yaml"),
		scenesDir:  filepath.Join(configDir, "scenes"),
		scenes:     make(map[string]*Scene),
	}
}

// Load reads global config and all scene files from disk.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure directories exist
	if err := os.MkdirAll(m.scenesDir, 0755); err != nil {
		return err
	}

	// Load global config
	m.config = DefaultConfig()
	if data, err := os.ReadFile(m.configPath); err == nil {
		if err := yaml.Unmarshal(data, &m.config); err != nil {
			return err
		}
	}

	// Load scenes
	entries, err := os.ReadDir(m.scenesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(m.scenesDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var scene Scene
		if err := yaml.Unmarshal(data, &scene); err != nil {
			continue
		}
		m.scenes[scene.Name] = &scene
	}

	// Create default scene if none exist
	if len(m.scenes) == 0 {
		def := DefaultScene()
		m.scenes[def.Name] = def
		if err := m.saveScene(def); err != nil {
			return err
		}
	}

	return nil
}

// SaveConfig writes the global config to disk.
func (m *Manager) SaveConfig() error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, err := yaml.Marshal(&m.config)
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}

// GetConfig returns a copy of the global config.
func (m *Manager) GetConfig() GlobalConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the global config and saves to disk.
func (m *Manager) UpdateConfig(updates GlobalConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = updates
	data, err := yaml.Marshal(&m.config)
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}

// GetScenes returns all scenes.
func (m *Manager) GetScenes() []Scene {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Scene, 0, len(m.scenes))
	for _, s := range m.scenes {
		result = append(result, *s)
	}
	return result
}

// GetScene returns a scene by name.
func (m *Manager) GetScene(name string) (*Scene, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.scenes[name]
	if !ok {
		return nil, false
	}
	return s, true
}

// SaveScene creates or updates a scene.
func (m *Manager) SaveScene(scene *Scene) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveScene(scene)
}

func (m *Manager) saveScene(scene *Scene) error {
	m.scenes[scene.Name] = scene
	data, err := yaml.Marshal(scene)
	if err != nil {
		return err
	}
	path := filepath.Join(m.scenesDir, scene.Name+".yaml")
	return os.WriteFile(path, data, 0644)
}

// DeleteScene removes a scene by name.
func (m *Manager) DeleteScene(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.scenes, name)
	path := filepath.Join(m.scenesDir, name+".yaml")
	return os.Remove(path)
}

// RenameScene renames a scene.
func (m *Manager) RenameScene(oldName, newName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	scene, ok := m.scenes[oldName]
	if !ok {
		return nil
	}
	delete(m.scenes, oldName)
	scene.Name = newName
	m.scenes[newName] = scene

	oldPath := filepath.Join(m.scenesDir, oldName+".yaml")
	os.Remove(oldPath)

	data, err := yaml.Marshal(scene)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(m.scenesDir, newName+".yaml"), data, 0644)
}
