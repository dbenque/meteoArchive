package meteoAPI

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"time"

	"code.google.com/p/biogo.store/kdtree"
)

// ------------------ Stations --------------------
const (
	rayonTerre = 6371.0
	coordDim   = 3
	// Randoms is the maximum number of random values to sample for calculation of median of
	// random elements
	nbRandoms = 1000
)

// POI point of interest
type POI struct {
	Name        string
	Altitude    float64
	Latitude    float64
	Longitude   float64
	coord       [coordDim]float64
	coordCached bool
}

const piOn180 = math.Pi / 180.

func toRad(d float64) float64 {
	return d * piOn180
}

// GetCoord return array of coordinates
func (poi *POI) GetCoord() [coordDim]float64 {
	if !poi.coordCached {
		poi.coord[0] = rayonTerre * math.Cos(toRad(poi.Latitude)) * math.Cos(toRad(poi.Longitude))
		poi.coord[1] = rayonTerre * math.Cos(toRad(poi.Latitude)) * math.Sin(toRad(poi.Longitude))
		poi.coord[2] = rayonTerre * math.Sin(toRad(poi.Latitude))
		poi.coordCached = true
	}
	return poi.coord
}

// Station meteo station information
type Station struct {
	POI
	Origin         string
	RemoteID       string
	RemoteMetadata map[string]interface{}
}

func (p *Station) PutMetadata(key string, value interface{}) {
	if p.RemoteMetadata == nil {
		p.RemoteMetadata = make(map[string]interface{})
	}
	p.RemoteMetadata[key] = value
}

// NewStation constructor for station
func NewStation(name string, alt, lat, lng float64) *Station {
	return &Station{POI{name, alt, lat, lng, [...]float64{0., 0., 0.}, false}, "", "", nil}
}

// NewStationFromPOI constructor for station based on POI
func NewStationFromPOI(poi POI) *Station {
	return &Station{poi, "", "", nil}
}

// Stations collection of stations
type Stations []Station

// ---------------Implementation of kdtree point interface for station --------------

// Compare compare stations on a given dimension
func (p Station) Compare(c kdtree.Comparable, d kdtree.Dim) float64 {
	q := c.(Station)
	return p.GetCoord()[d] - q.GetCoord()[d]
}

//Dims Return the number of dimension associated to a station
func (p Station) Dims() int { return coordDim }

//Distance compute the distance between 2 stations
func (p Station) Distance(c kdtree.Comparable) float64 {
	q := c.(Station)
	var sum float64
	for dim, c := range p.GetCoord() {
		d := c - q.GetCoord()[dim]
		sum += d * d
	}
	return math.Sqrt(sum)
}

//Index return the station associated to the given index
func (p Stations) Index(i int) kdtree.Comparable { return p[i] }

//Len return the number of stations stored in the list of stations
func (p Stations) Len() int { return len(p) }

//Pivot compute the pivot index on the given dimension
func (p Stations) Pivot(d kdtree.Dim) int { return stPlane{Stations: p, Dim: d}.Pivot() }

//Slice slicer
func (p Stations) Slice(start, end int) kdtree.Interface { return p[start:end] }

// An nbPlane is a wrapping type that allows a Points type be pivoted on a dimension.
type stPlane struct {
	kdtree.Dim
	Stations
}

func (p stPlane) Less(i, j int) bool {
	return p.Stations[i].GetCoord()[p.Dim] < p.Stations[j].GetCoord()[p.Dim]
}

func medianOf(list kdtree.SortSlicer) int {
	n := list.Len()
	kdtree.Select(list.Slice(0, n), n/2)
	return n / 2
}

//Pivot compute the pivot for the given plane
func (p stPlane) Pivot() int { return kdtree.Partition(p, medianOf(p)) }

//func (p stPlane) Pivot() int                             { return kdtree.Partition(p, kdtree.MedianOfRandoms(p, nbRandoms)) }

//Slice slicer
func (p stPlane) Slice(start, end int) kdtree.SortSlicer { p.Stations = p.Stations[start:end]; return p }

//Swap swapper
func (p stPlane) Swap(i, j int) {
	p.Stations[i], p.Stations[j] = p.Stations[j], p.Stations[i]
}

// ------------------- JSON and File helpers --------------------------

// StationsAsJSONFile serialize stations to file
func StationsAsJSONFile(filename string, stations Stations) error {
	dataj, _ := json.Marshal(stations)
	return ioutil.WriteFile(filename, dataj, 0644)
}

// StationsFromJSONFile serialize stations to file
func StationsFromJSONFile(filename string) (stations Stations, err error) {
	filecontent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(filecontent, &stations)

	return
}

// ------------------ Measure --------------------

//Measure set of data that are measured by stations
type Measure struct {
	ExtremeMin float32
	AverageMin float32
	Average    float32
	AverageMax float32
	ExtremeMax float32

	WhaterMilimeter float32
	SunHours        float32
}

// ----------------- Data --------------------------

//MonthlyReport average over a month for measure on a given station
type MonthlyReport struct {
	Measure
	Station
	Month time.Month
	Year  int
}
