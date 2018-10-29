FileSync
=========
    An auto sync the changes dir/file to remote server tool via ssh/sftp.

Wiki
----
    基于golang编写的文件自动同步工具,支持通过终端命令同步或删除指定文件或目录(use 'help' for usage)

前置环境
-----
    golang

支持平台
-----
    Linux/Windows

安装
-----
    1. go get github.com/xingshuo/file-sync
    2. go build

运行
-----
    run "file-sync" or "file-sync.exe"(windows)

Tips
-----
    1. go get过程中出现package golang.org/x/sys/unix: unrecognized import path "golang.org/x/sys/unix"报错,解决方案参考:
       https://javasgl.github.io/go-get-golang-x-packages/
    2. 由于fsnotify库实现问题，本工具只监听:文件/目录的create write remove事件(放弃rename事件监听)，创建文件/目录,修改文件,删除文件或目录操作均可自动同步远端.
       但文件/目录重命名仅会自动同步更名后的目标到远端，旧的文件/目录不会自动从远端移除，有需要可通过终端remove命令手动移除