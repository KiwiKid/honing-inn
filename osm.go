package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/serjvanilla/go-overpass"
)

// osmClient is a wrapper around the Overpass API client
type osmClient struct {
	client overpass.Client
}

type GeocodeResult struct {
	PlaceID     int      `json:"place_id"`
	Licence     string   `json:"licence"`
	OSMType     string   `json:"osm_type"`
	OSMID       int64    `json:"osm_id"`
	Lat         string   `json:"lat"`
	Lon         string   `json:"lon"`
	Category    string   `json:"category"`
	Type        string   `json:"type"`
	PlaceRank   int      `json:"place_rank"`
	Importance  float64  `json:"importance"`
	AddressType string   `json:"addresstype"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	BoundingBox []string `json:"boundingbox"`
}

// NewOSMClient creates a new instance of osmClient
func NewOSMClient() *osmClient {
	return &osmClient{
		client: overpass.New(),
	}
}

// FetchDrinkingWaterData queries the Overpass API for drinking water points
func (o *osmClient) FetchDrinkingWaterData() (*overpass.Result, error) {
	query := `[out:json];node["amenity"="drinking_water"](50.6,7.0,50.8,7.3);out body;`
	result, err := o.client.Query(query)
	if err != nil {
		log.Println("Error querying Overpass API:", err)
		return nil, err
	}
	return &result, nil
}

// GeocodeAddress performs an address lookup using the Nominatim API and returns latitude and longitude
func (o *osmClient) GeocodeAddress(address string) ([]GeocodeResult, error) {
	// Construct the Nominatim query URL
	endpoint := "https://nominatim.openstreetmap.org/search"
	query := url.Values{}
	query.Set("q", address)
	query.Set("format", "json")
	query.Set("limit", "10")

	// Make the request to Nominatim
	resp, err := http.Get(fmt.Sprintf("%s?%s", endpoint, query.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error querying Nominatim: %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var results []GeocodeResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("error decoding Nominatim response: %w", err)
	}

	// Check if we got any results
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", address)
	} else {
		for _, result := range results {
			log.Printf("Geocoded address: %+v", result)
		}
	}

	log.Printf("Geocoded %d results for address: %s", len(results), address)

	return results, nil
}
