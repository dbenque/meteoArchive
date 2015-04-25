package meteoServer

import (
	"errors"
	"net/http"
	"strconv"
	"time"
)

func readCountFromURL(r *http.Request) (count int, err error) {
	count = 3
	if counturl, ok := r.URL.Query()["count"]; ok {
		if count, err = strconv.Atoi(counturl[0]); err != nil {
			err = errors.New("count parameter is not an int")
		}
	}
	return
}
func readCityCountryFromURL(r *http.Request) (city, country string, err error) {

	country = "fr"

	city = ""

	if cityurl, ok := r.URL.Query()["city"]; ok {
		city = cityurl[0]
	} else {
		err = errors.New("Could not read city from GET Query.")
	}

	if countryurl, ok := r.URL.Query()["country"]; ok {
		country = countryurl[0]
	}

	return
}

func readLatitudeLongitudeFromURL(r *http.Request) (lat, lon float64, err error) {

	if laturl, ok := r.URL.Query()["lat"]; ok {
		lat, err = strconv.ParseFloat(laturl[0], 64)
		if err != nil {
			return
		}

		if lonurl, ok := r.URL.Query()["lon"]; ok {
			lon, err = strconv.ParseFloat(lonurl[0], 64)
			if err != nil {
				return
			}
		} else {
			err = errors.New("Missing Longitude")
			return
		}

		return
	}

	err = errors.New("missing Latitude")
	return

}

func readYearFromURL(r *http.Request) (year int, err error) {
	year = 0
	if yearStr, ok := r.URL.Query()["year"]; ok {
		year, err = strconv.Atoi(yearStr[0])
		if err == nil {
			return
		}
		if year > time.Now().Year() || year < 1900 {
			err = errors.New("Year must be between now and 1900 (included)")
			return
		}
	}

	err = errors.New("Year not (or not well) specified")
	return
}
