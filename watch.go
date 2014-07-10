package main

import (
	"code.google.com/p/go.exp/fsnotify"
	"errors"
	"log"
	"os"
	//	"time"
)

var watcher *fsnotify.Watcher
var watched map[string]struct{}

func watch(queue chan string) (chan *fsnotify.FileEvent, error) {

	watched = make(map[string]struct{})

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
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
						log.Println("removing path: " + ev.Name)
						err := watcher.RemoveWatch(ev.Name)
						if err != nil {
							go func() {
								watcher.Error <- err
							}()
							break
						}
					}
				} else if ev.IsCreate() {
					pi, err := os.Stat(ev.Name)
					if err != nil {
						go func() {
							watcher.Error <- err
						}()
						break
					}
					if pi.IsDir() {
						go func() {
							queue <- ev.Name
						}()
					}
				}
			case path := <-queue:
				_, ok := watched[path]
				if ok {
					log.Println("already watching " + path)
				} else {
					log.Println(len(watched), " ", path)
					watched[path] = struct{}{}
					we := watcher.Watch(path)
					if we != nil {
						go func() {
							watcher.Error <- we
						}()
						break
					}
					go func() {
						//time.Sleep(time.Millisecond * 1000)
						we = watch_walk(path, queue)
						if we != nil {
							go func() {
								watcher.Error <- we
							}()
							return
						}
					}()
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	return watcher.Event, nil
}

func watch_walk(path string, queue chan string) error {
	pi, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !pi.IsDir() {
		return errors.New("Not a directory: " + path)
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
			//log.Println("queuing", path+"/"+ffi.Name())
			queue <- path + "/" + ffi.Name()
			//log.Println("done")
		}
	}

	return nil
}
