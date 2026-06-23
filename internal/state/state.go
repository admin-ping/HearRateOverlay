package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// SharedState is the data shared between settings and overlay processes.
type SharedState struct {
	BPM        uint8                 `json:"bpm"`
	Timestamp  int64                 `json:"timestamp"`
	Connected  bool                  `json:"connected"`
	Style      string                 `json:"style"`
	StyleName  string                 `json:"style_name"`
	Component  string                 `json:"component"`
	StyleCfg   map[string]interface{} `json:"style_cfg"`
	MaxHR      int                    `json:"max_hr"`
	OverlayX   int                    `json:"overlay_x"`
	OverlayY   int                    `json:"overlay_y"`
	OverlayW   int                    `json:"overlay_w"`
	OverlayH   int                    `json:"overlay_h"`
}

// Store manages reading/writing the shared state file.
type Store struct {
	mu   sync.Mutex
	path string
}

// NewStore creates a state store using the given file path.
func NewStore(path string) *Store {
	return &Store{path: path}
}

// Write atomically writes the shared state to disk.
func (s *Store) Write(state *SharedState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	state.Timestamp = time.Now().UnixMilli()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	// Write to temp file then rename for atomicity
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Read reads the shared state from disk.
func (s *Store) Read() (*SharedState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var state SharedState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
