package ble

import (
	"testing"
)

func TestParseHeartRate8Bit(t *testing.T) {
	// Standard 8-bit heart rate: flags=0x00, bpm=72
	data := []byte{0x00, 72}
	bpm, err := ParseHeartRate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bpm != 72 {
		t.Errorf("expected 72, got %d", bpm)
	}
}

func TestParseHeartRate8BitWithContact(t *testing.T) {
	// Flags: bit1=1 (contact detected), bit2=1 (contact supported), bit0=0 (8-bit)
	// flags = 0x06, bpm = 85
	data := []byte{0x06, 85}
	bpm, contactSupported, contactDetected, err := ParseHeartRateWithContact(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bpm != 85 {
		t.Errorf("expected 85, got %d", bpm)
	}
	if !contactSupported {
		t.Error("expected contact supported")
	}
	if !contactDetected {
		t.Error("expected contact detected")
	}
}

func TestParseHeartRate16Bit(t *testing.T) {
	// 16-bit heart rate: flags=0x01, bpm=200 (0x00C8)
	data := []byte{0x01, 0xC8, 0x00}
	bpm, err := ParseHeartRate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bpm != 200 {
		t.Errorf("expected 200, got %d", bpm)
	}
}

func TestParseHeartRate16BitLarge(t *testing.T) {
	// 16-bit heart rate: flags=0x01, bpm=255 (0x00FF) - max for uint8
	data := []byte{0x01, 0xFF, 0x00}
	bpm, err := ParseHeartRate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bpm != 255 {
		t.Errorf("expected 255, got %d", bpm)
	}
}

func TestParseHeartRateTooShort(t *testing.T) {
	_, err := ParseHeartRate([]byte{0x00})
	if err == nil {
		t.Error("expected error for too-short data")
	}
}

func TestParseHeartRateEmpty(t *testing.T) {
	_, err := ParseHeartRate([]byte{})
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestParseHeartRateMaxUint8(t *testing.T) {
	// 8-bit max: flags=0x00, bpm=255
	data := []byte{0x00, 0xFF}
	bpm, err := ParseHeartRate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bpm != 255 {
		t.Errorf("expected 255, got %d", bpm)
	}
}
