package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/kiritoxkiriko/comical"

	"github.com/kiritoxkiriko/comical-kv"
)

var dbData = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

var db *comical_kv.Db

func initDB() {
	var err error
	db, err = comical_kv.NewDb(dbData, comical_kv.DefaultRdmLatencyMin, comical_kv.DefaultRdmLatencyMax)
	if err != nil {
		log.Fatalf("[DB] create db failed: %v", err)
	}
}

func createGroup() (*comical_kv.Group, error) {
	return comical_kv.NewGroup("scores", 2<<10, comical_kv.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db.Get(key); ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startKVServer(addr string, peerAddrs []string, ckv *comical_kv.Group) {
	pool := comical_kv.NewHTTPPool(fmt.Sprintf("http://%s", addr))
	pool.Set(peerAddrs...)
	ckv.RegisterPeers(pool)
	log.Println("comical kv is running at", addr)
	log.Fatal(http.ListenAndServe(addr, pool))
}

func startAPIServer(apiAddr string, ckv *comical_kv.Group) {
	engine := comical.New()
	engine.GET("/api/:key", func(c *comical.Context) {
		key := c.Param("key")
		if key == "" {
			c.Fail(http.StatusBadRequest, "key is needed")
			return
		}
		view, err := ckv.Get(key)
		if err != nil {
			c.Fail(http.StatusInternalServerError, err.Error())
			return
		}
		// return the view
		c.SetHeader("Content-Type", "application/octet-stream")
		c.Data(200, view.ByteSlice())
	})
	log.Println("comical kv api is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr, engine))
}

// for local test
func _main() {
	_, err := comical_kv.NewGroup("scores", 2<<10, comical_kv.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db.Get(key); ok {
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

func main() {
	initDB()
	// parse flags
	var port, apiPort int
	var api bool
	var peerAddr []string
	var peerAddrStr string
	flag.IntVar(&port, "port", 8999, "comical kv server port")
	flag.BoolVar(&api, "api", false, "start a comical kv api server")
	flag.IntVar(&apiPort, "apiPort", 9999, "comical kv api server port")
	flag.StringVar(&peerAddrStr, "peers", "", "comical kv peers, separated by comma")
	flag.Parse()

	peerAddr = strings.Split(strings.TrimSpace(peerAddrStr), ",")
	if len(peerAddr) == 0 {
		log.Println("no peers, use local")
	}

	// create a group
	ckv, err := createGroup()
	if err != nil {
		log.Fatal(err)
	}
	// start a kv server
	if api {
		go startAPIServer(fmt.Sprintf("localhost:%d", apiPort), ckv)
	}
	// start a kv server
	startKVServer(fmt.Sprintf("localhost:%d", port), peerAddr, ckv)
}
