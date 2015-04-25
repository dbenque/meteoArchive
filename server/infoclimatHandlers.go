package meteoServer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dbenque/meteoArchive/client"
	"github.com/dbenque/meteoArchive/infoclimat"
	"github.com/dbenque/meteoArchive/meteoAPI"
)

// Update the list of station for a given country
func handleInfoclimatUpdateStations(w http.ResponseWriter, r *http.Request) {
	// go func() {
	// 	defer func() { fmt.Println("UpdateStation from infoclimat completed/ended") }()

	inputCode := ""
	if country, ok := r.URL.Query()["country"]; ok {
		inputCode = country[0]
	}

	website := infoclimat.InfoClimatWebsite{}
	serverStorage.Initialize()
	website.UpdateStations(meteoClient.ClientFactory(r), serverStorage, inputCode)
	serverStorage.Persist()
	// }()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update Stations done"))
}

// Return the monthly serie for a given location (city/country  or  lon/lat) for a given year
func handleInfoclimatGetMonthlySerie(w http.ResponseWriter, r *http.Request) {

	// Define type that will be use for the output
	type resultData struct {
		meteoAPI.POI
		Km    int                          `json:"km,omitempty"`
		Serie meteoAPI.MonthlyMeasureSerie `json:"serie"`
	}

	// results is a list
	type resultsByCity struct {
		Results []resultData `json:"results,omitempty"`
	}

	// Retrieve the Year for the request [Manadatory]
	year, err := readYearFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Retrieve the nearest stations according to request parameters
	nearStations, err := getNearestFromRequest(r, 3)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	// Prepare struct for output
	resultsObj := resultsByCity{make([]resultData, len(nearStations))}

	// Loop over the retrieved stations to build the result
	for index, stationAndDist := range nearStations {

		// retrieve the serie from local storage, if none prepare a newserie for fetching
		serie := serverStorage.GetMonthlyMeasureSerie(stationAndDist.station)
		if serie == nil { // The Serie for the station does not even exist!
			newserie := make(meteoAPI.MonthlyMeasureSerie)
			serie = &newserie
		}

		// retrieve the serie
		if serie.GetMeasure(year, time.Month(1)) == nil { // looks like we have no input for that year let's fetch!
			infoclimat.CompleteMonthlyReports(meteoClient.ClientFactory(r), serie, stationAndDist.station, year)
			serverStorage.PutMonthlyMeasureSerie(stationAndDist.station, serie)
			serverStorage.Persist()
		} else { // let's reuse what we have in storage
			fmt.Println("Serie retrieved from storage")
		}

		resultsObj.Results[index] = resultData{stationAndDist.station.POI, int(stationAndDist.distance), serie.GetSerieIndexedByMonth(year)}
	}

	dataj, _ := json.Marshal(resultsObj)

	w.WriteHeader(http.StatusOK)
	w.Write(dataj)

}
