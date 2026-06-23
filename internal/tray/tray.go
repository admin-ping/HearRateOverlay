package tray

import (
	"sync"

	"github.com/getlantern/systray"
)

// Action represents a user action from the system tray menu.
type Action int

const (
	ActionShowSettings Action = iota
	ActionShowOverlay
	ActionQuit
)

// Manager wraps the system tray icon and menu.
type Manager struct {
	mu       sync.Mutex
	actions  chan Action
	ready    chan struct{}
	quit     chan struct{}
	iconData []byte
}

// New creates a new tray manager.
func New() *Manager {
	return &Manager{
		actions: make(chan Action, 10),
		ready:   make(chan struct{}),
		quit:    make(chan struct{}),
	}
}

// Start launches the system tray in a background goroutine.
// It blocks until the tray is ready.
func (m *Manager) Start() {
	go systray.Run(m.onReady, m.onExit)
	<-m.ready
}

// Actions returns the channel for receiving user actions.
func (m *Manager) Actions() <-chan Action {
	return m.actions
}

// Quit signals the tray to exit.
func (m *Manager) Quit() {
	close(m.quit)
	systray.Quit()
}

// UpdateOverlayItem updates the overlay menu item text.
func (m *Manager) UpdateOverlayItem(visible bool) {
	// Would need to store menu references to update dynamically
	// For now, the menu items are created once in onReady
	_ = visible
}

func (m *Manager) onReady() {
	systray.SetTitle("HeartRateOverlay")
	systray.SetTooltip("HeartRateOverlay - BLE Heart Rate Monitor")

	// Create menu items
	mShowSettings := systray.AddMenuItem("打开设置", "Open settings window")
	mShowOverlay := systray.AddMenuItem("显示悬浮窗", "Show heart rate overlay")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "Quit application")

	// Signal ready
	close(m.ready)

	// Handle menu clicks
	go func() {
		for {
			select {
			case <-mShowSettings.ClickedCh:
				m.actions <- ActionShowSettings
			case <-mShowOverlay.ClickedCh:
				m.actions <- ActionShowOverlay
			case <-mQuit.ClickedCh:
				m.actions <- ActionQuit
			case <-m.quit:
				return
			}
		}
	}()
}

func (m *Manager) onExit() {
	// Cleanup if needed
}
