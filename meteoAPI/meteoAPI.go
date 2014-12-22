package meteoAPI

import (
	"math"
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
	ExtremeMin float64
	AverageMin float64
	Average    float64
	AverageMax float64
	ExtremeMax float64

	WhaterMilimeter float64
	SunHours        float64
}

//MonthlyMeasureSerie Represent a serie of measure indexed by Months
type MonthlyMeasureSerie struct {
	Serie map[int]Measure // index computed as Year*100+Month
}

//MeasureStorage interface toward storage
type MeasureStorage interface {
}

//----------------------------- Methods and helpers for measure ---------------------------

func (m *Measure) mergeMeasures(source *Measure) {

	if source == nil || m == nil {
		return
	}

	if m.Average == 0 && source.Average != 0 {
		m.Average = source.Average
	}
	if m.AverageMin == 0 && source.AverageMin != 0 {
		m.AverageMin = source.AverageMin
	}
	if m.AverageMax == 0 && source.AverageMax != 0 {
		m.AverageMax = source.AverageMax
	}
	if m.ExtremeMin == 0 && source.ExtremeMin != 0 {
		m.ExtremeMin = source.ExtremeMin
	}
	if m.ExtremeMax == 0 && source.ExtremeMax != 0 {
		m.ExtremeMax = source.ExtremeMax
	}
	if m.WhaterMilimeter == 0 && source.WhaterMilimeter != 0 {
		m.WhaterMilimeter = source.WhaterMilimeter
	}
	if m.SunHours == 0 && source.SunHours != 0 {
		m.SunHours = source.SunHours
	}

}

// Return the index in the monthlyMeasureSerie
func getMeasureIndex(year int, month time.Month) int {
	return year*100 + int(month)
}

// PutMeasure create or merge non nil field into the MonthlyMeasureSerie
func (s *MonthlyMeasureSerie) PutMeasure(m Measure, year int, month time.Month) {
	index := getMeasureIndex(year, month)
	data, found := s.Serie[index]
	if !found {
		s.Serie[index] = m
	} else {
		data.mergeMeasures(&m)
		s.Serie[index] = data
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
}
