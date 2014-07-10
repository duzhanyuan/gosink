package main

import (
	"code.google.com/p/go.exp/fsnotify"
	"errors"
	"log"
	"os"
)

func watch(queue chan string) (chan *fsnotify.FileEvent, error) {

	watched := make(map[string]struct{})
	watcher, err := fsnotify.NewWatcher()
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
					go func() {
						if we != nil {
							watcher.Error <- we
							return
						}

						err := watch_walk(path, queue)
						if err != nil {
							watcher.Error <- err
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
			queue <- path + "/" + ffi.Name()
		}
	}

	return nil
}
