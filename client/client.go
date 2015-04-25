package meteoClient

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type URLGetter interface {
	Get(url string) (*http.Response, error)
}

// Factory Function for client
type URLGetterFactory func(*http.Request) URLGetter

var ClientFactory URLGetterFactory

// Util to Get a goquery.Document
func GetGoqueryDocument(getter URLGetter, url string) (*goquery.Document, error) {

	res, errGet := getter.Get(url)
	if errGet != nil {
		log.Fatal(errGet)
		return nil, errGet
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Fatal(err)
		return nil, errGet
	}

	return doc, err
}
