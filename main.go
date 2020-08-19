package airlcache

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	//traceMemStats()
	for i := 0; i < 10; i++ {
		go printOnce(100)
	}
	time.Sleep(time.Second)
}

var set = make(map[int]bool, 0)

var lock sync.Mutex

func printOnce(num int) {

	lock.Lock()
	defer lock.Unlock()

	if _, exist := set[num]; !exist {
		fmt.Println(num)
	}
	set[num] = true
}

func traceMemStats() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("Alloc:%d(bytes) HeapIdle:%d(bytes) HeapReleased:%d(bytes) sys:%d(bytes)", ms.Alloc, ms.HeapIdle, ms.HeapReleased, ms.Sys)
}
