package main

import (
	"code.google.com/p/go.exp/fsnotify"
	"log"
	"os"
	"errors"
)

var watcher *fsnotify.Watcher
var watched map[string]struct{}

func watch(queue chan string) chan *fsnotify.FileEvent {

	watched = make(map[string]struct{})

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
					log.Println("event:", ev)
					if ev.IsDelete() || ev.IsRename() {
						_, ok := watched[ev.Name]
						if ok {
							delete(watched, ev.Name)
							log.Println("removing path: "+ev.Name)
							err := watcher.RemoveWatch(ev.Name)
							if err != nil {
								watcher.Error<-err
								break
							}
						}
					} else if ev.IsCreate() {
						pi, err := os.Stat(ev.Name)
						if err != nil {
							watcher.Error<-err
							break
						}
						if pi.IsDir() {
							queue<-ev.Name
						}
					}
			case path := <-queue:
				_, ok := watched[path]
				if ok {
					log.Println("already watching "+path)
				} else {
					log.Println("adding path: "+path)
					watched[path] = struct{}{}
					we := watcher.Watch(path)
					if we != nil {
						watcher.Error<-we
						break
					}
					we = watch_walk(path, queue)
					if we != nil {
						watcher.Error<-we
						break
					}
				}
			case err := <-watcher.Error:
					log.Println("error:", err)
			}
		}
	}()

	return watcher.Event
}

func watch_walk(path string, queue chan string) error {
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
//			err := walk(path + "/" + ffi.Name())
//			if err != nil {
//				return err
//			}
			queue <- path + "/" + ffi.Name()
		}
	}

	return nil
}
