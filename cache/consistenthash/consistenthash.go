package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type hashMethod func([]byte) uint32

type Map struct {
	hashmethod hashMethod
	replicas   int
	keys       []int
	hashMap    map[int]string
}

func New(replicas int, hash hashMethod) *Map {
	var mapInfo = &Map{
		hashmethod: hash,
		replicas:   replicas,
		hashMap:    make(map[int]string),
	}

	if mapInfo.hashmethod == nil {
		mapInfo.hashmethod = crc32.ChecksumIEEE
	}

	return mapInfo
}

// 定义了函数类型 Hash，采取依赖注入的方式，允许用于替换成自定义的 Hash 函数，也方便测试时替换，默认为 crc32.ChecksumIEEE 算法。
// Map 是一致性哈希算法的主数据结构，包含 4 个成员变量：Hash 函数 hash；虚拟节点倍数 replicas；哈希环 keys；虚拟节点与真实节点的映射表 hashMap，
// 键是虚拟节点的哈希值，值是真实节点的名称。
// 构造函数 New() 允许自定义虚拟节点倍数和 Hash 函数。

//添加真实节点/机器
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			var hashVal = m.hashmethod([]byte(strconv.Itoa(i) + key))
			m.keys = append(m.keys, int(hashVal))
			m.hashMap[int(hashVal)] = key
		}
	}

	//记得排序
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	var hashVal = m.hashmethod([]byte(key))
	var index = sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= int(hashVal)
	})

	return m.hashMap[m.keys[index%len(m.keys)]]
}
