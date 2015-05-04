package moduleMeteo

import (
	"errors"

	"github.com/dbenque/meteoArchive/appengineStorage"
	"github.com/dbenque/meteoArchive/resource"
	"github.com/dbenque/meteoArchive/server"

	"net/http"
	"net/url"

	"appengine"
	"appengine/taskqueue"
	"appengine/urlfetch"
)

func createAppengineURLFetcher(r interface{}) (resource.URLGetter, error) {

	switch r.(type) {
	default:
		return nil, errors.New("Can't create appengine URLFetcher from that interface type")
	case *http.Request:
		return urlfetch.Client(appengine.NewContext(r.(*http.Request))), nil
	case appengine.Context:
		return urlfetch.Client(r.(appengine.Context)), nil

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

func createTaskQueue(r interface{}) (resource.TaskQueue, error) {
	switch r.(type) {
	default:
		return nil, errors.New("Can't create appengine TaskQueue from that interface type")
	case *http.Request:
		return &Tasker{appengine.NewContext(r.(*http.Request))}, nil
	case appengine.Context:
		return &Tasker{r.(appengine.Context)}, nil
	}
}

type Tasker struct {
	context appengine.Context
}

func (t *Tasker) AsTask(path string, params url.Values) error {

	task := taskqueue.NewPOSTTask(path, params)
	if _, err := taskqueue.Add(t.context, task, ""); err != nil {
		return err
	}
	t.context.Infof("Task added in appengine url:params %s:%v", path, params)
	return nil

}

func init() {

	resource.ResourceFactoryInstance.Client = createAppengineURLFetcher
	resource.ResourceFactoryInstance.Logger = createAppengineLogger
	resource.ResourceFactoryInstance.Storage = createAppengineStorage
	resource.ResourceFactoryInstance.TaskQueue = createTaskQueue
	// setup http handler using local storage
	meteoServer.ApplyHttpHandler()

}
