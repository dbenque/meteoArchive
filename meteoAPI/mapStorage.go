package meteoAPI

import (
	"encoding/json"
	"io/ioutil"
)

//MapStorage for local testing
type MapStorage struct {
	name          string
	Stations      map[string]Station
	MonthlySeries map[string]MonthlyMeasureSerie
}

//NewMapStorage initialize a new storage
func NewMapStorage(name string) *MapStorage {
	var storage MapStorage
	storage.name = name
	storage.Stations = make(map[string]Station)
	storage.MonthlySeries = make(map[string]MonthlyMeasureSerie)

	return &storage
}

//PutStation insert a station in the MapStorage
func (s *MapStorage) PutStation(p *Station) error {
	s.Stations[p.GetKey()] = *p
	return nil
}

//GetStation get a station from the mapStore
func (s *MapStorage) GetStation(key string) *Station {
	sta, found := s.Stations[key]
	if !found {
		return nil
	}

	instance := sta // create a clone to better mimic a DB

	return &instance
}

//PutMonthlyMeasureSerie store measure for a station
func (s *MapStorage) PutMonthlyMeasureSerie(p *Station, measures *MonthlyMeasureSerie) error {
	s.MonthlySeries[p.GetKey()] = *measures
	return nil
}

//GetMonthlyMeasureSerie get the monthly measure serie associated to the station
func (s *MapStorage) GetMonthlyMeasureSerie(p *Station) *MonthlyMeasureSerie {
	m, found := s.MonthlySeries[p.GetKey()]
	if !found {
		return nil
	}
	return &m
}

//Persist persist to file
func (s *MapStorage) Persist() {
	dataj, _ := json.Marshal(*s)
	ioutil.WriteFile(s.name+".json", dataj, 0644)
}

//Initialize retrieve from file
func (s *MapStorage) Initialize() {
	filecontent, err := ioutil.ReadFile(s.name + ".json")
	if err != nil {
		return
	}
	err = json.Unmarshal(filecontent, s)
}

//GetStations return all the stations in the storage
func (s *MapStorage) GetAllStations() *Stations {

	result := make(Stations, len(s.Stations), len(s.Stations))

	i := 0
	for _, v := range s.Stations {
		result[i] = v
		i++
	}
	return &result
}
