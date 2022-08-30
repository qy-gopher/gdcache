package lru

import "container/list"

type Cache struct {
	maxBytes int64 // 允许使用的最大内存，为0时不对内存做限制
	nbytes   int64 // 当前已使用内存
	ll       *list.List
	cache    map[string]*list.Element

	OnEvicted func(key string, value Value) // 记录被移除时的回调函数
}

// entry 对真实缓存值的封装
// 添加key是为了淘汰队首节点时，获取key删除map中相应的映射
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int // 值的字节数
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (Value, bool) {
	if el, ok := c.cache[key]; ok {
		c.ll.MoveToBack(el)
		kv := el.Value.(*entry)
		return kv.value, true
	}

	return nil, false
}

// Add 当key不存在时添加元素，存在时修改对应元素值
func (c *Cache) Add(key string, value Value) {
	if el, ok := c.cache[key]; ok {
		c.ll.MoveToBack(el)
		kv := el.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value

	} else {
		el := c.ll.PushBack(&entry{key, value})
		c.cache[key] = el
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	el := c.ll.Front()
	if el == nil {
		return
	}

	c.ll.Remove(el)
	kv := el.Value.(*entry)
	delete(c.cache, kv.key)
	c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
