package geoloc

import (
	"fmt"
	"testing"
)

func TestGeolocFromCity(T *testing.T) {

	if poi, err := GeolocFromCity("New York", "us", "fr"); err != nil {
		T.Fatal("TestGeolocFromCity:", err)
	} else {
		fmt.Println(poi)
	}
}
