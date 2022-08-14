package handlers

import (
	"LoRaWAN/app/db"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/go-redis/redis"
)

type Deveuis struct {
	Deveuis []string `json:"deveuis”`
}
type DeveuisSingle struct {
	Id         string `json:"Deveuis"`
	Registered bool   `json:"registered"`
}
type RespDeveuis struct {
	Ids []DeveuisSingle `json:"Deveuis"`
}

func NewDevice(db db.DataBase, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//check if there is a key
		key, ok := r.URL.Query()["id"]
		if !ok || len(key[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//check if id is redis key  if not create one
		val, err := GetFromRedis(key[0], client)
		if val == key[0] || err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Add("Conflict", "Devices already registered")
			w.WriteHeader(http.StatusConflict)
			return
		}

		fmt.Println("\n checked Redis \n ")
		//get body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		decode := json.NewDecoder(r.Body)
		decode.DisallowUnknownFields()
		var dev Deveuis
		err = decode.Decode(&dev)
		if err != nil {
			log.Fatalf("Unable to decode json err:%s", err)
		}

		//check dev is smaller that 100
		if len(dev.Deveuis) > 100 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ch := make(chan RespDeveuis, 10)
		var wg sync.WaitGroup
		var response RespDeveuis
		var tmp RespDeveuis
		for _, id := range dev.Deveuis {
			wg.Add(1)
			go worker(id, ch, db, &wg)
			wg.Wait()
			tmp = <-ch
			response.Ids = append(response.Ids, tmp.Ids...)
		}
		close(ch)

		//TODO stor in redis
		err = SetRedis(key[0], dev, client)
		if err != nil {
			log.Fatalf("Unable to save key to redis err: %s", err)
		}
		err = db.AddKey(key[0], dev)
		if err != nil {
			log.Fatalf("Unable to add  key to Redis err:%s", err)
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

func TestDevice(db db.DataBase, client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ids []string
		for i := 0; i < 100; i++ {
			ids = append(ids, RandStringBytesMaskImprSrcUnsafe(10))
		}

		if len(ids) > 100 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ch := make(chan RespDeveuis, 10)
		var wg sync.WaitGroup
		var response RespDeveuis
		var tmp RespDeveuis
		for _, id := range ids {
			wg.Add(1)
			go worker(id, ch, db, &wg)
			wg.Wait()
			tmp = <-ch
			response.Ids = append(response.Ids, tmp.Ids...)
		}
		close(ch)
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
	if err != nil || val == "" {
		return "", err
	}
	return key, err
}

func SetRedis(key string, val interface{}, r *redis.Client) error {
	val, _ = json.Marshal(val)
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

func worker(id string, ch chan RespDeveuis, db db.DataBase, wg *sync.WaitGroup) {
	var response RespDeveuis
	status, _ := db.GetDeviceStatus(id)
	if !status {
		success, _ := MakeRequestLorawan(id)
		if success == http.StatusOK {
			err := db.AddNewDevice(id, true)
			if err != nil {
				log.Fatalf("Unable to add Devuce to the Database err:%s", err)
			}
			response.Ids = append(response.Ids, DeveuisSingle{Id: id, Registered: true})
		} else if success != http.StatusOK {
			response.Ids = append(response.Ids, DeveuisSingle{Id: id, Registered: false})
			err := db.AddNewDevice(id, false)
			if err != nil {
				log.Fatalf("Unable to add Devuce to the Database err:%s", err)
			}
		}

	} else {
		success, _ := MakeRequestLorawan(id)
		if success == http.StatusOK {
			err := db.UpdateDevicesStatus(id, true)
			if err != nil {
				log.Fatalf("Unable to add Devuce to the Database err:%s", err)
			}
			response.Ids = append(response.Ids, DeveuisSingle{Id: id, Registered: true})
		}
	}
	ch <- response
	wg.Done()
}

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	var src = rand.NewSource(time.Now().UnixNano())
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
