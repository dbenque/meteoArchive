// Launch Local environment for Dev of meteoArchive module

docker run --rm -p 127.0.0.1:8080:8080 -p 127.0.0.1:8000:8000 -p 127.0.0.1:9000:9000 -v "$GOPATH/src/github.com/dbenque/meteoArchive/moduleMeteoArchive:/home/project/moduleMeteoArchive" -v "$GOPATH:/localgopath:ro" -e "LOCALGOPATH=code.google.com github.com golang.org" dbenque/goappengine

// Access Local datastore:

http://localhost:8000/datastore
