package main

import (
	"encoding/json"
	"net/http"
)

func geocodeAddressHandler(client *osmClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "address query parameter is required", http.StatusBadRequest)
			return
		}

		geoRes, err := client.GeocodeAddress(address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(geoRes)
	}
}
