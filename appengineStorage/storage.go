package appengineStorage

import (
	"encoding/json"

	"github.com/dbenque/goAppengineToolkit/datastoreEntity"
	"github.com/dbenque/meteoArchive/meteoAPI"

	"appengine"
	"appengine/datastore"
)

type AppEngineStorage struct {
	context appengine.Context
}

func NewAppengineStorage(context appengine.Context) *AppEngineStorage {
	return &AppEngineStorage{context}
}

func (s *AppEngineStorage) PutStation(p *meteoAPI.Station) error {
	return datastoreEntity.Store(s.context, p)
}

func (s *AppEngineStorage) GetStation(origin string, remoteId string) *meteoAPI.Station {
	sta := meteoAPI.Station{}
	sta.Origin = origin
	sta.RemoteID = remoteId
	if err := datastoreEntity.Retrieve(s.context, &sta); err != nil {
		return &sta
	}
	return nil
}

type MonthlyMeasureSerieInDatastore struct {
	station  *meteoAPI.Station
	measures *meteoAPI.MonthlyMeasureSerie
	Series   string `datastore:",noindex"`
}

//GetKey return a unique identifier for the station
func (p *MonthlyMeasureSerieInDatastore) GetKey() string {
	return p.station.GetKey()
}

//GetKey return a unique identifier for the station
func (p *MonthlyMeasureSerieInDatastore) GetKind() string {
	return meteoAPI.MonthlySerieKind
}

func (s *AppEngineStorage) PutMonthlyMeasureSerie(p *meteoAPI.Station, measures *meteoAPI.MonthlyMeasureSerie) error {

	dataj, _ := json.Marshal(*measures)
	v := MonthlyMeasureSerieInDatastore{p, measures, string(dataj)}
	return datastoreEntity.Store(s.context, &v)

}
func (s *AppEngineStorage) GetMonthlyMeasureSerie(p *meteoAPI.Station) *meteoAPI.MonthlyMeasureSerie {

	v := MonthlyMeasureSerieInDatastore{}
	v.station = p
	if err := datastoreEntity.Retrieve(s.context, &v); err == nil {
		m := meteoAPI.MonthlyMeasureSerie{}
		if err = json.Unmarshal([]byte(v.Series), &m); err != nil {
			return nil
		}
		return &m
	}
	return nil

}
func (s *AppEngineStorage) GetAllStations() (*meteoAPI.Stations, error) {

	stations := make(meteoAPI.Stations, 0, 0)

	q := datastore.NewQuery(meteoAPI.StationKind)

	for t := q.Run(s.context); ; {
		var x meteoAPI.Station
		_, err := t.Next(&x)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		stations = append(stations, x)
	}

	return &stations, nil

}
func (s *AppEngineStorage) Initialize() error {
	return nil
}
func (s *AppEngineStorage) Persist() error {
	return nil
}
