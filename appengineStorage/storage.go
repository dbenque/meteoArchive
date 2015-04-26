package appengineStorage

import (
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

func (s *AppEngineStorage) PutMonthlyMeasureSerie(p *meteoAPI.Station, measures *meteoAPI.MonthlyMeasureSerie) error {
	return nil
}
func (s *AppEngineStorage) GetMonthlyMeasureSerie(p *meteoAPI.Station) *meteoAPI.MonthlyMeasureSerie {
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
