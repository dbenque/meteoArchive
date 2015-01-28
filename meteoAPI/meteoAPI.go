package meteoAPI

import (
	"math"
	"strconv"
	"time"
)

// ------------------ Math for Longitude/Latitude to x,y,z coordinates --------------------
const (
	rayonTerre = 6371.0
	coordDim   = 3
	piOn180    = math.Pi / 180.
)

func toRad(d float64) float64 {
	return d * piOn180
}

// getCoord return array of coordinates
func (poi *POI) getCoord() [coordDim]float64 {
	if !poi.coordCached {
		poi.coord[0] = rayonTerre * math.Cos(toRad(poi.Latitude)) * math.Cos(toRad(poi.Longitude))
		poi.coord[1] = rayonTerre * math.Cos(toRad(poi.Latitude)) * math.Sin(toRad(poi.Longitude))
		poi.coord[2] = rayonTerre * math.Sin(toRad(poi.Latitude))
		poi.coordCached = true
	}
	return poi.coord
}

// ==================  Structures for Stations =================

// POI point of interest
type POI struct {
	Name        string
	Altitude    float64
	Latitude    float64
	Longitude   float64
	coord       [coordDim]float64
	coordCached bool
}

// Station meteo station information
type Station struct {
	POI
	Origin         string
	RemoteID       string
	RemoteMetadata map[string]interface{}
}

// Stations collection of stations
type Stations []Station

// ------------------------- Methods and Functions for stations -----------------------

//GetKey return a unique identifier for the station
func (p *Station) GetKey() string {
	return p.Origin + "." + p.RemoteID
}

//PutMetadata insert/modify a Metadata
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

// ======================= Measures structure ==========================

//Measure set of data that are measured by stations
type Measure struct {
	ExtremeMin *float64 `json:"emin,omitempty"`
	AverageMin *float64 `json:"amin,omitempty"`
	Average    *float64 `json:"a,omitempty"`
	AverageMax *float64 `json:"amax,omitempty"`
	ExtremeMax *float64 `json:"emax,omitempty"`

	WhaterMilimeter *float64 `json:"wmm,omitempty"`
	SunHours        *float64 `json:"sh,omitempty"`
}

//MonthlyMeasureSerie Represent erie of measure indexed by Months
type MonthlyMeasureSerie map[string]Measure // index computed as Year*100+Month

//----------------------------- Methods and helpers for measure ---------------------------
func (m *Measure) mergeMeasures(source *Measure) {

	if source == nil || m == nil {
		return
	}

	if m.Average == nil && source.Average != nil {
		m.Average = source.Average
	}
	if m.AverageMin == nil && source.AverageMin != nil {
		m.AverageMin = source.AverageMin
	}
	if m.AverageMax == nil && source.AverageMax != nil {
		m.AverageMax = source.AverageMax
	}
	if m.ExtremeMin == nil && source.ExtremeMin != nil {
		m.ExtremeMin = source.ExtremeMin
	}
	if m.ExtremeMax == nil && source.ExtremeMax != nil {
		m.ExtremeMax = source.ExtremeMax
	}
	if m.WhaterMilimeter == nil && source.WhaterMilimeter != nil {
		m.WhaterMilimeter = source.WhaterMilimeter
	}
	if m.SunHours == nil && source.SunHours != nil {
		m.SunHours = source.SunHours
	}

}

// Return the index in the monthlyMeasureSerie
func getMeasureIndex(year int, month time.Month) string {
	return strconv.Itoa(year) + "." + strconv.Itoa(int(month))
}

// PutMeasure create or merge non nil field into the MonthlyMeasureSerie
func (s *MonthlyMeasureSerie) PutMeasure(m *Measure, year int, month time.Month) {
	index := getMeasureIndex(year, month)
	data, found := (*s)[index]
	if !found {
		(*s)[index] = *m
	} else {
		data.mergeMeasures(m)
		(*s)[index] = data
	}
	return
}

//================= Storage ============

//Storage interface toward storage
type Storage interface {
	PutStation(p *Station) error
	GetStation(key string) *Station
	PutMonthlyMeasureSerie(p *Station, measures *MonthlyMeasureSerie) error
	GetMonthlyMeasureSerie(p *Station) *MonthlyMeasureSerie
	Persist()
	Initialize()
	GetAllStations() *Stations
}

//================= Web Grabber Interface ==

// MeteoWebsite interface that should  be implemented by a website grabber
type MeteoWebsite interface {
	//UpdateStations go to website and retrieve the list of stations and update the storage
	UpdateStations(s *Storage, inputCountryCode string)
	RetrieveMonthlyReports(station *Station, year int) MonthlyMeasureSerie
}
