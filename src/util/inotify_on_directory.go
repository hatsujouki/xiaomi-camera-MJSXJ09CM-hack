package main

import "log"
import "gopkg.in/fsnotify/fsnotify.v1"

func main() {
    log.Print("start")
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Close()

    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }
                log.Println("event:", event)
//                if event.Has(fsnotify.Write) {
//                    log.Println("modified file:", event.Name)
//                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("error:", err)
            }
        }
    }()

    err = watcher.Add("/tmp/longnamedrepofolder")
    if err != nil {
        log.Fatal(err)
    }
    <-make(chan struct{})

    log.Print("stop")
}
