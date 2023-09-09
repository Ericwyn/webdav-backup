# WebDavBackup
用来将远程 webdav 服务器的全部文件拉取到本地的工具

可用来基于 webdav 服务器做备份

比如我们搭建了一个 seafile ，之后希望可以在自己的另一个 Nas 上面做一个定期备份，那么就可以用得上这个工具

## 快速使用
```shell
./webdavBackup -c "/etc/webdavBackup.json" -m "backup"
```
配置文件不存在的情况下, 会自动创建默认配置文件, 文件如下

```json
{
  "BaseUrl": "https://test.dav",
  "User": "user-account",
  "Password": "user-password",
  "TargetDir": "target-local-backup-dir-full-path"
}
```
## 支持参数
```shell
  -c string
        config file path (shorthand) (default "./webdavBackup.json")
  -conf string
        config file path (default "./webdavBackup.json")
  -d    show debug log (shorthand)
  -debug
        show debug log
  -m string
        backup mode: "backup" or "mirror", (shorthand) (default "backup")
  -mode string
        backup mode: "backup" or "mirror"
        "backup": In this mode, if a file is deleted or moved on the webdav server,
                  the local copy of the file is still retained
        "mirror": In this mode, if a file is deleted or moved on the webdav server,
                  the corresponding local copy is also deleted,
                  effectively mirroring the state of the server  (default "backup")
  -v    show version msg (shorthand)
  -version
        show version msg

```

- `-conf` / `-c`
  - 配置文件位置, 配置文件为 json 格式

- `-mode` / `-m` 备份模式
  - backup: 增量备份
    - 源目录删除的文件，本地不会删除
    - 本地文件如果与源目录的文件大小和修改日志都一样，那么就会自动跳过
  - mirror: 镜像模式
    - 源目录删除的文件，本地也会删除

- `-version` / `-v`
  - 软件版本

- `-debug` / `-d`
  - 显示 debug 日志


## 参考脚本
- 可 cron 定期执行
- 如果有正在运行的线程可以会退出

```shell
#!/bin/bash

# 日志位置
log_path="/var/log/seafile-back-log"
log_file="${log_path}/$(date +%Y%m%d_%H_%M_%S).log"

# 检查其他 webdavBackup 线程是否正在运行
if pgrep -x "webdavBackup" > /dev/null
then
    echo "A webdavBackup process is already running - skipping execution." | tee -a "${log_file}"
    exit 1
else
    # 如果没有其他 webdavBackup 轮程运行，执行你的原始命令并将输出重定向到日志文件
    /opt/WebdavBackup/build/webdavBackup -c /opt/WebdavBackup/build/webdavBackup.json >> "${log_file}" 2>&1   
    exit 0
fi

```