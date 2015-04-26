package resource

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

//------------- Client to fetch URL

//URLGetter interface that define the method to retrieve URL
type URLGetter interface {
	Get(url string) (*http.Response, error)
}

// Factory Function for client
type URLGetterFactory func(c interface{}) (URLGetter, error)

// Util to Get a goquery.Document
func GetGoqueryDocument(getter URLGetter, url string) (*goquery.Document, error) {

	res, errGet := getter.Get(url)
	if errGet != nil {
		return nil, errGet
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, errGet
	}

	return doc, err
}

//------------- Logger

//Logger interface to abstract which logger to use (Appengine, golog, other)
type Logger interface {
	// Debugf formats its arguments according to the format, analogous to fmt.Printf,
	// and records the text as a log message at Debug level.
	Debugf(format string, args ...interface{})

	// Infof is like Debugf, but at Info level.
	Infof(format string, args ...interface{})

	// Warningf is like Debugf, but at Warning level.
	Warningf(format string, args ...interface{})

	// Errorf is like Debugf, but at Error level.
	Errorf(format string, args ...interface{})

	// Criticalf is like Debugf, but at Critical level.
	Criticalf(format string, args ...interface{})
}

// Factory Function for client
type LoggerFactory func(c interface{}) (Logger, error)

//------------- Storage

type Storage interface {
}

type StorageFactory func(c interface{}) (Storage, error)

//------------- Resource Factories

type ResourceFactory struct {
	Client  URLGetterFactory
	Logger  LoggerFactory
	Storage StorageFactory
}

var ResourceFactoryInstance ResourceFactory

//-------------- Resource Instances

type ResourceInstances struct {
	context   interface{}
	logger    Logger
	urlGetter URLGetter
	storage   Storage
}

func NewResources(context interface{}) *ResourceInstances {
	ri := ResourceInstances{}
	ri.context = context
	return &ri
}

func (r *ResourceInstances) Logger() Logger {
	if r.logger == nil {
		if l, err := ResourceFactoryInstance.Logger(r.context); err == nil {
			r.logger = l
		} else {
			return nil
		}
	}
	return r.logger
}

func (r *ResourceInstances) Client() URLGetter {
	if r.urlGetter == nil {
		if l, err := ResourceFactoryInstance.Client(r.context); err == nil {
			r.urlGetter = l
		} else {
			return nil
		}
	}
	return r.urlGetter
}

func (r *ResourceInstances) Storage() Storage {
	if r.storage == nil {
		if l, err := ResourceFactoryInstance.Storage(r.context); err == nil {
			r.storage = l
		} else {
			return nil
		}
	}
	return r.storage
}
