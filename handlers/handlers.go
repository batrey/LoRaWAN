package handlers

import (
	"LoRaWAN/db"
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"sync"

	"github.com/go-redis/redis"
)

type Deveuis struct {
	Id []string `json:"Deveuis"`
}

func NewDevice(db db.DataBase, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check if there is a key
		key, ok := r.URL.Query()["id"]
		if !ok || len(key[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
		}

		//check if id is redis key  if not create one
		val, _ := GetFromRedis(key[0], client)
		if val != "" {
			resp, err := json.Marshal(val)
			if err != nil {
				//TODO: handle error
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(resp)
		}
		//get body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		decode := json.NewDecoder(r.Body)
		decode.DisallowUnknownFields()
		var dev Deveuis
		err := decode.Decode(&dev)
		if err != nil {
			//TODO handle error
		}
		//check dev is smaller that 100
		if len(dev.Id) > 100 {
			w.WriteHeader(http.StatusBadRequest)
		}

		//make request to LoRaWABN server
		parallel := flag.Int("parallel", 10, "max parallel requests allowed")
		flag.Parse()

		results := make(chan string)
		var wg sync.WaitGroup

		for i := 0; i < *parallel; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, id := range dev.Id {
					status, _ := db.GetDeviceStatus(id)
					if !status {
						MakeRequestLorawan(id)
						db.AddNewDevice(id)
						// TODO: error handling
					}
				}

			}()
		}
		go func() {
			wg.Wait()
			close(results)
		}()

		//TODO stor in redis
		err = SetRedis(key[0], dev, client)
		if err != nil {
			//TODO:Error handling
		}
	}
}

func GetFromRedis(key string, r *redis.Client) (string, error) {
	val, err := r.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, err
}

func SetRedis(key string, val Deveuis, r *redis.Client) error {
	_, err := r.Set(key, val, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func MakeRequestLorawan(Deveuis string) error {
	url := os.Getenv("LORAWAN_URL")

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(Deveuis)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
