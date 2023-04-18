package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/brutella/can"
	"log"
	"sync"
)

type PanSettingsType struct {
	TargetPower       float64
	TargetBatteryHigh float64
	TargetBatteryLow  float64
	FuelCellOn        bool
}

type RunCommandType byte

const (
	StartUp = iota + 1
	ShutDown
)

const CanOutputControlMsg = 0x961088C1

type OutputControlType struct {
	FuelCellRunEnable RunCommandType // Run / Stop
	PowerDemand       uint8          // kW x 10
	ExhaustMode       byte           // ?
}

func (t *OutputControlType) Load(data [8]byte) {
	if data[0] == 1 {
		t.FuelCellRunEnable = StartUp
	} else {
		t.FuelCellRunEnable = ShutDown
	}
	t.PowerDemand = data[1]
	t.ExhaustMode = data[2]
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
	frame.Data[2] = t.ExhaustMode
	return bus.bus.Publish(frame)
}

const CanBatterVoltageLimitsMsg = 0x961088C2

type BatteryVoltageLimitsType struct {
	BMSHightVoltage uint16 //Battery high voltage setpoint
	BMSLowVoltage   uint16 //Battery low voltage setpoint
}

func (t *BatteryVoltageLimitsType) Load(data [8]byte) {
	t.BMSHightVoltage = binary.LittleEndian.Uint16(data[0:2])
	t.BMSLowVoltage = binary.LittleEndian.Uint16(data[2:4])
}

func (t *BatteryVoltageLimitsType) LoadLimits(HighLimit float64, LowLimit float64) {
	t.BMSHightVoltage = uint16(HighLimit * 10)
	t.BMSLowVoltage = uint16(LowLimit * 10)
	log.Printf("Fuel cell high limit = %d : fuel cell low limit  = %d", t.BMSHightVoltage, t.BMSLowVoltage)
}

func (t *BatteryVoltageLimitsType) GetBMSHighVoltage() float64 {
	return float64(t.GetBMSHighVoltage()) / 10.0
}

func (t *BatteryVoltageLimitsType) GetBMSLowVoltage() float64 {
	return float64(t.GetBMSLowVoltage()) / 10.0
}

func (t *BatteryVoltageLimitsType) UpdateFuelCell(bus *CANBus) error {
	var frame can.Frame

	frame.ID = CanBatterVoltageLimitsMsg
	frame.Length = 8
	binary.LittleEndian.PutUint16(frame.Data[0:2], t.BMSHightVoltage)
	binary.LittleEndian.PutUint16(frame.Data[2:4], t.BMSLowVoltage)
	return bus.bus.Publish(frame)
}

type PowerModeStateType byte

const (
	PMOff = iota
	PMInit
	PMH2Purge
	PMStartup
	PMAirPurge
	PMH2LeakCheck
	PMManual
	PMEmergencyShut
	PMFault
	PM_hutdown
)

func (pm PowerModeStateType) String() string {
	modeStates := [...]string{"Off", "Standby", "Hydrogen intake", "Start", "AirPurge", "Hydrogen leak cheak", "manual", "emergency stop", "fault", "shutdown"}
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

func (t *AlarmsType) Load(data [8]byte) {
	t.VoltageLow = (data[0] & 1) != 0
	t.H2concentration = (data[0] & 2) != 0
	t.CoolantTempOutDiff = (data[0] & 4) != 0
	t.CoolantTempHigh = (data[0] & 8) != 0
	t.WaterPumpFault = (data[0] & 0x10) != 0
	t.H2CirculatingPumpFault = (data[0] & 0x20) != 0
	t.H2PressureHigh = (data[0] & 0x40) != 0
	t.H2PressureSensorFault = (data[0] & 0x80) != 0

	t.DcdcCommunicationFault = (data[1] & 0x01) != 0
	t.DcdcFault = (data[1] & 0x02) != 0
	t.PtcFault = (data[1] & 0x04) != 0
	t.H2TankTemp = (data[1] & 0x08) != 0
	t.H2TankHighPressure = (data[1] & 0x10) != 0
	t.H2TankMidPressure = (data[1] & 0x20) != 0
	t.H2TankLowPressure = (data[1] & 0x40) != 0
	t.FcuToVcuFault = (data[1] & 0x80) != 0

	t.TempSensorFault = (data[2] & 0x01) != 0
	t.H2SPCheckFault = (data[2] & 0x02) != 0
	t.H2SOCLow = (data[2] & 0x04) != 0
	t.H2OutPressureLow = (data[2] & 0x08) != 0
	t.AirPressureLow = (data[2] & 0x10) != 0
	t.AirPressureHigh = (data[2] & 0x20) != 0
	t.AirTempHigh = (data[2] & 0x40) != 0
	t.CoolantPressureHigh = (data[2] & 0x80) != 0

	t.CellVoltageHigh = (data[3] & 0x01) != 0
	t.IsoLow = (data[3] & 0x02) != 0
	t.H2AirDiffHighMinus = (data[3] & 0x04) != 0
	t.H2AirDiffHighPlus = (data[3] & 0x08) != 0
	t.StartUploss = (data[3] & 0x10) != 0
	t.H2Leakageloss = (data[3] & 0x20) != 0
}

func (al *AlarmsType) Text() []string {
	alarmText := make([]string, 0)
	if al.AirPressureLow {
		alarmText = append(alarmText, "Low Air Pressure")
	}
	if al.AirPressureHigh {
		alarmText = append(alarmText, "High AirPressure")
	}
	if al.AirTempHigh {
		alarmText = append(alarmText, "High Air Temperature")
	}
	if al.CoolantTempOutDiff {
		alarmText = append(alarmText, "Coolant Temperature Output Differential")
	}
	if al.CoolantTempHigh {
		alarmText = append(alarmText, "High Coolant Temperature")
	}
	if al.CoolantPressureHigh {
		alarmText = append(alarmText, "High Coolant Pressure")
	}
	if al.CellVoltageHigh {
		alarmText = append(alarmText, "High Cell Voltage")
	}
	if al.DcdcFault {
		alarmText = append(alarmText, "DC to DCConverter Fault")
	}
	if al.DcdcCommunicationFault {
		alarmText = append(alarmText, "DC to DC Converter Communication Fault")
	}
	if al.FcuToVcuFault {
		alarmText = append(alarmText, "FCU to VCU Fault")
	}
	if al.H2AirDiffHighMinus {
		alarmText = append(alarmText, "H2 Air Differentinal High Minus")
	}
	if al.H2AirDiffHighPlus {
		alarmText = append(alarmText, "H2 Air Differential High Plus")
	}
	if al.H2concentration {
		alarmText = append(alarmText, "H2 Concentration")
	}
	if al.H2CirculatingPumpFault {
		alarmText = append(alarmText, "H2 Circulation Pump Fault")
	}
	if al.H2Leakageloss {
		alarmText = append(alarmText, "H2 Leakage Loss")
	}
	if al.H2OutPressureLow {
		alarmText = append(alarmText, "H2 Outlet Pressure Low")
	}
	if al.H2PressureHigh {
		alarmText = append(alarmText, "H2 Pressure High")
	}
	if al.H2PressureSensorFault {
		alarmText = append(alarmText, "H2 Pressure Sensor Fault")
	}
	if al.H2SOCLow {
		alarmText = append(alarmText, "H2 SOC Low")
	}
	if al.H2SPCheckFault {
		alarmText = append(alarmText, "H2 SP Check Fault")
	}
	if al.H2TankLowPressure {
		alarmText = append(alarmText, "H2 Tank Low Pressure")
	}
	if al.H2TankMidPressure {
		alarmText = append(alarmText, "H2 Tank Mid Pressure")
	}
	if al.H2TankHighPressure {
		alarmText = append(alarmText, "H2 Tank High Pressure")
	}
	if al.H2TankTemp {
		alarmText = append(alarmText, "H2 Tank Temperature")
	}
	if al.IsoLow {
		alarmText = append(alarmText, "Iso Low")
	}
	if al.PtcFault {
		alarmText = append(alarmText, "PCT Fault")
	}
	if al.StartUploss {
		alarmText = append(alarmText, "Startup Loss")
	}
	if al.TempSensorFault {
		alarmText = append(alarmText, "Temperature Sensor Fault")
	}
	if al.VoltageLow {
		alarmText = append(alarmText, "Voltage Low")
	}
	if al.WaterPumpFault {
		alarmText = append(alarmText, "Water Pump Fault")
	}
	return alarmText
}

const CanStackOutputMsg = 0x961088A7

type StackOutputType struct {
	Voltage uint16 //Stack voltage
	Current uint16 //Stack current
	Power   uint16 //Stack power
}

func (t *StackOutputType) Load(data [8]byte) {
	t.Voltage = binary.LittleEndian.Uint16(data[0:2])
	t.Current = binary.LittleEndian.Uint16(data[2:4])
	t.Power = binary.LittleEndian.Uint16(data[4:6])
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
}

func (t *CffMsgType) GetGasConcentration() int16 {
	return (int16(t.GasConcentration) * 500) - 5500
}

const CanInsulationMsg = 0x80000200

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
	StackCellVoltage           [32]uint16
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
}

func (t *StackCellsType) Load(id uint32, data [8]byte) {
	switch id {
	case CanStackCellsID1to4Msg:
		t.StackCellVoltage[0] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[1] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[2] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[3] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID5to8Msg:
		t.StackCellVoltage[4] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[5] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[6] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[7] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID9to12Msg:
		t.StackCellVoltage[8] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[9] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[10] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[11] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID13to16Msg:
		t.StackCellVoltage[12] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[13] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[14] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[15] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID17to20Msg:
		t.StackCellVoltage[16] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[17] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[18] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[19] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID21to24Msg:
		t.StackCellVoltage[20] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[21] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[22] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[23] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID25to28Msg:
		t.StackCellVoltage[24] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[25] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[26] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[27] = binary.LittleEndian.Uint16(data[6:8])
	case CanStackCellsID29to32Msg:
		t.StackCellVoltage[28] = binary.LittleEndian.Uint16(data[0:2])
		t.StackCellVoltage[29] = binary.LittleEndian.Uint16(data[2:4])
		t.StackCellVoltage[30] = binary.LittleEndian.Uint16(data[4:6])
		t.StackCellVoltage[31] = binary.LittleEndian.Uint16(data[6:8])
	case CanMaxMinCellsMsg:
		t.MaxCellVolts = binary.LittleEndian.Uint16(data[0:2])
		t.MinCellVolts = binary.LittleEndian.Uint16(data[2:4])
		t.AvgCellVolts = binary.LittleEndian.Uint16(data[4:6])
		t.IndexMaxVoltsCell = data[6]
		t.IndexMinVoltsCell = data[7]
	case CanTotalStackVoltageMsg:
		t.TotakStackVoltage = binary.LittleEndian.Uint16(data[0:2])
		t.StdDeviation = binary.LittleEndian.Uint16(data[2:4])
		t.Temperature = binary.LittleEndian.Uint16(data[4:6])
		t.StackControllerFaultStatus = data[6]
		t.LifeSignal = data[7]
	}
}

func (t *StackCellsType) GetStackCellVoltage(cell int) int16 {
	return int16(t.StackCellVoltage[cell]) - 5000
}

func (t *StackCellsType) GetMaxCellVoltage(cell int) int16 {
	return int16(t.MaxCellVolts) - 5000
}

func (t *StackCellsType) GetMinCellVoltage(cell int) int16 {
	return int16(t.MinCellVolts) - 5000
}

func (t *StackCellsType) GetAvgCellVoltage(cell int) int16 {
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
	DCTemp        uint8
	DCFaultLevel  uint8
	DCFaultCode   byte
	DCOutVolltage uint8
	DCOutCurrent  uint8
}

func (t *DCOutputType) Load(data [8]byte) {
	t.DCTemp = data[0]
	t.DCFaultLevel = data[1] & 0x0f
	t.DCFaultCode = data[2]
	t.DCOutVolltage = data[3]
	t.DCOutCurrent = data[4]
}

func (t *DCOutputType) GetTemperature() int16 {
	return int16(t.DCTemp) - 40
}

const CanBMSSettingsMsg = 0x9610AAAA

type BMSSettingsType struct {
	TargetPowerLevel uint8
	BMSHigh          uint8
	BMSLow           uint8
	CurrentPower     uint8
}

func (t *BMSSettingsType) Load(data [8]byte) {
	t.TargetPowerLevel = data[0]
	t.BMSHigh = data[1]
	t.BMSLow = data[2]
	t.CurrentPower = data[3]
}

type PANFuelCell struct {
	mu                   sync.Mutex
	bus                  *CANBus
	OutputControl        OutputControlType
	BatteryVoltageLimits BatteryVoltageLimitsType
	PowerMode            PowerModeType
	Pressures            PressuresType
	StackCoolant         StackCoolantType
	AirFlow              AirFlowType
	Alarms               AlarmsType
	StackOutput          StackOutputType
	CffMsg               CffMsgType
	Insulation           InsulationType
	StackCells           StackCellsType
	ATSCoolingFan        ATSCoolingFanType
	WaterPump            WaterPumpType
	DCDCConverter        DCDCConverterType
	DCOutput             DCOutputType
	BMSSettings          BMSSettingsType
	Control              PanSettingsType
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

func (fc *PANFuelCell) start() error {
	fc.Control.FuelCellOn = true
	return fc.updateOutput()
}

func (fc *PANFuelCell) stop() error {
	fc.Control.FuelCellOn = false
	return fc.updateOutput()
}

/**
updateSettings sends the can control messages to the fuel cell
*/
func (fc *PANFuelCell) updateSettings() error {
	var limits BatteryVoltageLimitsType

	limits.BMSHightVoltage = uint16(fc.Control.TargetBatteryHigh * 10)
	limits.BMSLowVoltage = uint16(fc.Control.TargetBatteryLow * 10)
	if fc.bus != nil {
		return limits.UpdateFuelCell(fc.bus)
	} else {
		return fmt.Errorf("No active CAN bus found for the fuel cell.")
	}
}

func (fc *PANFuelCell) updateOutput() error {
	var output OutputControlType

	output.PowerDemand = uint8(fc.Control.TargetPower * 10)
	if fc.Control.FuelCellOn {
		output.FuelCellRunEnable = StartUp
	} else {
		output.FuelCellRunEnable = ShutDown
	}
	if fc.bus != nil {
		return output.UpdateFuelCell(fc.bus)
	} else {
		return fmt.Errorf("No active CAN bus found for the fuel cell.")
	}
}

//func (fc *PANFuelCell) saveToSettingsFile() {
//	currentSettings.FuelCellBatteryLimits.PowerSetting = fc.Control.TargetPower
//	currentSettings.FuelCellBatteryLimits.HighBatterySetpoint = fc.Control.TargetBatteryHigh
//	currentSettings.FuelCellBatteryLimits.LowBatterySetpoint = fc.Control.TargetBatteryLow
//	currentSettings.SaveSettings(currentSettings.filepath)
//
//}

type PanStatus struct {
	H2Pressure        float64 // Hydrogen pressure
	AirPressure       float64 // Air pressure
	CoolantPressure   float64 // Coolant pressure
	H2AirPressureDiff float64 // Hydrogen air pressure difference
	CoolantInletTemp  float64
	CoolantOutletTemp float64
	AirTemp           float64
	AmbientTemp       float64
	AirFlow           float64
	StackVolts        float64
	StackCurrent      float64
	StackPower        float64
	DCInVolts         float64
	DCInAmps          float64
	DCOutVolts        float64
	DCOutAmps         float64
	BMSPower          float64
	BMSHigh           float64
	BMSLow            float64
	BMSCurrentPower   float64
	BMSTargetPower    float64
	BMSTargetHigh     float64
	BMSTargetLow      float64
	RunStatus         string
	Alarms            []string
}

// Send a status block from the fuel cell
func (fc *PANFuelCell) GetStatus() PanStatus {
	var status PanStatus
	fc.mu.Lock()
	defer fc.mu.Unlock()

	status.H2Pressure = float64(fc.Pressures.H2Pressure)
	status.AirPressure = float64(fc.Pressures.AirPressure)
	status.CoolantPressure = float64(fc.Pressures.CoolantPressure)
	status.H2AirPressureDiff = float64(fc.Pressures.H2AirPressureDiff)
	status.CoolantInletTemp = float64(fc.StackCoolant.CoolantInTemp) / 10.0
	status.CoolantOutletTemp = float64(fc.StackCoolant.CoolantOutTemp) / 10.0
	status.AirTemp = float64(fc.StackCoolant.AirTemp) / 10.0
	status.AmbientTemp = float64(fc.StackCoolant.AmbientTemp) / 10.0
	status.AirFlow = float64(fc.AirFlow.Flow) / 10.0
	status.StackVolts = float64(fc.StackOutput.Voltage) / 10.0
	status.StackCurrent = float64(fc.StackOutput.Current) / 10.0
	status.StackPower = float64(fc.StackOutput.Power) / 10.0
	status.DCInVolts = float64(fc.DCDCConverter.InputVoltage) / 100.0
	status.DCOutVolts = float64(fc.DCDCConverter.OutputVoltage) / 10.0
	status.DCInAmps = float64(fc.DCDCConverter.InputCurrent) / 10.0
	status.DCOutAmps = float64(fc.DCDCConverter.OutputCurrent) / 100.0
	status.BMSPower = float64(fc.BMSSettings.TargetPowerLevel)
	status.BMSHigh = float64(fc.BMSSettings.BMSHigh)
	status.BMSLow = float64(fc.BMSSettings.BMSLow)
	status.BMSCurrentPower = float64(fc.BMSSettings.CurrentPower)
	status.BMSTargetPower = fc.Control.TargetPower
	status.BMSTargetHigh = fc.Control.TargetBatteryHigh
	status.BMSTargetLow = fc.Control.TargetBatteryLow
	status.RunStatus = fc.PowerMode.PowerModeState.String()
	status.Alarms = fc.Alarms.Text()

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
