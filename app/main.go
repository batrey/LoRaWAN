package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var (
	host     = os.Getenv("POSTGRES_HOST")
	port     = os.Getenv("POSTGRES_PORT")
	user     = os.Getenv("POSTGRES_USER")
	password = os.Getenv("POSTGRES_PASSWORD")
	dbname   = os.Getenv("POSTGRES_DB")
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
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("could not connect to postgres database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("could not connect to postgres database: %v", err)
	}

	fmt.Println("PostgreSQL and Redis connected successfully...")
	// Output: PONG <nil>

}
