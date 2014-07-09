package main

import (
	"fmt"
	"os"
	"errors"
)

func main() {
	fmt.Println("starting up")
	queue := make(chan string, 8196)

	_ = watch(queue)

//	for {
		// every day, do a full re-index
//	}
//	fichan := make(chan []os.FileInfo, 4096)

	for _, path := range os.Args[1:] {
		queue<-path
	}


//		err := walk(path)
//		if err != nil {
//			fmt.Println("ERROR: ", err)
//		}
//		err := watch_add(path)
//		if err != nil {
//			fmt.Println("ERROR: ", err)
//		}
//	}

	select {}
}

func walk(path string) error {
	pi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !pi.IsDir() {
		return errors.New("Not a directory: "+path)
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	fi, err := f.Readdir(0)
	if err != nil {
		return err
	}

	for _, ffi := range fi {
		if ffi.IsDir() {
			err := walk(path + "/" + ffi.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

