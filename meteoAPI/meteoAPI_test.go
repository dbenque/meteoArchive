package meteoAPI

import (
	"os"
	"testing"
)

func TestStationsJSON(T *testing.T) {

	filename := "ParisNice.json"

	paris := Station{POI{"Paris", 42, 2.3488000, 48.8534100, [...]float64{0., 0., 0.}, false}, "test", "1"}
	nice := Station{POI{"Nice", 18, 7.2660800, 43.7031300, [...]float64{0., 0., 0.}, false}, "test", "2"}

	stationsOut := [...]Station{paris, nice}

	if StationsAsJSONFile(filename, stationsOut[:]) != nil {
		T.Fatal("StationsAsJSONFile test")
	}

	stationIn, err := StationsFromJSONFile(filename)

	if err != nil {
		T.Fatal("StationsFromJSONFile test")
	}

	if stationIn[0].Altitude != 42 {
		T.Fatal("Paris altitude is not 42")
	}

	os.Remove(filename)
}
