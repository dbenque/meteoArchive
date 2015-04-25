package moduleMeteo

import (
	"github.com/dbenque/meteoArchive/client"
	"github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/server"

	"net/http"

	"appengine"
	"appengine/urlfetch"
)

func createAppengineURLFetcher(r *http.Request) meteoClient.URLGetter {
	return urlfetch.Client(appengine.NewContext(r))
}

func init() {
	// setup http handler using local storage
	meteoServer.ApplyHttpHandler(meteoAPI.NewMapStorage("mapStorage"), createAppengineURLFetcher)

}
