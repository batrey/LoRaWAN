package handlers

import (
	"LoRaWAN/app/db"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
	"unsafe"

	"github.com/go-redis/redis"
)

type Deveuis struct {
	DevEuis []string `json:"deveuis”`
}

type DeveuisSingle struct {
	Id         string `json:"deveuis"`
	Registered bool   `json:"registered"`
}

type RespDeveuis struct {
	Ids []DeveuisSingle `json:"deveuis"`
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
		keyExist, _ := db.GetKey(key[0])
		val, _ := GetFromRedis(key[0], client)
		if val == key[0] || keyExist {
			w.Header().Add("Conflict", "Devices already registered")
			w.WriteHeader(http.StatusConflict)
			return
		}

		//get body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		decode := json.NewDecoder(r.Body)
		decode.DisallowUnknownFields()
		var dev Deveuis
		err := decode.Decode(&dev)
		if err != nil {
			fmt.Printf("Unable to decode json err:%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//check dev is smaller that 100
		if len(dev.DevEuis) > 100 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ch := make(chan RespDeveuis, 10)
		var wg sync.WaitGroup
		var response RespDeveuis
		var tmp RespDeveuis
		for _, id := range dev.DevEuis {
			wg.Add(1)
			go worker(id, ch, db, &wg)
			wg.Wait()
			tmp = <-ch
			response.Ids = append(response.Ids, tmp.Ids...)
		}
		close(ch)

		err = SetRedis(key[0], response, client)
		if err != nil {
			fmt.Printf("Unable to save key to redis err: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = db.AddKey(key[0])
		if err != nil {
			fmt.Printf("Unable to add  key to Redis err:%s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
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

		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(response); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
	}
}
func GetFromRedis(key string, r *redis.Client) (string, error) {
	val, err := r.Get(key).Result()
	if err != nil || val == "" {
		return "", err
	}
	return key, err
}

func SetRedis(key string, val RespDeveuis, r *redis.Client) error {
	_, err := r.Set(key, nil, 0).Result()
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
				fmt.Printf("Unable to add Deveuis to the Database err:%s", err)
			}
			response.Ids = append(response.Ids, DeveuisSingle{Id: id, Registered: true})
		} else {
			response.Ids = append(response.Ids, DeveuisSingle{Id: id, Registered: false})
			err := db.AddNewDevice(id, false)
			if err != nil {
				fmt.Printf("Unable to add Deveuis to the Database err:%s", err)
			}
		}

	} else {
		success, _ := MakeRequestLorawan(id)
		if success == http.StatusOK {
			err := db.UpdateDevicesStatus(id, true)
			if err != nil {
				fmt.Printf("Unable to add Devuce to the Database err:%s", err)
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
