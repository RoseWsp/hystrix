package hystrix

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestHystrix(t *testing.T) {
	Limiter := New(10, 1000, 1*time.Second)
	fmt.Println(Limiter.Max())
	fmt.Println(Limiter.Count())

	var wg sync.WaitGroup
	Limiter.Start()
	for i := 0; i < 20; i++ {
		wg.Add(1)
		goroutineNo := i
		go func(goroutineNo int) {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				rst := Limiter.Visit()
				fmt.Printf("goroutineNo:%d;Visit:%v\r\n", goroutineNo, rst)
				time.Sleep(1 * time.Microsecond)
			}
		}(goroutineNo)
	}

	wg.Wait()
	fmt.Println(Limiter.Count())
}
