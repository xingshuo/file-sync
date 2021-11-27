package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	g_ConfigFile  = "config.json"
	g_SyncCfg     SyncConfig
)

type SyncConfig struct { //json.Unmarshal struct must public var
	LocalDir  string //absolute path needed
	RemoteDir string

	SshHost     string
	SshPort     int
	SshUserName string
	SshPassword string

	IgnoreFiles []string
	IgnoreDirs  []string //relative path to LocalDir
	ReplaceRule map[string]string
}

func loadConfig() bool {
	flag.StringVar(&g_ConfigFile, "config", "config.json", "sync config file")
	flag.Parse()
	_, err := os.Stat(g_ConfigFile)
	if err != nil {
		log.Printf("Not Exist ConfigFile:%v\n", err)
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

	if !filepath.IsAbs(g_SyncCfg.LocalDir) {
		log.Print("LocalDir must be Abs Path\n")
		return false
	}
	log.Printf("---load cfg: %v----\n", g_SyncCfg)

	return true
}

func IsIgnoreDir(dirname string) bool {
	for _, prefix := range g_SyncCfg.IgnoreDirs {
		absPrefix := filepath.Join(g_SyncCfg.LocalDir, prefix)
		if strings.HasPrefix(dirname, absPrefix) {
			return true
		}
	}
	return false
}

func IsIgnoreFile(fpath string) bool {
	dirname, filename := filepath.Split(fpath)
	for _, suffix := range g_SyncCfg.IgnoreFiles {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	for _, prefix := range g_SyncCfg.IgnoreDirs {
		absPrefix := filepath.Join(g_SyncCfg.LocalDir, prefix)
		if strings.HasPrefix(dirname, absPrefix) {
			return true
		}
	}
	return false
}

func JoinRemotePath(localPath string) string { //remote abs dir or file path
	localPath = strings.Replace(localPath, "\\", "/", -1)
	localPath = strings.Replace(localPath, g_SyncCfg.LocalDir, "", -1)
	syncPath := filepath.ToSlash(localPath) //change platform dependent path delimiter to '/', example on windows '\' -> '/'
	return path.Join(g_SyncCfg.RemoteDir, syncPath)
}
