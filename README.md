meteoArchive
============

API to access meteo archive:


// Feed the DB with stations of one country
http://127.0.0.1:8080/infoclimat/updateStations/mapStorage?country=CL
// Refresh the kdtree
http://127.0.0.1:8080/kdtreeReload/mapStorage

// Geoloc API
http://127.0.0.1:8080/distance?d=200&city=strasbourg
http://127.0.0.1:8080/near?count=10&city=vinadelmar&country=CL
http://127.0.0.1:8080/near?lon=7.296262&lat=48.018517
http://127.0.0.1:8080/geoloc?d=200&city=strasbourg

// Meteo Data
http://127.0.0.1:8080/infoclimat/getMonthlySerie?city=84230&country=FR&year=2008
http://127.0.0.1:8080/infoclimat/getMonthlySerie?lon=7.296262&lat=48.018517&year=2008




TODO
====

- store the series with time stamp instead of string "year.month"
- do the storage on datastore
- do the appengine module
- create a job to feed db for one year or for one station
