package main

import (
	"log"
	"os"
	"runtime"
	"sync"
)

var (
	g_WaitGroup   sync.WaitGroup
	g_FileWatcher *FileWatcher
)

func init() {
	iCpuNum := runtime.NumCPU()
	runtime.GOMAXPROCS(iCpuNum)
}

func main() {
	ok := loadConfig()
	if !ok {
		os.Exit(-1)
	}

	g_FileWatcher = newFileWatcher()
	ok = g_FileWatcher.Init()
	if !ok {
		os.Exit(-1)
	}

	g_WaitGroup.Add(2)
	go g_FileSyncerService.Run()
	go g_FileWatcher.Run()

	g_WaitGroup.Wait()

	log.Printf("Safe Exit!\n")
}
