package meteoServer

import (
	"io"
	"net/http"

	"github.com/biogo/store/kdtree"
	//"code.google.com/p/biogo.store/kdtree"

	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"
	"github.com/gorilla/mux"
)

func serveError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, err.Error())
}

// Server Global Variable
var kdtreeOfStation *kdtree.Tree

func GetServerStorage(c interface{}) meteoAPI.Storage {
	res := resource.NewResources(c)
	si := res.Storage()
	return si.(meteoAPI.Storage)
}

//Serve serve rest API to work on station and meteo measure
func ApplyHttpHandler() {

	// serverStorage = storage
	// serverStorage.Initialize()
	// stations, _ := serverStorage.GetAllStations()
	//kdtreeOfStation = kdtree.New(stations, true)

	r := mux.NewRouter()
	r.HandleFunc("/meteo/geoloc", handleGetGeoloc)
	r.HandleFunc("/meteo/distance", handleDistance)
	r.HandleFunc("/meteo/near", handleNear)
	r.HandleFunc("/meteo/packStations", handlePackStation)
	r.HandleFunc("/meteo/kdtreeReload/{storageName}", handleKDTreeReload)
	r.HandleFunc("/meteo/infoclimat/updateStations/{storageName}", handleInfoclimatUpdateStations)
	r.HandleFunc("/meteo/infoclimat/getMonthlySerie", handleInfoclimatGetMonthlySerie)
	http.Handle("/", r)

}
