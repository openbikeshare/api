eackage main

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
	SystemId  string  `json:"systemId"`
}

var (
	db          *sql.DB
	pool        *redis.Pool
	redisServer = flag.String("redisServer", ":6379", "")
)

func setup() error {
	var err error
	connStr := "user=openbikeshare dbname=openbikeshare_data password=N!7bV29w@jKZMt!X sslmode=disable"
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
	values := r.URL.Query()

	swLat := values.Get("sw_lat")
	swLng := values.Get("sw_lng")
	neLat := values.Get("ne_lat")
	neLng := values.Get("ne_lng")

	if (swLat == "" || swLng == "" || neLat == "" || neLng == "") {
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
