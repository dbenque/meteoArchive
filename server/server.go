package meteoServer

import (
	"io"
	"net/http"

	"github.com/dbenque/meteoArchive/meteoAPI"

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
var serverStorage meteoAPI.Storage

//Serve serve rest API to work on station and meteo measure
func ApplyHttpHandler(storage meteoAPI.Storage) {

	serverStorage = storage
	serverStorage.Initialize()
	stations := serverStorage.GetAllStations()
	kdtreeOfStation = kdtree.New(stations, true)

	r := mux.NewRouter()
	r.HandleFunc("/geoloc", handleGetGeoloc)
	r.HandleFunc("/distance", handleDistance)
	r.HandleFunc("/near", handleNear)
	r.HandleFunc("/kdtreeReload/{storageName}", handleKDTreeReload)
	r.HandleFunc("/infoclimat/updateStations/{storageName}", handleInfoclimatUpdateStations)
	r.HandleFunc("/infoclimat/getMonthlySerie", handleInfoclimatGetMonthlySerie)
	http.Handle("/", r)

}
