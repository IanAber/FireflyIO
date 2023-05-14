package main

import (
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
)

type AnalogInputType struct {
	Name  string  `json:"Name"`
	Raw   uint16  `json:"Raw"`
	Value float32 `json:"Value"`
}

type AnalogInputsType struct {
	Inputs         [8]AnalogInputType `json:"Inputs"`
	Temperature    int16              `json:"Temperature"`
	RawTemperature uint16             `json:"RawTemperature"`
	VrefValue      uint16             `json:"VrefValue"`
	mu             sync.Mutex
}

func (ai *AnalogInputsType) InitAnalogInputs() {
	for idx := range ai.Inputs {
		ai.Inputs[idx].Name = fmt.Sprintf("input-%d", idx)
	}
}

func (ai *AnalogInputsType) SetAnanlog0To3(data [8]byte) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.Inputs[0].Raw = binary.LittleEndian.Uint16(data[0:2])
	ai.Inputs[0].Value = (float32(ai.Inputs[0].Raw) * currentSettings.AnalogChannels[0].calibrationMultiplier) + currentSettings.AnalogChannels[0].calibrationConstant
	ai.Inputs[1].Raw = binary.LittleEndian.Uint16(data[2:4])
	ai.Inputs[1].Value = (float32(ai.Inputs[1].Raw) * currentSettings.AnalogChannels[1].calibrationMultiplier) + currentSettings.AnalogChannels[1].calibrationConstant
	ai.Inputs[2].Raw = binary.LittleEndian.Uint16(data[4:6])
	ai.Inputs[2].Value = (float32(ai.Inputs[2].Raw) * currentSettings.AnalogChannels[2].calibrationMultiplier) + currentSettings.AnalogChannels[2].calibrationConstant
	ai.Inputs[3].Raw = binary.LittleEndian.Uint16(data[6:8])
	ai.Inputs[3].Value = (float32(ai.Inputs[3].Raw) * currentSettings.AnalogChannels[3].calibrationMultiplier) + currentSettings.AnalogChannels[3].calibrationConstant
}

func (ai *AnalogInputsType) SetAnanlog4To7(data [8]byte) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.Inputs[4].Raw = binary.LittleEndian.Uint16(data[0:2])
	ai.Inputs[4].Value = (float32(ai.Inputs[4].Raw) * currentSettings.AnalogChannels[4].calibrationMultiplier) + currentSettings.AnalogChannels[4].calibrationConstant
	ai.Inputs[5].Raw = binary.LittleEndian.Uint16(data[2:4])
	ai.Inputs[5].Value = (float32(ai.Inputs[5].Raw) * currentSettings.AnalogChannels[5].calibrationMultiplier) + currentSettings.AnalogChannels[5].calibrationConstant
	ai.Inputs[6].Raw = binary.LittleEndian.Uint16(data[4:6])
	ai.Inputs[6].Value = (float32(ai.Inputs[6].Raw) * currentSettings.AnalogChannels[6].calibrationMultiplier) + currentSettings.AnalogChannels[6].calibrationConstant
	ai.Inputs[7].Raw = binary.LittleEndian.Uint16(data[6:8])
	ai.Inputs[7].Value = (float32(ai.Inputs[7].Raw) * currentSettings.AnalogChannels[7].calibrationMultiplier) + currentSettings.AnalogChannels[7].calibrationConstant
}

func (ai *AnalogInputsType) SetAnanlogInternal(data [8]byte) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.Temperature = int16(binary.LittleEndian.Uint16(data[0:2]))
	ai.RawTemperature = binary.LittleEndian.Uint16(data[2:4])
	ai.VrefValue = binary.LittleEndian.Uint16(data[4:6])
}

func (ai *AnalogInputsType) SetInputName(port uint8, name string) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	ai.Inputs[port].Name = strings.ToLower(name)
}

func (ai *AnalogInputsType) GetInputByName(port string) (int32, error) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	port = strings.ToLower(port)
	switch port {
	case "temperature":
		return int32(ai.Temperature), nil
	case "raw temperature":
		return int32(ai.RawTemperature), nil
	case "vref":
		return int32(ai.VrefValue), nil
	default:
		for _, ip := range ai.Inputs {
			if ip.Name == port {
				return int32(ip.Value), nil
			}
		}
	}
	return 0, fmt.Errorf("invalid input name %s", port)
}

func (ai *AnalogInputsType) GetCPUTemperature() (rawValue uint16, celsiusValue float32) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	return ai.RawTemperature, float32(ai.Temperature) / 100
}

func (ai *AnalogInputsType) GetVREF() uint16 {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	return ai.VrefValue
}

func (ai *AnalogInputsType) GetInput(port uint8) (uint16, float32) {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	return ai.Inputs[port].Raw, ai.Inputs[port].Value
}

func (ai *AnalogInputsType) GetRawInput(port uint8) uint16 {
	ai.mu.Lock()
	defer ai.mu.Unlock()

	return ai.Inputs[port].Raw
}
