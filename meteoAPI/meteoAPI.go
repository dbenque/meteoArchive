package meteoAPI

import (
	"encoding/json"
	"math"
	"strconv"
	"time"
)

// ------------------ Math for Longitude/Latitude to x,y,z coordinates --------------------
const (
	rayonTerre       = 6371.0
	coordDim         = 3
	piOn180          = math.Pi / 180.
	StationKind      = "station"
	MonthlySerieKind = "monthlySerie"
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
	Name        string            `json:"name,omitempty"`
	Altitude    float64           `json:"alt,omitempty"`
	Latitude    float64           `json:"lat,omitempty"`
	Longitude   float64           `json:"lon,omitempty"`
	coord       [coordDim]float64 `datastore:"-"`
	coordCached bool              `datastore:"-"`
}

// Station meteo station information
type Station struct {
	POI
	Origin          string
	RemoteID        string
	remoteMetadata  map[string]interface{} `datastore:"-"`
	MetadataInStore string                 `datastore:"metedata,noindex"` // Public for datastore access. Don't play wit hthe value!
}

// Stations collection of stations
type Stations []Station

// ------------------------- Methods and Functions for stations -----------------------

//GetKey return a unique identifier for the station
func (p *Station) GetKey() string {
	return BuildStationKey(p.Origin, p.RemoteID)
}

//GetKey return a unique identifier for the station
func BuildStationKey(origin string, remoteId string) string {
	return origin + ":" + remoteId
}

//GetKey return a unique identifier for the station
func (p *Station) GetKind() string {
	return StationKind
}

//PutMetadata insert/modify a Metadata
func (p *Station) PutMetadata(key string, value interface{}) {
	if p.remoteMetadata == nil {
		p.remoteMetadata = make(map[string]interface{})
	}
	p.remoteMetadata[key] = value
	dataj, _ := json.Marshal(p.remoteMetadata)
	p.MetadataInStore = string(dataj)
}

//PutMetadata insert/modify a Metadata
func (p *Station) GetMetadata(key string) interface{} {
	if p.remoteMetadata == nil {
		if len(p.MetadataInStore) == 0 {
			return nil
		}
		json.Unmarshal([]byte(p.MetadataInStore), &(p.remoteMetadata))
	}

	return p.remoteMetadata[key]
}

// NewStation constructor for station
func NewStation(name string, alt, lat, lng float64) *Station {
	return &Station{POI{name, alt, lat, lng, [...]float64{0., 0., 0.}, false}, "", "", nil, ""}
}

// NewStationFromPOI constructor for station based on POI
func NewStationFromPOI(poi POI) *Station {
	return &Station{poi, "", "", nil, ""}
}

// ======================= Measures structure ==========================

//Measure set of data that are measured by stations
type Measure struct {
	ExtremeMin *float64 `json:"em,omitempty"`
	AverageMin *float64 `json:"am,omitempty"`
	Average    *float64 `json:"a,omitempty"`
	AverageMax *float64 `json:"aM,omitempty"`
	ExtremeMax *float64 `json:"eM,omitempty"`

	WhaterMilimeter *float64 `json:"w,omitempty"`
	SunHours        *float64 `json:"s,omitempty"`
}

//MonthlyMeasureSerie Represent serie of measure indexed by Months
type MonthlyMeasureSerie map[string]Measure // index computed thanks to func getMeasureIndex

//----------------------------- Methods and helpers for measure ---------------------------

//IsEmpty check that at least one value of the measure is valid
func (m *Measure) IsEmpty() bool {
	return (m.ExtremeMin == nil && m.AverageMin == nil && m.Average == nil && m.AverageMax == nil && m.ExtremeMax == nil && m.WhaterMilimeter == nil && m.SunHours == nil)
}

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

// GetMeasure from the MonthlyMeasureSerie. Nil if it does not exist
func (s *MonthlyMeasureSerie) GetMeasure(year int, month time.Month) *Measure {
	index := getMeasureIndex(year, month)
	data, found := (*s)[index]
	if !found {
		return nil
	}
	return &data

}

//GetSerieIndexedByMonth returns the serie index by month only and with empty values purged
func (s *MonthlyMeasureSerie) GetSerieIndexedByMonth(year int) MonthlyMeasureSerie {
	outputSerie := make(MonthlyMeasureSerie)
	for i := 1; i <= 12; i++ {
		if m := s.GetMeasure(year, time.Month(i)); m != nil && !m.IsEmpty() {
			// Only use the month as index
			outputSerie[strconv.Itoa(i)] = *m
		}
	}
	return outputSerie
}

//================= Storage ============

//Storage interface toward storage
type Storage interface {
	PutStation(p *Station) error
	GetStation(origin string, remoteId string) *Station
	PutMonthlyMeasureSerie(p *Station, measures *MonthlyMeasureSerie) error
	GetMonthlyMeasureSerie(p *Station) *MonthlyMeasureSerie
	Persist() error
	Initialize() error
	GetAllStations() (*Stations, error)
}

//================= Web Grabber Interface ==

// MeteoWebsite interface that should  be implemented by a website grabber
type MeteoWebsite interface {
	//UpdateStations go to website and retrieve the list of stations and update the storage
	UpdateStations(s *Storage, inputCountryCode string)
	RetrieveMonthlyReports(station *Station, year int) MonthlyMeasureSerie
}
