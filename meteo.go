package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dbenque/meteoArchive/logger"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"
	"github.com/dbenque/meteoArchive/server"
)

func createURLFetcher(r interface{}) (resource.URLGetter, error) {
	return &http.Client{CheckRedirect: nil}, nil
}

func createLogger(r interface{}) (resource.Logger, error) {
	return logger.New(), nil
}

func createStorage(r interface{}) (resource.Storage, error) {
	return meteoAPI.NewMapStorage("mapStorage"), nil
}

func createTaskQueue(r interface{}) (resource.TaskQueue, error) {
	return &Tasker{}, nil
}

func main() {

	resource.ResourceFactoryInstance.Client = createURLFetcher
	resource.ResourceFactoryInstance.Logger = createLogger
	resource.ResourceFactoryInstance.Storage = createStorage
	resource.ResourceFactoryInstance.Storage = createTaskQueue

	// setup http handler using local storage
	meteoServer.ApplyHttpHandler()

	// Serve
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	done := make(chan bool)
	<-done
	return
}

type Tasker struct {
}

func (t *Tasker) AsTask(path string, params url.Values) error {
	go func() {
		req, err := http.NewRequest("POST", path, strings.NewReader(params.Encode()))
		hc := http.Client{}
		resp, err := hc.Do(req)
	}()
	return nil
}
