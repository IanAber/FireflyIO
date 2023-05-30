package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

const version = "1.1.02"

/**********************************************************
CAN bus must be enabled before this service can be started
https://www.pragmaticlinux.com/2021/07/automatically-bring-up-a-socketcan-interface-on-boot/
*/
var (
	canBus           *CANBus
	CANInterface     string
	WebPort          string
	databaseServer   string
	databasePort     string
	databaseName     string
	databaseLogin    string
	databasePassword string
	Relays           RelaysType
	Outputs          DigitalOutputsType
	Inputs           DigitalInputsType
	AnalogInputs     AnalogInputsType
	ACMeasurements   [4]ACMeasurementsType
	DCMeasurements   [4]DCMeasurementsType
	jsonSettings     string
	currentSettings  *SettingsType
	webFiles         string
	pDB              *sql.DB
	FuelCell         PANFuelCell
	logFile          *os.File
	logFileName      string
)

func connectToDatabase() (*sql.Stmt, *sql.DB, error) {
	if pDB != nil {
		if closeErr := pDB.Close(); closeErr != nil {
			log.Println(closeErr)
		}
		pDB = nil
	}
	// Set the time zone to Local to correctly record times
	var sConnectionString = databaseLogin + ":" + databasePassword + "@tcp(" + databaseServer + ":" + databasePort + ")/" + databaseName + "?loc=Local"

	db, err := sql.Open("mysql", sConnectionString)
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}
	err = db.Ping()
	if err != nil {
		if closeErr := db.Close(); closeErr != nil {
			log.Println(closeErr)
		}
		pDB = nil
		return nil, nil, err
	}
	logAnalog, err := db.Prepare("INSERT INTO firefly.IOValues(a0, a1, a2, a3, a4, a5, a6, a7, vref, cpuTemp, rawCpuTemp, inputs, outputs, relays, ACVolts, ACAmps, ACWatts, ACHertz) VALUES  (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Println(err)
		if closeErr := db.Close(); closeErr != nil {
			log.Println(closeErr)
		}
		return nil, nil, err
	}
	dbRecord.stmt, err = db.Prepare(`INSERT INTO firefly.PANFuelCell (StackCurrent
	, StackVoltage
	, CoolantInlTemp
	, CoolantOutTemp
	, OutputVoltage
	, OutputCurrent
	, CoolantFanSpeed
	, CoolantPumpSpeed
	, CoolantPumpVolts
	, CoolantPumpAmps
	, InsulationResistance
	, HydrogenPressure
	, AirPressure
	, CoolantPressure
	, AirinletTemp
	, AmbientTemp
	, AirFlow
	, HydrogenConcentration
	, DCDCTemp, DCDCInVolts
	, DCDCOutVolts
	, DCDCInAmps
	, DCDCOutAmps
	, MinCellVolts
	, MaxCellVolts
	, AvgCellVolts
	, IdxMaxCell
	, IdxMinCell
	, RunStage
	, FaultLevel
	, PowerModeState
	, Cell00Volts
	, Cell01Volts
	, Cell02Volts
	, Cell03Volts
	, Cell04Volts
	, Cell05Volts
	, Cell06Volts
	, Cell07Volts
	, Cell08Volts
	, Cell09Volts
	, Cell10Volts
	, Cell11Volts
	, Cell12Volts
	, Cell13Volts
	, Cell14Volts
	, Cell15Volts
	, Cell16Volts
	, Cell17Volts
	, Cell18Volts
	, Cell19Volts
	, Cell20Volts
	, Cell21Volts
	, Cell22Volts
	, Cell23Volts
	, Cell24Volts
	, Cell25Volts
	, Cell26Volts
	, Cell27Volts
	, Cell28Volts
	, Cell29Volts
	, Cell30Volts
	, Cell31Volts
	, Alarms
	) 
		 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?);`)

	return logAnalog, db, err
}

func ConnectCANBus() *CANBus {
	if Bus, err := NewCANBus(CANInterface); err != nil {
		log.Println(err)
		return nil
	} else {
		return Bus
	}
}

func init() {

	flag.StringVar(&CANInterface, "can", "can0", "CAN Interface Name")
	flag.StringVar(&WebPort, "WebPort", "20080", "Web port")
	flag.StringVar(&jsonSettings, "jsonSettings", "/etc/FireFlyIO.json", "JSON file containing the system control parameters")
	flag.StringVar(&webFiles, "webFiles", "/FireflyIO/web", "Path to the WEB files location")
	flag.StringVar(&databaseServer, "sqlServer", "localhost", "MySQL Server")
	flag.StringVar(&databaseName, "database", "firefly", "Database name")
	flag.StringVar(&databaseLogin, "dbUser", "FireflyService", "Database login user name")
	flag.StringVar(&databasePassword, "dbPassword", "logger", "Database user password")
	flag.StringVar(&databasePort, "dbPort", "3306", "Database port")
	flag.StringVar(&logFileName, "logfile", "/var/log/FireflyIO", "Name of the log file")
	flag.Parse()

	// open log file
	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	// set log out put
	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	Relays.InitRelays()
	Outputs.InitOutputs()
	Inputs.InitInputs()
	AnalogInputs.InitAnalogInputs()

	log.Println("Loading the settings")
	currentSettings = NewSettings()
	if err := currentSettings.LoadSettings(jsonSettings); err != nil {
		log.Print(err)
	}

	log.Println("Connecting to can bus")
	canBus = ConnectCANBus()
	FuelCell.init(canBus)
	if err := FuelCell.setTargetBattHigh(currentSettings.FuelCellSettings.HighBatterySetpoint); err != nil {
		log.Print(err)
	}
	if err := FuelCell.setTargetBattLow(currentSettings.FuelCellSettings.LowBatterySetpoint); err != nil {
		log.Print(err)
	}
	if err := FuelCell.setTargetPower(currentSettings.FuelCellSettings.PowerSetting); err != nil {
		log.Print(err)
	}

	log.Println("Starting the WEB site.")
	go setUpWebSite()
}

/**
ClientLoop ticks every second and logs values to the database. It also broadcasts the values to any registered web socket clients.
*/
func ClientLoop() {
	// Set up the sync to send data to waiting web socket clients
	broadcastTime := time.NewTicker(time.Second)

	for {
		select {
		case <-broadcastTime.C:
			{
				if canBus == nil || canBus.bus == nil {
					log.Println("Adding the CAN bus monitor")
					if canBus != nil {
						if canBus.bus != nil {
							if err := canBus.bus.Disconnect(); err != nil {
								log.Println(err)
							}
						}
						canBus = nil
						FuelCell.bus = nil
					}
					canBus = ConnectCANBus()
					FuelCell.bus = canBus
				}

				//				log.Println("Broadcast")
				bytes, err := getJsonStatus()
				if err != nil {
					log.Print("Error marshalling the data - ", err)
				} else {
					select {
					case pool.Broadcast <- bytes:
					default:
						fmt.Println("Channel would block!")
					}
				}
			}
		}
	}
}

func DatabaseLogger() {
	var (
		err       error
		logAnalog *sql.Stmt
	)
	logAnalog, pDB, err = connectToDatabase()
	if err != nil {
		log.Println(err)
	}
	loggingTime := time.NewTicker(time.Second)

	for {
		select {
		case <-loggingTime.C:
			if pDB == nil {
				log.Println("Reconnect to the database")
				logAnalog, pDB, err = connectToDatabase()
				if err != nil {
					log.Println(err)
				}
			}
			if pDB != nil {
				//					log.Println("Logging data")
				rawTemp, cpuTemp := AnalogInputs.GetCPUTemperature()
				if _, err := logAnalog.Exec(AnalogInputs.GetRawInput(0), AnalogInputs.GetRawInput(1), AnalogInputs.GetRawInput(2), AnalogInputs.GetRawInput(3),
					AnalogInputs.GetRawInput(4), AnalogInputs.GetRawInput(5), AnalogInputs.GetRawInput(6), AnalogInputs.GetRawInput(7),
					AnalogInputs.GetVREF(), cpuTemp, rawTemp,
					Inputs.GetAllInputs(), Outputs.GetAllOutputs(), Relays.GetAllRelays(),
					ACMeasurements[0].getVolts(), ACMeasurements[0].getAmps(), ACMeasurements[0].getPower(), ACMeasurements[0].getFrequency(),
				); err != nil {
					log.Println(err)
					if closeErr := pDB.Close(); closeErr != nil {
						log.Println(closeErr)
					}
					pDB = nil
					logAnalog = nil
				}
				if err := dbRecord.saveToDatabase(); err != nil {
					log.Println(err)
					if closeErr := pDB.Close(); closeErr != nil {
						log.Println(closeErr)
					}
					pDB = nil
					dbRecord.stmt = nil
				}
			} else {
				log.Println("Database is not connected")
			}
		}
	} // Log data to the database
}

/*
CANHeartbeat sends CAN packets to the fuel cell
*/
func CANHeartbeat() {
	heartbeatTime := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-heartbeatTime.C:
			{
				if canBus != nil {
					Relays.UpdateRelays() // Heartbeat to the FireflyIO board. If we don't send this the board will turn all relays off after about a minute.
					if err := canBus.SetFlags(currentSettings.getModbusFlags(), 0, 0, 0, 0, 0, 0, 0); err != nil {
						log.Println(err)
					}
					if err := FuelCell.updateOutput(); err != nil {
						log.Print(err)
					}
					if err := FuelCell.updateSettings(); err != nil {
						log.Print(err)
					}
				} else {
					log.Println("No CAN bus available")
				}
				heartbeat++
			}
		}
	}
}

func main() {
	defer func() {
		if err := logFile.Close(); err != nil {
			_, _ = fmt.Fprint(os.Stderr, err)
		}
	}()

	// ToDo
	//	go AcquireElectrolysers()

	go CANHeartbeat()
	go DatabaseLogger()
	ClientLoop()
}
