package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

/**
GetTimeRange returns the start and end times passed as query parameters.
*/
func GetTimeRange(r *http.Request) (start time.Time, end time.Time, err error) {
	params := r.URL.Query()
	values := params["start"]
	if len(values) != 1 {
		err = fmt.Errorf("Exactly one 'start=' value must be supplied for start time")
		return
	}
	timeVal, err := time.Parse("2006-1-2 15:04", values[0])
	if err != nil {
		return
	} else {
		start = timeVal
	}

	values = params["end"]
	if len(values) != 1 {
		err = fmt.Errorf("Exactly one 'start=' value must be supplied for start time")
		return
	}
	timeVal, err = time.Parse("2006-1-2 15:04", values[0])
	if err != nil {
		return
	} else {
		end = timeVal
	}
	//	log.Println("Date/time requested from ", start, " to ", end)
	return
}

type DCDCData struct {
	Logged float64 `json:"logged"`
	VOut   float64 `json:"volts"`
	IOut   float64 `json:"amps"`
}

func getFuelCellData(w http.ResponseWriter, r *http.Request) {
	var (
		Results []*DCDCData
		rqst    string
	)

	const DeviceString = "DC-DC Data"

	start, end, err := GetTimeRange(r)
	if err != nil {
		ReturnJSONError(w, DeviceString, err, http.StatusBadRequest, false)
		return
	}

	if pDB == nil {
		ReturnJSONErrorString(w, DeviceString, "No Database", http.StatusInternalServerError, true)
		return
	}

	if end.Sub(start) > time.Hour {
		rqst = `select min(UNIX_TIMESTAMP(logged)) as logged
                       ,(avg(DCDCOutVolts) / 10) as voltage
                       ,(avg(DCDCOutAmps) / 100) as current
                   from PANFuelCell
                  where logged between ? and ?
	              group by UNIX_TIMESTAMP(logged) div 60`
	} else {
		rqst = `select UNIX_TIMESTAMP(logged) as logged
		              ,(DCDCOutVolts / 10) as voltage
		              ,(DCDCOutAmps / 100) as current
		          from PANFuelCell
		         where logged between ? and ?`
		//rqst = `select UNIX_TIMESTAMP(logged) as logged
		//               ,(DCDCOutVolts / 10) as voltage
		//               ,(DCDCOutAmps / 10) as current
		//           from PANFuelCell
		//          where logged between '2023-05-10 01:00' and '2023-05-10 01:15' and ? <> ?`
	}
	//	log.Println(rqst, start, end)
	if rows, err := pDB.Query(rqst, start, end); err != nil {
		ReturnJSONError(w, DeviceString, err, http.StatusInternalServerError, true)
		return
	} else {
		log.Println("Got data....")
		defer func() {
			if err := rows.Close(); err != nil {
				log.Print(err)
			}
		}()
		for rows.Next() {
			result := new(DCDCData)
			if err := rows.Scan(&result.Logged, &result.VOut, &result.IOut); err != nil {
				ReturnJSONError(w, DeviceString, err, http.StatusInternalServerError, true)
				return
			}
			Results = append(Results, result)
		}
		if resultJSON, err := json.Marshal(Results); err != nil {
			ReturnJSONError(w, DeviceString, err, http.StatusInternalServerError, true)
		} else {
			if _, err := fmt.Fprintf(w, string(resultJSON)); err != nil {
				log.Print(err)
			}
		}
	}
}
