package main

import "meteoArchive/server"

func main() {

	meteoServer.Serve()
	done := make(chan bool)
	<-done
	return
}
