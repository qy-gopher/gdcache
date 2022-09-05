package main

import (
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

func main() {
	gdcache.NewGroup("scores", 2<<10, gdcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB search key]", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}

			return nil, fmt.Errorf("%s not exits", key)
		}))

	addr := "localhost:8080"
	peers := gdcache.NewHTTPPool(addr)
	log.Println("gdcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
