package meteoServer

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/dbenque/meteoArchive/geoloc"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"

	"github.com/biogo/store/kdtree"
)

//"code.google.com/p/biogo.store/kdtree"
type stationAndDistance struct {
	station  *meteoAPI.Station
	distance float64
}

func getNearestByStr(res *resource.ResourceInstances, city string, country string, count int) (nearStations []stationAndDistance) {

	poi, err := geoloc.FromCity(res, city, country, "fr")
	if err != nil {
		return
	}
	return getNearest(poi, count)
}

func getNearestByCoord(latitude, longitute float64, count int) (nearStations []stationAndDistance) {

	poi := meteoAPI.POI{Latitude: latitude, Longitude: longitute}

	return getNearest(poi, count)
}

func getNearest(poi meteoAPI.POI, count int) (nearStations []stationAndDistance) {

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

func handleGetGeoloc(w http.ResponseWriter, r *http.Request) {

	res := resource.NewResources(r)

	city, country, err := readCityCountryFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	poi, err := geoloc.FromCity(res, city, country, "fr")
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dataj, _ := json.Marshal(poi)

	w.WriteHeader(http.StatusOK)
	w.Write(dataj) //[]byte(output)

}

func handleDistance(w http.ResponseWriter, r *http.Request) {

	res := resource.NewResources(r)
	ensureKdtreeLoaded(res)

	w.Write([]byte("Distance"))

	city, country, err := readCityCountryFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	poi, err := geoloc.FromCity(res, city, country, "fr")
	if err != nil {
		w.Write([]byte("Geolo not found"))
		return
	}

	rayon := 100.0
	if rayonStr, ok := r.URL.Query()["d"]; ok {
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

func getNearestFromRequest(r *http.Request, defaultCount int) (nearStations []stationAndDistance, err error) {

	res := resource.NewResources(r)
	ensureKdtreeLoaded(res)

	// Retrieve the count
	count, err := readCountFromURL(r)
	if err != nil {
		if defaultCount > 0 {
			count = defaultCount
		} else {
			return
		}
	}

	// Retrieve the location
	var city, country string
	latitude, longitute, err := readLatitudeLongitudeFromURL(r)
	if err != nil {
		err = nil
		city, country, err = readCityCountryFromURL(r)
		if err != nil {
			return
		}
		nearStations = getNearestByStr(res, city, country, count)

	} else {
		nearStations = getNearestByCoord(latitude, longitute, count)
	}

	if len(nearStations) == 0 {
		err = errors.New("Geolo not found")
		return
	}

	return
}

func handleNear(w http.ResponseWriter, r *http.Request) {

	res := resource.NewResources(r)
	ensureKdtreeLoaded(res)

	if nearStations, err := getNearestFromRequest(r, 3); err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	} else {

		output := "Result:\n"
		for _, s := range nearStations {
			output = output + "Station at " + strconv.Itoa(int(s.distance)) + " km, " + s.station.Name + "\n"
		}

		w.Write([]byte(output))
	}
}

func ensureKdtreeLoaded(res *resource.ResourceInstances) error {
	if kdtreeOfStation == nil {
		return kdtreeReload(res)
	}
	return nil
}

func kdtreeReload(res *resource.ResourceInstances) error {

	if err := GetServerStorage(res.Context).Initialize(); err != nil {
		return err
	}

	stations, err := GetServerStorage(res.Context).GetAllStations()
	if err != nil {
		res.Logger().Errorf("Fail to load all stations from the store: %s", err.Error())
		return err
	}
	res.Logger().Infof("Loading KDTree with %d stations", len(*stations))

	kdtreeOfStation = kdtree.New(stations, true)

	return nil
}

func handleKDTreeReload(w http.ResponseWriter, r *http.Request) {

	res := resource.NewResources(r)

	if err := kdtreeReload(res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update on going for kdtree"))
}
