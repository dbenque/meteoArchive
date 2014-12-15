package ncdc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	api "meteo/meteoAPI"
	"net/http"
)

// GetStations get all the stations from NCDC website
func GetStations() (stations api.Stations, err error) {

	client := &http.Client{CheckRedirect: nil}
	req, err := http.NewRequest("GET", "http://www.ncdc.noaa.gov/cdo-web/api/v2/stations", nil)
	req.Header.Add("token", "UbzuUUCoziZyNHPjqIjWEbBjgdsTgQva")

	fmt.Println("sending request:", *req)
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	fmt.Println("reading request")
	responseStr, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	fmt.Println("json unmarshall")
	var dataj map[string]interface{}
	err = json.Unmarshal(responseStr, dataj)
	fmt.Println(string(responseStr))
	fmt.Println("prepare iteration")
	results, ok := dataj["results"].([]interface{})
	if !ok {
		return nil, errors.New("bad cast") // I was stucked with this error since hte token distributed by the website allowed me to do only one call per day! Mail sent to support.
	}

	fmt.Println(results[0])

	// stations = make(api.Stations, len(results), len(results))
	// fmt.Println("Iterations")
	// for i, s := range results {
	//
	// 	stations[i] = api.Station{api.POI{s["name"].(string), s["elevation"].(float32), s["latitude"].(float32), s["longitude"].(float32)}, "NCDC"}
	// 	fmt.Println("Iteration done ", i)
	// }

	return stations, nil
}
