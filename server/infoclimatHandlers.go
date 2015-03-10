package meteoServer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

func handleInfoclimatGetMonthlySerie(w http.ResponseWriter, r *http.Request) {

	city, country, err := readCityCountryFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	year, err := readYearFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	output := "Resulat:\n"

	for _, stationAndDist := range getNearestByStr(city, country, 3) {

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

		dataj, _ := json.Marshal(*serie)
		output = output + stationAndDist.station.Name + "(" + strconv.Itoa(int(stationAndDist.distance)) + "): " + string(dataj) + "\n"

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}
