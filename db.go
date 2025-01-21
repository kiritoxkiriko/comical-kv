package comical_kv

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	DefaultRdmLatencyMin = 100 * time.Millisecond
	DefaultRdmLatencyMax = 500 * time.Millisecond
)

type Db struct {
	data          map[string]string
	lock          sync.RWMutex
	rdmLatencyMin time.Duration
	rdmLatencyMax time.Duration
}

func NewDb(data map[string]string, rdmLatMin, rdmLatMax time.Duration) (*Db, error) {
	if rdmLatMin < 0 || rdmLatMax < 0 {
		return nil, fmt.Errorf("latency must be a positive number")
	}
	if rdmLatMin > rdmLatMax {
		return nil, fmt.Errorf("min latency is greater than max latency")
	}
	return &Db{
		data:          data,
		lock:          sync.RWMutex{},
		rdmLatencyMin: rdmLatMin,
		rdmLatencyMax: rdmLatMax,
	}, nil
}

func (d *Db) InitData(data map[string]string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data = data
}

func (d *Db) Del(key string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.data, key)
}

func (d *Db) Get(key string) (string, bool) {
	d.lock.RLock()
	defer d.lock.RUnlock()
	// simulate random latency
	time.Sleep(time.Duration(d.rdmLatencyMin + time.Duration(rand.Intn(int(d.rdmLatencyMax-d.rdmLatencyMin)))))
	if v, ok := d.data[key]; ok {
		return v, true
	}
	return "", false
}

func (d *Db) Set(key, value string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data[key] = value
}
