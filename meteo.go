package main

import (
	"log"
	"net/http"

	"github.com/dbenque/meteoArchive/client"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/server"
)

func createURLFetcher(r *http.Request) meteoClient.URLGetter {
	return &http.Client{CheckRedirect: nil}
}

func main() {

	// setup http handler using local storage
	meteoServer.ApplyHttpHandler(meteoAPI.NewMapStorage("mapStorage"), createURLFetcher)

	// Serve
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	done := make(chan bool)
	<-done
	return
}
