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
	//	pool = NewPool()
	pool.Init()
	go pool.Start()

	log.Println("Staring the WEB site on port ", WebPort)
	router := mux.NewRouter().StrictSlash(true)
	// Register with the WebSocket which will then push a JSON payload with data to keep the displayed data up to date. No polling necessary.
	router.HandleFunc("/ws", startDataWebSocket).Methods("GET")

	router.HandleFunc("/setRelay/{relay}/{on}", setRelay).Methods("PUT")
	router.HandleFunc("/setOutput/{output}/{on}", setOutput).Methods("PUT")
	router.HandleFunc("/getSettings", getSettings).Methods("GET")
	router.HandleFunc("/setSettings", setSettings).Methods("POST")
	router.HandleFunc("/getStatus", getStatus).Methods("GET")
	router.HandleFunc("/getFuelCell", getFuelCell).Methods("GET")                          // Returns the current status of the fuel cell only
	router.HandleFunc("/setFuelCell/TargetPower/{power}", setFcPower).Methods("PUT")       // Set the target power output
	router.HandleFunc("/setFuelCell/TargetBattHigh/{volts}", setFcBattHigh).Methods("PUT") // Set the battery high voltage setpoint
	router.HandleFunc("/setFuelCell/TargetBattLow/{volts}", setFcBatLow).Methods("PUT")    // Set the batery low voltage set point
	router.HandleFunc("/setFuelCell/Start", startFc).Methods("PUT")                        // Start the fuel cell
	router.HandleFunc("/setFuelCell/Stop", stopFc).Methods("PUT")                          // Stop the fuel cell
	router.HandleFunc("/setFuelCellSettings", setFuelCellSettings).Methods("POST")         // Submit a form with setpoints and power level
	router.HandleFunc("/setFuelCell/ExhaustOpen", exhaustOpen).Methods("PUT")              // Start the water pump on high and beginn air removal
	router.HandleFunc("/setFuelCell/ExhaustClose", exhaustClose).Methods("PUT")            // Stop the exhaust function
	router.HandleFunc("/setFuelCell/Enable", enableFc).Methods("PUT")                      // Enable CAN communications to the fuel cell (we are always listening but may not be sending)
	router.HandleFunc("/setFuelCell/Disable", disableFc).Methods("PUT")                    // Disable CAN communications to the fuel cell so it can be controlled locally by its own user interface
	router.HandleFunc("/unknown", getUnknownFrames).Methods("GET")

	router.HandleFunc("/FuelCellData/DCDC", getFuelCellData).Methods("GET")

	fileServer := http.FileServer(neuteredFileSystem{http.Dir(webFiles)})
	router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))

	log.Println("Starting WEB server")
	port := fmt.Sprintf(":%s", WebPort)
	log.Fatal(http.ListenAndServe(port, router))
}

func startDataWebSocket(w http.ResponseWriter, r *http.Request) {
	//	fmt.Println("WebSocket Endpoint Hit")
	conn, err := Upgrade(w, r)
	if err != nil {
		_, err = fmt.Fprintf(w, "%+v\n", err)
		if err != nil {
			log.Println(err)
		}
	}

	client := &Client{
		ID:   r.RemoteAddr,
		Conn: conn,
		//		Pool: pool,
	}

	pool.Register <- client
}

func enableFc(w http.ResponseWriter, r *http.Request) {
	currentSettings.FuelCellSettings.Enabled = true
	log.Println("Enabled")
	if err := currentSettings.SaveSettings(currentSettings.filepath); err != nil {
		ReturnJSONError(w, "Enable Fuel Cell", err, http.StatusInternalServerError, true)
		return
	} else {
		getFuelCell(w, r)
	}
}

func disableFc(w http.ResponseWriter, r *http.Request) {

	currentSettings.FuelCellSettings.Enabled = false
	log.Println("Disabled")
	if err := currentSettings.SaveSettings(currentSettings.filepath); err != nil {
		ReturnJSONError(w, "Enable Fuel Cell", err, http.StatusInternalServerError, true)
		return
	} else {
		getFuelCell(w, r)
	}
}

func setFcPower(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	request := vars["power"]
	const function = "Set Fuel Cell Power"

	fPower, err := strconv.ParseFloat(request, 64)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
		return
	}
	log.Println("set fuel cell power to ", fPower)
	err = FuelCell.setTargetPower(fPower)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
		return
	}
	if err := FuelCell.updateOutput(); err != nil {
		ReturnJSONError(w, "Set Fuel Cell Power", err, http.StatusInternalServerError, true)
		return
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
		return
	}
	log.Println("set fuel cell high battery limit to ", fVolts)
	err = FuelCell.setTargetBattHigh(fVolts)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
		return
	}
	if err = FuelCell.updateSettings(); err != nil {
		ReturnJSONError(w, function, err, http.StatusInternalServerError, true)
		return
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
		return
	}
	log.Println("set fuel cell low battery limit to ", fVolts)
	err = FuelCell.setTargetBattLow(fVolts)
	if err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
		return
	}
	if err = FuelCell.updateSettings(); err != nil {
		ReturnJSONError(w, function, err, http.StatusInternalServerError, true)
		return
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

func exhaustOpen(w http.ResponseWriter, r *http.Request) {
	FuelCell.exhaustOpen()
	getFuelCell(w, r)
}

func exhaustClose(w http.ResponseWriter, r *http.Request) {
	FuelCell.exhaustClose()
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
		return
	} else {
		currentSettings.FuelCellSettings.PowerSetting = floatval
		FuelCell.Control.TargetPower = floatval
	}
	if floatval, err := strconv.ParseFloat(r.FormValue("LowBattDemand"), 64); err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	} else {
		currentSettings.FuelCellSettings.LowBatterySetpoint = floatval
		FuelCell.Control.TargetBatteryLow = floatval
	}
	if floatval, err := strconv.ParseFloat(r.FormValue("HighBattDemand"), 64); err != nil {
		ReturnJSONError(w, function, err, http.StatusBadRequest, true)
	} else {
		currentSettings.FuelCellSettings.HighBatterySetpoint = floatval
		FuelCell.Control.TargetBatteryHigh = floatval
	}
	if err := currentSettings.SaveSettings(currentSettings.filepath); err != nil {
		log.Print(err)
	}
	if FuelCell.SystemInfo.Run {
		if err := FuelCell.updateSettings(); err != nil { // Update the battery limit settings
			log.Print(err)
		}
	}
	if err := FuelCell.updateOutput(); err != nil { // Update the power setting
		log.Print(err)
	}
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
			return
		}
	} else {
		if (relayNum >= 0) && (relayNum < int64(len(Relays.Relays))) {
			Relays.SetRelay(uint8(relayNum), bOn)
		} else {
			ReturnJSONErrorString(w, "setRelay", fmt.Sprintf("Invalid relay number - %d", relayNum), http.StatusBadRequest, true)
			return
		}
	}
	getFuelCell(w, r)
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
			return
		}
	} else {
		if (outputNum >= 0) && (outputNum < int64(len(Outputs.Outputs))) {

			Outputs.SetOutput(uint8(outputNum), bOn)
		} else {
			ReturnJSONErrorString(w, "setOutput", fmt.Sprintf("Invalid output number - %d", outputNum), http.StatusBadRequest, true)
			return
		}
	}
	getFuelCell(w, r)
}

type ACValuesType struct {
	Name          string
	ACVolts       float32
	ACAmps        float32
	ACWatts       float32
	ACWattHours   uint32
	ACHertz       float32
	ACPowerFactor float32
	Error         string
}
type DCValuesType struct {
	Name    string
	DCVolts float32
	DCAmps  float32
	Error   string
}
type JsonDataType struct {
	System            string
	Version           string
	Relays            *RelaysType
	Analog            *AnalogInputsType
	DigitalOut        *DigitalOutputsType
	DigitalIn         *DigitalInputsType
	ACMeasurements    []ACValuesType
	DCMeasurements    []DCValuesType
	PanFuelCellStatus PanStatus
}

func getJsonStatus() ([]byte, error) {
	var data JsonDataType

	data.System = currentSettings.Name
	data.Version = version
	data.Relays = &Relays
	data.DigitalIn = &Inputs
	data.DigitalOut = &Outputs
	data.Analog = &AnalogInputs
	count := 0
	for idx := range ACMeasurements {
		if ACMeasurements[idx].Name != "" {
			count++
		}
	}
	data.ACMeasurements = make([]ACValuesType, count)
	i := 0
	for idx := range ACMeasurements {
		if ACMeasurements[idx].Name != "" {
			data.ACMeasurements[i].Name = ACMeasurements[idx].Name
			data.ACMeasurements[i].ACVolts = ACMeasurements[idx].getVolts()
			data.ACMeasurements[i].ACAmps = ACMeasurements[idx].getAmps()
			data.ACMeasurements[i].ACWatts = ACMeasurements[idx].getPower()
			data.ACMeasurements[i].ACWattHours = ACMeasurements[idx].getEnergy()
			data.ACMeasurements[i].ACHertz = ACMeasurements[idx].getFrequency()
			data.ACMeasurements[i].ACPowerFactor = ACMeasurements[idx].getPowerFactor()
			data.ACMeasurements[i].Error = ACMeasurements[idx].getError()
			i++
		}
	}

	count = 0
	for idx := range DCMeasurements {
		if DCMeasurements[idx].Name != "" {
			count++
		}
	}
	data.DCMeasurements = make([]DCValuesType, count)
	i = 0
	for i := range DCMeasurements {
		if DCMeasurements[i].Name != "" {
			data.DCMeasurements[i].Name = DCMeasurements[i].Name
			data.DCMeasurements[i].DCVolts = DCMeasurements[i].getVolts()
			data.DCMeasurements[i].DCAmps = DCMeasurements[i].getAmps()
			data.DCMeasurements[i].Error = DCMeasurements[i].getError()
			i++
		}
	}
	data.PanFuelCellStatus = FuelCell.GetStatus()

	JSONBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	} else {
		return JSONBytes, nil
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
	if r.FormValue("isoLowBehaviour") == "true" {
		currentSettings.FuelCellSettings.IgnoreIsoLow = true
	} else {
		currentSettings.FuelCellSettings.IgnoreIsoLow = false
	}
	currentSettings.ACMeasurement[0].Name = strings.TrimSpace(r.FormValue("ACMeasurement20"))
	currentSettings.ACMeasurement[0].SlaveID = 20
	currentSettings.ACMeasurement[1].Name = strings.TrimSpace(r.FormValue("ACMeasurement21"))
	currentSettings.ACMeasurement[1].SlaveID = 21
	currentSettings.ACMeasurement[2].Name = strings.TrimSpace(r.FormValue("ACMeasurement22"))
	currentSettings.ACMeasurement[2].SlaveID = 22
	currentSettings.ACMeasurement[3].Name = strings.TrimSpace(r.FormValue("ACMeasurement23"))
	currentSettings.ACMeasurement[3].SlaveID = 23
	currentSettings.DCMeasurement[0].Name = strings.TrimSpace(r.FormValue("DCMeasurement10"))
	currentSettings.DCMeasurement[0].SlaveID = 10
	currentSettings.DCMeasurement[1].Name = strings.TrimSpace(r.FormValue("DCMeasurement11"))
	currentSettings.DCMeasurement[1].SlaveID = 11
	currentSettings.DCMeasurement[2].Name = strings.TrimSpace(r.FormValue("DCMeasurement12"))
	currentSettings.DCMeasurement[2].SlaveID = 12
	currentSettings.DCMeasurement[3].Name = strings.TrimSpace(r.FormValue("DCMeasurement13"))
	currentSettings.DCMeasurement[3].SlaveID = 13

	if err := currentSettings.SaveSettings(currentSettings.filepath); err != nil {
		log.Print(err)
	}
	if err := currentSettings.LoadSettings(currentSettings.filepath); err != nil {
		log.Print(err)
	}

	http.Redirect(w, r, "/config.html", http.StatusTemporaryRedirect)
}

func getStatus(w http.ResponseWriter, _ *http.Request) {
	sJSON, err := getJsonStatus()
	setContentTypeHeader(w)
	_, err = fmt.Fprint(w, string(sJSON))
	if err != nil {
		log.Println("failed to send the status - ", err)
		return
	}
}

func getFuelCell(w http.ResponseWriter, _ *http.Request) {
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
