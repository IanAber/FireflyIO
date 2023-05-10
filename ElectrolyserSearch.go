package main

//
///*
//GetOurIP will returns the preferred ip address of the host on which we are running.
//*/
//func GetOurIP() (net.IP, error) {
//	conn, err := net.Dial("udp", "8.8.8.8:80")
//	if err != nil {
//		return nil, err
//	}
//	defer func() {
//		if err := conn.Close(); err != nil {
//			log.Print(err)
//		}
//	}()
//
//	localAddr := conn.LocalAddr().(*net.UDPAddr)
//
//	return localAddr.IP, nil
//}
//
///*
//SearchForElectrolyser will turn on the relevant relay and search the subnet that we are in for an electorlyser to come on line.
//If a new electrolyser is found it adds it to the chain.
//*/
//func SearchForElectrolyser(device int, testIP string) error {
//	OurIP, err := GetOurIP()
//	if err != nil {
//		return err
//	}
//
//	// First we lock the electrolysers so they do not get turned off when we are searching
//	SystemStatus.ElectrolyserLock = true
//	defer func() { SystemStatus.ElectrolyserLock = false }()
//	// Switch just the one electrolyser on
//	switch device {
//	case 0:
//		relays.SetElectrolyser1(Off)
//		relays.SetElectrolyser2(Off)
//		relays.SetElectrolyser0(On)
//	case 1:
//		relays.SetElectrolyser0(Off)
//		relays.SetElectrolyser2(Off)
//		relays.SetElectrolyser1(On)
//	case 2:
//		relays.SetElectrolyser0(Off)
//		relays.SetElectrolyser1(Off)
//		relays.SetElectrolyser2(On)
//	default:
//		return fmt.Errorf("we already have two electrolysers registered")
//	}
//
//	// Delay for 15 seconds to let the electrolyser power up.
//	time.Sleep(time.Second * 15)
//
//	// See if we can find the electrolyser. Start at what we recorded as this devices IP last time
//	if IP := scan(OurIP, testIP); IP != nil {
//		SystemStatus.Electrolysers = append(SystemStatus.Electrolysers, NewElectrolyser(IP))
//	}
//	return nil
//}
//
///*
//AcquireElectrolysers attempts to find two electrolysers
//*/
//func AcquireElectrolysers() {
//	// Wait for the ModbusRTU system to get started so we can turn the relays on.
//
//	// Set all Electrolysers to off.
//
//	relays.SetElectrolyser0(Off)
//	relays.SetElectrolyser1(Off)
//
//	el0IP := ""
//	el1IP := ""
//	if len(params.Electrolysers) > 0 {
//		el0IP = params.Electrolysers[0].IP
//	}
//	if len(params.Electrolysers) > 1 {
//		el1IP = params.Electrolysers[1].IP
//	}
//
//	// Clear any existing electrolyser registrations
//	params.Electrolysers = nil
//	// Make sure we turn the electrolysers off when we are done.
//	defer func() {
//		relays.SetElectrolyser0(Off)
//		relays.SetElectrolyser1(Off)
//	}()
//	// Search for the first electrolyser
//	if err := SearchForElectrolyser(0, el0IP); err == nil {
//		// Give it 5 seconds then get the serial number
//		time.Sleep(time.Second * 10)
//
//		el := new(ElectrolyserConfig)
//		el.Serial = SystemStatus.Electrolysers[0].GetSerial()
//		el.IP = SystemStatus.Electrolysers[0].GetIPString()
//		params.Electrolysers = append(params.Electrolysers, el)
//	} else {
//		log.Print(err)
//	}
//
//	// Search for a second electrolyser
//	if err := SearchForElectrolyser(1, el1IP); err == nil {
//		time.Sleep(time.Second * 10)
//
//		el := new(ElectrolyserConfig)
//		el.Serial = SystemStatus.Electrolysers[1].GetSerial()
//		el.IP = SystemStatus.Electrolysers[1].GetIPString()
//		params.Electrolysers = append(params.Electrolysers, el)
//	} else {
//		log.Print(err)
//	}
//
//	if len(params.Electrolysers) > 0 {
//		if err := params.WriteSettings(); err != nil {
//			log.Print(err)
//		}
//	}
//	plural := ""
//	if len(SystemStatus.Electrolysers) > 1 {
//		plural = "s"
//	}
//	log.Printf("Found %d electrolyser%s", len(SystemStatus.Electrolysers), plural)
//}
//
///**
//tryConnect attepmts to connect to the IP on the given port. err is nil if it succeeds
//*/
//func tryConnect(host net.IP, port int) error {
//	timeout := time.Millisecond * 250
//
//	log.Println("Scanning", net.JoinHostPort(host.String(), fmt.Sprint(port)))
//	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host.String(), fmt.Sprint(port)), timeout)
//	if err != nil {
//		return err
//	}
//	if conn != nil {
//		defer func() {
//			if err := conn.Close(); err != nil {
//				log.Print(err)
//			}
//		}()
//		return nil
//	}
//	return fmt.Errorf("unknown error")
//}
//
///**
//CheckForElectrolyser tests for an electrolyser.
//Given that we can connect on the Modbus port, test to see if this looks like an electrolyser by
//checking for the serial number.
//*/
//func CheckForElectrolyser(ip net.IP) error {
//	var config modbus.ClientConfiguration
//	config.Timeout = 1 * time.Second // 1 second timeout
//	config.URL = "tcp://" + ip.String() + ":502"
//	if Client, err := modbus.NewClient(&config); err == nil {
//		if err := Client.Open(); err != nil {
//			return err
//		}
//		defer func() {
//			if err := Client.Close(); err != nil {
//				log.Print(err)
//			}
//		}()
//		model, err := Client.ReadUint32(0, modbus.INPUT_REGISTER)
//		if err != nil {
//			log.Println("Error getting serial number - ", err)
//			return err
//		}
//		// Is this an EL21?
//		if model != 0x454C3231 {
//			return fmt.Errorf("not an EL21")
//		}
//		return nil
//	} else {
//		return err
//
//	}
//}
//
//func ipRegistered(ip byte) bool {
//	for _, el := range SystemStatus.Electrolysers {
//		if el.ip[3] == ip {
//			return true
//		}
//	}
//	return false
//}
//
//func scan(OurIP net.IP, testIP string) net.IP {
//	IP := OurIP
//
//	// First look for an electrolyser at the given IP if we got one
//	if len(testIP) > 0 {
//		TestIP := net.ParseIP(testIP)
//		if tryConnect(TestIP, 502) == nil {
//			// Something is there and responding on the Modbus port.
//			if CheckForElectrolyser(TestIP) == nil {
//				// It is an electrolyser so return its IP
//				return TestIP
//			}
//		}
//	}
//
//	// Electrolyser was not found or no IP was given so we need to search the whole subnet
//	for ip := byte(254); ip > 1; ip-- {
//		if (ip != OurIP[3]) && !ipRegistered(ip) {
//			IP[3] = ip
//			if tryConnect(IP, 502) == nil {
//				// Something is there and responding on the Modbus port.
//				if CheckForElectrolyser(IP) == nil {
//					// It is an electrolyser so return its IP
//					return IP
//				}
//			}
//		}
//	}
//	return nil
//}
