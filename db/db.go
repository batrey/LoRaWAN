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
	fmt.Println(url)
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
	_, err := db.Conn.Exec("INSERT INTO registered (DevEUI,status) VALUES ($1,$2)", DevEUII, status)
	if err != nil {
		return err
	}
	return nil
}

func (db DataBase) AddKey(key string, data []string) error {
	_, err := db.Conn.Exec("INSERT INTO Idempotency (IdempotencyKey,data) VALUES ($1,$2)", key, data)
	if err != nil {
		return err
	}
	return nil
}

func (db DataBase) GetDeviceStatus(DevEUII string) (bool, error) {
	var status bool
	err := db.Conn.QueryRow("SELECT status FROM registered WHERE DevEUI=$1", DevEUII).Scan(&status)
	if err != nil {
		return false, err
	}
	return status, nil
}

func (db DataBase) UpdateDevicesStatus(DevEUI string, status bool) error {
	_, err := db.Conn.Exec("UPDATE registered SET status=$1,user =$2 WHERE DevEUI=$3", status, DevEUI)
	if err != nil {
		return err
	}
	return nil
}

// func (db DataBase) AddUser(user User) error {
// 	_, err := db.Conn.Exec("INSERT INTO users (username,password) VALUES ($1,crypt($2,gen_salt('bf',8)))", user.UserName, user.Password)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (db DataBase) LogIn(users User) (bool, error) {
// 	var user string
// 	err := db.Conn.QueryRow("SELECT user FROM users WHERE password=ccrypt($1,gen_salt('bf',8)) AND username = $2", users.Password, users.UserName).Scan(&user)
// 	if err != nil && err != sql.ErrNoRows {
// 		return false, err
// 	} else {
// 		return true, nil
// 	}
// }
