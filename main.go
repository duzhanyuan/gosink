package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("starting up")
	queue := make(chan string)

	watch(queue)

	go func() {
		for _, path := range os.Args[1:] {
			queue <- path
		}
	}()

	select {}
}
