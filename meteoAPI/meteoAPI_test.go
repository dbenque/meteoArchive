package meteoAPI

import (
	"os"
	"testing"
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
