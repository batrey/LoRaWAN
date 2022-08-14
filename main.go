package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	deviceHandle := http.HandlerFunc(device.NewDevice(database, client))
	testHandle := http.HandlerFunc(device.TestDevice(database, client))
	mux.Handle("/device", middleWare(deviceHandle))
	mux.Handle("/test", middleWare(testHandle))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Server started on port 8080")
	<-done
	log.Println("Server Stopped on port 8080")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")

}
