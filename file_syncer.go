package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type FileSyncer struct {
	sftpClient  *sftp.Client
	syncEvent   chan string
	removeEvent chan string
	doneEvent   chan struct{}
}

func (s *FileSyncer) Connect() bool {
	auth := make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(g_SyncCfg.SshPassword))
	clientConfig := &ssh.ClientConfig{
		User:    g_SyncCfg.SshUserName,
		Auth:    auth,
		Timeout: 15 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	addr := fmt.Sprintf("%s:%d", g_SyncCfg.SshHost, g_SyncCfg.SshPort)
	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		log.Fatalf("connect [%s] failed:%v\n", addr, err)
		return false
	}
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		log.Fatalf("new sftp client failed:%v \n", err)
		return false
	}
	_, err = sftpClient.Stat(g_SyncCfg.RemoteDir)
	if err != nil {
		sftpClient.Close()
		log.Fatalf("RemoteDir Error: %v\n", err)
		return false
	}
	s.sftpClient = sftpClient
	return true
}

func (s *FileSyncer) Disconnect() {
	s.sftpClient.Close()
}

func (s *FileSyncer) Run() {
	defer s.Disconnect()
	defer g_WaitGroup.Done()
	for {
		select {
		case localSyncPath := <-s.syncEvent:
			{
				file, err := os.Stat(localSyncPath)
				if err != nil { //when delete dir, the remove event notify two times, maybe you can ignore this err
					log.Printf("NoExist Sync Path:%s :%v\n", localSyncPath, err)
					break
				}
				if file.IsDir() {
					s.SyncDir(localSyncPath)
				} else {
					s.SyncFile(localSyncPath)
				}
			}
		case localRemovePath := <-s.removeEvent:
			{
				remoteRemovePath := s.JoinRemotePath(localRemovePath)
				file, err := s.sftpClient.Stat(remoteRemovePath)
				if err != nil {
					log.Printf("NoExist Remove Path:%s :%v\n", localRemovePath, err)
					break
				}
				if file.IsDir() {
					s.RemoveDir(remoteRemovePath)
				} else {
					s.RemoveFile(remoteRemovePath)
				}
			}
		case <-s.doneEvent:
			{
				return
			}
		}
	}
}

func (s *FileSyncer) JoinRemotePath(localPath string) string { //remote abs dir or file path
	localPath = strings.Replace(localPath, g_SyncCfg.LocalDir, "", -1)
	syncPath := filepath.ToSlash(localPath)
	return path.Join(g_SyncCfg.RemoteDir, syncPath)
}

func (s *FileSyncer) SyncFile(localFilePath string) error {
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Printf("sync file %s failed: %v\n", localFilePath, err)
		return err
	}
	defer srcFile.Close()
	remoteFilePath := s.JoinRemotePath(localFilePath)
	dstFile, err := s.sftpClient.Create(remoteFilePath)
	if err != nil {
		fmt.Printf("create remote file %s failed: %v\n", remoteFilePath, err)
		return err
	}
	defer dstFile.Close()
	stream, err := ioutil.ReadAll(srcFile)
	if err != nil {
		fmt.Printf("read localFile %s failed: %v\n", localFilePath, err)
		return err
	}
	dstFile.Write(stream)
	log.Printf("sync file: %s -> %s ok\n", localFilePath, remoteFilePath)
	return nil
}

func (s *FileSyncer) SyncDir(localDirPath string) error {
	localFiles, err := ioutil.ReadDir(localDirPath)
	if err != nil {
		fmt.Printf("sync dir %s failed: %v\n", localDirPath, err)
		return err
	}
	remoteJoinDir := s.JoinRemotePath(localDirPath)
	s.sftpClient.Mkdir(remoteJoinDir)
	for _, file := range localFiles {
		subSyncPath := path.Join(localDirPath, file.Name())
		if file.IsDir() {
			s.SyncDir(subSyncPath)
		} else {
			s.SyncFile(subSyncPath)
		}
	}
	log.Printf("sync dir: %s -> %s ok\n", localDirPath, remoteJoinDir)
	return nil
}

func (s *FileSyncer) RemoveFile(remoteFilePath string) error {
	err := s.sftpClient.Remove(remoteFilePath)
	if err != nil {
		log.Printf("remove remote file: %s err: %v\n", remoteFilePath, err)
	} else {
		log.Printf("remove remote file: %s ok\n", remoteFilePath)
	}
	return err
}

func (s *FileSyncer) RemoveDir(remoteRemoveDir string) error {
	remoteFiles, err := s.sftpClient.ReadDir(remoteRemoveDir)
	if err != nil {
		log.Printf("remove remote dir: %s err: %v\n", remoteRemoveDir, err)
		return err
	}
	for _, file := range remoteFiles {
		subRemovePath := path.Join(remoteRemoveDir, file.Name())
		if file.IsDir() {
			s.RemoveDir(subRemovePath)
		} else {
			s.RemoveFile(subRemovePath)
		}
	}
	s.sftpClient.RemoveDirectory(remoteRemoveDir) //must empty dir to remove
	log.Printf("remove remote dir: %s ok\n", remoteRemoveDir)
	return nil
}

func (s *FileSyncer) IsIgnoreFile(filename string) bool {
	for _, suffix := range g_SyncCfg.IgnoreFiles {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}

func newFileSyncer() *FileSyncer {
	return &FileSyncer{
		sftpClient:  nil,
		syncEvent:   make(chan string),
		removeEvent: make(chan string),
		doneEvent:   make(chan struct{}),
	}
}