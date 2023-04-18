package main

import (
	"fmt"
	"strings"
	"sync"
)

type DigitalOutputType struct {
	Name string `json:name`
	Pin  bool   `json:value`
}

type DigitalOutputsType struct {
	Outputs [6]DigitalOutputType `json:outputs`
	mu      sync.Mutex
}

func (do *DigitalOutputsType) InitOutputs() {
	do.mu.Lock()
	defer do.mu.Unlock()

	for idx := range do.Outputs {
		do.Outputs[idx].Name = fmt.Sprintf("output-%d", idx)
		do.Outputs[idx].Pin = false
	}
}

func (do *DigitalOutputsType) SetAllOutputs(settings uint8) {
	do.mu.Lock()
	defer do.mu.Unlock()

	for idx, _ := range Outputs.Outputs {
		do.Outputs[idx].Pin = (settings & 1) != 0
		settings >>= 1
	}
}

func (do *DigitalOutputsType) GetOutput(port uint8) bool {
	do.mu.Lock()
	defer do.mu.Unlock()

	return do.Outputs[port].Pin
}

func (do *DigitalOutputsType) GetOutputName(port uint8) string {
	do.mu.Lock()
	defer do.mu.Unlock()

	return do.Outputs[port].Name
}

func (do *DigitalOutputsType) GetOutputByName(port string) (bool, error) {
	do.mu.Lock()
	defer do.mu.Unlock()

	port = strings.ToLower(port)
	for _, op := range do.Outputs {
		if op.Name == port {
			return op.Pin, nil
		}
	}
	return false, fmt.Errorf("Invalid output port name - %s", port)
}

func (do *DigitalOutputsType) SetOutputName(port uint8, name string) {
	do.mu.Lock()
	defer do.mu.Unlock()

	do.Outputs[port].Name = name
}

func (do *DigitalOutputsType) GetAllOutputs() uint8 {
	do.mu.Lock()
	defer do.mu.Unlock()

	var val uint8

	for _, op := range do.Outputs {
		val >>= 1
		if op.Pin {
			val += 0x20
		}
	}
	return val
}

func (do *DigitalOutputsType) SetOutput(pin uint8, on bool) {
	op := do.GetAllOutputs()
	fmt.Printf("Outputs = %b\n", op)
	if on {
		op |= uint8(1) << pin
	} else {
		op &= ^(uint8(1) << pin)
	}
	do.SetAllOutputs(op)
	fmt.Printf("New outputs = %b\n", op)
	canBus.SetDigitalOutputs(op)
}

func (do *DigitalOutputsType) SetOutputByName(pin string, on bool) error {
	pin = strings.ToLower(pin)
	for idx, op := range do.Outputs {
		if op.Name == pin {
			do.SetOutput(uint8(idx), on)
			return nil
		}
	}
	return fmt.Errorf("Invalid output port - %s", pin)
}
