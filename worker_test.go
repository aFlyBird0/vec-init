package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/xxjwxc/gowp/workpool"
)

func TestWorker(t *testing.T) {
	wp := workpool.New(5)      //设置最大线程数
	for i := 0; i < 100; i++ { //开启20个请求
		ii := i
		wp.Do(func() error {
			for j := 0; j < 5; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				time.Sleep(1 * time.Second)
			}
			return nil
		})

		fmt.Println(wp.IsDone()) //判断是否完成
	}
	wp.Wait()
	fmt.Println(wp.IsDone())
	fmt.Println("done")
}
