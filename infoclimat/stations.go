package infoclimat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"meteo/geoloc"
	api "meteo/meteoAPI"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var mapsOfPOI = make(map[string](api.Station))
var exceptionsStations = map[string]string{
	"Île d'Amsterdam - Martin de Viviès (999)":  "",
	"Île de la Possession - Alfred Faure (999)": "",
	"Ile de Ré - Pointe des baleines (17)":      "Ile de Re",
	"Îles Kerguelen - Port-Aux-Francais (999)":  "Ile Kerguelen",
	"LANNAERO (22)":                             "",
	"Maopoopo Ile Futuna  (FR)":                 "futuna",
	"Moy-de-l'Aisne (02)":                       "Moy-de-Aisne",
	"Rikitea (987)":                             "Rikitea",
	"Serge-Frolow Ile Tromelin (974)":           "Ile Tromelin",
	"Sisco - Cap Sagro (2B)":                    "Sisco",
	"Terre Adélie - Dumont d'Urville (999)":     "", // Arctantique !
	"VATRY AERO (FR)":                           "Chalons Vatry",
}

func readPOIMap() error {
	filecontent, err := ioutil.ReadFile("POI.json")
	if err != nil {
		return err
	}

	return json.Unmarshal(filecontent, &mapsOfPOI)

}

func writePOIMap() error {

	dataj, _ := json.Marshal(mapsOfPOI)
	return ioutil.WriteFile("POI.json", dataj, 0644)

}

func getExceptions() *map[string]string {
	return &exceptionsStations
}

func getFinalName(initialName string) string {

	if v, found := (*getExceptions())[initialName]; found {
		return v
	}

	return initialName
}

func getPOI(cell string) (station api.Station, cache bool, err error) {

	if v, found := (*getExceptions())[cell]; found {
		cell = v
		if v == "" {
			return station, true, errors.New("Known Exception POI")
		}
	}

	cache = false
	if s, ok := mapsOfPOI[cell]; ok {
		return s, true, nil
	}

	reformat := strings.Replace(strings.Replace(strings.Replace(strings.Replace(cell, " ", "+", -1), "(", "+", -1), ")", "+", -1), "''", "%27", -1)

	poi, err := geoloc.GeolocFromCity(reformat, "fr", "fr")
	station.POI = poi
	station.Origin = "Infoclimat"
	station.RemoteID = getFinalName(cell)

	if err == nil {
		mapsOfPOI[cell] = station
	}

	return
}

func getCountry() map[string]string {
	doc, err := goquery.NewDocument("http://www.infoclimat.fr/observations-meteo/temps-reel/bac-can/48810.html")
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

func getCities(countryCode string) int {

	url := "http://www.infoclimat.fr/stations-meteo/cache/select/_" + countryCode + ".html"
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	count := 0

	doc.Find("option").Each(func(i int, option *goquery.Selection) {

		stationID, found1 := option.Attr("value")
		stationPath, found2 := option.Attr("data-seo")

		if found1 && found2 {
			getStation(stationID, stationPath)
			count++
		}
	})

	return count
	// infoclimat.fr/climatologie/anne/2014/{city}/valeurs/{id}.html

}

func getStation(stationID string, stationPath string) {

	url := "http://www.infoclimat.fr/include/ajax/stations.php?q=" + stationID

	client := &http.Client{CheckRedirect: nil}
	req, err := http.NewRequest("GET", url, nil)
	response, err := client.Do(req)
	if err != nil {
		return
	}
	responseStr, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	response.Body.Close()

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
			station.Origin = "infoclimat"
			station.PutMetadata("path", stationPath)
			station.PutMetadata("country", sta["pays"].(string))
			if sta["miny"].(float64) > 0 {
				station.PutMetadata("minYear", int(sta["miny"].(float64)))
			}
			if sta["maxy"].(float64) > 0 {
				station.PutMetadata("maxYear", int(sta["maxy"].(float64)))
			}

			k := "infoclimat." + stationID
			mapsOfPOI[k] = *station

		}
	}

}

func GetStations(cacheOnly bool) (stations api.Stations, err error) {

	if cacheOnly {
		readPOIMap()
	} else {

		for code, country := range getCountry() {

			fmt.Print("[" + code + "] ")
			count := getCities(code)
			fmt.Println("Country:" + country + ", adding " + strconv.Itoa(count))

		}
	}
	stations = make(api.Stations, len(mapsOfPOI), len(mapsOfPOI))
	i := 0
	for _, v := range mapsOfPOI {
		stations[i] = v
		i++
	}

	fmt.Println("Number of stations: " + strconv.Itoa(i))

	writePOIMap()

	return

}

//GetStationsOld browse the page retrieve POI and meteo Data
func GetStationsOld(cacheOnly bool) (stations api.Stations, err error) {

	readPOIMap()

	if !cacheOnly {
		doc, err := goquery.NewDocument("http://www.infoclimat.fr/stations-meteo/analyses-mensuelles.php")
		if err != nil {
			log.Fatal(err)
		}

		doc.Find("#tableau-releves").Each(func(i int, s *goquery.Selection) { //Tableau
			rows := s.Find("TR")
			rows.Each(func(r int, row *goquery.Selection) { //Ligne
				if r == 0 {
					return // skip header
				}
				cachedPOI := false
				row.Find("TD").Each(func(c int, cell *goquery.Selection) { //Colonne
					switch c {
					case 0:
						_, cachedPOI, err = getPOI(cell.Text())
						if err != nil {
							fmt.Println("Error for station " + cell.Text() + " while POI retrieve: " + err.Error())
						}
						if !cachedPOI {
							fmt.Println("An new station was found: ", cell.Text())
							time.Sleep(time.Millisecond * 300) // Throttle a bit in order tostay in google policy with Geoloc API
						}
					}
				})
			})
		})
	}
	writePOIMap()

	stations = make(api.Stations, len(mapsOfPOI), len(mapsOfPOI))
	i := 0
	for _, v := range mapsOfPOI {
		stations[i] = v
		i++
	}
	return
}

//GetStations2 browse the page retrieve POI and meteo Data
func GetStations2() (stations api.Stations, err error) {

	readPOIMap()

	//?mois=1&annee=2014
	doc, err := goquery.NewDocument("http://www.infoclimat.fr/stations-meteo/analyses-mensuelles.php")
	if err != nil {
		log.Fatal(err)
	}

	var reportMatrix []api.MonthlyReport

	doc.Find("#tableau-releves").Each(func(i int, s *goquery.Selection) { //Tableau
		rows := s.Find("TR")
		reportMatrix = make([]api.MonthlyReport, rows.Length(), rows.Length())
		rows.Each(func(r int, row *goquery.Selection) { //Ligne
			if r == 0 {
				return // skip header
			}
			cachedPOI := false
			row.Find("TD").Each(func(c int, cell *goquery.Selection) { //Colonne
				switch c {
				case 0:
					//reportMatrix[r].Name = cell.Text()
					reportMatrix[r].Station, cachedPOI, err = getPOI(cell.Text())
					if err != nil {
						fmt.Println("Error for station " + cell.Text() + " while POI retrieve: " + err.Error())
					}
				case 1:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].ExtremeMin = float32(v)
				case 2:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].AverageMin = float32(v)
				case 3:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].Average = float32(v)
				case 4:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].AverageMax = float32(v)
				case 5:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].ExtremeMax = float32(v)
				case 6:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].WhaterMilimeter = float32(v)
				case 7:
					v, _ := strconv.ParseFloat(cell.Text(), 32)
					reportMatrix[r].SunHours = float32(v)

				}
			})
			if !cachedPOI {
				writePOIMap()
				time.Sleep(time.Millisecond * 300)
			}
		})
	})

	return
}
