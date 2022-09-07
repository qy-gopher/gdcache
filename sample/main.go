package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/qy-gopher/gdcache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gdcache.Group {
	return gdcache.NewGroup("scores", 2<<10, gdcache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}

		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs []string, gd *gdcache.Group) {
	peers := gdcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gd.RegisterPeers(peers)

	log.Println("gdcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gd *gdcache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := gd.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool

	flag.IntVar(&port, "port", 8081, "GdCache server port")
	flag.BoolVar(&api, "api", false, "strat a api server")
	flag.Parse()

	apiAddr := "http://localhost:9090"
	addrMap := map[int]string{
		8081: "http://localhost:8081",
		8082: "http://localhost:8082",
		8083: "http://localhost:8083",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gd := createGroup()

	if api {
		go startAPIServer(apiAddr, gd)
	}

	startCacheServer(addrMap[port], addrs, gd)
}
