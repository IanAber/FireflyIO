package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

func setUpWebSite() {
	log.Println("Staring the WEB site on port ", WebPort)
	router := mux.NewRouter().StrictSlash(true)
	// Register with the WebSocket which will then push a JSON payload with data to keep the displayed data up to date. No polling necessary.
	router.HandleFunc("/ws", startDataWebSocket).Methods("GET")

	router.HandleFunc("/setRelay/{relay}/{on}", setRelay).Methods("PUT")
	router.HandleFunc("/setOutput/{output}/{on}", setOutput).Methods("PUT")
	router.HandleFunc("/getSettings", getSettings).Methods("GET")
	router.HandleFunc("/setSettings", setSettings).Methods("POST")
	router.HandleFunc("/getStatus", getStatus).Methods("GET")
	router.HandleFunc("/getFuelCell", getFuelCell).Methods("GET")
	router.HandleFunc("/setFuelCell/TargetPower/{power}", setFcPower).Methods("PUT")
	router.HandleFunc("/setFuelCell/TargetBattHigh/{volts}", setFcBattHigh).Methods("PUT")
	router.HandleFunc("/setFuelCell/TargetBattLow/{volts}", setFcBatLow).Methods("PUT")
	router.HandleFunc("/setFuelCell/Start", startFc).Methods("PUT")
	router.HandleFunc("/setFuelCell/Stop}", stopFc).Methods("PUT")
	router.HandleFunc("/setFuelCellSettings", setFuelCellSettings).Methods("POST")
	router.HandleFunc("/unknown", getUnknownFrames).Methods("GET")

	fileServer := http.FileServer(neuteredFileSystem{http.Dir(webFiles)})
	router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))

	log.Println("Starting WEB server")
	port := fmt.Sprintf(":%s", WebPort)
	log.Fatal(http.ListenAndServe(port, router))
}

func setFcPower(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	request := vars["power"]
	const function = "Set Fuel Cell Power"

	fPower, err := strconv.ParseFloat(request, 64)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	}
	log.Println("set fuel cell power to ", fPower)
	err = FuelCell.setTargetPower(fPower)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	}
	getFuelCell(w, r)
}

func setFcBattHigh(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	request := vars["volts"]
	const function = "Set Fuel Cell Batt High"

	fVolts, err := strconv.ParseFloat(request, 64)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	}
	log.Println("set fuel cell high battery limit to ", fVolts)
	err = FuelCell.setTargetBattHigh(fVolts)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	}
	if err = FuelCell.updateSettings(); err != nil {
		ReturnJSONError(w, function, err, http.StatusInternalServerError, true)
	}
	getFuelCell(w, r)
}

func setFcBatLow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	request := vars["volts"]
	const function = "Set Fuel Cell Batt Low"

	fVolts, err := strconv.ParseFloat(request, 64)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	}
	log.Println("set fuel cell low battery limit to ", fVolts)
	err = FuelCell.setTargetBattLow(fVolts)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	}
	if err = FuelCell.updateSettings(); err != nil {
		ReturnJSONError(w, function, err, http.StatusInternalServerError, true)
	}
	getFuelCell(w, r)
}

func startFc(w http.ResponseWriter, r *http.Request) {
	FuelCell.start()
	getFuelCell(w, r)
}

func stopFc(w http.ResponseWriter, r *http.Request) {
	FuelCell.stop()
	getFuelCell(w, r)
}

func setFuelCellSettings(w http.ResponseWriter, r *http.Request) {
	const function = "Set Fuel Cell Settings"
	if err := r.ParseForm(); err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
		return
	}
	if floatval, err := strconv.ParseFloat(r.FormValue("PowerDemand"), 64); err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	} else {
		currentSettings.FuelCellBatteryLimits.PowerSetting = floatval
		FuelCell.Control.TargetPower = floatval
	}
	if floatval, err := strconv.ParseFloat(r.FormValue("LowBattDemand"), 64); err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	} else {
		currentSettings.FuelCellBatteryLimits.LowBatterySetpoint = floatval
		FuelCell.Control.TargetBatteryLow = floatval
	}
	if floatval, err := strconv.ParseFloat(r.FormValue("HighBattDemand"), 64); err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	} else {
		currentSettings.FuelCellBatteryLimits.HighBatterySetpoint = floatval
		FuelCell.Control.TargetBatteryHigh = floatval
	}
	currentSettings.SaveSettings(currentSettings.filepath)
	FuelCell.updateSettings() // Update the battery limit settings
	FuelCell.updateOutput()   // Update the power setting
	http.Redirect(w, r, "/FuelCellSettings.html", http.StatusTemporaryRedirect)

}

func setRelay(w http.ResponseWriter, r *http.Request) {
	var bOn bool
	vars := mux.Vars(r)
	relay := vars["relay"]
	on := vars["on"]

	on = strings.ToLower(on)
	if (on == "on") || (on == "true") || (on == "1") {
		bOn = true
	} else if (on == "off") || (on == "false") || (on == "0") {
		bOn = false
	} else {
		ReturnJSONErrorString(w, "setRelay", "Invalid value given for relay setting. Valid values are on, true, 1, off, false or 0", http.StatusBadRequest, true)
		return
	}
	relayNum, err := strconv.ParseInt(relay, 10, 8)
	if err != nil {
		if err := Relays.SetRelayByName(relay, bOn); err != nil {
			ReturnJSONError(w, "setRelay", err, http.StatusBadRequest, true)
		}
	} else {
		if (relayNum >= 0) && (relayNum < int64(len(Relays.Relays))) {
			Relays.SetRelay(uint8(relayNum), bOn)
		} else {
			ReturnJSONErrorString(w, "setRelay", fmt.Sprintf("Invalid relay number - %d", relayNum), http.StatusBadRequest, true)
		}
	}
}

func setOutput(w http.ResponseWriter, r *http.Request) {
	var bOn bool
	vars := mux.Vars(r)
	output := vars["output"]
	on := vars["on"]

	on = strings.ToLower(on)
	if (on == "on") || (on == "true") || (on == "1") {
		bOn = true
	} else if (on == "off") || (on == "false") || (on == "0") {
		bOn = false
	} else {
		ReturnJSONErrorString(w, "setOutput", "Invalid value given for output setting. Valid values are on, true, 1, off, false or 0", http.StatusBadRequest, true)
		return
	}
	outputNum, err := strconv.ParseInt(output, 10, 8)
	if err != nil {
		if err := Outputs.SetOutputByName(output, bOn); err != nil {
			ReturnJSONError(w, "setOutput", err, http.StatusBadRequest, true)
		}
	} else {
		if (outputNum >= 0) && (outputNum < int64(len(Outputs.Outputs))) {

			Outputs.SetOutput(uint8(outputNum), bOn)
		} else {
			ReturnJSONErrorString(w, "setOutput", fmt.Sprintf("Invalid output number - %d", outputNum), http.StatusBadRequest, true)
		}
	}
}

type JsonDataType struct {
	System            string
	Version           string
	Relays            *RelaysType
	Analog            *AnalogInputsType
	DigitalOut        *DigitalOutputsType
	DigitalIn         *DigitalInputsType
	ACVolts           float32
	ACAmps            float32
	ACWatts           float32
	ACWattHours       uint32
	ACHertz           float32
	ACPowerFactor     float32
	PanFuelCellStatus PanStatus
}

func getJsonStatus() (string, error) {
	var data JsonDataType

	data.System = currentSettings.Name
	data.Version = version
	data.Relays = &Relays
	data.DigitalIn = &Inputs
	data.DigitalOut = &Outputs
	data.Analog = &AnalogInputs
	data.ACVolts = ACMeasurements.getVolts()
	data.ACAmps = ACMeasurements.getAmps()
	data.ACWatts = ACMeasurements.getPower()
	data.ACWattHours = ACMeasurements.getEnergy()
	data.ACHertz = ACMeasurements.getFrequency()
	data.ACPowerFactor = ACMeasurements.getPowerFactor()
	data.PanFuelCellStatus = FuelCell.GetStatus()

	JSONBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	} else {
		return string(JSONBytes), nil
	}
}

func getSettings(w http.ResponseWriter, _ *http.Request) {
	currentSettings.SendSettingsJSON(w)
}

func setSettings(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ReturnJSONError(w, "setSettings", err, http.StatusBadRequest, true)
		return
	}
	// System name
	currentSettings.Name = r.FormValue("name")
	// Relay names
	for relay := 0; relay < 16; relay++ {
		currentSettings.Relays[relay].Name = r.FormValue(fmt.Sprintf("relay%dname", relay))
	}
	// Digital Input names
	for din := 0; din < 4; din++ {
		currentSettings.DigitalInputs[din].Name = r.FormValue(fmt.Sprintf("di%dname", din))
	}
	// Digital output names
	for dout := 0; dout < 4; dout++ {
		currentSettings.DigitalOutputs[dout].Name = r.FormValue(fmt.Sprintf("do%dname", dout))
	}
	// Analogue names and settings
	for analog := range currentSettings.AnalogChannels {
		currentSettings.AnalogChannels[analog].Name = r.FormValue(fmt.Sprintf("a%dname", analog))
		if f, err := strconv.ParseFloat(r.FormValue(fmt.Sprintf("a%dLowVal", analog)), 32); err != nil {
			log.Println(err)
		} else {
			currentSettings.AnalogChannels[analog].LowerCalibrationActual = float32(f)
		}
		if f, err := strconv.ParseFloat(r.FormValue(fmt.Sprintf("a%dHighVal", analog)), 32); err != nil {
			log.Println(err)
		} else {
			currentSettings.AnalogChannels[analog].UpperCalibrationActual = float32(f)
		}
		if f, err := strconv.ParseInt(r.FormValue(fmt.Sprintf("a%dLowA2D", analog)), 10, 32); err != nil {
			log.Println(err)
		} else {
			currentSettings.AnalogChannels[analog].LowerCalibrationAtoD = uint16(f)
		}
		if f, err := strconv.ParseInt(r.FormValue(fmt.Sprintf("a%dHighA2D", analog)), 10, 32); err != nil {
			log.Println(err)
		} else {
			currentSettings.AnalogChannels[analog].UpperCalibrationAtoD = uint16(f)
		}
		currentSettings.AnalogChannels[analog].calculateConstants()
	}
	currentSettings.SaveSettings(currentSettings.filepath)
	currentSettings.LoadSettings(currentSettings.filepath)

	http.Redirect(w, r, "/config.html", http.StatusTemporaryRedirect)
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	sJSON, err := getJsonStatus()
	setContentTypeHeader(w)
	_, err = fmt.Fprint(w, sJSON)
	if err != nil {
		log.Println("failed to send the status - ", err)
		return
	}
}

func getFuelCell(w http.ResponseWriter, r *http.Request) {
	strStatus, err := FuelCell.GetStatusAsJSON()
	setContentTypeHeader(w)
	if err != nil {
		ReturnJSONError(w, "FuelCell Status", err, http.StatusInternalServerError, true)
	}
	if _, err := fmt.Fprint(w, strStatus); err != nil {
		log.Println(err)
	}
}

func setContentTypeHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
