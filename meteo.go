package main

import "github.com/dbenque/meteoArchive/server"

func main() {

	meteoServer.Serve()
	done := make(chan bool)
	<-done
	return
}
