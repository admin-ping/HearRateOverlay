package style

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"gopkg.in/yaml.v3"
)

// SafeDirName converts a style name to a filesystem-safe directory name.
func SafeDirName(name string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return re.ReplaceAllString(name, "_")
}

// Definition describes a single style from its style.yaml file.
type Definition struct {
	Name      string            `yaml:"name" json:"name"`
	Version   int               `yaml:"version" json:"version"`
	Component string            `yaml:"component" json:"component"`
	Author    string            `yaml:"author,omitempty" json:"author,omitempty"`
	Default   map[string]interface{} `yaml:"default" json:"default"`
	Path      string            `yaml:"-" json:"-"` // filesystem path to the style dir
}

// Engine manages loading and switching styles.
type Engine struct {
	mu        sync.RWMutex
	stylesDir string
	styles    map[string]*Definition
	current   string
}

// NewEngine creates a style engine that scans the given directory.
func NewEngine(stylesDir string) *Engine {
	return &Engine{
		stylesDir: stylesDir,
		styles:    make(map[string]*Definition),
	}
}

// Load scans the styles directory and loads all style definitions.
func (e *Engine) Load() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(e.stylesDir, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(e.stylesDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		stylePath := filepath.Join(e.stylesDir, entry.Name())
		yamlPath := filepath.Join(stylePath, "style.yaml")
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			continue
		}
		var def Definition
		if err := yaml.Unmarshal(data, &def); err != nil {
			continue
		}
		def.Path = stylePath
		e.styles[def.Name] = &def
	}

	return nil
}

// GetAll returns all loaded style definitions.
func (e *Engine) GetAll() []Definition {
	e.mu.RLock()
	defer e.mu.RUnlock()
	result := make([]Definition, 0, len(e.styles))
	for _, s := range e.styles {
		result = append(result, *s)
	}
	return result
}

// Get returns a style by name.
func (e *Engine) Get(name string) (*Definition, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s, ok := e.styles[name]
	return s, ok
}

// SetCurrent sets the current active style.
func (e *Engine) SetCurrent(name string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	if _, ok := e.styles[name]; ok {
		e.current = name
		return true
	}
	return false
}

// GetCurrent returns the current active style name.
func (e *Engine) GetCurrent() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.current
}

// EmbeddedStyles returns a set of built-in style definitions for first-run.
func EmbeddedStyles() []Definition {
	return []Definition{
		{
			Name:      "极简数字",
			Version:   1,
			Component: "MinimalistNumber",
			Author:    "HeartRateOverlay",
			Default: map[string]interface{}{
				"font_size":  80,
				"font_color": "#FFFFFF",
				"bg_color":   "transparent",
				"show_unit":  true,
				"animation":  "pulse",
			},
		},
		{
			Name:      "环形进度",
			Version:   1,
			Component: "RingProgress",
			Author:    "HeartRateOverlay",
			Default: map[string]interface{}{
				"ring_color":  "#00FF88",
				"ring_width":  8,
				"font_size":   48,
				"font_color":  "#FFFFFF",
				"bg_color":    "transparent",
				"show_percent": true,
			},
		},
		{
			Name:      "速度表盘",
			Version:   1,
			Component: "SpeedGauge",
			Author:    "HeartRateOverlay",
			Default: map[string]interface{}{
				"gauge_color": "#FF4444",
				"needle_color": "#FFFFFF",
				"font_size":   36,
				"font_color":  "#FFFFFF",
				"max_value":   220,
			},
		},
		{
			Name:      "心电图曲线",
			Version:   1,
			Component: "ECGWave",
			Author:    "HeartRateOverlay",
			Default: map[string]interface{}{
				"line_color":   "#00FF00",
				"line_width":   2,
				"bg_color":     "rgba(0,0,0,0.3)",
				"grid_color":   "rgba(255,255,255,0.1)",
				"history_secs": 10,
			},
		},
		{
			Name:      "LED点阵",
			Version:   1,
			Component: "LEDMatrix",
			Author:    "HeartRateOverlay",
			Default: map[string]interface{}{
				"led_color_on":  "#FF0000",
				"led_color_off": "#330000",
				"dot_size":      8,
				"font_size":     64,
				"bg_color":      "transparent",
			},
		},
		{
			Name:      "霓虹",
			Version:   1,
			Component: "NeonGlow",
			Author:    "HeartRateOverlay",
			Default: map[string]interface{}{
				"glow_color":  "#FF00FF",
				"glow_radius": 20,
				"font_size":   72,
				"font_color":  "#FFFFFF",
				"bg_color":    "transparent",
			},
		},
	}
}
