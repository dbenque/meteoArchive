package meteoServer

import (
	"io"
	"log"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"net/http"

	"code.google.com/p/biogo.store/kdtree"

	"github.com/gorilla/mux"
)

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, "Internal Server Error")
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
