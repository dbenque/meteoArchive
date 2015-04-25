package meteoServer

import (
	"io"
	"net/http"

	"code.google.com/p/biogo.store/kdtree"
	"github.com/dbenque/meteoArchive/client"
	"github.com/dbenque/meteoArchive/meteoAPI"

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
func ApplyHttpHandler(storage meteoAPI.Storage, clientFactory meteoClient.URLGetterFactory) {

	meteoClient.ClientFactory = clientFactory
	serverStorage = storage
	serverStorage.Initialize()
	stations := serverStorage.GetAllStations()
	kdtreeOfStation = kdtree.New(stations, true)

	r := mux.NewRouter()
	r.HandleFunc("/meteo/geoloc", handleGetGeoloc)
	r.HandleFunc("/meteo/distance", handleDistance)
	r.HandleFunc("/meteo/near", handleNear)
	r.HandleFunc("/meteo/kdtreeReload/{storageName}", handleKDTreeReload)
	r.HandleFunc("/meteo/infoclimat/updateStations/{storageName}", handleInfoclimatUpdateStations)
	r.HandleFunc("/meteo/infoclimat/getMonthlySerie", handleInfoclimatGetMonthlySerie)
	http.Handle("/", r)

}
