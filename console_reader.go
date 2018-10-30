package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func handleConsole() {
	defer close(g_FileWatcher.doneEvent)
	defer close(g_FileSyncer.doneEvent)
	defer g_WaitGroup.Done()
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, err := inputReader.ReadString('\n')
		if err != nil {
			log.Printf("read input:%s error:%v\n", input, err)
			break
		}
		input = strings.Replace(input, "\r\n", "", -1)
		input = strings.Replace(input, "\n", "", -1)
		cmds := strings.Split(input, " ")
		switch {
		case cmds[0] == "sync":
			{
				if len(cmds) < 2 {
					log.Print("sync lost dir/file path args\n")
					break
				}
				syncPath := filepath.Join(g_SyncCfg.LocalDir, cmds[1])
				log.Printf("usr cmd sync:%s\n", syncPath)
				g_FileSyncer.syncEvent <- syncPath
			}
		case cmds[0] == "remove":
			{
				if len(cmds) < 2 {
					log.Print("remove lost dir/file path args\n")
					break
				}
				removePath := filepath.Join(g_SyncCfg.LocalDir, cmds[1])
				log.Printf("usr cmd remove:%s\n", removePath)
				g_FileSyncer.removeEvent <- removePath
			}
		case cmds[0] == "help":
			{
				log.Print("sync local dir/file to remote: 	sync dirpath/filepath\n")
				log.Print("remove remote dir/file: 	 remote dirpath/filepath\n")
				log.Print("quit app:	exit or quit\n")
			}
		case cmds[0] == "exit" || cmds[0] == "quit":
			{
				return
			}
		default:
			{
				if len(input) > 0 {
					log.Printf("unkonw cmd:%s\nenter 'help' for usage\n", input)
				}
			}
		}
	}
}
