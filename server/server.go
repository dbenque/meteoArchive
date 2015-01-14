package meteoServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"meteoArchive/geoloc"
	"meteoArchive/infoclimat"
	"meteoArchive/meteoAPI"
	"net/http"
	"strconv"

	"code.google.com/p/biogo.store/kdtree"

	"github.com/gorilla/mux"
)

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
}

func handleDistance(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Distance"))

	city, country, _, err := readCityCountryCountFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	poi, err := geoloc.FromCity(city, country, "fr")
	if err != nil {
		w.Write([]byte("Geolo not found"))
		return
	}

	rayon := 100.0
	if rayonStr, ok := r.URL.Query()["rayon"]; ok {
		rayonInt, _ := strconv.Atoi(rayonStr[0])
		rayon = float64(rayonInt)
	}

	keeper := kdtree.NewDistKeeper(rayon * rayon)
	kdtreeOfStation.NearestSet(keeper, meteoAPI.NewStationFromPOI(poi))
	output := "Result:\n"
	for keeper.Len() > 0 {
		v := keeper.Heap.Pop()

		if c, ok := v.(kdtree.ComparableDist); ok {
			if s, ok := c.Comparable.(meteoAPI.Station); ok {
				output = output + "Station at " + strconv.Itoa(int(math.Sqrt(c.Dist))) + " km, " + s.Name
			}
		}
		output = output + "\n"
	}
	w.Write([]byte(output))

}

type stationAndDistance struct {
	station  *meteoAPI.Station
	distance float64
}

func getNearest(city string, country string, count int) (nearStations []stationAndDistance) {

	poi, err := geoloc.FromCity(city, country, "fr")
	if err != nil {
		return
	}

	keeper := kdtree.NewNKeeper(count)
	kdtreeOfStation.NearestSet(keeper, meteoAPI.NewStationFromPOI(poi))

	for keeper.Len() > 0 {
		v := keeper.Heap.Pop()

		if c, ok := v.(kdtree.ComparableDist); ok {
			if s, ok := c.Comparable.(meteoAPI.Station); ok {
				nearStations = append(nearStations, stationAndDistance{&s, math.Sqrt(c.Dist)})
			}
		}
	}
	return
}

func readCityCountryCountFromURL(r *http.Request) (city, country string, count int, err error) {

	country = "fr"
	count = 4
	city = ""

	if cityurl, ok := r.URL.Query()["city"]; ok {
		city = cityurl[0]
	} else {
		errors.New("Could not read city from GET Query.")
	}

	if countryurl, ok := r.URL.Query()["country"]; ok {
		country = countryurl[0]
	}

	if counturl, ok := r.URL.Query()["count"]; ok {
		if count, err = strconv.Atoi(counturl[0]); err != nil {
			errors.New("count parameter is not an int")
		}
	}
	return
}

func handleNear(w http.ResponseWriter, r *http.Request) {

	city, country, count, err := readCityCountryCountFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	nearStations := getNearest(city, country, count)

	if len(nearStations) == 0 {
		w.Write([]byte("Geolo not found"))
		return
	}

	output := "Result:\n"
	for _, s := range nearStations {
		output = output + "Station at " + strconv.Itoa(int(s.distance)) + " km, " + s.station.Name + "\n"
	}

	w.Write([]byte(output))

}

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

func handleKDTreeReload(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	switch vars["storageName"] {
	case "mapStorage":
		go func() {

			defer func() { fmt.Println("Update kdtree completed/ended") }()

			storage := meteoAPI.NewMapStorage("mapStorage")
			storage.Initialize()
			stations := (*storage).GetAllStations()
			fmt.Println("Station Count:", len(*stations))

			kdtreeOfStation = kdtree.New(stations, true)
		}()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Update on going for kdtree"))
	default:
		w.WriteHeader(http.StatusBadRequest)
	}

}

func handleInfoclimatGetMonthlySerie(w http.ResponseWriter, r *http.Request) {

	city, country, _, err := readCityCountryCountFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	output := "Resulat:\n"

	for _, stationAndDist := range getNearest(city, country, 3) {
		serie := infoclimat.RetrieveMonthlyReports(stationAndDist.station, 2013)
		dataj, _ := json.Marshal(*serie)
		output = output + stationAndDist.station.Name + "(" + strconv.Itoa(int(stationAndDist.distance)) + "): " + string(dataj) + "\n"

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))

}

// Server Global Variable
var kdtreeOfStation *kdtree.Tree

//Serve serve rest API to work on station and meteo measure
func Serve() {

	storage := meteoAPI.NewMapStorage("mapStorage")
	storage.Initialize()
	stations := (*storage).GetAllStations()
	kdtreeOfStation = kdtree.New(stations, true)

	r := mux.NewRouter()
	r.HandleFunc("/distance", handleDistance)
	r.HandleFunc("/near", handleNear)
	r.HandleFunc("/kdtreeReload/{storageName}", handleKDTreeReload)
	r.HandleFunc("/infoclimat/updateStations/{storageName}", handleInfoclimatUpdateStations)
	r.HandleFunc("/infoclimat/getMonthlySerie", handleInfoclimatGetMonthlySerie)
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
