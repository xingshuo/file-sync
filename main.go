package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type SyncConfig struct { //json.Unmarshal struct must public var
	LocalDir  string
	RemoteDir string

	SshHost     string
	SshPort     int
	SshUserName string
	SshPassword string

	IgnoreFiles []string
}

const (
	g_ConfigFile = "config.json"
)

var (
	g_WaitGroup   sync.WaitGroup
	g_SyncCfg     SyncConfig
	g_FileSyncer  *FileSyncer
	g_FileWatcher *FileWatcher
)

func loadConfig() bool {
	_, err := os.Stat(g_ConfigFile)
	if err != nil {
		return false
	}
	configJson, err := ioutil.ReadFile(g_ConfigFile)
	if err != nil {
		log.Printf("ReadFile Error:%v\n", err)
		return false
	}
	err = json.Unmarshal(configJson, &g_SyncCfg)
	if err != nil {
		log.Printf("json.Unmarshal Error:%v\n", err)
		return false
	}

	g_SyncCfg.LocalDir, err = filepath.Abs(g_SyncCfg.LocalDir) //change to abs path
	if err != nil {
		log.Printf("Change LocalDir Abs Error:%v\n", err)
		return false
	}

	return true
}

func init() {
	iCpuNum := runtime.NumCPU()
	runtime.GOMAXPROCS(iCpuNum)
}

func main() {
	ok := loadConfig()
	if !ok {
		os.Exit(-1)
	}
	g_FileSyncer = newFileSyncer()
	ok = g_FileSyncer.Connect()
	if !ok {
		os.Exit(-1)
	}
	g_FileWatcher = newFileWatcher()
	ok = g_FileWatcher.Init()
	if !ok {
		os.Exit(-1)
	}

	g_WaitGroup.Add(3)
	go g_FileSyncer.Run()
	go g_FileWatcher.Run()
	go handleConsole()

	g_WaitGroup.Wait()

	log.Printf("Safe Exit!\n")
}
