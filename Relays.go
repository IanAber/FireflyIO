package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

type RelayType struct {
	Name string `json:"name"`
	On   bool   `json:"value"`
}

type RelaysType struct {
	Relays [16]RelayType `json:"relays"`
	mu     sync.Mutex
}

func (rl *RelaysType) InitRelays() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for idx := range rl.Relays {
		rl.Relays[idx].Name = fmt.Sprintf("relay-%d", idx)
		rl.Relays[idx].On = false
	}
}

func (rl *RelaysType) SetAllRelays(settings uint16) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for relay := range rl.Relays {
		rl.Relays[relay].On = (settings & 1) != 0
		settings >>= 1
	}
}

func (rl *RelaysType) GetAllRelays() uint16 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	var val uint16

	for _, relay := range rl.Relays {
		val >>= 1
		if relay.On {
			val += 0x8000
		}
	}
	return val
}

func (rl *RelaysType) GetRelay(relay uint8) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.Relays[relay].On
}

func (rl *RelaysType) GetRelayName(relay uint8) string {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.Relays[relay].Name
}

func (rl *RelaysType) SetRelayName(relay uint8, name string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.Relays[relay].Name = name
}

func (rl *RelaysType) SetRelay(relay uint8, on bool) {
	// Get the current relay settings
	relays := rl.GetAllRelays()
	// Set or reset the supplied relay
	if on {
		relays |= uint16(1) << relay
	} else {
		relays &= ^(uint16(1) << relay)
	}
	// Set the hardware
	if err := canBus.SetRelays(relays); err != nil {
		log.Print(err)
	}
	// Update the local copy
	rl.SetAllRelays(relays)
}

func (rl *RelaysType) SetRelayByName(relay string, on bool) error {
	relay = strings.ToLower(relay)
	for idx, r := range rl.Relays {
		if strings.ToLower(r.Name) == relay {
			rl.SetRelay(uint8(idx), on)
			return nil
		}
	}
	return fmt.Errorf("invalid relay name - %s", relay)
}

/*
UpdateRelays retransmits the current relay settings as a heartbeat signal to the Firefly IO board
*/
func (rl *RelaysType) UpdateRelays() {
	relays := rl.GetAllRelays()
	if err := canBus.SetRelays(relays); err != nil {
		log.Print(err)
	}
}
