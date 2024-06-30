package comical_kv

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kiritoxkiriko/comical"
)

const (
	DefaultBasePath = "/_comical-kv/"
)

// HTTPPool is a pool of HTTP servers
type HTTPPool struct {
	// self is the address of this node
	self string
	// basePath is the base path of the node
	basePath string
	// engine comical http engine
	engine *comical.Engine
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
