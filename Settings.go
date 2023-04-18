package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type AnalogSettingType struct {
	Name                   string
	Port                   uint8
	LowerCalibrationActual float32
	LowerCalibrationAtoD   uint16
	UpperCalibrationActual float32
	UpperCalibrationAtoD   uint16
	calibrationConstant    float32
	calibrationMultiplier  float32
}

type PortNameType struct {
	Name string
	Port uint8
}

type FuelCellBatteryLimitsType struct {
	HighBatterySetpoint float64
	LowBatterySetpoint  float64
	PowerSetting        float64
}

type SettingsType struct {
	Name                  string
	AnalogChannels        [8]AnalogSettingType
	DigitalInputs         [4]PortNameType
	DigitalOutputs        [6]PortNameType
	Relays                [16]PortNameType
	FuelCellBatteryLimits FuelCellBatteryLimitsType
	filepath              string
}

func NewSettings() *SettingsType {
	settings := new(SettingsType)
	settings.Name = "FireflyIO"
	for idx := range settings.AnalogChannels {
		settings.AnalogChannels[idx].Port = uint8(idx)
		settings.AnalogChannels[idx].Name = fmt.Sprintf("Analog-%d", idx)
		settings.AnalogChannels[idx].UpperCalibrationActual = 1024
		settings.AnalogChannels[idx].UpperCalibrationAtoD = 1024
		settings.AnalogChannels[idx].LowerCalibrationActual = 0
		settings.AnalogChannels[idx].LowerCalibrationAtoD = 0
		settings.AnalogChannels[idx].calculateConstants()
	}
	for idx := range settings.DigitalInputs {
		settings.DigitalInputs[idx].Port = uint8(idx)
		settings.DigitalInputs[idx].Name = fmt.Sprintf("Intput-%d", idx)
	}

	for idx := range settings.DigitalOutputs {
		settings.DigitalOutputs[idx].Port = uint8(idx)
		settings.DigitalOutputs[idx].Name = fmt.Sprintf("Output-%d", idx)
	}

	for idx := range settings.Relays {
		settings.Relays[idx].Port = uint8(idx)
		settings.Relays[idx].Name = fmt.Sprintf("Relay-%d", idx)
	}
	return settings
}

func (settings *SettingsType) LoadSettings(filepath string) error {
	if file, err := ioutil.ReadFile(filepath); err != nil {
		log.Println(err)
		if err := settings.SaveSettings(filepath); err != nil {
			return err
		}
	} else {
		settings.filepath = filepath
		if err := json.Unmarshal(file, settings); err != nil {
			return err
		}
	}
	settings.filepath = filepath
	settings.calculateConstants()
	for _, rl := range settings.Relays {
		Relays.Relays[rl.Port].Name = rl.Name
	}
	for _, op := range settings.DigitalOutputs {
		Outputs.Outputs[op.Port].Name = op.Name
	}
	for _, ip := range settings.DigitalInputs {
		Inputs.Inputs[ip.Port].Name = ip.Name
	}
	for _, analog := range settings.AnalogChannels {
		AnalogInputs.Inputs[analog.Port].Name = analog.Name
	}
	return nil
}

func (settings *SettingsType) SaveSettings(filepath string) error {
	settings.filepath = filepath
	if bData, err := json.Marshal(settings); err != nil {
		log.Println("Error converting settings to text -", err)
		return err
	} else {
		if err = ioutil.WriteFile(settings.filepath, bData, 0644); err != nil {
			log.Println("Error writing JSON settings file -", err)
			return err
		}
	}
	return nil
}

func (AnalogSetting *AnalogSettingType) calculateConstants() {
	AnalogSetting.calibrationMultiplier = (AnalogSetting.UpperCalibrationActual - AnalogSetting.LowerCalibrationActual) / float32(AnalogSetting.UpperCalibrationAtoD-AnalogSetting.LowerCalibrationAtoD)
	AnalogSetting.calibrationConstant = AnalogSetting.LowerCalibrationActual - (float32(AnalogSetting.LowerCalibrationAtoD) * AnalogSetting.calibrationMultiplier)
}

func (settings *SettingsType) calculateConstants() {
	for idx := range settings.AnalogChannels {
		settings.AnalogChannels[idx].calculateConstants()
	}
}

func (settings *SettingsType) SendSettingsJSON(w http.ResponseWriter) {
	if bData, err := json.Marshal(settings); err != nil {
		log.Println("Error converting settings to text -", err)
	} else {
		fmt.Fprint(w, string(bData))
	}
}

func (settings *SettingsType) LoadFromJSON(jsonData []byte) error {
	if err := json.Unmarshal(jsonData, settings); err != nil {
		return err
	} else {
		return nil
	}
}
