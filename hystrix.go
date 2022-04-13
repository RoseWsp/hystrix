package hystrix

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Limiter interface {
	Visit() bool
	Start()
	Stop()
	Max() int
	Count() int
}

var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

type reallimit struct {
	buckets []int32
	max     int           //支持最大请求数量
	windows time.Duration //窗口时间

	count int32 // 当前最大请求数量

	index  int //当前执行的桶
	bucket int // 桶数量

	isStart uint32
	mu      sync.Mutex

	ticker *time.Timer
	close  chan struct{}
}

func New(bucket, max int, windows time.Duration) Limiter {
	return &reallimit{
		buckets: make([]int32, bucket),
		max:     max,
		windows: windows,
		index:   0,
		bucket:  bucket,
		isStart: 0,
	}
}

func (r *reallimit) Max() int {
	return r.max
}

func (r *reallimit) Count() int {
	return int(r.count)
}

func (r *reallimit) Visit() bool {
	if r.count >= int32(r.max) {
		return false
	}
	if atomic.LoadInt32(&r.count) >= int32(r.max) {
		return false
	}
	atomic.AddInt32(&r.count, 1)
	atomic.AddInt32(&r.buckets[r.index], 1)
	if atomic.LoadInt32(&r.count) > int32(r.max) {
		return false
	}
	return true
}

func (r *reallimit) Start() {
	if atomic.LoadUint32(&r.isStart) == 1 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ticker = time.NewTimer(r.windows / time.Duration(r.bucket))
	go tick(r)
	defer atomic.AddUint32(&r.isStart, 1)
}

func tick(r *reallimit) {
	fmt.Println("tick start")
	for {
		select {
		case <-r.close:
			r.ticker.Stop()
			return
		case <-r.ticker.C:
			slide(r)
			r.ticker.Reset(r.windows / time.Duration(r.bucket))
		}
	}
}

func slide(r *reallimit) {
	r.index = (r.index + 1) % r.bucket
	val := atomic.LoadInt32(&r.buckets[r.index])

	if val != 0 {
		// release
		atomic.AddInt32(&r.buckets[r.index], -val)
		atomic.AddInt32(&r.count, -val)
	}
}

func (r *reallimit) Stop() {
	if r.close == nil {
		r.close = closedchan
	}
}
