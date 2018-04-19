package mm

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	geoip2 "github.com/oschwald/geoip2-golang"
)

//go:generate ffjson --nodecoder $GOFILE

type City struct {
	Name string `json:"name,omitempty"`
}

type Continent struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}

type Country struct {
	IsInEuropeanUnion bool   `json:"is_in_european_union,omitempty"`
	IsoCode           string `json:"iso_code,omitempty"`
	Name              string `json:"name,omitempty"`
}

type Location struct {
	AccuracyRadius uint16  `json:"accuracy_radius,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	MetroCode      uint    `json:"metro_code,omitempty"`
	TimeZone       string  `json:"time_zone,omitempty"`
}

type Postal struct {
	Code string `json:"code,omitempty"`
}

type RepresentedCountry struct {
	IsInEuropeanUnion bool   `json:"is_in_european_union,omitempty"`
	IsoCode           string `json:"iso_code,omitempty"`
	Name              string `json:"name,omitempty"`
	Type              string `json:"type,omitempty"`
}

type Subdivision struct {
	IsoCode string `json:"iso_code,omitempty"`
	Name    string `json:"name,omitempty"`
}

type Traits struct {
	IsAnonymousProxy    bool `json:"is_anonymous_proxy,omitempty"`
	IsSatelliteProvider bool `json:"is_satellite_provider,omitempty"`
}

type GeoData struct {
	City               `json:"city,omitempty"`
	Continent          `json:"continent,omitempty"`
	Country            `json:"country,omitempty"`
	Location           `json:"location,omitempty"`
	Postal             `json:"postal,omitempty"`
	RegisteredCountry  Country `json:"registered_country,omitempty"`
	RepresentedCountry `json:"represented_country,omitempty"`
	Subdivisions       []Subdivision `json:"subdivisions,omitempty"`
	Subdivision        Subdivision   `json:"subdivision,omitempty"`
	Traits             `json:"traits,omitempty"`
}

type JsonResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Data    *GeoData `json:"data"`
}

func NewFromGeoIp2City(record *geoip2.City) *GeoData {
	var subdivisions []Subdivision
	for _, sub := range record.Subdivisions {
		subdivisions = append(subdivisions, Subdivision{
			IsoCode: sub.IsoCode,
			Name:    sub.Names["en"],
		})
	}

	var data = &GeoData{
		City: City{
			Name: record.City.Names["en"],
		},

		Continent: Continent{
			Code: record.Continent.Code,
			Name: record.Continent.Names["en"],
		},

		Country: Country{
			IsInEuropeanUnion: record.Country.IsInEuropeanUnion,
			IsoCode:           record.Country.IsoCode,
			Name:              record.Country.Names["en"],
		},

		Location: Location{
			AccuracyRadius: record.Location.AccuracyRadius,
			Latitude:       record.Location.Latitude,
			Longitude:      record.Location.Longitude,
			MetroCode:      record.Location.MetroCode,
			TimeZone:       record.Location.TimeZone,
		},

		Postal: Postal{
			Code: record.Postal.Code,
		},

		RegisteredCountry: Country{
			IsInEuropeanUnion: record.RegisteredCountry.IsInEuropeanUnion,
			IsoCode:           record.RegisteredCountry.IsoCode,
			Name:              record.RegisteredCountry.Names["en"],
		},

		RepresentedCountry: RepresentedCountry{
			IsInEuropeanUnion: record.RepresentedCountry.IsInEuropeanUnion,
			IsoCode:           record.RepresentedCountry.IsoCode,
			Name:              record.RepresentedCountry.Names["en"],
			Type:              record.RepresentedCountry.Type,
		},

		Subdivisions: subdivisions,

		Traits: Traits{
			IsAnonymousProxy:    record.Traits.IsAnonymousProxy,
			IsSatelliteProvider: record.Traits.IsSatelliteProvider,
		},
	}

	if len(subdivisions) > 0 {
		data.Subdivision = subdivisions[len(subdivisions)-1]
	}

	return data
}

func (gd *GeoData) Unknown() bool {
	if gd.City.Name == "" && gd.Continent.Name == "" &&
		gd.Country.Name == "" && gd.Location.AccuracyRadius == 0 &&
		gd.Postal.Code == "" && len(gd.Subdivisions) == 0 {
		return true
	}
	return false
}

// ffjson: skip
type Database struct {
	Reader *geoip2.Reader
}

var dbInstances map[string]*Database = map[string]*Database{}

var once sync.Once

func GetDatabase(path string) (*Database, error) {
	once.Do(func() {
		database, err := geoip2.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		dbInstances[path] = &Database{Reader: database}
	})
	return dbInstances[path], nil
}

func CloseDatabases() {
	// Not really thread safe, but doesn't really matter in this app
	for _, db := range dbInstances {
		db.Close()
	}
}

func (db *Database) Close() {
	db.Reader.Close()
}

func (db *Database) Lookup(ipText string) (*GeoData, error) {
	ip := net.ParseIP(ipText)
	if ip == nil {
		return nil, errors.New(fmt.Sprintf("unable to decode ip `%s`", ipText))
	}

	record, err := db.Reader.City(ip)
	if err != nil {
		return nil, err
	}

	return NewFromGeoIp2City(record), nil
}
