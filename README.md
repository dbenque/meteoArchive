meteoArchive
============

Feed the DB with stations of one country for infoclimat website


http://127.0.0.1:8080/meteo/infoclimat/updateStations/mapStorage?country=FR

Pack all the stations so that we can retrieve them within a low number of queries. This serialize the slice of all stations and build chunks of 1Mo into the datastore.

http://127.0.0.1:8080/meteo/packStation

Refresh the kdtree

http://127.0.0.1:8080/meteo/kdtreeReload/mapStorage

Geoloc API

http://127.0.0.1:8080/meteo/distance?d=200&city=strasbourg
http://127.0.0.1:8080/meteo/near?count=10&city=vinadelmar&country=CL
http://127.0.0.1:8080/meteo/near?lon=7.296262&lat=48.018517
http://127.0.0.1:8080/meteo/geoloc?d=200&city=strasbourg

Monthly Meteo Data

http://127.0.0.1:8080/meteo/infoclimat/getMonthlySerie?city=84230&country=FR&year=2008
http://127.0.0.1:8080/meteo/infoclimat/getMonthlySerie?lon=7.296262&lat=48.018517&year=2008



TODO
====

- create a job to feed db for one year or for one station
- do the appengine module
