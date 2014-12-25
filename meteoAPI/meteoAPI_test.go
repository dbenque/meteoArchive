package meteoAPI

import (

	"os"
	"testing"
	"time"
)

func TestStationsJSON(T *testing.T) {

	storageName := "ParisNice"

	paris := NewStation("Paris", 42, 2.3488000, 48.8534100)
	nice := NewStation("Nice", 18, 7.2660800, 43.7031300)

	paris.Origin = "test"
	paris.RemoteID = "0"
	nice.Origin = "test"
	nice.RemoteID = "1"

	mapStorage := NewMapStorage(storageName)
	mapStorage.PutStation(paris)
	mapStorage.PutStation(nice)
	mapStorage.Persist()

	anotherStorage := NewMapStorage(storageName)
	anotherStorage.Initialize()

	p := NewStation("", 0, 0, 0)
	p.Origin = "test"
	p.RemoteID = "0"

	paris2 := anotherStorage.GetStation(p.GetKey())

	if paris2.Altitude != 42 {
		T.Fatal("Paris altitude is not 42")
	}

	os.Remove(storageName + ".json")
}

func TestPutMeasure(T *testing.T) {
	serie := make(MonthlyMeasureSerie)
	m := new(Measure)
	a := 10.5
	m.Average = &a

	serie.PutMeasure(m, 2014, time.Month(1))


	m2 := new(Measure)
	aa := 106.5
	m2.SunHours = &aa

	serie.PutMeasure(m2, 2014, time.Month(1))

	if v,f := serie[100*2014+1]; !f {
		T.Fatal("nil in serie for hardcoded index")
		}else {
			if *(v.Average)!=10.5 {
				T.Fatal("not 10.5")
			}
			if *(v.SunHours)!=106.5 {
				T.Fatal("not 106.5")
			}

		}

}
