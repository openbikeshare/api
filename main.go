package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

type cycleLocation struct {
	Id        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Category  string  `json:"category"`
	Updated   string  `json:"updated"`
	SystemId  string  `json:"systemId"`
}

var (
	db *sql.DB
)

func setup() error {
	var err error
	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		log.Fatal("DB_URL not set")
	}
	db, err = sql.Open("postgres", connStr)
	return err
}

func cycleLocations(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	swLat := values.Get("sw_lat")
	swLng := values.Get("sw_lng")
	neLat := values.Get("ne_lat")
	neLng := values.Get("ne_lng")

	if swLat == "" || swLng == "" || neLat == "" || neLng == "" {
		w.WriteHeader(400)
		return
	}
	rows, err := db.Query(`SELECT id, ST_Y(location::geometry), ST_X(location::geometry), type, system_id
	FROM cycle_location
	WHERE location && ST_MakeEnvelope($1, $2, $3, $4, 4326)`, swLng, swLat, neLng, neLat)
	if err != nil {
		log.Fatal(err)
	}

	cycleLocations := []cycleLocation{}
	for rows.Next() {
		var cycleLocation cycleLocation
		rows.Scan(&cycleLocation.Id, &cycleLocation.Latitude, &cycleLocation.Longitude, &cycleLocation.Category, &cycleLocation.SystemId)
		cycleLocations = append(cycleLocations, cycleLocation)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cycleLocations)
}

func cycleCoverage(w http.ResponseWriter, r *http.Request) {

//	n, err := conn.Do("GET", "coverage")
//	if err != nil {
//		log.Print(err)
//	}
        rows, err := db.Query(`
	SELECT attributes -> 'coverage' AS coverage 
        FROM cache 
        WHERE name = 'geocache'
        `)
        if err != nil {
		log.Fatal(err)
	}
        var raw_data []byte
        var data interface{}
        rows.Next()
        rows.Scan(&raw_data)
        json.Unmarshal(raw_data, &data)
        w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/cycles", cycleLocations).Methods("GET")
	r.HandleFunc("/coverage", cycleCoverage).Methods("GET")
	log.Fatal(http.ListenAndServe(":3000", r))
}
