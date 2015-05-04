package infoclimat

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"

	api "github.com/dbenque/meteoArchive/meteoAPI"
	"github.com/dbenque/meteoArchive/resource"

	"github.com/PuerkitoBio/goquery"
)

const (
	OriginStr = "infoclimat"
)

//InfoClimatWebsite empty function receiver for MeteoWebsite interface
type InfoClimatWebsite struct {
}

func GetCountry(res *resource.ResourceInstances) map[string]string {

	doc, err := resource.GetGoqueryDocument(res.Client(), "http://www.infoclimat.fr/observations-meteo/temps-reel/bac-can/48810.html")
	if err != nil {
		log.Fatal(err)
	}

	result := make(map[string]string)

	doc.Find("#select_pays").Each(func(i int, s *goquery.Selection) { //Tableau
		rows := s.Find("option")
		rows.Each(func(r int, row *goquery.Selection) { //Ligne

			if countryCode, found := row.Attr("value"); found {
				result[countryCode] = row.Text()
			}
		})

	})
	return result
}

func getCities(res *resource.ResourceInstances, countryCode string, storage *api.Storage) int {

	url := "http://www.infoclimat.fr/stations-meteo/cache/select/_" + countryCode + ".html"
	doc, err := resource.GetGoqueryDocument(res.Client(), url)
	if err != nil {
		res.Logger().Errorf(err.Error())
	}

	count := 0

	doc.Find("option").Each(func(i int, option *goquery.Selection) {

		stationID, found1 := option.Attr("value")
		stationPath, found2 := option.Attr("data-seo")

		if found1 && found2 {
			getStation(res, stationID, stationPath, storage)
			count++
		}
	})

	return count

}

func getStation(res *resource.ResourceInstances, stationID string, stationPath string, storage *api.Storage) {

	url := "http://www.infoclimat.fr/include/ajax/stations.php?q=" + stationID

	response, err := res.Client().Get(url)
	defer response.Body.Close()

	responseStr, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	var dataj map[string]interface{}
	if err = json.Unmarshal(responseStr, &dataj); err != nil {
		return
	}

	results, ok := dataj["data"].([]interface{})
	if !ok {
		return
	}

	for _, s := range results {
		if sta, ok := s.(map[string]interface{}); ok {
			station := api.NewStation(sta["name"].(string), 0, sta["latitude"].(float64), sta["longitude"].(float64))
			station.RemoteID = stationID
			station.Origin = OriginStr
			station.PutMetadata("path", stationPath)
			station.PutMetadata("country", sta["pays"].(string))
			if sta["miny"].(float64) > 0 {
				station.PutMetadata("minYear", int(sta["miny"].(float64)))
			}
			if sta["maxy"].(float64) > 0 {
				station.PutMetadata("maxYear", int(sta["maxy"].(float64)))
			}

			(*storage).PutStation(station)

		}
	}

}

//UpdateStations update the given storage with the Infoclimat website's stations
func (website *InfoClimatWebsite) UpdateStations(res *resource.ResourceInstances, s api.Storage, code string) error {

	if len(code) == 0 {
		res.Logger().Warningf("No Country code provided to infoclimate::UpdateStations")
		return errors.New("No Country code provided to infoclimate::UpdateStations")
	}

	res.Logger().Infof("[%s] grab stations on going...", code)
	count := getCities(res, code, &s)
	res.Logger().Infof("[%s] , adding %d", code, count)

	return nil
}
