package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	db "LoRaWAN/db"
	device "LoRaWAN/handlers"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

func middleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//check content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		//TODO: check authorization token and user
		next.ServeHTTP(w, r)

	})
}

func main() {
	//Load env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	//Connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	//Connect to postgres

	database, err := db.ConnectDb()
	if err != nil {
		log.Fatalf("Could not set up database: %v", err)
	}
	//close db connection
	defer database.Conn.Close()

	fmt.Println("PostgreSQL and Redis connected successfully...")

	mux := http.NewServeMux()
	deviceHandle := http.HandlerFunc(device.NewDevice(database, client))
	mux.Handle("/device", middleWare(deviceHandle))
	log.Println("Server started on port 8080")
	err = http.ListenAndServe(":8080", mux)
	log.Fatal(err)

}
