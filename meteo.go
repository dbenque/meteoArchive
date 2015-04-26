package main

import (
	"log"
	"net/http"

	"github.com/dbenque/meteoArchive/logger"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"
	"github.com/dbenque/meteoArchive/server"
)

func createURLFetcher(r interface{}) (resource.URLGetter, error) {
	return &http.Client{CheckRedirect: nil}, nil
}

func createAppengineLogger(r interface{}) (resource.Logger, error) {
	return logger.New(), nil
}

func main() {

	resource.ResourceFactoryInstance.Client = createURLFetcher
	resource.ResourceFactoryInstance.Logger = createAppengineLogger

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
