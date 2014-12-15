package main

import (
	"meteo/infoclimat"
	"testing"
	//  "meteo/meteoAPI"

	"code.google.com/p/biogo.store/kdtree"
)

func BenchmarkMeteo(b *testing.B) {

	stations, _ := infoclimat.GetStations(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		kdtree.New(stations, true)
	}

}
