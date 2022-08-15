package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DataBase struct {
	Conn *sql.DB
}

type Deveuis struct {
	Id         string `json:"deveuis"`
	Registered bool   `json:"registered"`
}

type User struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

// Makes a connection to the DB
func ConnectDb() (DataBase, error) {
	db := DataBase{}
	var err error
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"))
	db.Conn, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatalf("could not connect to postgres database: %v", err)
		return db, err
	}
	err = db.Conn.Ping()
	if err != nil {
		return db, err
	}
	fmt.Println(url)
	return db, nil
}

// Add New Device to the DB
func (db DataBase) AddNewDevice(DevEUII string, status bool) error {
	_, err := db.Conn.Exec("INSERT INTO registered (dev_eui,status) VALUES ($1,$2)", DevEUII, status)
	if err != nil {
		return err
	}
	return nil
}

// Adds idempotency key  to DB
func (db DataBase) AddKey(key string) error {
	_, err := db.Conn.Exec("INSERT INTO idempotency (key) VALUES ($1)", key)
	if err != nil {
		return err
	}
	return nil
}

// Gets idempotency key  stored in DB
func (db DataBase) GetKey(key string) (bool, error) {
	var status string
	err := db.Conn.QueryRow("SELECT key FROM idempotency WHERE key=$1", key).Scan(&status)
	if err != sql.ErrNoRows {
		return true, err
	}
	return false, err

}

// Gets Status fo Device
func (db DataBase) GetDeviceStatus(DevEUII string) (bool, error) {
	var status bool
	err := db.Conn.QueryRow("SELECT status FROM registered WHERE dev_eui=$1", DevEUII).Scan(&status)
	if err != nil {
		return status, err
	}
	return status, err
}

// Updates the Status of a Device
func (db DataBase) UpdateDevicesStatus(DevEUI string, status bool) error {
	_, err := db.Conn.Exec("UPDATE registered SET status=$1 WHERE dev_eui=$2", status, DevEUI)
	if err != nil {
		return err
	}
	return nil
}
