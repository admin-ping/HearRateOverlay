package ble

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"tinygo.org/x/bluetooth"
)

var (
	HeartRateServiceUUID     = bluetooth.ServiceUUIDHeartRate
	HeartRateMeasurementUUID = bluetooth.CharacteristicUUIDHeartRateMeasurement
)

// State represents the BLE connection state.
type State int32

const (
	StateIdle         State = iota
	StateScanning
	StateConnecting
	StateConnected
	StateSubscribed
	StateDisconnected
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateScanning:
		return "scanning"
	case StateConnecting:
		return "connecting"
	case StateConnected:
		return "connected"
	case StateSubscribed:
		return "subscribed"
	case StateDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

// DeviceInfo holds basic info about a discovered BLE device.
type DeviceInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	RSSI    int16  `json:"rssi"`
}

// ScanProgressCallback is called for each discovered device during scanning.
type ScanProgressCallback func(devices []DeviceInfo)

// HeartRateCallback is called when a new heart rate value is received.
type HeartRateCallback func(bpm uint8)

// BLEManager handles BLE scanning, connection, and heart rate subscription.
type BLEManager struct {
	adapter         *bluetooth.Adapter
	device          *bluetooth.Device
	state           atomic.Int32
	mu              sync.RWMutex
	onHeartRate     HeartRateCallback
	onScanProgress  ScanProgressCallback
	reconnectCancel context.CancelFunc
	scanStop        chan struct{}
	lastAddress     string
	lastDeviceName  string
}

// NewBLEManager creates a new BLE manager.
func NewBLEManager() *BLEManager {
	return &BLEManager{}
}

// Enable initializes the BLE adapter.
func (m *BLEManager) Enable() error {
	adapter := bluetooth.DefaultAdapter
	if adapter == nil {
		return errors.New("no BLE adapter found")
	}
	err := adapter.Enable()
	if err != nil {
		return err
	}
	m.adapter = adapter
	m.setState(StateIdle)
	return nil
}

// SetHeartRateCallback sets the callback for heart rate updates.
func (m *BLEManager) SetHeartRateCallback(cb HeartRateCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onHeartRate = cb
}

// GetState returns the current BLE state.
func (m *BLEManager) GetState() State {
	return State(m.state.Load())
}

func (m *BLEManager) setState(s State) {
	m.state.Store(int32(s))
}

// SetScanProgressCallback sets the callback for scan progress updates.
func (m *BLEManager) SetScanProgressCallback(cb ScanProgressCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onScanProgress = cb
}

// ScanAndDiscover scans for BLE devices advertising the Heart Rate service.
// It scans for the specified duration and returns discovered devices.
func (m *BLEManager) ScanAndDiscover(duration time.Duration) ([]DeviceInfo, error) {
	if m.adapter == nil {
		return nil, errors.New("adapter not initialized, call Enable first")
	}

	m.setState(StateScanning)

	var devices []DeviceInfo
	var mu sync.Mutex
	seen := make(map[string]bool)

	// Start scan
	err := m.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		mu.Lock()
		defer mu.Unlock()

		addr := result.Address.MAC.String()
		if seen[addr] {
			return
		}

		name := result.LocalName()

		// Include devices with HR service or with a name
		if result.HasServiceUUID(HeartRateServiceUUID) || name != "" {
			seen[addr] = true
			devices = append(devices, DeviceInfo{
				Name:    name,
				Address: addr,
				RSSI:    result.RSSI,
			})
		}
	})

	if err != nil {
		m.setState(StateIdle)
		return devices, err
	}

	// Wait for scan duration, then stop
	time.Sleep(duration)
	m.adapter.StopScan()

	// Small delay to let pending callbacks finish
	time.Sleep(100 * time.Millisecond)

	m.setState(StateIdle)
	return devices, nil
}

// StartAsyncScan begins scanning in a goroutine and returns immediately.
// Results are delivered via the ScanProgressCallback. Call StopAsyncScan to stop.
func (m *BLEManager) StartAsyncScan() error {
	if m.adapter == nil {
		return errors.New("adapter not initialized, call Enable first")
	}

	m.mu.Lock()
	m.scanStop = make(chan struct{})
	m.mu.Unlock()

	m.setState(StateScanning)

	go func() {
		var devices []DeviceInfo
		var mu sync.Mutex
		seen := make(map[string]bool)
		lastFlush := time.Now()

		err := m.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			mu.Lock()
			defer mu.Unlock()

			addr := result.Address.MAC.String()
			if seen[addr] {
				return
			}

			name := result.LocalName()
			if result.HasServiceUUID(HeartRateServiceUUID) || name != "" {
				seen[addr] = true
				devices = append(devices, DeviceInfo{
					Name:    name,
					Address: addr,
					RSSI:    result.RSSI,
				})

				// Flush to callback every 500ms
				if time.Since(lastFlush) > 500*time.Millisecond {
					m.mu.RLock()
					cb := m.onScanProgress
					m.mu.RUnlock()
					if cb != nil {
						snapshot := make([]DeviceInfo, len(devices))
						copy(snapshot, devices)
						cb(snapshot)
					}
					lastFlush = time.Now()
				}
			}
		})

		// Final flush
		m.mu.RLock()
		cb := m.onScanProgress
		m.mu.RUnlock()
		if cb != nil {
			cb(devices)
		}

		if err != nil {
			m.setState(StateIdle)
		} else {
			m.setState(StateIdle)
		}
	}()

	return nil
}

// StopAsyncScan stops the ongoing async scan.
func (m *BLEManager) StopAsyncScan() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.scanStop != nil {
		close(m.scanStop)
		m.scanStop = nil
	}
	m.adapter.StopScan()
	m.setState(StateIdle)
}

// Connect connects to a BLE device by its MAC address string.
func (m *BLEManager) Connect(deviceID string) error {
	if m.adapter == nil {
		return errors.New("adapter not initialized, call Enable first")
	}

	m.setState(StateConnecting)

	// Parse address string to MAC
	mac, err := bluetooth.ParseMAC(deviceID)
	if err != nil {
		m.setState(StateIdle)
		return err
	}

	addr := bluetooth.Address{
		MACAddress: bluetooth.MACAddress{
			MAC: mac,
		},
	}

	device, err := m.adapter.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		m.setState(StateIdle)
		return err
	}

	m.mu.Lock()
	m.device = &device
	m.lastAddress = deviceID
	m.mu.Unlock()

	m.setState(StateConnected)
	return nil
}

// SubscribeHeartRate discovers the heart rate service and subscribes to notifications.
func (m *BLEManager) SubscribeHeartRate() error {
	m.mu.RLock()
	device := m.device
	m.mu.RUnlock()

	if device == nil {
		return errors.New("not connected to a device")
	}

	// Discover heart rate service
	services, err := device.DiscoverServices([]bluetooth.UUID{HeartRateServiceUUID})
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return errors.New("heart rate service not found on device")
	}

	// Discover heart rate measurement characteristic
	chars, err := services[0].DiscoverCharacteristics([]bluetooth.UUID{HeartRateMeasurementUUID})
	if err != nil {
		return err
	}
	if len(chars) == 0 {
		return errors.New("heart rate measurement characteristic not found")
	}

	// Subscribe to notifications
	err = chars[0].EnableNotifications(func(buf []byte) {
		bpm, err := ParseHeartRate(buf)
		if err != nil {
			return
		}
		m.mu.RLock()
		cb := m.onHeartRate
		m.mu.RUnlock()
		if cb != nil {
			cb(bpm)
		}
	})
	if err != nil {
		return err
	}

	m.setState(StateSubscribed)
	return nil
}

// Disconnect disconnects from the current device.
func (m *BLEManager) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.reconnectCancel != nil {
		m.reconnectCancel()
		m.reconnectCancel = nil
	}

	if m.device != nil {
		err := m.device.Disconnect()
		m.device = nil
		m.setState(StateDisconnected)
		return err
	}
	return nil
}

// StartAutoReconnect starts automatic reconnection using exponential backoff.
func (m *BLEManager) StartAutoReconnect() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.reconnectCancel != nil {
		m.reconnectCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.reconnectCancel = cancel

	go m.autoReconnect(ctx)
}

func (m *BLEManager) autoReconnect(ctx context.Context) {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		state := m.GetState()
		if state == StateConnected || state == StateSubscribed || state == StateConnecting {
			backoff = 1 * time.Second
			continue
		}

		m.mu.RLock()
		addr := m.lastAddress
		m.mu.RUnlock()

		if addr == "" {
			return
		}

		if err := m.Connect(addr); err == nil {
			if err := m.SubscribeHeartRate(); err == nil {
				backoff = 1 * time.Second
				continue
			}
		}

		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// GetLastDevice returns the last connected device address and name.
func (m *BLEManager) GetLastDevice() (address, name string) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastAddress, m.lastDeviceName
}
