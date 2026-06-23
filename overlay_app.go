package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"HearRateOverlay/internal/state"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// OverlayApp is the minimal app for the overlay window process.
type OverlayApp struct {
	ctx       context.Context
	statePath string
	store     *state.Store
}

// NewOverlayApp creates the overlay application instance.
func NewOverlayApp(statePath string) *OverlayApp {
	return &OverlayApp{
		statePath: statePath,
		store:     state.NewStore(statePath),
	}
}

func (a *OverlayApp) startup(ctx context.Context) {
	a.ctx = ctx
	runtime.LogInfo(ctx, "Overlay process started, state path: "+a.statePath)
}

func (a *OverlayApp) shutdown(ctx context.Context) {}

// GetState reads the latest shared state and returns it to the frontend.
func (a *OverlayApp) GetState() (map[string]interface{}, error) {
	s, err := a.store.Read()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"bpm":        s.BPM,
		"timestamp":  s.Timestamp,
		"connected":  s.Connected,
		"style":      s.Style,
		"styleName":  s.StyleName,
		"component":  s.Component,
		"styleCfg":   s.StyleCfg,
		"maxHR":      s.MaxHR,
		"overlayX":   s.OverlayX,
		"overlayY":   s.OverlayY,
		"overlayW":   s.OverlayW,
		"overlayH":   s.OverlayH,
	}, nil
}

// CloseOverlay closes the overlay window.
func (a *OverlayApp) CloseOverlay() {
	runtime.Quit(a.ctx)
}

// SavePosition saves the overlay window position to shared state.
func (a *OverlayApp) SavePosition(x, y int) error {
	s, err := a.store.Read()
	if err != nil {
		return err
	}
	s.OverlayX = x
	s.OverlayY = y
	return a.store.Write(s)
}

// PollState continuously reads state and emits to frontend every ~150ms.
func (a *OverlayApp) PollState() {
	go func() {
		var lastTimestamp int64
		for {
			time.Sleep(150 * time.Millisecond)
			s, err := a.store.Read()
			if err != nil {
				continue
			}
			if s.Timestamp != lastTimestamp {
				lastTimestamp = s.Timestamp
				runtime.EventsEmit(a.ctx, "state-update", map[string]interface{}{
					"bpm":       s.BPM,
					"connected": s.Connected,
					"style":     s.Style,
					"styleName": s.StyleName,
					"component": s.Component,
					"styleCfg":  s.StyleCfg,
					"maxHR":     s.MaxHR,
				})
			}
		}
	}()

	// Also report position changes
	go func() {
		var lastX, lastY int
		for {
			time.Sleep(2 * time.Second)
			x, y := runtime.WindowGetPosition(a.ctx)
			if x != lastX || y != lastY {
				lastX, lastY = x, y
				a.SavePosition(x, y)
			}
		}
	}()
}

// CloseWindow is bound to the ✕ button in overlay.
func (a *OverlayApp) CloseWindow() {
	runtime.Quit(a.ctx)
}

// ReadStateFile reads and returns the state for debugging.
func readOverlayState(path string) (*state.SharedState, error) {
	store := state.NewStore(path)
	return store.Read()
}

// WriteOverlayState writes state for debugging.
func writeOverlayState(path string, s *state.SharedState) error {
	store := state.NewStore(path)
	return store.Write(s)
}

// LaunchOverlayProcess spawns the overlay as a separate process.
func LaunchOverlayProcess(statePath string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	attr := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	proc, err := os.StartProcess(exe, []string{exe, "--overlay", statePath}, attr)
	if err != nil {
		return fmt.Errorf("start overlay: %w", err)
	}
	// Release the process (don't wait for it)
	proc.Release()
	return nil
}
