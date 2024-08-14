package comical_kv

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/kiritoxkiriko/comical"

	"github.com/kiritoxkiriko/comical-kv/consistent_hash"
)

const (
	// DefaultBasePath is the default base path of the HTTPPool
	DefaultBasePath = "/_comical-kv/"
	// DefaultReplica is the default replica count of the HTTPPool
	DefaultReplica = 50
)

// HTTPPool is a pool of HTTP servers
type HTTPPool struct {
	// self is the address of this node
	self string
	// basePath is the base path of the node
	basePath string
	// engine comical http engine
	engine *comical.Engine
	// lock guards peers and httpGetters
	lock sync.RWMutex
	// peers consistent hash map
	peers *consistent_hash.Map
	// httpGetters http getter map
	httpGetters map[string]*httpGetter
}

// NewHTTPPool creates a new HTTPPool
func NewHTTPPool(self string) *HTTPPool {
	h := &HTTPPool{
		self:     self,
		basePath: DefaultBasePath,
		engine:   comical.New(),
	}
	h.registerRoute()
	return h
}

// Log logs a message
func (p *HTTPPool) Log(format string, v ...any) {
	log.Printf("[Comical-KV Server %s] %s\n", p.self, fmt.Sprintf(format, v...))
}

// registerRoute registers the route for the HTTPPool
func (p *HTTPPool) registerRoute() {
	// register the route for the view
	p.engine.GET(p.basePath+":groupName/:key", func(c *comical.Context) {
		p.Log("%s %s", c.Method, c.Path)

		// get the group name and key from the path
		groupName := c.Param("groupName")
		if groupName == "" {
			c.Fail(http.StatusBadRequest, "bad request, no group name")
			return
		}
		key := c.Param("key")
		if key == "" {
			c.Fail(http.StatusBadRequest, "bad request, no key")
			return
		}

		// get the group
		group, ok := GetGroup(groupName)
		if !ok {
			c.Fail(http.StatusNotFound, fmt.Sprintf("no such group: %s", groupName))
			return
		}

		// get the view
		view, err := group.Get(key)
		if err != nil {
			c.Err(err)
			return
		}

		// return the view
		c.SetHeader("Content-Type", "application/octet-stream")
		c.Data(200, view.ByteSlice())
	})
}

// ServeHTTP serves the HTTPPool
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.engine.ServeHTTP(w, r)
}

// Set sets the peers of the HTTPPool
func (p *HTTPPool) Set(peers ...string) {
	// acquire lock
	p.lock.Lock()
	defer p.lock.Unlock()
	// create a consistent-hash map
	p.peers = consistent_hash.New(DefaultReplica, nil)
	p.peers.Add(peers...)
	// create http getter map
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseUrl: peer + p.basePath}
	}
}

// PickPeer picks a peer for a key
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	// acquire lock
	p.lock.RLock()
	defer p.lock.RUnlock()
	// if no peers, return false
	if p.peers == nil {
		return nil, false
	}
	// pick a peer
	peer := p.peers.Get(key)
	if peer == "" || peer == p.self {
		return nil, false
	}
	return p.httpGetters[peer], true
}

var (
	// check implementation
	_ PeerGetter = (*httpGetter)(nil)
)

// httpGetter is a Getter that gets data from an HTTP server
type httpGetter struct {
	// baseUrl is the base URL of the HTTP server
	baseUrl string
}

// Get gets data from an HTTP server
func (h *httpGetter) Get(group, key string) ([]byte, error) {
	// parse url use url parser
	u, err := url.Parse(h.baseUrl)
	if err != nil {
		return nil, err
	}
	// add path param use join path
	u.JoinPath(group, key)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	// check status code
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	// read body
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return bytes, nil
}
