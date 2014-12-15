package meteoServer

import (
	"io"
	"log"
	"meteo/geoloc"
	"meteo/infoclimat"
	"meteo/meteoAPI"
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
		poi, err := geoloc.GeolocFromCity(city[0], country, "fr")
		if err != nil {
			w.Write([]byte("Geolo not found"))
			return
		}

		rayon := 100.0
		if rayonStr, ok := r.URL.Query()["rayon"]; ok {
			rayonInt, _ := strconv.Atoi(rayonStr[0])
			rayon = float64(rayonInt)
		}

		stations, _ := infoclimat.GetStations(true)

		tree := kdtree.New(stations, true)
		keeper := kdtree.NewDistKeeper(rayon)
		tree.NearestSet(keeper, meteoAPI.NewStationFromPOI(poi))
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

func Serve() {

	r := mux.NewRouter()
	r.HandleFunc("/distance", handleDistance)
	http.Handle("/", r)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
