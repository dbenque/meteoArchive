package meteoServer

import (
	"encoding/json"
	"fmt"
	"meteoArchive/infoclimat"
	"meteoArchive/meteoAPI"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func handleInfoclimatUpdateStations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	switch vars["storageName"] {
	case "mapStorage":
		go func() {

			defer func() { fmt.Println("UpdateStation from infoclimat completed/ended") }()

			inputCode := ""
			if country, ok := r.URL.Query()["country"]; ok {
				inputCode = country[0]
			}

			website := infoclimat.InfoClimatWebsite{}
			mapStorage := meteoAPI.NewMapStorage("mapStorage")
			mapStorage.Initialize()
			website.UpdateStations(mapStorage, inputCode)
			mapStorage.Persist()
		}()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Update on going with Infoclimat website"))
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
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
		serie := infoclimat.RetrieveMonthlyReports(stationAndDist.station, year)
		dataj, _ := json.Marshal(*serie)
		output = output + stationAndDist.station.Name + "(" + strconv.Itoa(int(stationAndDist.distance)) + "): " + string(dataj) + "\n"

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}
