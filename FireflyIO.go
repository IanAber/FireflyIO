package main

import (
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
	"time"
)

const version = "1.0.05"

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
	ACMeasurements   ACMeasurementsType
	jsonSettings     string
	currentSettings  *SettingsType
	webFiles         string
	pDB              *sql.DB
	FuelCell         PANFuelCell
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
	logAnalog, err := db.Prepare("INSERT INTO firefly.IOValues(a0, a1, a2, a3, a4, a5, a6, a7, inputs, outputs, relays, ACVolts, ACAmps, ACWatts, ACHertz) VALUES  (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Println(err)
		if closeErr := db.Close(); closeErr != nil {
			log.Println(closeErr)
		}
		return nil, nil, err
	}

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
	flag.Parse()
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	Relays.InitRelays()
	Outputs.InitOutputs()
	Inputs.InitInputs()
	AnalogInputs.InitAnalogInputs()

	log.Println("Loading the settings")
	currentSettings = NewSettings()
	currentSettings.LoadSettings(jsonSettings)

	log.Println("Connecting to can bus")
	canBus = ConnectCANBus()
	FuelCell.bus = canBus
	FuelCell.setTargetBattHigh(currentSettings.FuelCellBatteryLimits.HighBatterySetpoint)
	FuelCell.setTargetBattLow(currentSettings.FuelCellBatteryLimits.LowBatterySetpoint)
	FuelCell.setTargetPower(currentSettings.FuelCellBatteryLimits.PowerSetting)

	log.Println("Starting the WEB site.")
	go setUpWebSite()
}

/**
loggingLoop ticks every second and logs values to the database. It also broadcasts the values to any registered web socket clients.
*/
func loggingLoop() {
	loops := 0
	done := make(chan bool)
	loggingTime := time.NewTicker(time.Second)
	logAnalog, pDB, err := connectToDatabase()
	if err != nil {
		log.Println(err)
	}

	for {
		select {
		case <-done:
			log.Println("Exiting the logging loop. This should not happen!")
			return
		case <-loggingTime.C:
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
				dataSignal.Broadcast() // Signal to broadcast values to registered web socket clients

				loops++
				if loops >= 5 {
					//					log.Println("Updating the relays")
					Relays.UpdateRelays()     // Heartbeat to the FireflyIO board. If we don't send this the board will turn all relays off after about a minute.
					FuelCell.updateSettings() // Update the battery limit settings
					FuelCell.updateOutput()   // Update the power setting
					loops = 0
				}

				// Log data to the database
				if pDB == nil {
					log.Println("Reconnect to the database")
					logAnalog, pDB, err = connectToDatabase()
					if err != nil {
						log.Println(err)
					}
				}
				if pDB != nil {
					//					log.Println("Logging data")
					if _, err := logAnalog.Exec(AnalogInputs.GetRawInput(0), AnalogInputs.GetRawInput(1), AnalogInputs.GetRawInput(2), AnalogInputs.GetRawInput(3),
						AnalogInputs.GetRawInput(4), AnalogInputs.GetRawInput(5), AnalogInputs.GetRawInput(6), AnalogInputs.GetRawInput(7),
						Inputs.GetAllInputs(), Outputs.GetAllOutputs(), Relays.GetAllRelays(),
						ACMeasurements.getVolts(), ACMeasurements.getAmps(), ACMeasurements.getPower(), ACMeasurements.getFrequency()); err != nil {
						log.Println(err)
						if closeErr := pDB.Close(); closeErr != nil {
							log.Println(closeErr)
							pDB = nil
							logAnalog = nil
						}
					}
				} else {
					log.Println("Database is not connected")
				}
			}
		}
	}
}

func main() {
	// Set up the sync to send data to waiting web socket clients
	dataSignal = sync.NewCond(&sync.Mutex{})

	// ToDo
	//	go AcquireElectrolysers()

	// Start the logging loop
	loggingLoop()
}
