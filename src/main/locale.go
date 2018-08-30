package main

import (
	"html/template"
	"log"
	"net"
	"net/http"

	maxminddb "github.com/oschwald/maxminddb-golang"
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

// Lang contains the languages
var Lang *i18n.I18n

// GeoIP is the database of IPs
var GeoIP *maxminddb.Reader

// GuestLocaleCache is indexed by their address and the value of their locale. This is to stop looking up everytime.
var GuestLocaleCache map[string]string

var record struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func initLocale() {

	// Load translations
	Lang = i18n.New(yaml.New("locale"))

	// Open ip database
	db, err := maxminddb.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Println("[!!] Error opening geo ip database:", err)
	}
	GeoIP = db

	// declare map
	GuestLocaleCache = make(map[string]string)
}

// T translates a string
func T(locale string, key string, args ...interface{}) template.HTML {
	return Lang.Fallbacks("US").T(locale, key, args...)
}

// GetLocale gets the locale of a request
func GetLocale(r *http.Request) string {
	ip := net.ParseIP(GetIP(r))

	// local
	if ip == nil {
		return "US"
	}

	cachedLocale := GuestLocaleCache[ip.String()]
	if cachedLocale != "" {
		return cachedLocale
	}

	err := GeoIP.Lookup(ip, &record)
	if err != nil {
		log.Println("[!!] Error looking up IP:", ip, err)
		GuestLocaleCache[ip.String()] = "US"
		return "US"
	}

	GuestLocaleCache[ip.String()] = record.Country.ISOCode
	return record.Country.ISOCode
}
