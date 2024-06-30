package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kiritoxkiriko/comical-kv"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	_, err := comical_kv.NewGroup("scores", 2<<10, comical_kv.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	if err != nil {
		panic(err)
	}
	addr := ":9999"

	peers := comical_kv.NewHTTPPool(addr)
	log.Println("comical kv is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
