package meteoServer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dbenque/meteoArchive/infoclimat"
	"github.com/dbenque/meteoArchive/meteoAPI"
)

func handleInfoclimatUpdateStations(w http.ResponseWriter, r *http.Request) {
	go func() {
		defer func() { fmt.Println("UpdateStation from infoclimat completed/ended") }()

		inputCode := ""
		if country, ok := r.URL.Query()["country"]; ok {
			inputCode = country[0]
		}

		website := infoclimat.InfoClimatWebsite{}
		serverStorage.Initialize()
		website.UpdateStations(serverStorage, inputCode)
		serverStorage.Persist()
	}()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update Stations on going with Infoclimat website"))
}

type resultData struct {
	meteoAPI.POI
	Km    int                          `json:"km,omitempty"`
	Serie meteoAPI.MonthlyMeasureSerie `json:"serie,omitempty"`
}

type resultsByCity struct {
	Results []resultData `json:"results,omitempty"`
}

func handleInfoclimatGetMonthlySerie(w http.ResponseWriter, r *http.Request) {

	var nearStations []stationAndDistance

	year, err := readYearFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nbResult := 3

	var city, country string
	latitude, longitute, err := readLatitudeLongitudeFromURL(r)
	if err != nil {
		err = nil
		city, country, err = readCityCountryFromURL(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Println("By Str")
		nearStations = getNearestByStr(city, country, nbResult)

	} else {
		fmt.Println("By Coord")
		nearStations = getNearestByCoord(latitude, longitute, nbResult)
	}

	resultsObj := resultsByCity{make([]resultData, nbResult)}
	for index, stationAndDist := range nearStations {

		serie := serverStorage.GetMonthlyMeasureSerie(stationAndDist.station)
		if serie == nil { // The Serie for the station does not even exist!
			newserie := make(meteoAPI.MonthlyMeasureSerie)
			serie = &newserie
		}

		if serie.GetMeasure(year, time.Month(1)) == nil { // looks like we have no input for that year ...
			infoclimat.CompleteMonthlyReports(serie, stationAndDist.station, year)
			serverStorage.PutMonthlyMeasureSerie(stationAndDist.station, serie)
			serverStorage.Persist()
		} else { // let's reuse what we have in storage
			fmt.Println("Serie retrieved from storage")
		}

		resultsObj.Results[index] = resultData{stationAndDist.station.POI, int(stationAndDist.distance), *serie}
	}

	dataj, _ := json.Marshal(resultsObj)

	w.WriteHeader(http.StatusOK)
	w.Write(dataj)

}
