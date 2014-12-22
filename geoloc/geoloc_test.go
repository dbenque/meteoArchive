package geoloc

import (
	"fmt"
	"testing"
)

func TestGeolocFromCity(T *testing.T) {

	if poi, err := FromCity("New York", "us", "fr"); err != nil {
		T.Fatal("TestGeolocFromCity:", err)
	} else {
		fmt.Println(poi)
	}
}
