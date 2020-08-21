package cache

import (
	"airlsubject/airlcache/cache/consistenthash"
	"bufio"
	"fmt"

	//"io"
	//"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type HttpPool struct {
	//自身ip+端口地址
	self string
	//基础前缀
	basePath string

	mu sync.Mutex
	//新增成员变量 peers，类型是一致性哈希算法的 Map，用来根据具体的 key 选择节点。
	peers *consistenthash.Map
	//新增成员变量 httpGetters，映射远程节点与对应的 httpGetter。
	//每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关。
	httpGetters map[string]*httpGetter
}

const (
	basePathStr     = "/airlcache/"
	defaultReplicas = 50
)

func NewHttpPool(s string) *HttpPool {
	return &HttpPool{
		self:     s,
		basePath: basePathStr,
	}
}

// HTTPPool 只有 2 个参数，一个是 self，用来记录自己的地址，包括主机名/IP 和端口。
// 另一个是 basePath，作为节点间通讯地址的前缀，默认是 /_geecache/，那么 http://example.com/_geecache/ 开头的请求，
// 就用于节点间的访问。因为一个主机上还可能承载其他的服务，加一段 Path 是一个好习惯。比如，大部分网站的 API 接口，一般以 /api 作为前缀。

func (h *HttpPool) log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

//服务器逻辑
func (h *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		http.Error(w, "访问地址不存在", http.StatusNotFound)
		return
	}

	var paths = strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(paths) != 2 {
		http.Error(w, "访问地址不存在", http.StatusNotFound)
		return
	}

	var groupName = paths[0]
	var key = paths[1]

	var group = GetGroup(groupName)
	if group == nil {
		http.Error(w, "group不存在", http.StatusInternalServerError)
		return
	}

	val, err := group.Get(key)
	if err != nil {
		http.Error(w, "key不存在", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset = utf-8")
	w.Write(val.ByteSlice())
}

//实现PeerGetter
type httpGetter struct {
	baseURL string
}

//实现PeerGetter
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	var u = fmt.Sprintf("%s%s/%s", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	var r []byte
	//三种读取方式，bufio总体比较快   有缓冲
	//1 bufio
	var reader = bufio.NewReader(resp.Body)

	if _, err = reader.Read(r); err != nil {
		return nil, err
	}

	//2 ioutil
	// if _, err = ioutil.ReadAll(resp.Body, r); err != nil {
	// 	return nil, err
	// }

	//3 自身reader
	// if _, err = resp.Body.Read(r); err != nil {
	// 	return nil, err
	// }

	return r, nil
}

//设置分布式节点相关信息
// Set() 方法实例化了一致性哈希算法，并且添加了传入的节点。
// 并为每一个节点创建了一个 HTTP 客户端 httpGetter。
func (h *HttpPool) Set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.peers = consistenthash.New(defaultReplicas, nil)
	h.peers.Add(peers...)
	h.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, p := range peers {
		h.httpGetters[p] = &httpGetter{baseURL: p + h.basePath}
	}
}

//选择节点
//PickerPeer() 包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func (h *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if peer := h.peers.Get(key); peer != "" {
		log.Printf("获取到远程节点：%s", peer)
		if peer != h.self {
			return h.httpGetters[peer], true
		}
		log.Printf("远程节点为自身服务节点：%s", peer)
	}
	return nil, false
}
