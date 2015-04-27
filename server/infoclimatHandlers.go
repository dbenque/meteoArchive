package meteoServer

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dbenque/meteoArchive/infoclimat"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"
)

// Update the list of station for a given country
func handleInfoclimatUpdateStations(w http.ResponseWriter, r *http.Request) {
	// go func() {
	// 	defer func() { fmt.Println("UpdateStation from infoclimat completed/ended") }()

	inputCode := ""
	if country, ok := r.URL.Query()["country"]; ok {
		inputCode = country[0]
	}

	res := resource.NewResources(r)
	res.Logger().Infof("Updating stations for country %s. On going ...", inputCode)

	website := infoclimat.InfoClimatWebsite{}
	GetServerStorage(r).Initialize()
	website.UpdateStations(res, GetServerStorage(r), inputCode)
	GetServerStorage(r).Persist()
	res.Logger().Infof("Stations updated for country %s. Completed", inputCode)
	// }()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update Stations done"))
}

// Return the monthly serie for a given location (city/country  or  lon/lat) for a given year
func handleInfoclimatGetMonthlySerie(w http.ResponseWriter, r *http.Request) {

	res := resource.NewResources(r)
	ensureKdtreeLoaded(res)

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
		serie := GetServerStorage(r).GetMonthlyMeasureSerie(stationAndDist.station)
		if serie == nil { // The Serie for the station does not even exist!
			newserie := make(meteoAPI.MonthlyMeasureSerie)
			serie = &newserie
			res.Logger().Infof("Creating New Serie")
		}

		// retrieve the serie
		if serie.GetMeasure(year, time.Month(1)) == nil { // looks like we have no input for that year let's fetch!
			infoclimat.CompleteMonthlyReports(res, serie, stationAndDist.station, year)
			GetServerStorage(r).PutMonthlyMeasureSerie(stationAndDist.station, serie)
			GetServerStorage(r).Persist()
		} else { // let's reuse what we have in storage
			res.Logger().Infof("Monthly serie retrieved from storage")
		}

		resultsObj.Results[index] = resultData{stationAndDist.station.POI, int(stationAndDist.distance), serie.GetSerieIndexedByMonth(year)}
	}

	dataj, _ := json.Marshal(resultsObj)

	w.WriteHeader(http.StatusOK)
	w.Write(dataj)

}
