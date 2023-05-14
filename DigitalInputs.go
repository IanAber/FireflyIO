package main

import (
	"fmt"
	"strings"
	"sync"
)

type DigitalInputType struct {
	Name string `json:"Name"`
	Pin  bool   `json:"Value"`
}

type DigitalInputsType struct {
	Inputs [4]DigitalInputType `json:"Inputs"`
	mu     sync.Mutex
}

func (di *DigitalInputsType) InitInputs() {
	di.mu.Lock()
	defer di.mu.Unlock()

	for idx := range di.Inputs {
		di.Inputs[idx].Name = fmt.Sprintf("input-%d", idx)
		di.Inputs[idx].Pin = false
	}
}

func (di *DigitalInputsType) SetAllInputs(settings uint8) {
	di.mu.Lock()
	defer di.mu.Unlock()

	for ip := range di.Inputs {
		di.Inputs[ip].Pin = (settings & 1) != 0
		settings >>= 1
	}
}

func (di *DigitalInputsType) GetInput(port uint8) bool {
	di.mu.Lock()
	defer di.mu.Unlock()

	return di.Inputs[port].Pin
}

func (di *DigitalInputsType) GetInputName(port uint8) string {
	di.mu.Lock()
	defer di.mu.Unlock()

	return di.Inputs[port].Name
}

func (di *DigitalInputsType) SetInputName(port uint8, name string) {
	di.mu.Lock()
	defer di.mu.Unlock()

	di.Inputs[port].Name = strings.ToLower(name)
}

func (di *DigitalInputsType) GetAllInputs() uint8 {
	di.mu.Lock()
	defer di.mu.Unlock()

	var val uint8

	for _, op := range di.Inputs {
		val <<= 1
		if op.Pin {
			val += 1
		}
	}
	return val
}

func (di *DigitalInputsType) GetInputByName(port string) (bool, error) {
	port = strings.ToLower(port)
	for _, input := range di.Inputs {
		if input.Name == port {
			return input.Pin, nil
		}
	}
	return false, fmt.Errorf("invalid input pin name - %s", port)
}
