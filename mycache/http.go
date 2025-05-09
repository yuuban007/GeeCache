package mycache

import (
	"fmt"
	"mycache/consistenthash"
	pb "mycache/mycachepb"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"
)

const defaultBasePath = "/_geecache/"
const defaultReplicas = 50

// HTTP pool implements PeerPicker for a pool of HTTP peer
// HTTPPool implements a pool of HTTP peers that can be used for distributed caching.
type HTTPPool struct {
	// self is the base URL of this peer, e.g. "https://example.net:800".
	self string
	// basePath is the base path of the cache server, e.g. "/cache/".
	basePath string
	// mu is a mutex to protect the peers and httpGetter maps.
	mu sync.Mutex
	// peers is a consistent hash map that stores the URLs of all the peers in the pool.
	peers *consistenthash.Map
	// httpGetter is a map that stores the HTTP client for each peer URL.
	httpGetter map[string]*httpGetter
}

type httpGetter struct {
	baseURL string
}

// NewHTTPoll initializes an HTTP pool of peers
func NewHTTPPool(self string) *HTTPPool {

	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServerHTTP handle all http request
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool severing unexpected path " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basePath>/<groupName>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad Request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the value to the response body as a protobuf message
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set sets the list of peers in the HTTPPool.
// It takes a variadic parameter `peers` which represents the list of peers to be added.
// The method uses a consistent hash algorithm to distribute the peers across the hash ring.
// It also initializes the `httpGetter` map with the base URL for each peer.
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetter = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetter[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

// PickPeer picks a peer according to key.
// mainly use consistenthash map Get() function
func (p *HTTPPool) PickPeer(key string) (peer PeerGetter, ok bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// peer not refer to p itself
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("pick peer %s", peer)
		return p.httpGetter[peer], true
	}
	return nil, false
}

// Get retrieves the value associated with the given group and key from the remote cache server.
// It returns the value as a byte slice and an error if any occurred.
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	group := in.GetGroup()
	key := in.GetKey()
	u := fmt.Sprintf("%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)

	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server return: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response: %v", err)
	}
	return nil
}
