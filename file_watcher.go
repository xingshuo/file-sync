package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
)

type FileWatcher struct {
	handler   *fsnotify.Watcher
	doneEvent chan struct{}
}

func (w *FileWatcher) Init() bool {
	_, err := os.Stat(g_SyncCfg.LocalDir)
	if err != nil {
		log.Printf("os.Stat LocalDir %s error:v\n", g_SyncCfg.LocalDir, err)
		return false
	}
	log.Printf("Run Watch: %s\n", g_SyncCfg.LocalDir)
	filepath.Walk(g_SyncCfg.LocalDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			path, err := filepath.Abs(path)
			if err != nil {
				log.Fatalf("Walk filepath:%s err1:%v\n", path, err)
			}
			if IsIgnoreDir(path) {
				// log.Printf("Ignore path: %s\n", path)
				return nil
			}
			err = w.handler.Add(path)
			if err != nil {
				log.Fatalf("Walk filepath:%s err2:%v\n", path, err)
			}
			log.Printf("Watching path: %s\n", path)
		}

		return nil
	})
	log.Printf("Watch: %s Ok!\n", g_SyncCfg.LocalDir)
	return true
}

func (w *FileWatcher) Run() {
	defer g_WaitGroup.Done()
	for {
		select {
		case event := <-w.handler.Events:
			{
				if (event.Op & fsnotify.Create) == fsnotify.Create {
					log.Printf("----create event (name:%s) (op:%v)\n", event.Name, event.Op)
					file, err := os.Stat(event.Name)
					if err != nil {
						log.Printf("----create no exist path:%s (op:%v)\n", event.Name, event.Op)
						break
					}
					if file.IsDir() {
						w.handler.Add(event.Name)
					}
					g_ChangeQueue.changedEvents.Store(event.Name,  NewFileSyncInfo(SyncType_Sync))
				}

				if (event.Op & fsnotify.Write) == fsnotify.Write {
					log.Printf("----write event (name:%s) (op:%v)\n", event.Name, event.Op)
					g_ChangeQueue.changedEvents.Store(event.Name, NewFileSyncInfo(SyncType_Sync))
				}

				if (event.Op & fsnotify.Remove) == fsnotify.Remove {
					log.Printf("----remove event (name:%s) (op:%v)\n", event.Name, event.Op)
					file, err := os.Stat(event.Name)
					if err == nil && file.IsDir() {
						w.handler.Remove(event.Name)
					}
					g_ChangeQueue.changedEvents.Store(event.Name, NewFileSyncInfo(SyncType_Del))
				}

				if (event.Op & fsnotify.Rename) == fsnotify.Rename {
					log.Printf("----Rename event (name:%s) (op:%v)\n", event.Name, event.Op)
					file, err := os.Stat(event.Name)
					if err == nil && file.IsDir() {
						w.handler.Remove(event.Name)
					}
					g_ChangeQueue.changedEvents.Store(event.Name, NewFileSyncInfo(SyncType_Del))
				}
			}
		case err := <-w.handler.Errors:
			{
				log.Printf("File Watch Error: %v\n", err)
			}
		case <-w.doneEvent:
			{
				return
			}
		}
	}
}

func newFileWatcher() *FileWatcher {
	fw, _ := fsnotify.NewWatcher()
	return &FileWatcher{
		handler:   fw,
		doneEvent: make(chan struct{}),
	}
}
