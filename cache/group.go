package cache

import (
	"airlsubject/airlcache/cache/singleflight"
	"errors"
	"log"
	"sync"
)

// 是
// 接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
// |  否                         是
// |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
// |  否
// |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

//主体结构 Group
//Group 是 GeeCache 最核心的数据结构，负责与用户的交互，并且控制缓存值存储和获取的流程

// 我们思考一下，如果缓存不存在，应从数据源（文件，数据库等）获取数据并添加到缓存中。GeeCache 是否应该支持多种数据源的配置呢？
// 不应该，一是数据源的种类太多，没办法一一实现；二是扩展性不好。如何从源头获取数据，应该是用户决定的事情，我们就把这件事交给用户好了。
// 因此，我们设计了一个回调函数(callback)，在缓存不存在时，调用这个函数，得到源数据。

// 定义接口 Getter 和 回调函数 Get(key string)([]byte, error)，参数是 key，返回值是 []byte。
// 定义函数类型 GetterFunc，并实现 Getter 接口的 Get 方法。

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	if g == nil {
		return []byte{}, errors.New("未设置本地获取方法")
	}

	return g(key)
}

//参数可配置
type Option func(*Group)

func GroupName(name string) Option {
	return func(g *Group) {
		g.name = name
	}
}

func GroupCacheBytes(cachebytes int64) Option {
	return func(g *Group) {
		g.mainCache = cache{cacheBytes: cachebytes}
	}
}

func GroupGetter(getter Getter) Option {
	return func(g *Group) {
		g.getter = getter
	}
}

type Group struct {
	name       string
	getter     Getter
	mainCache  cache
	peerPicker PeerPicker
	//限制并发请求次数，一次请求内，不管并发多少个请求，都只会获取一次
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// func NewGroup(name string, cachebytes int64, getter Getter) *Group {
// 	mu.Lock()
// 	defer mu.Unlock()
// 	var group = &Group{
// 		name:      name,
// 		getter:    getter,
// 		mainCache: cache{cacheBytes: cachebytes},
// 	}
// 	groups[name] = group

// 	return group
// }

func NewGroup(opt ...Option) *Group {
	mu.Lock()
	defer mu.Unlock()

	var defaultGroup = &Group{
		name:      "default",
		getter:    nil,
		mainCache: cache{cacheBytes: 2 << 10},
		loader:    new(singleflight.Group),
	}

	for _, o := range opt {
		o(defaultGroup)
	}

	groups[defaultGroup.name] = defaultGroup

	return defaultGroup
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	return groups[name]
}

// 一个 Group 可以认为是一个缓存的命名空间，每个 Group 拥有一个唯一的名称 name。
// 比如可以创建三个 Group，缓存学生的成绩命名为 scores，缓存学生信息的命名为 info，缓存学生课程的命名为 courses。
// 第二个属性是 getter Getter，即缓存未命中时获取源数据的回调(callback)。
// 第三个属性是 mainCache cache，即一开始实现的并发缓存。
// 构建函数 NewGroup 用来实例化 Group，并且将 group 存储在全局变量 groups 中。
// GetGroup 用来特定名称的 Group，这里使用了只读锁 RLock()，因为不涉及任何冲突变量的写操作。

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	if v, ok := g.mainCache.Get(key); ok {
		log.Println("从当前缓存中获取到数据")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	//远程获取
	//使用loader    限制并发次数
	bvRes, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peerPicker != nil {
			if peergetter, ok := g.peerPicker.PickPeer(key); ok {
				log.Println("获取到数据存储节点：", peergetter.(*httpGetter).baseURL)
				if res, err := peergetter.Get(g.name, key); err == nil {
					log.Println("返回数据：", string(res))
					return ByteView{res}, nil
				}

				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getlocal(key)
	})

	if err == nil {
		return bvRes.(ByteView), nil
	}
	return
}

func (g *Group) getlocal(key string) (ByteView, error) {
	v, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	var val = ByteView{cloneBytes(v)}
	g.populateCache(key, val)
	log.Println("从本地缓存方法中获取到数据", string(v))

	return val, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}

// Get 方法实现了上述所说的流程 ⑴ 和 ⑶。
// 流程 ⑴ ：从 mainCache 中查找缓存，如果存在则返回缓存值。
// 流程 ⑶ ：缓存不存在，则调用 load 方法，load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取），
// getLocally 调用用户回调函数 g.getter.Get() 获取源数据，并且将源数据添加到缓存 mainCache 中（通过 populateCache 方法）

func (g *Group) RegisterPeerPick(peerpick PeerPicker) {
	if g.peerPicker != nil {
		log.Println("RegisterPeerPicker called more than once")
		return
	}

	g.peerPicker = peerpick
}
