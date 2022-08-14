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
	Id         string `json:"Deveuis"`
	Registered bool   `json:"registered"`
}

type User struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

func ConnectDb() (DataBase, error) {
	db := DataBase{}
	var err error
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"))
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

func (db DataBase) AddNewDevice(DevEUII string, status bool) error {
	_, err := db.Conn.Exec("INSERT INTO registered (deveui,status) VALUES ($1,$2)", DevEUII, status)
	if err != nil {
		return err
	}
	return nil
}

func (db DataBase) AddKey(key string) error {
	_, err := db.Conn.Exec("INSERT INTO idempotency (idempotencykey) VALUES ($1)", key)
	if err != nil {
		return err
	}
	return nil
}

func (db DataBase) GetKey(key string) (bool, error) {

	err := db.Conn.QueryRow("SELECT idempotencykey FROM idempotency WHERE idempotencykey=$1", key).Scan()
	if err != sql.ErrNoRows {
		return true, err
	}
	return false, err

}

func (db DataBase) GetDeviceStatus(DevEUII string) (bool, error) {
	var status bool
	err := db.Conn.QueryRow("SELECT status FROM registered WHERE deveui=$1", DevEUII).Scan(&status)
	if err != nil {
		return status, err
	}
	fmt.Println("STATUS HERE", status)
	return status, err
}

func (db DataBase) UpdateDevicesStatus(DevEUI string, status bool) error {
	_, err := db.Conn.Exec("UPDATE registered SET status=$1 WHERE deveui=$2", status, DevEUI)
	if err != nil {
		return err
	}
	return nil
}
