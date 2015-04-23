package meteoServer

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/dbenque/meteoArchive/geoloc"
	"github.com/dbenque/meteoArchive/meteoAPI"

	"code.google.com/p/biogo.store/kdtree"
)

type stationAndDistance struct {
	station  *meteoAPI.Station
	distance float64
}

func getNearestByStr(city string, country string, count int) (nearStations []stationAndDistance) {

	poi, err := geoloc.FromCity(city, country, "fr")
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

	city, country, err := readCityCountryFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	poi, err := geoloc.FromCity(city, country, "fr")
	if err != nil {
		w.Write([]byte("Geolo not found"))
		w.WriteHeader(http.StatusNotFound)
	}

	dataj, _ := json.Marshal(poi)

	w.WriteHeader(http.StatusOK)
	w.Write(dataj) //[]byte(output)

}

func handleDistance(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Distance"))

	city, country, err := readCityCountryFromURL(r)
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

func handleNear(w http.ResponseWriter, r *http.Request) {

	var nearStations []stationAndDistance

	count, err := readCountFromURL(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	var city, country string
	latitude, longitute, err := readLatitudeLongitudeFromURL(r)
	if err != nil {
		err = nil
		city, country, err = readCityCountryFromURL(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		fmt.Println("By Str")
		nearStations = getNearestByStr(city, country, count)

	} else {
		fmt.Println("By Coord")
		nearStations = getNearestByCoord(latitude, longitute, count)
	}

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

func handleKDTreeReload(w http.ResponseWriter, r *http.Request) {

	go func() {

		defer func() { fmt.Println("Update kdtree completed/ended") }()

		serverStorage.Initialize()
		stations := serverStorage.GetAllStations()
		fmt.Println("Station Count:", len(*stations))

		kdtreeOfStation = kdtree.New(stations, true)
	}()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Update on going for kdtree"))
}
