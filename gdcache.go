package gdcache

import (
	"fmt"
	"log"
	"sync"

	"github.com/qy-gopher/gdcache/singleflight"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string // 缓存命名空间的名称
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache
	peers     PeerPicker

	loader *singleflight.Group // 确保相同的key只从源加载一次
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g

	return g
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}

	g.peers = peers
}

// Get method 从缓存中得到数据
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[gdCache] hit")
		return v, nil
	}

	return g.load(key)
}

// load method 缓存不存在时从源加载数据
func (g *Group) load(key string) (ByteView, error) {
	view, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFormPeer(peer, key); err == nil {
					return value, err

				} else {
					log.Println("[GeeCache] Failed to get from peer: ", err)
				}

			}
		}

		return g.getLocally(key)
	})

	if err != nil {
		return ByteView{}, err
	}

	return view.(ByteView), err
}

func (g *Group) getFormPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{b: bytes}, nil
}

// getLocally method 从本机节点加载数据
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)

	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	return groups[name]
}

// Getter interface  定义获取源数据的回调函数接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 实现Getter接口
type GetterFunc func(key string) ([]byte, error)

func (gf GetterFunc) Get(key string) ([]byte, error) {
	return gf(key)
}
