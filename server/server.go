package meteoServer

import (
	"io"
	"net/http"
	"net/url"

	"github.com/biogo/store/kdtree"
	//"code.google.com/p/biogo.store/kdtree"

	"github.com/dbenque/meteoArchive/infoclimat"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"
	"github.com/gorilla/mux"

	"appengine"
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
	r.HandleFunc("/{meteo}/geoloc", handleGetGeoloc)
	r.HandleFunc("/{meteo}/distance", handleDistance)
	r.HandleFunc("/{meteo}/near", handleNear)
	r.HandleFunc("/{meteo}/packStations/asTask", handlePackStation)
	r.HandleFunc("/{meteo}/packStations", createTaskPackStation)
	r.HandleFunc("/{meteo}/kdtreeReload", handleKDTreeReload)
	r.HandleFunc("/{meteo}/infoclimat/updateStations/asTask", handleInfoclimatUpdateStations)
	r.HandleFunc("/{meteo}/infoclimat/updateStations", createTaskInfoclimatUpdateStations)
	r.HandleFunc("/{meteo}/infoclimat/getMonthlySerie", handleInfoclimatGetMonthlySerie)
	http.Handle("/", r)

}

func createTaskPackStation(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	res := resource.NewResources(c)

	if err := res.TaskQueue().AsTask(r.URL.Path+"/asTask", url.Values{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task created for stations packing"))

}

func createTaskInfoclimatUpdateStations(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)
	res := resource.NewResources(c)

	if len(r.URL.Query()) == 0 {
		for code, country := range infoclimat.GetCountry(res) {

			if err := res.TaskQueue().AsTask(r.URL.Path+"/asTask", url.Values{"country": {code}}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Logger().Infof("Task added for station update: [" + code + "] " + country)
		}
	} else {

		if ccc, ok := r.URL.Query()["country"]; ok {
			if len(ccc) == 0 {
				http.Error(w, "No country defined in query", http.StatusBadRequest)
				return
			}
			if err := res.TaskQueue().AsTask(r.URL.Path+"/asTask", r.URL.Query()); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			res.Logger().Infof("Task added for station update: [" + ccc[0] + "] ")
		}
	}
}
