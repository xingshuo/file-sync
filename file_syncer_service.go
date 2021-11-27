package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	g_FileSyncerService = newFileSyncerService()
)

type SyncType int
const (
	SyncType_Invalid SyncType = iota
	SyncType_Sync
	SyncType_Del
)

type FileSyncInfo struct {
	SType SyncType
}

func NewFileSyncInfo (syncType SyncType) *FileSyncInfo {
	return &FileSyncInfo{
		SType: syncType,
	}
}

type FileSyncerService struct {
}

func (fsync* FileSyncerService) Run()  {
	defer g_WaitGroup.Done()
	for {
		timeBegin := time.Now()
		g_ChangeQueue.changedEvents.Range(func (key, value interface{}) bool {
			var err error
			var file os.FileInfo
			localPath, _ := key.(string)
			fsInfo, _ := value.(*FileSyncInfo)
			if fsInfo.SType == SyncType_Sync {
				file, err = os.Stat(key.(string))
				if err != nil { //when delete dir, the remove event notify two times, maybe you can ignore this err
					log.Printf("NoExist Sync Path:%s :%v\n", localPath, err)
					g_ChangeQueue.changedEvents.Delete(key)
					return true
				}
				if !g_FileSyncer.Connect() {
					return false
				}
				defer g_FileSyncer.Disconnect()

				if file.IsDir() {
					err = g_FileSyncer.SyncDir(localPath)
				} else {
					err = g_FileSyncer.SyncFile(localPath)
				}
			} else if fsInfo.SType == SyncType_Del {
				if !g_FileSyncer.Connect() {
					return false
				}
				defer g_FileSyncer.Disconnect()

				remoteRemovePath := JoinRemotePath(localPath)
				file, err = g_FileSyncer.sftpClient.Stat(remoteRemovePath)
				if err != nil {
					log.Printf("NoExist Remove Path:%s :%v\n", localPath, err)
					g_ChangeQueue.changedEvents.Delete(key)
					return true
				}
				if file.IsDir() {
					err = g_FileSyncer.RemoveDir(remoteRemovePath)
				} else {
					err = g_FileSyncer.RemoveFile(remoteRemovePath)
				}
			} else {
				fmt.Println("something wrong")
			}
			if err != nil {
				fmt.Println("err:", err)
			}
			g_ChangeQueue.changedEvents.Delete(key)
			return true
		})

		timeEnd := time.Now()
		cost := timeEnd.Sub(timeBegin).Milliseconds()
		frameTimeMS := int64(100)
		if cost < frameTimeMS {
			frameTimeMS = frameTimeMS - cost
			time.Sleep(time.Millisecond * time.Duration(frameTimeMS))
		}
	}
}

func newFileSyncerService() *FileSyncerService {
	return &FileSyncerService{

	}
}