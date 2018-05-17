package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
)

type cycleLocation struct {
	Id        string  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Category  string  `json:"category"`
	Updated   string  `json:"updated"`
}

var (
	db          *sql.DB
	pool        *redis.Pool
	redisServer = flag.String("redisServer", ":6379", "")
)

func setup() error {
	var err error
	connStr := "user=openbikeshare dbname=openbikeshare_data sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	return err
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func cycleLocations(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, ST_Y(location::geometry), ST_X(location::geometry), type, last_time_updated FROM cycle_location")
	if err != nil {
		log.Fatal(err)
	}

	var cycleLocations []cycleLocation
	for rows.Next() {
		var cycleLocation cycleLocation
		rows.Scan(&cycleLocation.Id, &cycleLocation.Latitude, &cycleLocation.Longitude, &cycleLocation.Category, &cycleLocation.Updated)
		cycleLocations = append(cycleLocations, cycleLocation)
	}

	json.NewEncoder(w).Encode(cycleLocations)
}

func cycleCoverage(w http.ResponseWriter, r *http.Request) {
	conn := pool.Get()
	defer conn.Close()

	n, err := conn.Do("GET", "coverage")
	if err != nil {
		log.Print(err)
	}
        var data interface{}
        json.Unmarshal(n.([]byte), &data)

	json.NewEncoder(w).Encode(data)
}

func main() {
	err := setup()
	if err != nil {
		log.Fatal(err)
	}
	pool = newPool(*redisServer)

	r := mux.NewRouter()
	r.HandleFunc("/cycles", cycleLocations).Methods("GET")
	r.HandleFunc("/coverage", cycleCoverage).Methods("GET")
	log.Fatal(http.ListenAndServe(":3000", r))
}
