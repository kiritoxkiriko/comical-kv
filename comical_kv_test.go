package comical_kv

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	f := GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("aaa")
	if val, _ := f.Get("aaa"); !reflect.DeepEqual(val, expect) {
		t.Fatalf("call back failed")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	ckv, err := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range db {
		if view, err := ckv.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %s\n", k)
		} // load from callback function
		if _, err := ckv.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // cache hit
	}

	if view, err := ckv.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
