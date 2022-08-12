package main

import (
	"fmt"
	"log"
	"os"

	db "LoRaWAN/db"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

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

}
