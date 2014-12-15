package main

import (
	"fmt"
	"meteo/infoclimat"
	"meteo/server"

	"code.google.com/p/biogo.store/kdtree"
)

func main() {

	stations, _ := infoclimat.GetStations(true)

	tree := kdtree.New(stations, true)
	fmt.Println(tree.Len())
	meteoServer.Serve()
	done := make(chan bool)
	<-done
	return
}
