package moduleMeteo

import (
	"errors"

	"github.com/dbenque/meteoArchive/appengineStorage"
	"github.com/dbenque/meteoArchive/resource"
	"github.com/dbenque/meteoArchive/server"

	"net/http"

	"appengine"
	"appengine/urlfetch"
)

func createAppengineURLFetcher(r interface{}) (resource.URLGetter, error) {

	switch r.(type) {
	default:
		return nil, errors.New("Can't create appengine URLFetcher from that interface type")
	case *http.Request:
		return urlfetch.Client(appengine.NewContext(r.(*http.Request))), nil
	}
}

func createAppengineLogger(r interface{}) (resource.Logger, error) {

	switch r.(type) {
	default:
		return nil, errors.New("Can't create appengine Logger from that interface type")
	case *http.Request:
		return appengine.NewContext(r.(*http.Request)), nil
	case appengine.Context:
		return r.(appengine.Context), nil
	}

}

func createAppengineStorage(r interface{}) (resource.Storage, error) {
	switch r.(type) {
	default:
		return nil, errors.New("Can't create appengine Storage from that interface type")
	case *http.Request:
		return appengineStorage.NewAppengineStorage(appengine.NewContext(r.(*http.Request))), nil
	case appengine.Context:
		return appengineStorage.NewAppengineStorage(r.(appengine.Context)), nil
	}

}

func init() {

	resource.ResourceFactoryInstance.Client = createAppengineURLFetcher
	resource.ResourceFactoryInstance.Logger = createAppengineLogger
	resource.ResourceFactoryInstance.Storage = createAppengineStorage
	// setup http handler using local storage
	meteoServer.ApplyHttpHandler()

}
