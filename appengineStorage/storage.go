package appengineStorage

import (
	"encoding/json"
	"strconv"

	"github.com/dbenque/goAppengineToolkit/datastoreEntity"
	"github.com/dbenque/meteoArchive/meteoAPI"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
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

//GetAllStations retireve all the stations that have been previously packed
func (s *AppEngineStorage) GetAllStations() (*meteoAPI.Stations, error) {

	stations := make(meteoAPI.Stations, 0, 0)
	blob := make([]byte, 0, 0)

	// retrieve thanks to memcached
	item, err := memcache.Get(s.context, MaxChunckIndexKey)
	if err == nil && err != memcache.ErrCacheMiss {

		s.context.Infof("Retrieving stations chunks index from memcached")

		maxChunk, _ := strconv.Atoi(string(item.Value))
		for i := 0; i <= maxChunk; i++ {
			x := PackedStationsChunk{}
			x.Index = i
			if errd := datastoreEntity.Retrieve(s.context, &x); errd != nil {
				return nil, errd
			}
			blob = append(blob, x.Chunk...)
		}
	} else {

		s.context.Infof("Retrieving stations chunks querying datastore")

		// retrieve chunks and aggregate blob
		q := datastore.NewQuery(PackedStationsChunkKind).Order("Index")
		maxIndex := 0
		for t := q.Run(s.context); ; {
			var x PackedStationsChunk
			_, err := t.Next(&x)
			if err == datastore.Done {
				break
			}
			if err != nil {
				return nil, err
			}
			blob = append(blob, x.Chunk...)
			maxIndex++
		}

		saveMaxChunckIndexinMemcached(s.context, maxIndex-1)

	}

	err = json.Unmarshal(blob, &stations)

	return &stations, err

}
func (s *AppEngineStorage) Initialize() error {
	return nil
}
func (s *AppEngineStorage) Persist() error {
	return nil
}

// ---------------------------------------------
// Chunks --------------------------------------
// ---------------------------------------------
const chunkSize = 1024 * 600 // 1Mo which is the max for entity size - buffer for index and 96 bytes reserved for the memcache internals (https://cloud.google.com/appengine/docs/go/memcache/)
const PackedStationsChunkKind = "PackedStationsChunk"
const MaxChunckIndexKey = "MaxChunckIndex"

// PackedStationsChunk -------------------------
type PackedStationsChunk struct {
	Chunk []byte `datastore:",noindex"`
	Index int
}

func (s *PackedStationsChunk) GetKey() string {
	return strconv.Itoa((*s).Index)
}

func (s *PackedStationsChunk) GetKind() string {
	return PackedStationsChunkKind
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func newPackedStationsChunk(s *[]byte, nb int) *PackedStationsChunk {
	return &PackedStationsChunk{(*s)[min(len(*s), nb*chunkSize):min(len(*s), (nb+1)*chunkSize)], nb}
}

//PackStations create chunks contain the json serialization of all the stations
func (s *AppEngineStorage) PackStations() error {

	// Get all the stations
	stations := make(meteoAPI.Stations, 0, 0)
	q := datastore.NewQuery(meteoAPI.StationKind)
	for t := q.Run(s.context); ; {
		var x meteoAPI.Station
		_, err := t.Next(&x)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		stations = append(stations, x)
	}

	// build new blob
	dataj, _ := json.Marshal(stations)
	s.context.Infof("Packing %d stations as %d bytes", len(stations), len(dataj))

	// delete all previous chunck
	qChunk := datastore.NewQuery(PackedStationsChunkKind).KeysOnly()
	if keys, err := qChunk.GetAll(s.context, nil); err == nil && keys != nil {
		s.context.Infof("Deleting previous %d chunk(s)", len(keys))
		if datastore.DeleteMulti(s.context, keys) != nil {
			s.context.Errorf("unable to perform delete of previous chunks: %s", err.Error())
		}
	} else {
		if err != nil {
			s.context.Errorf("Can't query previous chunks: %s", err.Error())
		}

		if keys == nil {
			s.context.Warningf("Looks like there was no chunks before...")
		}

	}

	// create new chunks
	l := len(dataj) / chunkSize
	s.context.Infof("Creating new %d chunk(s)", l+1)
	for i := 0; i <= l; i++ {
		pchunk := newPackedStationsChunk(&dataj, i)
		datastoreEntity.Store(s.context, pchunk)
	}

	saveMaxChunckIndexinMemcached(s.context, l)

	return nil
}

func saveMaxChunckIndexinMemcached(c appengine.Context, index int) {
	itemChunkIndex := &memcache.Item{
		Key:   MaxChunckIndexKey,
		Value: []byte(strconv.Itoa(index)),
	}
	memcache.Set(c, itemChunkIndex)
}
