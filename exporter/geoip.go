package exporter

import (
	"log"
	"net"

	"github.com/oschwald/geoip2-golang"
)

var (
	dbPath   string
	database = &GeoIPDatabase{}
)

type GeoIPDatabase struct {
	db          *geoip2.Reader
	fileHandler string
}

func SetGeoIPPath(path string) {
	dbPath = path
}

func openIPDatabase() (*GeoIPDatabase, error) {
	if database.db != nil {
		return database, nil
	}

	var err error
	database.db, err = geoip2.Open(dbPath)
	database.fileHandler = "open"
	return database, err
}

func (db *GeoIPDatabase) Close() {
	if db.fileHandler == "open" {
		db.db.Close()
		db.fileHandler = "closed"
	}
}

func GetIpLocationDetails(ipAddress string) (city *geoip2.City, err error) {
	db, err := openIPDatabase()
	if err != nil {
		log.Fatal(err)
	}

	ip := net.ParseIP(ipAddress)
	city, err = db.db.City(ip)
	if err != nil {
		return city, err
	}

	return city, nil
}
