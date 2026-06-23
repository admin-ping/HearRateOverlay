package ble

import (
	"errors"
)

// ParseHeartRate parses a raw heart rate measurement value according to
// the Bluetooth Heart Rate Service specification.
//
// The first byte is a flags field:
//   - Bit 0: Heart Rate Value Format (0 = uint8, 1 = uint16)
//   - Bit 1: Sensor Contact Status
//   - Bit 2: Sensor Contact Support
//   - Bit 3: Energy Expended Status
//   - Bit 4: RR-Interval
//
// Returns the heart rate in beats per minute (BPM).
func ParseHeartRate(data []byte) (uint8, error) {
	if len(data) < 2 {
		return 0, errors.New("heart rate data too short (min 2 bytes)")
	}

	flags := data[0]
	offset := 1

	var bpm uint16

	// Check heart rate value format
	if flags&0x01 != 0 {
		// 16-bit heart rate value
		if len(data) < offset+2 {
			return 0, errors.New("heart rate data too short for 16-bit value")
		}
		bpm = uint16(data[offset]) | (uint16(data[offset+1]) << 8)
		offset += 2
	} else {
		// 8-bit heart rate value
		bpm = uint16(data[offset])
		offset++
	}

	return uint8(bpm), nil
}

// ParseHeartRateWithContact parses heart rate data and also returns
// the sensor contact status if available.
func ParseHeartRateWithContact(data []byte) (bpm uint8, contactSupported bool, contactDetected bool, err error) {
	if len(data) < 2 {
		return 0, false, false, errors.New("heart rate data too short")
	}

	flags := data[0]

	// Check contact support (bit 2) and contact status (bit 1)
	contactSupported = (flags & 0x04) != 0
	if contactSupported {
		contactDetected = (flags & 0x02) != 0
	}

	bpm, err = ParseHeartRate(data)
	return bpm, contactSupported, contactDetected, err
}
