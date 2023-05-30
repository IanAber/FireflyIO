package main

import (
	"encoding/binary"
	"fmt"
	"github.com/brutella/can"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"
)

var UnknownFrames map[uint32]time.Time

func init() {
	UnknownFrames = make(map[uint32]time.Time)
}

func getUnknownFrames(w http.ResponseWriter, _ *http.Request) {
	const deviceString = "GetUnknownFrames"
	setContentTypeHeader(w)
	if _, err := fmt.Fprint(w, `{
  "UnknownFrames" : {
`); err != nil {
		ReturnJSONError(w, deviceString, err, http.StatusInternalServerError, true)
	}
	for id, when := range UnknownFrames {
		if _, err := fmt.Fprintf(w, `    "0x%08x" : "%v"
`, id-0x80000000, when); err != nil {
			ReturnJSONError(w, deviceString, err, http.StatusInternalServerError, true)
		}
	}
	if _, err := fmt.Fprint(w, `  }
}`); err != nil {
		ReturnJSONError(w, deviceString, err, http.StatusInternalServerError, true)
	}
}

type FrameHandler func(frame can.Frame, canBus *CANBus)

//type CANHandler struct {
//	CANFrameID uint16
//	Handler    FrameHandler
//}

type CANBus struct {
	FrameHandlers  map[uint32]FrameHandler
	bus            *can.Bus
	Analog         [8]uint16
	Temperature    float32
	RawTemperature uint16
	VDD            float32
	RawVDD         uint16
	mu             sync.Mutex
}

const FlagsCanId = 0x010

const RelaysAndDigitalOutCanId = 0x011
const RelaysOutputsAndHeartbeat = 0x016

//  const DigitalInCanId = 0x012
const AnalogInputs0to3CanId = 0x013
const AnalogInputs4to7CanId = 0x014
const AnalogInputsInternalCanId = 0x015

const AcVoltsAmpsCanId0 = 0x018
const AcPowerEnergyCanId0 = 0x019
const AcHertzPfCanId0 = 0x01A
const AcErrorCanId0 = 0x1B
const DcVoltsAmpsCanId0 = 0x1C
const DcErrorCanId0 = 0x1F

const AcVoltsAmpsCanId1 = 0x028
const AcPowerEnergyCanId1 = 0x029
const AcHertzPfCanId1 = 0x02A
const AcErrorCanId1 = 0x2B
const DcVoltsAmpsCanId1 = 0x2C
const DcErrorCanId1 = 0x2F

const AcVoltsAmpsCanId2 = 0x038
const AcPowerEnergyCanId2 = 0x039
const AcHertzPfCanId2 = 0x03A
const AcErrorCanId2 = 0x3B
const DcVoltsAmpsCanId2 = 0x3C
const DcErrorCanId2 = 0x3F

const AcVoltsAmpsCanId3 = 0x048
const AcPowerEnergyCanId3 = 0x049
const AcHertzPfCanId3 = 0x04A
const AcErrorCanId3 = 0x4B
const DcVoltsAmpsCanId3 = 0x4C
const DcErrorCanId3 = 0x4F

// handleCANFrame figures out what to do with each CAN frame received
func (canBus *CANBus) handleCANFrame(frm can.Frame) {
	handler := canBus.FrameHandlers[frm.ID]
	if handler != nil {
		handler(frm, canBus)
	} else if frm.ID < 255 {
		log.Printf("Frame %d received with data %v\n", frm.ID, frm.Data)
	} else {
		UnknownFrames[frm.ID] = time.Now()
		//		log.Printf("0x%x", frm.ID)
	}
}

/*
NewCANBus
 connects to the given interface and starts receiving frames.
*/
func NewCANBus(interfaceName string) (*CANBus, error) {
	canBus := new(CANBus)
	var err error

	canBus.bus, err = can.NewBusForInterfaceWithName(interfaceName)
	canBus.FrameHandlers = make(map[uint32]FrameHandler)
	if err != nil {
		log.Println("CAN interface not available.", err)
	} else {
		canBus.bus.SubscribeFunc(canBus.handleCANFrame)
		canBus.FrameHandlers[FlagsCanId] = flagsHandler
		canBus.FrameHandlers[RelaysOutputsAndHeartbeat] = relayHandler
		canBus.FrameHandlers[AnalogInputs0to3CanId] = analogInputs0to3Handler
		canBus.FrameHandlers[AnalogInputs4to7CanId] = analogInputs4to7Handler
		canBus.FrameHandlers[AnalogInputsInternalCanId] = analogInputsInternalHandler
		canBus.FrameHandlers[AcVoltsAmpsCanId0] = acVoltsAndAmpsHandler0
		canBus.FrameHandlers[AcPowerEnergyCanId0] = acPowerAndEnergyHandler0
		canBus.FrameHandlers[AcHertzPfCanId0] = acPowerFactorAndFrequencyHandler0
		canBus.FrameHandlers[AcErrorCanId0] = acErrorHandler0
		canBus.FrameHandlers[AcVoltsAmpsCanId1] = acVoltsAndAmpsHandler1
		canBus.FrameHandlers[AcPowerEnergyCanId1] = acPowerAndEnergyHandler1
		canBus.FrameHandlers[AcHertzPfCanId1] = acPowerFactorAndFrequencyHandler1
		canBus.FrameHandlers[AcErrorCanId1] = acErrorHandler1
		canBus.FrameHandlers[AcVoltsAmpsCanId2] = acVoltsAndAmpsHandler2
		canBus.FrameHandlers[AcPowerEnergyCanId2] = acPowerAndEnergyHandler2
		canBus.FrameHandlers[AcHertzPfCanId2] = acPowerFactorAndFrequencyHandler2
		canBus.FrameHandlers[AcErrorCanId2] = acErrorHandler2
		canBus.FrameHandlers[AcVoltsAmpsCanId3] = acVoltsAndAmpsHandler3
		canBus.FrameHandlers[AcPowerEnergyCanId3] = acPowerAndEnergyHandler3
		canBus.FrameHandlers[AcHertzPfCanId3] = acPowerFactorAndFrequencyHandler3
		canBus.FrameHandlers[AcErrorCanId3] = acErrorHandler3
		canBus.FrameHandlers[DcVoltsAmpsCanId0] = dcVoltsAndAmpsHandler0
		canBus.FrameHandlers[DcErrorCanId0] = dcErrorHandler0
		canBus.FrameHandlers[DcVoltsAmpsCanId1] = dcVoltsAndAmpsHandler1
		canBus.FrameHandlers[DcErrorCanId1] = dcErrorHandler1
		canBus.FrameHandlers[DcVoltsAmpsCanId2] = dcVoltsAndAmpsHandler2
		canBus.FrameHandlers[DcErrorCanId2] = dcErrorHandler2
		canBus.FrameHandlers[DcVoltsAmpsCanId3] = dcVoltsAndAmpsHandler3
		canBus.FrameHandlers[DcErrorCanId3] = dcErrorHandler3

		//canBus.FrameHandlers[CanOutputControlMsg] = CanOutputControlHandler
		//canBus.FrameHandlers[CanBatterVoltageLimitsMsg] = CanBatterVoltageLimitsHandler
		canBus.FrameHandlers[CanPowerModeMsg] = CanPowerModeHandler
		canBus.FrameHandlers[CanPressuresMsg] = CanPressuresHandler
		canBus.FrameHandlers[CanStackCoolantMsg] = CanStackCoolantHandler
		canBus.FrameHandlers[CanAirFlowMsg] = CanAirFlowHandler
		canBus.FrameHandlers[CanAlarmsMsg] = CanAlarmsHandler
		canBus.FrameHandlers[CanStackOutputMsg] = CanStackOutputHandler
		canBus.FrameHandlers[CanCff1Msg] = CanCff1Handler
		canBus.FrameHandlers[CanInsulationMsg] = CanInsulationHanddler
		canBus.FrameHandlers[CanStackCellsID1to4Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID5to8Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID9to12Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID13to16Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID17to20Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID21to24Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID25to28Msg] = CanStackHandler
		canBus.FrameHandlers[CanStackCellsID29to32Msg] = CanStackHandler
		canBus.FrameHandlers[CanMaxMinCellsMsg] = CanStackHandler
		canBus.FrameHandlers[CanTotalStackVoltageMsg] = CanStackHandler
		canBus.FrameHandlers[CanATSCoolingFanMsg] = CanATSCoolingFanHandler
		canBus.FrameHandlers[CanWaterPumpMsg] = CanWaterPumpHandler
		canBus.FrameHandlers[CanDCDCConverterMsg] = CanDCDCConverterHandler
		canBus.FrameHandlers[CanDCOutputMsg] = CanDCOutputHandler
		canBus.FrameHandlers[CanBMSSettingsMsg] = CanBMSSettingsHandler
		canBus.FrameHandlers[CanKeyOnMsg] = CanKeyOnHandler
		canBus.FrameHandlers[CanRunTimeMsg] = CanRunTimeHandler

		go ConnectAndPublish(canBus)
	}
	log.Println("Logging CAN bus messages")
	return canBus, err
}

func ConnectAndPublish(canBus *CANBus) {
	if err := canBus.bus.ConnectAndPublish(); err != nil {
		// The CAN bus has stopped working!
		log.Println(err)
	}
	if disconnectErr := canBus.bus.Disconnect(); disconnectErr != nil {
		log.Println(disconnectErr)
	}
	canBus.bus = nil
}

func flagsHandler(_ can.Frame, _ *CANBus) {

}

func CanKeyOnHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.SystemInfo.Run = frame.Data[0] != 0
}

func CanRunTimeHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.SystemInfo.SetRunTime(frame.Data[2], frame.Data[3])
}

func CanPowerModeHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.PowerMode.Load(frame.Data)
}
func CanPressuresHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.Pressures.Load(frame.Data)
}
func CanStackCoolantHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.StackCoolant.Load(frame.Data)
}
func CanAirFlowHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.AirFlow.Load(frame.Data)
}
func CanAlarmsHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.Alarms.Load(frame.Data)
}
func CanStackOutputHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.StackOutput.Load(frame.Data)
}
func CanCff1Handler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.CffMsg.Load(frame.Data)
}
func CanInsulationHanddler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.Insulation.Load(frame.Data)
}
func CanStackHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.StackCells.Load(frame.ID, frame.Data)
}
func CanATSCoolingFanHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.ATSCoolingFan.Load(frame.Data)
}

func CanWaterPumpHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.WaterPump.Load(frame.Data)
}

func CanDCDCConverterHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.DCDCConverter.Load(frame.Data)
}
func CanDCOutputHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.DCOutput.Load(frame.Data)
}
func CanBMSSettingsHandler(frame can.Frame, _ *CANBus) {
	FuelCell.mu.Lock()
	defer FuelCell.mu.Unlock()
	FuelCell.BMSSettings.Load(frame.Data)
	if (frame.Data[6] != 0) != FuelCell.SystemInfo.exhaustLastValue {
		// Set the flag if it has changed since the last time we saw it. A timer resets it if it does not keep changing
		FuelCell.SystemInfo.SetExhaustFlag()
	}
	FuelCell.SystemInfo.exhaustLastValue = frame.Data[6] != 0
}

func relayHandler(frame can.Frame, _ *CANBus) {
	Relays.SetAllRelays(binary.LittleEndian.Uint16(frame.Data[0:2]))
	Outputs.SetAllOutputs(frame.Data[2])
	returnedHeartbeat = binary.LittleEndian.Uint16(frame.Data[4:6])
}

func analogInputs0to3Handler(frame can.Frame, _ *CANBus) {
	AnalogInputs.SetAnanlog0To3(frame.Data)
}

func analogInputs4to7Handler(frame can.Frame, _ *CANBus) {
	AnalogInputs.SetAnanlog4To7(frame.Data)
}

func analogInputsInternalHandler(frame can.Frame, _ *CANBus) {
	AnalogInputs.SetAnanlogInternal(frame.Data)
	Inputs.SetAllInputs(frame.Data[6] & 0xf)
}

func acVoltsAndAmpsHandler(device uint8, frame can.Frame) {
	ACMeasurements[device].setVolts(binary.LittleEndian.Uint16(frame.Data[0:2]))
	ACMeasurements[device].setAmps(binary.LittleEndian.Uint32(frame.Data[2:6]))
	ACMeasurements[device].setError(0)
}

func acVoltsAndAmpsHandler0(frame can.Frame, _ *CANBus) {
	acVoltsAndAmpsHandler(0, frame)
}
func acVoltsAndAmpsHandler1(frame can.Frame, _ *CANBus) {
	acVoltsAndAmpsHandler(1, frame)
}
func acVoltsAndAmpsHandler2(frame can.Frame, _ *CANBus) {
	acVoltsAndAmpsHandler(2, frame)
}
func acVoltsAndAmpsHandler3(frame can.Frame, _ *CANBus) {
	acVoltsAndAmpsHandler(3, frame)
}

func acPowerAndEnergyHandler(device uint8, frame can.Frame) {
	ACMeasurements[device].setPower(binary.LittleEndian.Uint32(frame.Data[0:4]))
	ACMeasurements[device].setEnergy(binary.LittleEndian.Uint32(frame.Data[4:8]))
	ACMeasurements[device].setError(0)
}
func acPowerAndEnergyHandler0(frame can.Frame, _ *CANBus) {
	acPowerAndEnergyHandler(0, frame)
}
func acPowerAndEnergyHandler1(frame can.Frame, _ *CANBus) {
	acPowerAndEnergyHandler(1, frame)
}
func acPowerAndEnergyHandler2(frame can.Frame, _ *CANBus) {
	acPowerAndEnergyHandler(2, frame)
}
func acPowerAndEnergyHandler3(frame can.Frame, _ *CANBus) {
	acPowerAndEnergyHandler(3, frame)
}

func acErrorHandler(device uint8, frame can.Frame) {
	ACMeasurements[device].setError(frame.Data[0])
}
func acErrorHandler0(frame can.Frame, _ *CANBus) {
	acErrorHandler(0, frame)
}
func acErrorHandler1(frame can.Frame, _ *CANBus) {
	acErrorHandler(1, frame)
}
func acErrorHandler2(frame can.Frame, _ *CANBus) {
	acErrorHandler(2, frame)
}
func acErrorHandler3(frame can.Frame, _ *CANBus) {
	acErrorHandler(3, frame)
}

func acPowerFactorAndFrequencyHandler(device uint8, frame can.Frame) {
	ACMeasurements[device].setFrequency(binary.LittleEndian.Uint16(frame.Data[0:2]))
	ACMeasurements[device].setPowerFactor(binary.LittleEndian.Uint16(frame.Data[2:4]))
	ACMeasurements[device].setError(0)
}
func acPowerFactorAndFrequencyHandler0(frame can.Frame, _ *CANBus) {
	acPowerFactorAndFrequencyHandler(0, frame)
}
func acPowerFactorAndFrequencyHandler1(frame can.Frame, _ *CANBus) {
	acPowerFactorAndFrequencyHandler(1, frame)
}
func acPowerFactorAndFrequencyHandler2(frame can.Frame, _ *CANBus) {
	acPowerFactorAndFrequencyHandler(2, frame)
}
func acPowerFactorAndFrequencyHandler3(frame can.Frame, _ *CANBus) {
	acPowerFactorAndFrequencyHandler(3, frame)
}

func dcVoltsAndAmpsHandler(device uint8, frame can.Frame) {
	DCMeasurements[device].setVolts(binary.LittleEndian.Uint16(frame.Data[0:2]))
	DCMeasurements[device].setAmps(binary.LittleEndian.Uint32(frame.Data[2:6]))
	DCMeasurements[device].setError(0)
}

func dcVoltsAndAmpsHandler0(frame can.Frame, _ *CANBus) {
	dcVoltsAndAmpsHandler(0, frame)
}
func dcVoltsAndAmpsHandler1(frame can.Frame, _ *CANBus) {
	dcVoltsAndAmpsHandler(1, frame)
}
func dcVoltsAndAmpsHandler2(frame can.Frame, _ *CANBus) {
	dcVoltsAndAmpsHandler(2, frame)
}
func dcVoltsAndAmpsHandler3(frame can.Frame, _ *CANBus) {
	dcVoltsAndAmpsHandler(3, frame)
}

func dcErrorHandler(device uint8, frame can.Frame) {
	DCMeasurements[device].setError(frame.Data[0])
}
func dcErrorHandler0(frame can.Frame, _ *CANBus) {
	dcErrorHandler(0, frame)
}
func dcErrorHandler1(frame can.Frame, _ *CANBus) {
	dcErrorHandler(1, frame)
}
func dcErrorHandler2(frame can.Frame, _ *CANBus) {
	dcErrorHandler(2, frame)
}
func dcErrorHandler3(frame can.Frame, _ *CANBus) {
	dcErrorHandler(3, frame)
}

func (bus *CANBus) SetRelays(relays uint16) error {
	var frame can.Frame
	binary.LittleEndian.PutUint16(frame.Data[:], relays)
	frame.Data[2] = Outputs.GetAllOutputs()
	binary.LittleEndian.PutUint16(frame.Data[4:6], heartbeat)
	frame.ID = RelaysAndDigitalOutCanId
	frame.Length = 8
	if err := bus.bus.Publish(frame); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (bus *CANBus) SetDigitalOutputs(outputs uint8) error {
	var frame can.Frame
	binary.LittleEndian.PutUint16(frame.Data[:], Relays.GetAllRelays())
	frame.Data[2] = outputs
	frame.ID = RelaysAndDigitalOutCanId
	if err := bus.bus.Publish(frame); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (bus *CANBus) SetFlags(flag0 uint8, flag1 uint8, flag2 uint8, flag3 uint8, flag4 uint8, flag5 uint8, flag6 uint8, flag7 uint8) error {
	var frame can.Frame
	frame.Data[0] = flag0
	frame.Data[1] = flag1
	frame.Data[2] = flag2
	frame.Data[3] = flag3
	frame.Data[4] = flag4
	frame.Data[5] = flag5
	frame.Data[6] = flag6
	frame.Data[7] = flag7
	frame.ID = FlagsCanId
	frame.Length = 8
	if err := bus.bus.Publish(frame); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func init() {
	go MonitorCANBusComms()
}

func MonitorCANBusComms() {
	heartbeatTimer := time.NewTicker(time.Second * 5)
	for {
		<-heartbeatTimer.C
		diff := heartbeat - returnedHeartbeat

		if diff > 10 {
			log.Printf("CAN Heartbeat has been lost. Resetting the USB port. Heartbeat = %d | returnedHeartbeat = %d\n", heartbeat, returnedHeartbeat)
			heartbeat = 0
			returnedHeartbeat = 0
			// Reset the CAN bus interface
			cmd := exec.Command("usbreset", "1d50:606f")
			if err := cmd.Start(); err != nil {
				log.Println("Failed to reset the CAN bus.", err)
			}
		} else {
			heartbeat++
		}
	}
}
