package infoclimat

import (
	"fmt"
	"meteoArchive/meteoAPI"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// regular expression to apply
var filterRegExp = func() *regexp.Regexp {
	r, _ := regexp.Compile("^([\\+]*[\\-]*[0-9]+([,\\.]{1}[0-9]+){0,1})")
	return r
}()

func purgeCellToGetMeasure(cellText string) *float64 {
	if matches := filterRegExp.FindStringSubmatch(cellText); matches != nil {
		s := strings.Replace(matches[1], ",", ".", 1)
		if f, e := strconv.ParseFloat(s, 64); e == nil {
			return &f
		}

	}
	return nil
}

const (
	skipStr = "SKIP"
)

var mapRowTitleToFieldName = map[string]string{
	"Tempé. maxiextrême":     "ExtremeMax",
	"Tempé. maximoyennes":    "AverageMax",
	"Tempé. moymoyennes":     "Average",
	"Tempé. minimoyennes":    "AverageMin",
	"Tempé. miniextrême":     "ExtremeMin",
	"Ensoleillement(heures)": "SunHours",
	"CumulPrécips":           "WhaterMilimeter",
	"":                       skipStr,
	"Tempé. maximinimale":       skipStr,
	"Tempé. minimaximale":       skipStr,
	"DJU(chauffagiste)":         skipStr,
	"DJU(climaticien)":          skipStr,
	"Max en 24hde précips":      skipStr,
	"Max en 5jde précips":       skipStr,
	"Moyenne ≥ 1de précips [?]": skipStr,
	"Neige au solmaximale":      skipStr,
	"Rafalemaximale":            skipStr,
	"Pressionminimale":          skipStr,
	"Pressionmaximale":          skipStr,
}

//RetrieveMonthlyReports go to infoclimat website and get the monthly report
func RetrieveMonthlyReports(station *meteoAPI.Station, year int) *meteoAPI.MonthlyMeasureSerie {
	// infoclimat.fr/climatologie/anne/2014/{city}/valeurs/{id}.html
	fmt.Println(*station)

	if station.Origin != OriginStr {
		return nil
	}

	url := "http://www.infoclimat.fr/climatologie/annee/" + strconv.Itoa(year) + "/" + station.RemoteMetadata["path"].(string) + "/valeurs/" + station.RemoteID + ".html"

	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println(url)

	serie := make(meteoAPI.MonthlyMeasureSerie)
	doc.Find("#tableau-releves").Each(func(i int, s *goquery.Selection) { //Tableau

		rows := s.Find("tr")
		rows.Each(func(r int, row *goquery.Selection) {
			//			fmt.Println("tr:", row.Text())
			cells := row.Find("td")

			if fieldName, found := mapRowTitleToFieldName[cells.First().Text()]; found {
				if fieldName != skipStr {
					for i := range cells.Nodes {
						if i == 0 || i > 12 { // Skip row title (0) and year average (13)
							continue
						}

						s := cells.Eq(i)

						if f := purgeCellToGetMeasure((*s).Text()); f != nil {
							m := new(meteoAPI.Measure)
							reflect.ValueOf(m).Elem().FieldByName(fieldName).Set(reflect.ValueOf(f))
							serie.PutMeasure(m, year, time.Month(i))
						} else {
							if len((*s).Text()) > 0 {
								fmt.Print("can decode value (skip):", (*s).Text())
							}
						}

					}
				}
			} else {
				fmt.Println("Unknow measure(skip):", cells.First().Text())
			}
		})
	})
	return &serie

}
