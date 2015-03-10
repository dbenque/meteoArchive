package main

import (
	"log"
	"net/http"

	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/server"
)

func main() {

	// setup http handler using local storage
	meteoServer.ApplyHttpHandler(meteoAPI.NewMapStorage("mapStorage"))

	// Serve
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	done := make(chan bool)
	<-done
	return
}
