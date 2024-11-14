package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/web-recommendation-service/config"
	"github.com/web-recommendation-service/meander"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatal("Unable to load env: ", err)
	}
	http.HandleFunc("/journeys", cors(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": meander.Journeys,
		})
	}))

	http.HandleFunc("/recommedations", cors(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		loc := strings.Split(query.Get("loc"), ",")
		if len(loc) != 2 {
			http.Error(w, "Pass valid query paramters", http.StatusBadRequest)
			return
		}
		lng, _ := strconv.ParseFloat(loc[0], 64)
		lat, _ := strconv.ParseFloat(loc[1], 64)
		radius, _ := strconv.Atoi(query.Get("radius"))
		limit, _ := strconv.Atoi(query.Get("limit"))
		q := &meander.Query{
			Lat:     lat,
			Lng:     lng,
			Journey: query.Get("journey"),
			Radius:  radius,
			Limit:   limit,
		}
		data := q.Run()
		respond(w, data)
	}))
	log.Println("Serving API on :8000")
	http.ListenAndServe(":8000", http.DefaultServeMux)
}

func cors(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		f(w, r)
	}
}

func respond(w http.ResponseWriter, data []interface{}) error {
	publicData := make([]interface{}, len(data))
	for i, d := range data {
		publicData[i] = meander.Public(d)
	}
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"data": publicData,
	})
}
