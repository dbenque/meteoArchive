package meteoServer

import (
	"encoding/json"
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
	if city, ok := r.URL.Query()["city"]; ok {
		country := "fr"
		if countryQ, ok := r.URL.Query()["country"]; ok {
			country = countryQ[0]
		}
		poi, err := geoloc.FromCity(city[0], country, "fr")
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
	} else {
		w.Write([]byte("Could not read city from GET Query."))
	}
}

func handleNear(w http.ResponseWriter, r *http.Request) {

	if city, ok := r.URL.Query()["city"]; ok {
		country := "fr"
		if countryQ, ok := r.URL.Query()["country"]; ok {
			country = countryQ[0]
		}
		poi, err := geoloc.FromCity(city[0], country, "fr")
		if err != nil {
			w.Write([]byte("Geolo not found"))
			return
		}

		count := 4

		if c, ok := r.URL.Query()["count"]; ok {
			count, _ = strconv.Atoi(c[0])
		}

		keeper := kdtree.NewNKeeper(count)
		kdtreeOfStation.NearestSet(keeper, meteoAPI.NewStationFromPOI(poi))
		output := "Result:\n"
		for keeper.Len() > 0 {
			v := keeper.Heap.Pop()

			if c, ok := v.(kdtree.ComparableDist); ok {
				if s, ok := c.Comparable.(meteoAPI.Station); ok {
					output = output + "Station at " + strconv.Itoa(int(c.Dist)) + " km, " + s.Name
				}
			}
			output = output + "\n"
		}
		w.Write([]byte(output))
	} else {
		w.Write([]byte("Could not read city from GET Query."))
	}

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

func handleInfoclimatUpdateMonthlySerie(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	switch vars["storageName"] {
	case "mapStorage":
		go func() {

			defer func() { fmt.Println("Update Monthly completed/ended") }()

			storage := meteoAPI.NewMapStorage("mapStorage")
			storage.Initialize()
			stations := (*storage).GetAllStations()

			serie := infoclimat.RetrieveMonthlyReports(&(*stations)[100], 2013)

			//fmt.Println(len(*serie))
			dataj, _ := json.Marshal(*serie)
			fmt.Println(string(dataj))

		}()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("On Going"))
	default:
		w.WriteHeader(http.StatusBadRequest)
	}

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
	r.HandleFunc("/infoclimat/updateMonthlySerie/{storageName}", handleInfoclimatUpdateMonthlySerie)
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
