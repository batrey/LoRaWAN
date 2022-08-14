package handlers

import (
	"LoRaWAN/db"
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/go-redis/redis"
)

type Reg struct {
	Id []string `json:"Deveuis"`
}
type Deveuis struct {
	Id         string `json:"Deveuis"`
	Registered bool   `json:"registered"`
}
type RespDeveuis struct {
	Ids []Deveuis `json:"Deveuis"`
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
		var dev Reg
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
		var response RespDeveuis
		for i := 0; i < *parallel; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, id := range dev.Id {
					status, _ := db.GetDeviceStatus(id)
					if !status {
						success, err := MakeRequestLorawan(id)
						if success == http.StatusOK {
							err = db.AddNewDevice(id, true)
							if err != nil {
								//TODO: handle error
							}
							response.Ids = append(response.Ids, Deveuis{Id: id, Registered: true})
						} else if err != nil {
							// TODO: error handling
							response.Ids = append(response.Ids, Deveuis{Id: id, Registered: false})
							err = db.AddNewDevice(id, false)
							if err != nil {
								//TODO: handle error
							}
						}

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
		err = db.AddKey(key[0], dev.Id)
		if err != nil {
			//TODO add error handling
		}

		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("Error marshalling json Err:%s", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResp)
	}
}

func GetFromRedis(key string, r *redis.Client) (string, error) {
	val, err := r.Get(key).Result()
	if err != nil {
		return "", err
	}
	return val, err
}

func SetRedis(key string, val Reg, r *redis.Client) error {
	_, err := r.Set(key, val, 0).Result()
	if err != nil {
		return err
	}
	return nil
}

func MakeRequestLorawan(Deveuis string) (int, error) {
	url := os.Getenv("LORAWAN_URL")

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(Deveuis)))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return resp.StatusCode, err
	}

	defer resp.Body.Close()
	return resp.StatusCode, err
}
