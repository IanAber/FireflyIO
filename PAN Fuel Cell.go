package main

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/brutella/can"
	"log"
	"time"

	//	"log"
	"sync"
)

type PanSettingsType struct {
	TargetPower       float64 // Power output requested
	TargetBatteryHigh float64 // High voltage target value
	TargetBatteryLow  float64 // Low voltage target value
	FuelCellOn        bool    // Flag to tell the unit to turn on
	Exhaust           bool    // Flag to indicate that an Exhaust command is requested
	PumpActive        bool    // Flag to show if the water pump is running
	//	PumpTimer         *time.Timer // The timer to detect that no water pump messages have been received
}

type RunCommandType byte

var heartbeat uint16
var returnedHeartbeat uint16

const (
	_ = iota
	StartUp
	ShutDown
)

type ExhaustModeType byte

const (
	ExhaustClosed = iota
	ExhaustOpen
)

const CanOutputControlMsg = 0x961088C1

type OutputControlType struct {
	FuelCellRunEnable RunCommandType  // Startup / Shutdown
	PowerDemand       uint8           // kW x 10
	ExhaustMode       ExhaustModeType // Open / Closed
}

func (t *OutputControlType) GetPowerDemand() float64 {
	return float64(t.PowerDemand) / 10.0
}

func (t *OutputControlType) UpdateFuelCell(bus *CANBus) error {
	var frame can.Frame

	frame.ID = CanOutputControlMsg
	frame.Length = 8
	frame.Data[0] = byte(t.FuelCellRunEnable)
	frame.Data[1] = t.PowerDemand
	frame.Data[2] = byte(t.ExhaustMode)
	return bus.bus.Publish(frame)
}

const CanBatteryVoltageLimitsMsg = 0x961088C2

type BatteryVoltageLimitsType struct {
	BMSHighVoltage uint16 //Battery high voltage setpoint
	BMSLowVoltage  uint16 //Battery low voltage setpoint
	IsoFlag        bool   //Set true to suppress stack isolation tests.
}

// UpdateFuelCell sends the frame to the CAN bus
func (t *BatteryVoltageLimitsType) UpdateFuelCell(bus *CANBus) error {
	var frame can.Frame

	frame.ID = CanBatteryVoltageLimitsMsg
	frame.Length = 8
	binary.LittleEndian.PutUint16(frame.Data[0:2], t.BMSHighVoltage)
	binary.LittleEndian.PutUint16(frame.Data[2:4], t.BMSLowVoltage)
	if t.IsoFlag {
		frame.Data[4] = 1
	} else {
		frame.Data[4] = 0
	}
	return bus.bus.Publish(frame)
}

type PowerModeStateType byte

//const (
//	PMOff = iota
//	PMInit
//	PMH2Purge
//	PMStartup
//	PMAirPurge
//	PMH2LeakCheck
//	PMManual
//	PMEmergencyShut
//	PMFault
//	PM_Shutdown
//)

func (pm PowerModeStateType) String() string {
	modeStates := [...]string{"Off", "Standby", "Hydrogen intake", "Start", "AirPurge", "Hydrogen leak check", "manual", "emergency stop", "fault", "shutdown"}
	return modeStates[pm]
}

const CanPowerModeMsg = 0x961088A1

type PowerModeType struct {
	PowerModeState PowerModeStateType
	FaultLevel     byte
	FaultCode      uint16
	RunStage       byte
}

func (t *PowerModeType) Load(data [8]byte) {
	t.PowerModeState = PowerModeStateType(data[0])
	t.FaultLevel = data[1]
	t.FaultCode = binary.LittleEndian.Uint16(data[2:4])
	t.RunStage = data[4]
	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.PowerModeState = data[0]
	dbRecord.RunStage = data[4]
	dbRecord.FaultLevel = data[1]
}

const CanPressuresMsg = 0x961088A2

type PressuresType struct {
	H2Pressure        uint16 // Hydrogen pressure
	AirPressure       uint16 // Air pressure
	CoolantPressure   uint16 // Coolant pressure
	H2AirPressureDiff uint16 // Hydrogen air pressure difference
}

func (t *PressuresType) Load(data [8]byte) {
	t.H2Pressure = binary.LittleEndian.Uint16(data[0:2])
	t.AirPressure = binary.LittleEndian.Uint16(data[2:4])
	t.CoolantPressure = binary.LittleEndian.Uint16(data[4:6])
	t.H2AirPressureDiff = binary.LittleEndian.Uint16(data[6:8])
	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.HydrogenPressure = t.H2Pressure
	dbRecord.AirPressure = t.AirPressure
	dbRecord.CoolantPressure = t.CoolantPressure
}

func (t *PressuresType) GetH2Pressure() float64 {
	return float64(t.H2Pressure) / 10.0
}

func (t *PressuresType) GetAirPressure() float64 {
	return float64(t.AirPressure) / 10.0
}

func (t *PressuresType) GetCoolantPressure() float64 {
	return float64(t.CoolantPressure) / 10.0
}

func (t *PressuresType) GetH2AirPressureDiff() float64 {
	return float64(t.H2AirPressureDiff) / 10.0
}

const CanStackCoolantMsg = 0x961088A3

type StackCoolantType struct {
	CoolantInTemp  uint16 // Coolant temperature at the inlet of the stack
	CoolantOutTemp uint16 // Coolant temperature at the outlet of the stack
	AirTemp        uint16 // Air temperature at the inlet of the stack
	AmbientTemp    uint16 // Ambient temperature
}

func (t *StackCoolantType) Load(data [8]byte) {
	t.CoolantInTemp = binary.LittleEndian.Uint16(data[0:2])
	t.CoolantOutTemp = binary.LittleEndian.Uint16(data[2:4])
	t.AirTemp = binary.LittleEndian.Uint16(data[4:6])
	t.AmbientTemp = binary.LittleEndian.Uint16(data[6:8])
	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.CoolantInlTemp = t.CoolantInTemp
	dbRecord.CoolantOutTemp = t.CoolantOutTemp
	dbRecord.AirinletTemp = t.AirTemp
	dbRecord.AmbientTemp = t.AmbientTemp
}

func (t *StackCoolantType) GetCoolantInTemp() float64 {
	return float64(t.CoolantInTemp) / 10.0
}

func (t *StackCoolantType) GetCoolantOutTemp() float64 {
	return float64(t.CoolantOutTemp) / 10.0
}

func (t *StackCoolantType) GetAirTemp() float64 {
	return float64(t.AirTemp) / 10.0
}

func (t *StackCoolantType) GetAmbientTemp() float64 {
	return float64(t.AmbientTemp) / 10.0
}

const CanAirFlowMsg = 0x961088A4

type AirFlowType struct {
	Flow uint16 //Air flow Lpm * 10
}

func (t *AirFlowType) Load(data [8]byte) {
	t.Flow = binary.LittleEndian.Uint16(data[4:6])
	dbRecord.mu.Lock()
	dbRecord.mu.Unlock()
	dbRecord.AirFlow = t.Flow
}

func (t *AirFlowType) GetFlow() float64 {
	return float64(t.Flow) / 10.0
}

const CanAlarmsMsg = 0x961088A5

type AlarmsType struct {
	VoltageLow             bool
	H2concentration        bool
	CoolantTempOutDiff     bool
	CoolantTempHigh        bool
	WaterPumpFault         bool
	H2CirculatingPumpFault bool
	H2PressureHigh         bool
	H2PressureSensorFault  bool
	DcdcCommunicationFault bool
	DcdcFault              bool
	PtcFault               bool
	H2TankTemp             bool
	H2TankHighPressure     bool
	H2TankMidPressure      bool
	H2TankLowPressure      bool
	FcuToVcuFault          bool
	TempSensorFault        bool
	H2SPCheckFault         bool
	H2SOCLow               bool
	H2OutPressureLow       bool
	AirPressureLow         bool
	AirPressureHigh        bool
	AirTempHigh            bool
	CoolantPressureHigh    bool
	CellVoltageHigh        bool
	IsoLow                 bool
	H2AirDiffHighMinus     bool
	H2AirDiffHighPlus      bool
	StartUploss            bool
	H2Leakageloss          bool
}

func (al *AlarmsType) Load(data [8]byte) {
	al.VoltageLow = (data[0] & 1) != 0
	al.H2concentration = (data[0] & 2) != 0
	al.CoolantTempOutDiff = (data[0] & 4) != 0
	al.CoolantTempHigh = (data[0] & 8) != 0
	al.WaterPumpFault = (data[0] & 0x10) != 0
	al.H2CirculatingPumpFault = (data[0] & 0x20) != 0
	al.H2PressureHigh = (data[0] & 0x40) != 0
	al.H2PressureSensorFault = (data[0] & 0x80) != 0

	al.DcdcCommunicationFault = (data[1] & 0x01) != 0
	al.DcdcFault = (data[1] & 0x02) != 0
	al.PtcFault = (data[1] & 0x04) != 0
	al.H2TankTemp = (data[1] & 0x08) != 0
	al.H2TankHighPressure = (data[1] & 0x10) != 0
	al.H2TankMidPressure = (data[1] & 0x20) != 0
	al.H2TankLowPressure = (data[1] & 0x40) != 0
	al.FcuToVcuFault = (data[1] & 0x80) != 0

	al.TempSensorFault = (data[2] & 0x01) != 0
	al.H2SPCheckFault = (data[2] & 0x02) != 0
	al.H2SOCLow = (data[2] & 0x04) != 0
	al.H2OutPressureLow = (data[2] & 0x08) != 0
	al.AirPressureLow = (data[2] & 0x10) != 0
	al.AirPressureHigh = (data[2] & 0x20) != 0
	al.AirTempHigh = (data[2] & 0x40) != 0
	al.CoolantPressureHigh = (data[2] & 0x80) != 0

	al.CellVoltageHigh = (data[3] & 0x01) != 0
	al.IsoLow = (data[3] & 0x02) != 0
	al.H2AirDiffHighMinus = (data[3] & 0x04) != 0
	al.H2AirDiffHighPlus = (data[3] & 0x08) != 0
	al.StartUploss = (data[3] & 0x10) != 0
	al.H2Leakageloss = (data[3] & 0x20) != 0
}

func (al *AlarmsType) Text() []string {
	alarmText := make([]string, 0)
	if al.AirPressureLow {
		alarmText = append(alarmText, "Abnormal low air pressure")
	}
	if al.AirPressureHigh {
		alarmText = append(alarmText, "Abnormal high air pressure")
	}
	if al.AirTempHigh {
		alarmText = append(alarmText, "Abnormally high air temperature")
	}
	if al.CoolantTempOutDiff {
		alarmText = append(alarmText, "Abnormal temperature difference between inlet and outlet")
	}
	if al.CoolantTempHigh {
		alarmText = append(alarmText, "Abnormally high outlet water temperature")
	}
	if al.CoolantPressureHigh {
		alarmText = append(alarmText, "Abnormal high cooling water pressure")
	}
	if al.CellVoltageHigh {
		alarmText = append(alarmText, "Stack cell high voltage abnormality")
	}
	if al.DcdcFault {
		alarmText = append(alarmText, "DC to DC Converter Fault")
	}
	if al.DcdcCommunicationFault {
		alarmText = append(alarmText, "DC to DC Converter Communication Fault")
	}
	if al.FcuToVcuFault {
		alarmText = append(alarmText, "FCU communication abnormal")
	}
	if al.H2AirDiffHighMinus {
		alarmText = append(alarmText, "Abnormal large hydrogen-air pressure difference (negative direction)")
	}
	if al.H2AirDiffHighPlus {
		alarmText = append(alarmText, "Abnormal hydrogen-air pressure difference (forward direction)")
	}
	if al.H2concentration {
		alarmText = append(alarmText, "Abnormal hydrogen concentration in the module")
	}
	if al.H2CirculatingPumpFault {
		alarmText = append(alarmText, "Abnormal hydrogen pump")
	}
	if al.H2Leakageloss {
		alarmText = append(alarmText, "Hydrogen leak check failed")
	}
	if al.H2OutPressureLow {
		alarmText = append(alarmText, "H2 Outlet Pressure Low")
	}
	if al.H2PressureHigh {
		alarmText = append(alarmText, "Abnormally high hydrogen pressure")
	}
	if al.H2PressureSensorFault {
		alarmText = append(alarmText, "The hydrogen outlet pressure sensor is abnormal")
	}
	if al.H2SOCLow {
		alarmText = append(alarmText, "Hydrogen tank SOC is too low")
	}
	if al.H2SPCheckFault {
		alarmText = append(alarmText, "Hydrogen pressure sensor self-test is abnormal")
	}
	if al.H2TankLowPressure {
		alarmText = append(alarmText, "Abnormal low pressure of hydrogen tank")
	}
	if al.H2TankMidPressure {
		alarmText = append(alarmText, "Abnormal pressure in the hydrogen tank")
	}
	if al.H2TankHighPressure {
		alarmText = append(alarmText, "Abnormal high pressure of hydrogen tank")
	}
	if al.H2TankTemp {
		alarmText = append(alarmText, "Abnormal temperature of hydrogen tank")
	}
	if al.IsoLow {
		alarmText = append(alarmText, "Abnormal low insulation")
	}
	if al.PtcFault {
		alarmText = append(alarmText, "Heater failure")
	}
	if al.StartUploss {
		alarmText = append(alarmText, "Low starting hydrogen pressure (below 20KPA)")
	}
	if al.TempSensorFault {
		alarmText = append(alarmText, "Abnormal temperature sensor")
	}
	if al.VoltageLow {
		alarmText = append(alarmText, "Single cell voltage undervoltage")
	}
	if al.WaterPumpFault {
		alarmText = append(alarmText, "Water pump failure")
	}
	return alarmText
}

const CanStackOutputMsg = 0x961088A7

type StackOutputType struct {
	Voltage uint16 //Stack voltage
	Current uint16 //Stack current
	Power   uint32 //Stack power
}

func (t *StackOutputType) Load(data [8]byte) {
	t.Voltage = binary.LittleEndian.Uint16(data[0:2])
	t.Current = binary.LittleEndian.Uint16(data[2:4])
	// Power is actually a 24 bit value
	t.Power = uint32(data[4]) | uint32(data[5])<<8 | uint32(data[6])<<16

	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.StackVoltage = t.Voltage
	dbRecord.StackCurrent = t.Current
}

func (t *StackOutputType) GetVoltage() float64 {
	return float64(t.Voltage) / 10.0
}

func (t *StackOutputType) GetCurrent() float64 {
	return float64(t.Current) / 10.0
}

func (t *StackOutputType) GetPower() float64 {
	return float64(t.Power) / 10.0
}

const CanCff1Msg = 0x8CFF1C91

type CffMsgType struct {
	GasConcentration uint8
	MSBSide          byte
	CycleCounter     uint8
	SensorFaultCode  byte
	LSBCheckSumq     byte
}

func (t *CffMsgType) Load(data [8]byte) {
	t.GasConcentration = data[0]
	t.MSBSide = data[1]
	t.CycleCounter = data[2] & 0x0f
	t.SensorFaultCode = (data[2] & 0x30) >> 4
	t.LSBCheckSumq = data[4]

	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.HydrogenConcentration = t.GasConcentration
}

func (t *CffMsgType) GetGasConcentration() int16 {
	return (int16(t.GasConcentration) * 500) - 5500
}

const CanInsulationMsg = 0x98FEA3B2

type InsulationType struct {
	InsulationStatusCode byte
	InsulationStatus     byte
	InsulationResistance uint16
	IsolationBattVolt    uint16
	IsolationLife        uint8
}

func (t *InsulationType) Load(data [8]byte) {
	t.InsulationStatusCode = data[0] & 0x0f
	t.InsulationStatus = (data[0] & 0x30) >> 4
	t.InsulationResistance = binary.LittleEndian.Uint16(data[1:3])
	t.IsolationBattVolt = binary.LittleEndian.Uint16(data[3:5])
	t.IsolationLife = data[7]

	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.InsulationResistance = t.InsulationResistance
}

func (t *InsulationType) getStatus() string {
	if t.InsulationStatus == 0 {
		return ""
	}
	switch t.InsulationStatusCode {
	case 0x0010:
		return "Normal Operation"
	case 0b0100:
		return "Wiring Fault"
	case 0b0101:
		return "The high voltage positive electrode has a small insulation resistance to the ground"
	case 0b0110:
		return "The high voltage negative electrode has a small insulation resistance to the ground"
	default:
		return "Unknown Status"
	}
}

func (t *InsulationType) getFault() string {
	switch t.InsulationStatus {
	case 0b11:
		return "Device Fault"
	case 0b01:
		return "level 1 alarm(resistance<100K)"
	case 0b10:
		return "level 2 alarm(resistance is between 100K-500K)"
	default:
		return "Normal"
	}
}

const CanStackCellsID1to4Msg = 0x9810A7B1
const CanStackCellsID5to8Msg = 0x9811A7B1
const CanStackCellsID9to12Msg = 0x9812A7B1
const CanStackCellsID13to16Msg = 0x9813A7B1
const CanStackCellsID17to20Msg = 0x9814A7B1
const CanStackCellsID21to24Msg = 0x9815A7B1
const CanStackCellsID25to28Msg = 0x9816A7B1
const CanStackCellsID29to32Msg = 0x9817A7B1
const CanMaxMinCellsMsg = 0x9801A7B1
const CanTotalStackVoltageMsg = 0x9802A7B1

type StackCellsType struct {
	StackCellVoltage           [5][32]uint16
	TotakStackVoltage          uint16
	StdDeviation               uint16
	Temperature                uint16
	StackControllerFaultStatus byte
	LifeSignal                 byte
	MinCellVolts               uint16
	MaxCellVolts               uint16
	AvgCellVolts               uint16
	IndexMaxVoltsCell          uint8
	IndexMinVoltsCell          uint8
	loop                       uint8
}

func (t *StackCellsType) Load(id uint32, data [8]byte) {
	switch id {
	case CanStackCellsID1to4Msg:
		t.loop++
		if t.loop > 4 {
			t.loop = 0
		}
		t.StackCellVoltage[t.loop][0] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][1] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][2] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][3] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID5to8Msg:
		t.StackCellVoltage[t.loop][4] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][5] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][6] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][7] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID9to12Msg:
		t.StackCellVoltage[t.loop][8] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][9] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][10] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][11] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID13to16Msg:
		t.StackCellVoltage[t.loop][12] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][13] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][14] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][15] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID17to20Msg:
		t.StackCellVoltage[t.loop][16] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][17] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][18] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][19] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID21to24Msg:
		t.StackCellVoltage[t.loop][20] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][21] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][22] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][23] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID25to28Msg:
		t.StackCellVoltage[t.loop][24] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][25] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][26] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][27] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID29to32Msg:
		t.StackCellVoltage[t.loop][28] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[t.loop][29] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[t.loop][30] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[t.loop][31] = binary.LittleEndian.Uint16(data[6:8])
	case CanMaxMinCellsMsg:
		t.MaxCellVolts = binary.LittleEndian.Uint16(data[0:2])
		t.MinCellVolts = binary.LittleEndian.Uint16(data[2:4])
		t.AvgCellVolts = binary.LittleEndian.Uint16(data[4:6])
		t.IndexMaxVoltsCell = data[6]
		t.IndexMinVoltsCell = data[7]
		dbRecord.mu.Lock()
		defer dbRecord.mu.Unlock()
		dbRecord.MaxCellVolts = t.MaxCellVolts
		dbRecord.MinCellVolts = t.MinCellVolts
		dbRecord.IdxMinCell = t.IndexMinVoltsCell
		dbRecord.IdxMaxCell = t.IndexMaxVoltsCell
		if t.loop == 0 {
			for i := 0; i < len(t.StackCellVoltage[0]); i++ {
				dbRecord.CellVoltages[i] = t.GetStackCellVoltage(i)
			}
		}
	case CanTotalStackVoltageMsg:
		t.TotakStackVoltage = binary.LittleEndian.Uint16(data[0:2])
		t.StdDeviation = binary.LittleEndian.Uint16(data[2:4])
		t.Temperature = binary.LittleEndian.Uint16(data[4:6])
		t.StackControllerFaultStatus = data[6]
		t.LifeSignal = data[7]
	}
}

func (t *StackCellsType) GetStackCellVoltage(cell int) int16 {
	var volts = int32(0)
	for idx := 0; idx < 5; idx++ {
		volts += int32(t.StackCellVoltage[idx][cell])
	}
	return int16(volts/5) - 5000
}

func (t *StackCellsType) GetMaxCellVoltage() int16 {
	return int16(t.MaxCellVolts) - 5000
}

func (t *StackCellsType) GetMinCellVoltage() int16 {
	return int16(t.MinCellVolts) - 5000
}

func (t *StackCellsType) GetAvgCellVoltage() int16 {
	return int16(t.AvgCellVolts) - 5000
}

const CanATSCoolingFanMsg = 0x99BBB701

type ATSCoolingFanType struct {
	Enable uint16
	Speed  uint16
}

func (t *ATSCoolingFanType) Load(data [8]byte) {
	t.Enable = binary.LittleEndian.Uint16(data[0:2])
	t.Speed = binary.LittleEndian.Uint16(data[2:4])

	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.CoolantFanSpeed = t.Speed
}

const CanWaterPumpMsg = 0x98FAC503

type WaterPumpType struct {
	Speed   uint16
	Voltage uint8
	Current uint8
}

func (t *WaterPumpType) Load(data [8]byte) {
	t.Speed = binary.LittleEndian.Uint16(data[0:2])
	t.Voltage = data[2]
	t.Current = data[3]

	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.CoolantPumpSpeed = t.Speed
	dbRecord.CoolantPumpAmps = t.Current
	dbRecord.CoolantPumpVolts = t.Voltage
}

func (t *WaterPumpType) getVoltage() float64 {
	return float64(t.Voltage) * 0.2
}

func (t *WaterPumpType) getCurrent() float64 {
	return float64(t.Current) * 0.2
}

const CanDCDCConverterMsg = 0x9029FF00

type DCDCConverterType struct {
	InputCurrent  uint16
	InputVoltage  uint16
	OutputCurrent uint16
	OutputVoltage uint16
}

func (t *DCDCConverterType) Load(data [8]byte) {
	t.InputCurrent = binary.LittleEndian.Uint16(data[0:2])
	t.InputVoltage = binary.LittleEndian.Uint16(data[2:4])
	t.OutputCurrent = binary.LittleEndian.Uint16(data[4:6])
	t.OutputVoltage = binary.LittleEndian.Uint16(data[6:8])
	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.DCDCInVolts = t.InputVoltage
	dbRecord.DCDCOutVolts = t.OutputVoltage
	dbRecord.DCDCInAmps = t.InputCurrent
	dbRecord.DCDCOutAmps = t.OutputCurrent
}

func (t *DCDCConverterType) GetInputCurrent() float64 {
	return float64(t.InputCurrent) / 10.0
}

func (t *DCDCConverterType) GetOutputCurrent() float64 {
	return float64(t.OutputCurrent) / 10.0
}

func (t *DCDCConverterType) GetInputVoltage() float64 {
	return float64(t.InputVoltage) / 100.0
}

func (t *DCDCConverterType) GetOutputVoltage() float64 {
	return float64(t.OutputVoltage) / 100.0
}

const CanDCOutputMsg = 0x98FFB587

type DCOutputType struct {
	Temp         uint8
	Status       uint8
	FaultLevel   uint8
	ErrorCode    byte
	OutVolltage  uint8
	OutCurrent   uint8
	InputVoltage uint8
	InternalTest uint8
	LIFE         uint8
}

func (t *DCOutputType) Load(data [8]byte) {
	t.Temp = data[0]
	t.Status = data[1] & 0x0f
	t.FaultLevel = (data[1] & 0xf0) >> 4
	t.ErrorCode = data[2]
	t.OutVolltage = data[3]
	t.OutCurrent = data[4]
	t.InputVoltage = data[5]
	t.InternalTest = data[6]
	t.LIFE = data[7]
	dbRecord.mu.Lock()
	defer dbRecord.mu.Unlock()
	dbRecord.DCDCTemp = t.Temp
}

func (t *DCOutputType) GetTemperature() int16 {
	return int16(t.Temp) - 40
}

func (t *DCOutputType) GetStatus() string {
	switch t.Status {
	case 0:
		return "Stop"
	case 1:
		return "Running"
	case 2:
		return "discharge/soft off state"
	default:
		return "fault"
	}
}

func (t *DCOutputType) GetFaultCode() string {
	switch t.ErrorCode {
	case 0x01:
		return "Communication failure 1 (master-slave 1)"
	case 0x02:
		return "Communication failure 2 (master-slave 2)"
	case 0x03:
		return "Customer order input error"
	case 0x04:
		return "output overcurrent"
	case 0x06:
		return "output voltage overvoltage"
	case 0x07:
		return "output voltage undervoltage"
	case 0x08:
		return "Input total current overcurrent"
	case 0x09:
		return "input overvoltage"
	case 0x0A:
		return "input undervoltage"
	case 0x0B:
		return "overheating"
	case 0x0C:
		return "Voltage relationship protection"
	case 0x0D:
		return "Maximum power protection"
	case 0x0E:
		return "Bus communication failure"
	case 0x0F:
		return "negative current protection"
	case 0x11:
		return "input precharge failed 1"
	case 0x12:
		return "input precharge failed 2"
	case 0x13:
		return "Output precharge failed 1"
	case 0x14:
		return "Output precharge failed 2"
	case 0x15:
		return "Input short circuit fault"
	case 0x16:
		return "Output short circuit fault"
	case 0xFF:
		return "Auxiliary electrical failure"
	case 0x18:
		return "BUS overvoltage"
	default:
		return ""
	}
}

const CanKeyOnMsg = 0x961088AD
const CanRunTimeMsg = 0x9610AAAB

type SystemInfoType struct {
	Run              bool
	ExhaustFlag      bool
	Hours            uint8
	Mins             uint8
	exhaustFlagTimer *time.Timer
	exhaustLastValue bool
}

func (t *SystemInfoType) SetRunFlag(data byte) {
	t.Run = data != 0
}

func (t *SystemInfoType) SetRunTime(hours byte, mins byte) {
	t.Hours = hours
	t.Mins = mins
}

func (t *SystemInfoType) SetExhaustFlag() {
	log.Println("Set Exhaust")
	t.ExhaustFlag = true
	t.exhaustFlagTimer.Reset(time.Second)
}

const CanBMSSettingsMsg = 0x9610AAAA

type BMSSettingsType struct {
	TargetPowerLevel uint8
	BMSHigh          uint16
	BMSLow           uint16
	CurrentPower     uint8
}

func (t *BMSSettingsType) Load(data [8]byte) {
	t.BMSHigh = binary.LittleEndian.Uint16(data[0:2])
	t.BMSLow = binary.LittleEndian.Uint16(data[2:4])
	t.TargetPowerLevel = data[4]
	t.CurrentPower = data[5]
}

type PANFuelCell struct {
	mu            sync.Mutex
	bus           *CANBus
	SystemInfo    SystemInfoType
	PowerMode     PowerModeType
	Pressures     PressuresType
	StackCoolant  StackCoolantType
	AirFlow       AirFlowType
	Alarms        AlarmsType
	StackOutput   StackOutputType
	CffMsg        CffMsgType
	Insulation    InsulationType
	StackCells    StackCellsType
	ATSCoolingFan ATSCoolingFanType
	WaterPump     WaterPumpType
	DCDCConverter DCDCConverterType
	DCOutput      DCOutputType
	BMSSettings   BMSSettingsType
	Control       PanSettingsType
}

func (fc *PANFuelCell) init(canBus *CANBus) {
	fc.bus = canBus
	fc.SystemInfo.exhaustFlagTimer = time.AfterFunc(time.Second, func() { fc.SystemInfo.ExhaustFlag = false })
}

func (fc *PANFuelCell) getJSON() (string, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	jsonBytes, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return "", err
	} else {
		return string(jsonBytes), nil
	}
}

func (fc *PANFuelCell) setTargetPower(kw float64) error {
	if (kw <= 10.0) && (kw >= 0) {
		fc.Control.TargetPower = kw
		return nil
	}
	return fmt.Errorf("valid range for target power is 0kW to 10kW. %01fkW was requested", kw)
}

func (fc *PANFuelCell) setTargetBattHigh(volts float64) error {
	if (volts >= 35) && (volts <= 70) && (volts >= fc.Control.TargetBatteryLow) {
		fc.Control.TargetBatteryHigh = volts
		return nil
	}
	return fmt.Errorf("valid range for battery voltage high is 35V to 70V and must be above or equal to battery voltage low. %01fV was requested", volts)
}

func (fc *PANFuelCell) setTargetBattLow(volts float64) error {
	if (volts >= 35) && (volts <= 70) {
		fc.Control.TargetBatteryLow = volts
		if volts > fc.Control.TargetBatteryHigh {
			fc.Control.TargetBatteryHigh = fc.Control.TargetBatteryLow
		}
		return nil
	}
	return fmt.Errorf("valid range for battery voltage low is 35V to 70V and must be below or equal to battery voltage high. %01fV was requested", volts)
}

func (fc *PANFuelCell) start() {
	fc.Control.FuelCellOn = true
	if err := fc.updateOutput(); err != nil {
		log.Println(err)
	}
	log.Println("Start the fuel cell")
}

func (fc *PANFuelCell) stop() {
	fc.Control.FuelCellOn = false
	if err := fc.updateOutput(); err != nil {
		log.Println(err)
	}
	log.Println("Stop the fuel cell")
}

func (fc *PANFuelCell) exhaustOpen() {
	fc.Control.Exhaust = true
	if err := fc.updateOutput(); err != nil {
		log.Println(err)
	}
	log.Println("Exhaust is open")
}

func (fc *PANFuelCell) exhaustClose() {
	fc.Control.Exhaust = false
	if err := fc.updateOutput(); err != nil {
		log.Println(err)
	}
	log.Println("Exhaust is closed")
}

/**
updateSettings sends the can control messages to the fuel cell
*/
func (fc *PANFuelCell) updateSettings() error {
	var limits BatteryVoltageLimitsType

	// Only send settings if the fuel cell is enabled
	if currentSettings.FuelCellSettings.Enabled {
		limits.BMSHighVoltage = uint16(fc.Control.TargetBatteryHigh * 10)
		limits.BMSLowVoltage = uint16(fc.Control.TargetBatteryLow * 10)
		limits.IsoFlag = currentSettings.FuelCellSettings.IgnoreIsoLow
		if fc.bus != nil {
			return limits.UpdateFuelCell(fc.bus)
		} else {
			return fmt.Errorf("no active CAN bus found for the fuel cell")
		}
	}
	return nil
}

func (fc *PANFuelCell) updateOutput() error {
	var output OutputControlType

	// Only send commands if the fuel cell is enabled
	if currentSettings.FuelCellSettings.Enabled {
		if fc.Control.FuelCellOn {
			output.FuelCellRunEnable = StartUp
			output.PowerDemand = uint8(fc.Control.TargetPower * 10)
		} else {
			output.FuelCellRunEnable = ShutDown
			output.PowerDemand = 0
		}
		if fc.Control.Exhaust {
			output.ExhaustMode = ExhaustOpen
		} else {
			output.ExhaustMode = ExhaustClosed
		}
		if fc.bus != nil {
			return output.UpdateFuelCell(fc.bus)
		} else {
			return fmt.Errorf("no active CAN bus found for the fuel cell")
		}
	}
	return nil
}

type PanStatus struct {
	SystemName           string
	RunTimeHours         uint16
	RunTimeMinutes       uint8
	RunState             bool
	H2Pressure           float64 // Hydrogen pressure
	AirPressure          float64 // Air pressure
	CoolantPressure      float64 // Coolant pressure
	H2AirPressureDiff    float64 // Hydrogen air pressure difference
	CoolantInletTemp     float64
	CoolantOutletTemp    float64
	AirTemp              float64
	AmbientTemp          float64
	AirFlow              float64
	StackVolts           float64
	StackCurrent         float64
	StackPower           float64
	DCInVolts            float64
	DCInAmps             float64
	DCOutVolts           float64
	DCOutAmps            float64
	BMSPower             float64
	BMSHigh              float64
	BMSLow               float64
	BMSCurrentPower      float64
	BMSTargetPower       float64
	BMSTargetHigh        float64
	BMSTargetLow         float64
	RunStatus            string
	Alarms               []string
	DCOutputStatus       string
	DCOutputFaultCode    string
	Start                bool
	ExhaustOpen          bool
	Enable               bool
	InsulationResistance uint16
	InsulationStatus     string
	InsulationFault      string
	WaterPumpSpeed       uint16
	WaterPumpActive      bool
	CoolingFanSpeed      uint16
}

/*
GetStatus sends a status block from the fuel cell
*/
func (fc *PANFuelCell) GetStatus() PanStatus {
	var status PanStatus
	fc.mu.Lock()
	defer fc.mu.Unlock()

	status.SystemName = currentSettings.Name
	status.RunTimeHours = uint16(fc.SystemInfo.Hours)
	status.RunTimeMinutes = fc.SystemInfo.Mins
	status.ExhaustOpen = fc.SystemInfo.ExhaustFlag
	status.RunState = fc.SystemInfo.Run
	status.Enable = currentSettings.FuelCellSettings.Enabled
	status.H2Pressure = (float64(fc.Pressures.H2Pressure) - 500) / 10.0
	status.AirPressure = (float64(fc.Pressures.AirPressure) - 500) / 10.0
	status.CoolantPressure = (float64(fc.Pressures.CoolantPressure) - 500) / 10.0
	status.H2AirPressureDiff = (float64(fc.Pressures.H2AirPressureDiff) - 50) / 10.0
	status.CoolantInletTemp = (float64(fc.StackCoolant.CoolantInTemp) - 400) / 10.0
	status.CoolantOutletTemp = (float64(fc.StackCoolant.CoolantOutTemp) - 400) / 10.0
	status.AirTemp = (float64(fc.StackCoolant.AirTemp) - 400) / 10.0
	status.AmbientTemp = (float64(fc.StackCoolant.AmbientTemp) - 400) / 10.0
	status.AirFlow = float64(fc.AirFlow.Flow) / 10.0
	status.StackVolts = float64(fc.StackOutput.Voltage) / 10.0
	status.StackCurrent = float64(fc.StackOutput.Current) / 10.0
	status.StackPower = float64(fc.StackOutput.Power) / 10.0
	status.DCInVolts = float64(fc.DCDCConverter.InputVoltage) / 100.0
	status.DCOutVolts = float64(fc.DCDCConverter.OutputVoltage) / 10.0
	status.DCInAmps = float64(fc.DCDCConverter.InputCurrent) / 10.0
	status.DCOutAmps = float64(fc.DCDCConverter.OutputCurrent) / 100.0
	status.BMSPower = float64(fc.BMSSettings.TargetPowerLevel)
	status.BMSHigh = float64(fc.BMSSettings.BMSHigh) / 10.0
	status.BMSLow = float64(fc.BMSSettings.BMSLow) / 10.0
	status.BMSCurrentPower = float64(fc.BMSSettings.CurrentPower)
	status.BMSTargetPower = fc.Control.TargetPower
	status.BMSTargetHigh = fc.Control.TargetBatteryHigh
	status.BMSTargetLow = fc.Control.TargetBatteryLow
	status.RunStatus = fc.PowerMode.PowerModeState.String()
	status.Alarms = fc.Alarms.Text()
	status.DCOutputStatus = fc.DCOutput.GetStatus()
	status.DCOutputFaultCode = fc.DCOutput.GetFaultCode()
	status.Start = fc.Control.FuelCellOn
	status.InsulationResistance = fc.Insulation.InsulationResistance
	status.InsulationStatus = fc.Insulation.getStatus()
	status.InsulationFault = fc.Insulation.getFault()
	status.WaterPumpSpeed = fc.WaterPump.Speed
	status.WaterPumpActive = fc.Control.PumpActive
	status.CoolingFanSpeed = fc.ATSCoolingFan.Speed
	return status
}

func (fc *PANFuelCell) GetStatusAsJSON() (string, error) {

	jsonBytes, err := json.MarshalIndent(fc.GetStatus(), "", "  ")
	if err != nil {
		return "", err
	} else {
		return string(jsonBytes), nil
	}
}

type PANDatabaseRecordType struct {
	StackCurrent          uint16
	StackVoltage          uint16
	CoolantInlTemp        uint16
	CoolantOutTemp        uint16
	OutputVoltage         uint16
	OutputCurrent         uint16
	CoolantFanSpeed       uint16
	CoolantPumpSpeed      uint16
	CoolantPumpVolts      uint8
	CoolantPumpAmps       uint8
	InsulationResistance  uint16
	HydrogenPressure      uint16
	AirPressure           uint16
	CoolantPressure       uint16
	AirinletTemp          uint16
	AmbientTemp           uint16
	AirFlow               uint16
	HydrogenConcentration uint8
	DCDCTemp              uint8
	DCDCInVolts           uint16
	DCDCOutVolts          uint16
	DCDCInAmps            uint16
	DCDCOutAmps           uint16
	MinCellVolts          uint16
	MaxCellVolts          uint16
	AvgCellVolts          uint16
	IdxMaxCell            uint8
	IdxMinCell            uint8
	RunStage              byte
	FaultLevel            byte
	PowerModeState        byte
	CellVoltages          [32]int16
	mu                    sync.Mutex
	stmt                  *sql.Stmt
}

var dbRecord PANDatabaseRecordType

func (rec *PANDatabaseRecordType) saveToDatabase() error {
	rec.mu.Lock()
	defer rec.mu.Unlock()
	_, err := rec.stmt.Exec(rec.StackCurrent, rec.StackVoltage, rec.CoolantInlTemp, rec.CoolantOutTemp, rec.OutputVoltage,
		rec.OutputCurrent, rec.CoolantFanSpeed, rec.CoolantPumpSpeed, rec.CoolantPumpVolts, rec.CoolantPumpAmps,
		rec.InsulationResistance, rec.HydrogenPressure, rec.AirPressure, rec.CoolantPressure, rec.AirinletTemp,
		rec.AmbientTemp, rec.AirFlow, rec.HydrogenConcentration, rec.DCDCTemp, rec.DCDCInVolts, rec.DCDCOutVolts,
		rec.DCDCInAmps, rec.DCDCOutAmps, rec.MinCellVolts, rec.MaxCellVolts, rec.AvgCellVolts, rec.IdxMaxCell,
		rec.IdxMinCell, rec.RunStage, rec.FaultLevel, rec.PowerModeState,
		rec.CellVoltages[0], rec.CellVoltages[1], rec.CellVoltages[2], rec.CellVoltages[3], rec.CellVoltages[4],
		rec.CellVoltages[5], rec.CellVoltages[6], rec.CellVoltages[7], rec.CellVoltages[8], rec.CellVoltages[9],
		rec.CellVoltages[10], rec.CellVoltages[11], rec.CellVoltages[12], rec.CellVoltages[13], rec.CellVoltages[14],
		rec.CellVoltages[15], rec.CellVoltages[16], rec.CellVoltages[17], rec.CellVoltages[18], rec.CellVoltages[19],
		rec.CellVoltages[20], rec.CellVoltages[21], rec.CellVoltages[22], rec.CellVoltages[23], rec.CellVoltages[24],
		rec.CellVoltages[25], rec.CellVoltages[26], rec.CellVoltages[27], rec.CellVoltages[28], rec.CellVoltages[29],
		rec.CellVoltages[30], rec.CellVoltages[31])
	if err != nil {
		log.Println(err)
	}
	return err
}
