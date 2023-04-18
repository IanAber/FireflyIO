package main

import "sync"

type ACMeasurementsType struct {
	Volts       float32
	Amps        float32
	Power       float32
	WattHours   uint32
	Frequency   float32
	PowerFactor float32
	mu          sync.Mutex
}

func (ac *ACMeasurementsType) InitACMeasurement() {
	ac.Volts = 0.0
	ac.Amps = 0.0
	ac.Power = 0.0
	ac.WattHours = 0
	ac.Frequency = 60.0
	ac.PowerFactor = 1.0
}

func (ac *ACMeasurementsType) setVolts(v uint16) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.Volts = float32(v) / 10.0
}

func (ac *ACMeasurementsType) setAmps(i uint32) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.Amps = float32(i) / 1000.0
}

func (ac *ACMeasurementsType) setPower(p uint32) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.Power = float32(p) / 10.0
}

func (ac *ACMeasurementsType) setEnergy(whr uint32) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.WattHours = whr
}

func (ac *ACMeasurementsType) setFrequency(f uint16) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.Frequency = float32(f) / 10.0
}

func (ac *ACMeasurementsType) setPowerFactor(pf uint16) {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.PowerFactor = float32(pf) / 100
}

func (ac *ACMeasurementsType) getVolts() float32 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return (ac.Volts)
}

func (ac *ACMeasurementsType) getAmps() float32 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return (ac.Amps)
}

func (ac *ACMeasurementsType) getPower() float32 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return (ac.Power)
}

func (ac *ACMeasurementsType) getEnergy() uint32 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return ac.WattHours
}

func (ac *ACMeasurementsType) getFrequency() float32 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return ac.Frequency
}

func (ac *ACMeasurementsType) getPowerFactor() float32 {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	return ac.PowerFactor
}

func (ac *ACMeasurementsType) setError() {
	ac.mu.Lock()
	defer ac.mu.Unlock()

	ac.Power = 0
	ac.Frequency = 0
	ac.Amps = 0
	ac.PowerFactor = 0
	ac.Volts = 0
	ac.WattHours = 0
}
