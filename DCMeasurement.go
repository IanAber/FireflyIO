package main

import (
	"fmt"
	"math"
	"sync"
)

type DCMeasurementsType struct {
	Name  string
	Volts float32
	Amps  float32
	Error uint8
	mu    sync.Mutex
}

func (dc *DCMeasurementsType) InitDCMeasurement() {
	dc.Name = ""
	dc.Volts = 0.0
	dc.Amps = 0.0
}

func (dc *DCMeasurementsType) setVolts(v uint16) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Volts = float32(v) / 100.0
}

func (dc *DCMeasurementsType) setAmps(i uint32) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.Amps = float32(math.Round(float64(int32(i)-554416)/60.0) / 100)
}

func (dc *DCMeasurementsType) getVolts() float32 {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	return (dc.Volts)
}

func (dc *DCMeasurementsType) getAmps() float32 {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	return (dc.Amps)
}

func (dc *DCMeasurementsType) getPower() float32 {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	return (dc.Volts * dc.Amps)
}

func (dc *DCMeasurementsType) getError() string {
	switch dc.Error {
	case 0:
		return ""
	case 0xFF:
		return "Not a Master"
	case 0xFE:
		return "Polling Error"
	case 0xFD:
		return "Buffer Overflow"
	case 0xFC:
		return "Bad CRC"
	case 0xFB:
		return "Exception"
	case 0xFA:
		return "Bad Size"
	case 0xF9:
		return "Bad Address"
	case 0xF8:
		return "Timeout"
	case 0xF7:
		return "Bad Slave ID"
	case 0xF6:
		return "Bad TCP ID"
	default:
		return fmt.Sprintf("Error = %0x", dc.Error)
	}
}

func (dc *DCMeasurementsType) setError(err uint8) {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	dc.Error = err
}
