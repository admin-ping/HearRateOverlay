package stats

import (
	"container/ring"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Zone represents a heart rate training zone.
type Zone struct {
	Name    string  `json:"name"`
	MinPct  float64 `json:"min_pct"` // % of max HR
	MaxPct  float64 `json:"max_pct"`
}

// DefaultZones returns the standard 5-zone model.
func DefaultZones() []Zone {
	return []Zone{
		{Name: "热身", MinPct: 0.50, MaxPct: 0.60},
		{Name: "燃脂", MinPct: 0.60, MaxPct: 0.70},
		{Name: "有氧", MinPct: 0.70, MaxPct: 0.80},
		{Name: "无氧", MinPct: 0.80, MaxPct: 0.90},
		{Name: "极限", MinPct: 0.90, MaxPct: 1.00},
	}
}

// Session represents a single heart rate monitoring session.
type Session struct {
	ID         string     `json:"id"`
	StartTime  time.Time  `json:"start"`
	EndTime    *time.Time `json:"end,omitempty"`
	MaxHR      uint8      `json:"max_hr"`
	MinHR      uint8      `json:"min_hr"`
	AvgHR      float64    `json:"avg_hr"`
	TotalBeats int64      `json:"total_beats"`

	mu          sync.RWMutex
	heartRates  []uint8           `json:"-"` // all recorded heart rates
	ringBuffer  *ring.Ring        `json:"-"` // last 10 seconds of data
	zoneCounts  map[string]int    `json:"-"` // counts per zone
	maxHR       int               `json:"-"` // user's max HR for zone calc
	zones       []Zone            `json:"-"`
}

// NewSession creates a new session with the given user max HR.
func NewSession(maxHR int) *Session {
	return &Session{
		ID:         uuid.New().String(),
		StartTime:  time.Now(),
		MaxHR:      0,
		MinHR:      255,
		maxHR:      maxHR,
		ringBuffer: ring.New(20), // ~10 seconds at 2 samples/sec
		zoneCounts: make(map[string]int),
		zones:      DefaultZones(),
		heartRates: make([]uint8, 0, 3600), // ~1 hour at 1 sample/sec
	}
}

// RecordHeartRate records a new heart rate value.
func (s *Session) RecordHeartRate(bpm uint8) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.heartRates = append(s.heartRates, bpm)
	s.ringBuffer.Value = bpm
	s.ringBuffer = s.ringBuffer.Next()
	s.TotalBeats++

	// Update max/min
	if bpm > s.MaxHR {
		s.MaxHR = bpm
	}
	if bpm < s.MinHR {
		s.MinHR = bpm
	}

	// Update running average
	n := float64(len(s.heartRates))
	s.AvgHR = (s.AvgHR*(n-1) + float64(bpm)) / n

	// Update zone counts
	zone := s.classifyZone(bpm)
	s.zoneCounts[zone]++
}

// classifyZone determines which training zone a BPM falls into.
func (s *Session) classifyZone(bpm uint8) string {
	if s.maxHR <= 0 {
		return "未知"
	}
	pct := float64(bpm) / float64(s.maxHR)
	for _, z := range s.zones {
		if pct >= z.MinPct && pct < z.MaxPct {
			return z.Name
		}
	}
	if pct >= 1.0 {
		return "极限"
	}
	return "未知"
}

// GetRecentAverage returns the average of the last n seconds of data.
func (s *Session) GetRecentAverage(seconds int) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	sum := 0.0
	s.ringBuffer.Do(func(v interface{}) {
		if v != nil {
			sum += float64(v.(uint8))
			count++
		}
	})

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

// GetZonePercentages returns the percentage of time spent in each zone.
func (s *Session) GetZonePercentages() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]float64)
	total := len(s.heartRates)
	if total == 0 {
		return result
	}

	for _, z := range s.zones {
		count := s.zoneCounts[z.Name]
		result[z.Name] = float64(count) / float64(total) * 100
	}
	return result
}

// End marks the session as ended.
func (s *Session) End() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EndTime = &now
}

// Duration returns the session duration.
func (s *Session) Duration() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.EndTime != nil {
		return s.EndTime.Sub(s.StartTime)
	}
	return time.Since(s.StartTime)
}

// GetHeartRates returns a copy of all heart rate values.
func (s *Session) GetHeartRates() []uint8 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]uint8, len(s.heartRates))
	copy(result, s.heartRates)
	return result
}

// SessionSummary holds summary info for a session.
type SessionSummary struct {
	ID        string     `json:"id"`
	StartTime time.Time  `json:"start"`
	EndTime   *time.Time `json:"end,omitempty"`
	Duration  string     `json:"duration"`
	AvgHR     float64    `json:"avg_hr"`
	MaxHR     uint8      `json:"max_hr"`
	MinHR     uint8      `json:"min_hr"`
}

// Summary returns a summary of the session.
func (s *Session) Summary() SessionSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SessionSummary{
		ID:        s.ID,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
		Duration:  s.Duration().Round(time.Second).String(),
		AvgHR:     math.Round(s.AvgHR*10) / 10,
		MaxHR:     s.MaxHR,
		MinHR:     s.MinHR,
	}
}
