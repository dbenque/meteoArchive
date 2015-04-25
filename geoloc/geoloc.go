package geoloc

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/dbenque/meteoArchive/client"
	"github.com/dbenque/meteoArchive/meteoAPI"
)

const (
	googleAPIKey = "AIzaSyAEeofXwSEw12js8ft9xHxY-2bi5s0K5go"
	googleAPIURL = "https://maps.googleapis.com/maps/api/geocode/json?"
)

//FromCity retrieve the longitude latitude via google API https://developers.google.com/maps/documentation/geocoding/index
func FromCity(getter meteoClient.URLGetter, city string, region string, language string) (poi meteoAPI.POI, err error) {

	url := googleAPIURL + "key=" + googleAPIKey + "&address=" + city + "&region=" + region + "&language=" + language

	// client := &http.Client{CheckRedirect: nil}
	// req, err := http.NewRequest("GET", url, nil)
	// response, err := client.Do(req)
	// if err != nil {
	// 	return
	// }
	// response.Body.Close()

	response, err := getter.Get(url)
	defer response.Body.Close()

	responseStr, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var dataj map[string]interface{}
	if err = json.Unmarshal(responseStr, &dataj); err != nil {
		return poi, err
	}

	results, ok := dataj["results"].([]interface{})
	if !ok {
		return poi, errors.New("bad cast dataj")
	}

	if len(results) == 0 {
		return poi, errors.New("No result for geoloc")
	}

	obj, ok := results[0].(map[string]interface{})
	if !ok {
		return poi, errors.New("bad cast result0")
	}

	geo, ok := obj["geometry"].(map[string]interface{})
	if !ok {
		return poi, errors.New("bad cast geometry")
	}
	location, ok := geo["location"].(map[string]interface{})
	if !ok {
		return poi, errors.New("bad cast location")
	}

	poi.Latitude = location["lat"].(float64)
	poi.Longitude = location["lng"].(float64)
	poi.Name, _ = obj["formatted_address"].(string)
	return

}
