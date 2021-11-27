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

var (
	g_FileSyncer = newFileSyncer()
)

type FileSyncer struct {
	sftpClient  *sftp.Client
}

func (s *FileSyncer) Connect() bool {
	auth := make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(g_SyncCfg.SshPassword))
	clientConfig := &ssh.ClientConfig{
		User:    g_SyncCfg.SshUserName,
		Auth:    auth,
		Timeout: 20 * time.Second,
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
	if s.sftpClient != nil {
		s.sftpClient.Close()
		s.sftpClient = nil
	}
}

func (s *FileSyncer) Run() {
}

func (s *FileSyncer) SyncFile(localFilePath string) error {
	if IsIgnoreFile(localFilePath) {
		fmt.Printf("ignore sync file: %s\n", localFilePath)
		return nil
	}
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Printf("sync file %s failed: %v\n", localFilePath, err)
		return err
	}
	defer srcFile.Close()
	remoteFilePath := JoinRemotePath(localFilePath)
	fmt.Println("localFilePath:%v", localFilePath)
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
	for old, new := range g_SyncCfg.ReplaceRule {
		stream = []byte(strings.Replace(string(stream), old, new, -1))
	}
	dstFile.Write(stream)
	log.Printf("sync file: %s -> %s ok\n", localFilePath, remoteFilePath)
	return nil
}

func (s *FileSyncer) SyncDir(localDirPath string) error {
	if IsIgnoreDir(localDirPath) {
		fmt.Printf("ignore sync dir: %s\n", localDirPath)
		return nil
	}
	localFiles, err := ioutil.ReadDir(localDirPath)
	if err != nil {
		fmt.Printf("sync dir %s failed: %v\n", localDirPath, err)
		return err
	}

	if strings.HasSuffix(localDirPath, "新建文件夹") {
		return nil
	}
	if strings.Contains(localDirPath, "新建文本文档.txt") {
		return nil
	}

	remoteJoinDir := JoinRemotePath(localDirPath)
	s.sftpClient.Mkdir(remoteJoinDir)
	for _, file := range localFiles {
		subSyncPath := filepath.Join(localDirPath, file.Name())
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
	if IsIgnoreFile(remoteFilePath) {
		fmt.Printf("ignore remove file: %s\n", remoteFilePath)
		return nil
	}
	err := s.sftpClient.Remove(remoteFilePath)
	if err != nil {
		log.Printf("remove remote file: %s err: %v\n", remoteFilePath, err)
	} else {
		log.Printf("remove remote file: %s ok\n", remoteFilePath)
	}
	return err
}

func (s *FileSyncer) RemoveDir(remoteRemoveDir string) error {
	if IsIgnoreDir(remoteRemoveDir) {
		fmt.Printf("ignore remove dir: %s\n", remoteRemoveDir)
		return nil
	}
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

func newFileSyncer() *FileSyncer {
	return &FileSyncer{
		sftpClient:  nil,
	}
}
