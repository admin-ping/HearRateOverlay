package i18n

import (
	"sync"
)

// Lang represents a language identifier.
type Lang string

const (
	ZH Lang = "zh"
	EN Lang = "en"
)

// Messages holds localized strings.
type Messages map[string]string

// Bundle holds translations for multiple languages.
type Bundle struct {
	mu         sync.RWMutex
	current    Lang
	translations map[Lang]Messages
}

// NewBundle creates a new i18n bundle with the default language.
func NewBundle(defaultLang Lang) *Bundle {
	b := &Bundle{
		current:      defaultLang,
		translations: make(map[Lang]Messages),
	}
	b.loadBuiltin()
	return b
}

func (b *Bundle) loadBuiltin() {
	b.translations[ZH] = Messages{
		"app.title":              "心率悬浮窗",
		"app.scanning":           "扫描中...",
		"app.scan_devices":       "扫描设备",
		"app.disconnect":         "断开",
		"app.select_device":      "-- 选择设备 --",
		"app.connect":            "连接",
		"app.status":             "状态",
		"app.styles":             "样式",
		"app.scenes":             "场景",
		"app.new_scene":          "新场景名",
		"app.create_scene":       "创建场景",
		"app.stats":              "统计",
		"app.current":            "当前",
		"app.avg_10s":            "10秒均",
		"app.avg":                "平均",
		"app.max":                "最大",
		"app.min":                "最小",
		"app.duration":           "时长",
		"app.reset_session":      "重置会话",
		"app.settings":           "设置",
		"app.max_hr":             "最大心率",
		"app.close_settings":     "关闭设置",
		"app.connected":          "已连接",
		"app.disconnected":       "已断开",
		"app.connecting":         "连接中...",
		"app.scanning_status":    "扫描中...",
		"app.idle":               "空闲",
		"app.error":              "错误",
		"zone.warmup":            "热身",
		"zone.fat_burn":          "燃脂",
		"zone.aerobic":           "有氧",
		"zone.anaerobic":         "无氧",
		"zone.maximum":           "极限",
	}

	b.translations[EN] = Messages{
		"app.title":              "HeartRateOverlay",
		"app.scanning":           "Scanning...",
		"app.scan_devices":       "Scan Devices",
		"app.disconnect":         "Disconnect",
		"app.select_device":      "-- Select Device --",
		"app.connect":            "Connect",
		"app.status":             "Status",
		"app.styles":             "Styles",
		"app.scenes":             "Scenes",
		"app.new_scene":          "New Scene",
		"app.create_scene":       "Create Scene",
		"app.stats":              "Statistics",
		"app.current":            "Current",
		"app.avg_10s":            "10s Avg",
		"app.avg":                "Average",
		"app.max":                "Maximum",
		"app.min":                "Minimum",
		"app.duration":           "Duration",
		"app.reset_session":      "Reset Session",
		"app.settings":           "Settings",
		"app.max_hr":             "Max HR",
		"app.close_settings":     "Close Settings",
		"app.connected":          "Connected",
		"app.disconnected":       "Disconnected",
		"app.connecting":         "Connecting...",
		"app.scanning_status":    "Scanning...",
		"app.idle":               "Idle",
		"app.error":              "Error",
		"zone.warmup":            "Warm Up",
		"zone.fat_burn":          "Fat Burn",
		"zone.aerobic":           "Aerobic",
		"zone.anaerobic":         "Anaerobic",
		"zone.maximum":           "Maximum",
	}
}

// T translates a key to the current language.
func (b *Bundle) T(key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if msgs, ok := b.translations[b.current]; ok {
		if val, ok := msgs[key]; ok {
			return val
		}
	}
	// Fallback to English
	if b.current != EN {
		if msgs, ok := b.translations[EN]; ok {
			if val, ok := msgs[key]; ok {
				return val
			}
		}
	}
	return key
}

// SetLang changes the current language.
func (b *Bundle) SetLang(l Lang) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = l
}

// GetLang returns the current language.
func (b *Bundle) GetLang() Lang {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.current
}
