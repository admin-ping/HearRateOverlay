package stats

import (
	"testing"
)

func TestNewSession(t *testing.T) {
	s := NewSession(185)
	if s.ID == "" {
		t.Error("expected non-empty session ID")
	}
	if s.MaxHR != 0 {
		t.Error("expected initial MaxHR to be 0")
	}
	if s.MinHR != 255 {
		t.Errorf("expected initial MinHR to be 255, got %d", s.MinHR)
	}
}

func TestRecordHeartRate(t *testing.T) {
	s := NewSession(185)

	s.RecordHeartRate(72)
	s.RecordHeartRate(80)
	s.RecordHeartRate(75)

	if s.TotalBeats != 3 {
		t.Errorf("expected 3 beats, got %d", s.TotalBeats)
	}
	if s.MaxHR != 80 {
		t.Errorf("expected max 80, got %d", s.MaxHR)
	}
	if s.MinHR != 72 {
		t.Errorf("expected min 72, got %d", s.MinHR)
	}

	// Average: (72+80+75)/3 = 75.666...
	expected := (72.0 + 80.0 + 75.0) / 3.0
	if s.AvgHR < expected-0.1 || s.AvgHR > expected+0.1 {
		t.Errorf("expected avg ~%.1f, got %.1f", expected, s.AvgHR)
	}
}

func TestZoneClassification(t *testing.T) {
	s := NewSession(200)

	tests := []struct {
		bpm  uint8
		zone string
	}{
		{100, "热身"},  // 50-60%
		{130, "燃脂"},  // 60-70%
		{150, "有氧"},  // 70-80%
		{170, "无氧"},  // 80-90%
		{190, "极限"},  // 90-100%
		{210, "极限"},  // >100%
	}

	for _, tt := range tests {
		zone := s.classifyZone(tt.bpm)
		if zone != tt.zone {
			t.Errorf("for bpm %d, expected zone %q, got %q", tt.bpm, tt.zone, zone)
		}
	}
}

func TestGetZonePercentages(t *testing.T) {
	s := NewSession(200)

	// Record 10 samples: 5 in 热身, 3 in 燃脂, 2 in 有氧
	for i := 0; i < 5; i++ {
		s.RecordHeartRate(110) // 热身 55%
	}
	for i := 0; i < 3; i++ {
		s.RecordHeartRate(130) // 燃脂 65%
	}
	for i := 0; i < 2; i++ {
		s.RecordHeartRate(150) // 有氧 75%
	}

	zones := s.GetZonePercentages()
	if zones["热身"] < 49 || zones["热身"] > 51 {
		t.Errorf("expected 热身 ~50%%, got %.1f%%", zones["热身"])
	}
	if zones["燃脂"] < 29 || zones["燃脂"] > 31 {
		t.Errorf("expected 燃脂 ~30%%, got %.1f%%", zones["燃脂"])
	}
	if zones["有氧"] < 19 || zones["有氧"] > 21 {
		t.Errorf("expected 有氧 ~20%%, got %.1f%%", zones["有氧"])
	}
}

func TestGetRecentAverage(t *testing.T) {
	s := NewSession(185)

	for i := 0; i < 20; i++ {
		s.RecordHeartRate(100)
	}

	avg := s.GetRecentAverage(10)
	if avg < 99 || avg > 101 {
		t.Errorf("expected recent avg ~100, got %.1f", avg)
	}
}

func TestSessionSummary(t *testing.T) {
	s := NewSession(185)
	s.RecordHeartRate(80)
	s.RecordHeartRate(90)
	s.End()

	summary := s.Summary()
	if summary.AvgHR < 84.5 || summary.AvgHR > 85.5 {
		t.Errorf("expected avg ~85, got %.1f", summary.AvgHR)
	}
	if summary.MaxHR != 90 {
		t.Errorf("expected max 90, got %d", summary.MaxHR)
	}
	if summary.MinHR != 80 {
		t.Errorf("expected min 80, got %d", summary.MinHR)
	}
}
