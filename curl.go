package main

import (
	"fmt"
	"sync"

	"github.com/parnurzeal/gorequest"
)

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go curl(&wg)
	}
	wg.Wait()
}

func curl(wg *sync.WaitGroup) {
	_, body, errs := gorequest.New().Get("http://127.0.0.1:8789/").End()
	println(body)
	for _, err := range errs {
		fmt.Println(err)
	}
	wg.Done()
}
