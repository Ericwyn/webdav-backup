package main

import (
	"flag"
	"fmt"
	"github.com/Ericwyn/webdav-backup/conf"
	"github.com/Ericwyn/webdav-backup/log"
	"github.com/studio-b12/gowebdav"
	"io"
	"os"
	"path/filepath"
	"time"
)

var webdavClient *gowebdav.Client

type BackupMode string

var ModeBackup BackupMode = "backup"
var ModeMirror BackupMode = "mirror"

var configFilePath string
var mode string
var debugFlag bool
var versionFlag bool

var banner = `
=====================================================================
██╗    ██╗██████╗ ██████╗  █████╗  ██████╗██╗  ██╗██╗   ██╗██████╗   
██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔════╝██║ ██╔╝██║   ██║██╔══██╗  
██║ █╗ ██║██║  ██║██████╔╝███████║██║     █████╔╝ ██║   ██║██████╔╝  
██║███╗██║██║  ██║██╔══██╗██╔══██║██║     ██╔═██╗ ██║   ██║██╔═══╝   
╚███╔███╔╝██████╔╝██████╔╝██║  ██║╚██████╗██║  ██╗╚██████╔╝██║       
 ╚══╝╚══╝ ╚═════╝ ╚═════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝
========================= ` + appVersion + ` =========================
`
var appVersion = `V1.0-230909`

var scanDirCount int64 = 0
var scanFileCount int64 = 0
var realBackupFileCount int64 = 0

func init() {
	// -conf, -c
	flag.StringVar(&configFilePath, "conf", "./webdavBackup.json", "config file path")
	flag.StringVar(&configFilePath, "c", "./webdavBackup.json", "config file path (shorthand)")

	// -mode, -m
	flag.StringVar(&mode, "mode", "backup", "backup mode: \"backup\" or \"mirror\" \n"+
		"\"backup\": In this mode, if a file is deleted or moved on the webdav server, \n"+
		"          the local copy of the file is still retained \n"+
		"\"mirror\": In this mode, if a file is deleted or moved on the webdav server, \n"+
		"          the corresponding local copy is also deleted, \n"+
		"          effectively mirroring the state of the server ")
	flag.StringVar(&mode, "m", "backup", "backup mode: \"backup\" or \"mirror\", (shorthand)")

	// -debugFlag, -d
	flag.BoolVar(&debugFlag, "debug", false, "show debug log")
	flag.BoolVar(&debugFlag, "d", false, "show debug log (shorthand)")
	// -versionFlag, -v
	flag.BoolVar(&versionFlag, "version", false, "show version msg")
	flag.BoolVar(&versionFlag, "v", false, "show version msg (shorthand)")

	flag.Parse()

	if versionFlag {
		fmt.Println(banner)
		os.Exit(0)
	}
	log.Init("WDBackup")
	if debugFlag {
		// info / err 默认打印
		// debugFlag 默认不打印
		log.SetLogLevel(log.LevelDebug)
	}
}

func main() {

	config := conf.LoadConfig(configFilePath)

	log.I("=================================================")
	log.I("app version: ", appVersion)
	log.I("run config path : ", configFilePath)
	log.I("run config mode : ", mode)
	log.I("backup webdav : ", config.BaseUrl)
	log.I("backup target dir: ", config.TargetDir)
	log.I("log level: ", log.GetLogLevel())
	log.I("=================================================")
	log.I()

	webdavClient = gowebdav.NewClient(config.BaseUrl, config.User, config.Password)
	log.D("start connect webdav")
	err := webdavClient.Connect()
	if err != nil {
		log.E("webdav connect error, " + err.Error())
		return
	} else {
		log.D("webdav connect success")
	}

	timeStartBackup := time.Now()
	// 开始备份
	//rootDir, _ := webdavClient.ReadDir("")
	backupDir("")

	duration := time.Now().Sub(timeStartBackup)
	log.I()
	log.I("=================================================")
	log.I("backup finish, use : " + formatDuration(duration))
	log.I("scan dir count: ", scanDirCount)
	log.I("scan file count: ", scanFileCount)
	log.I("real backup count: ", realBackupFileCount)
	log.I("=================================================", "\n\n")

}

func backupDir(davBasePath string) {
	log.D("start backup dir: " + davBasePath)
	scanDirCount++

	files, _ := webdavClient.ReadDir(filepath.ToSlash(davBasePath))

	if mode == string(ModeMirror) {
		localDirPath := filepath.Join(conf.GetTargetBackupRootDir(), davBasePath)
		LocalSyncDirFile(localDirPath, files)
	}

	// 直接搜索看看需要跳过多少文件
	needBackupDir := make([]os.FileInfo, 0)
	needBackupFile := make([]os.FileInfo, 0)

	var davFilePathTemp, localFilePathTemp string
	var skipFileCount = 0

	// 先备份文件，再往子目录备份
	for _, file := range files {
		davFilePathTemp = filepath.ToSlash(filepath.Join(davBasePath, file.Name()))
		localFilePathTemp = filepath.Join(conf.GetTargetBackupRootDir(), davFilePathTemp)

		if !file.IsDir() {
			scanFileCount++
			// 文件处理
			// 看看是否需要跳过文件
			// 判断是否需要复制
			if checkNeedToCopy(localFilePathTemp, file) {
				// 删除本地文件
				err := os.Remove(localFilePathTemp)
				if err != nil {
					log.E("delete local file error:", err)
				}
				needBackupFile = append(needBackupFile, file)
				realBackupFileCount++
			} else {
				log.D("skip copy file: " + davFilePathTemp)
				skipFileCount++
			}
		} else {
			// 文件夹处理
			needBackupDir = append(needBackupDir, file)
		}
	}
	if skipFileCount != 0 {
		if len(needBackupFile) == 0 && len(needBackupDir) == 0 {
			log.I("backup dir, skip hold dir: [", davBasePath, "]")
		} else {
			log.I("backup dir: [", davBasePath, "]",
				", skip ", skipFileCount, " files"+
					", backup file count: ", len(needBackupFile),
				", back dir count: ", len(needBackupDir))
		}
	}

	for _, fileInfo := range needBackupFile {
		// 创建文件
		LocalCopyFile(davBasePath, fileInfo)
	}

	for _, dirInfo := range needBackupDir {
		backupDir(filepath.Join(davBasePath, dirInfo.Name()))
	}

}

//------------------- 本机文件操作

func LocalSyncDirFile(localDirPath string, files []os.FileInfo) {
	// List files within local directory
	localFiles, err := os.ReadDir(localDirPath)
	if err != nil {
		log.E("Error reading local directory:", err)
		return
	}

	// Map remote files for easy lookup
	remoteFiles := make(map[string]bool)
	for _, file := range files {
		remoteFiles[file.Name()] = true
	}

	// Go through local files and delete if not present in remote files
	for _, file := range localFiles {
		_, existsInRemote := remoteFiles[file.Name()]
		if !existsInRemote {
			fullPath := filepath.Join(localDirPath, file.Name())
			if file.IsDir() {
				// Deletes the directory with all its contents if it's a directory
				err := os.RemoveAll(fullPath)
				if err != nil {
					log.E("MIRROR mode, delete local dir error:", fullPath, err)
				}
			} else {
				// Deletes the file directly if it's a file
				err := os.Remove(fullPath)
				if err != nil {
					log.E("MIRROR mode, delete local file error:", err)
				}
			}
			log.I("MIRROR mode, delete local path: " + fullPath)
		}
	}

}

func LocalCreateDir(localDirPath string) {
	if _, err := os.Stat(localDirPath); os.IsNotExist(err) {
		err := os.MkdirAll(localDirPath, 0755)
		if err != nil {
			log.E("create dir error: "+localDirPath, err.Error())
		}
	}
}

// LocalCopyFile
// fileInfo 是从 github.com/studio-b12/gowebdav 里面读取得到的
func LocalCopyFile(davBasePath string, fileInfo os.FileInfo) {
	davFilePath := filepath.ToSlash(filepath.Join(davBasePath, fileInfo.Name()))
	log.D("copyfile, dav file path: ", davFilePath)

	// Creating new tarFile with the same permissions as original file
	localFilePath := filepath.Join(conf.GetTargetBackupRootDir(), davFilePath)
	// 判断 dir 是否已经创建了
	LocalCreateDir(filepath.Dir(localFilePath))

	// Open original file
	sourceStream, err := webdavClient.ReadStream(davFilePath)
	if err != nil {
		log.E("copy webdav file error: ", davFilePath, ", can't read source file, err: ", err.Error())
		return
	}

	tarFile, err := os.OpenFile(localFilePath, os.O_RDWR|os.O_CREATE, fileInfo.Mode())
	if err != nil {
		sourceStream.Close()
		log.E("copy webdav file error: ", davFilePath, ", can't create target file, err: ", err.Error())
		return
	}
	defer tarFile.Close()

	start := time.Now()

	// Copy the bytes to destination from source
	bytesCopied, err := io.Copy(tarFile, sourceStream)
	end := time.Now()

	sourceStream.Close()
	if err != nil {
		log.E("copy webdav file error: ", davFilePath, ", io copy err: ", err.Error())
		return
	}
	duration := end.Sub(start)
	// calculate the speed in MB/s: size in bytes / duration in seconds / 1e6 to get MB
	speed := float64(bytesCopied) / duration.Seconds() / 1e6

	log.I("copy file, speed:",
		"[", fmt.Sprintf("%.2f Mb/s", speed), "], ",
		"[", davFilePath, "]", " => ", "[", localFilePath, "]")

	// Set the original file's information to the new file
	err = os.Chtimes(localFilePath, fileInfo.ModTime(), fileInfo.ModTime())
	if err != nil {
		log.E("copy webdav file error: ", davFilePath, " , set file ModTime err : ", err.Error())
		return
	}
	return
}

// checkNeedToCopy
// localFilePath 本地文件路径
// fileInfo 远程 webdav 上面文件的信息
// 如果本地文件大小与 webdav 文件大小不一致或者修改日期不一致，那么就用远程文件来替换掉本地文件
func checkNeedToCopy(localFilePath string, fileInfo os.FileInfo) bool {
	localFileInfo, err := os.Stat(localFilePath)
	//if err != nil {
	//	log.E(err)
	//	return false
	//}
	if os.IsNotExist(err) {
		// 本地文件不存在，需要复制远程文件
		return true
	} else if err != nil {
		log.E("check need to copy error : ", localFilePath, ", ", err.Error())
		return false
	}

	if localFileInfo.Size() != fileInfo.Size() {
		// 文件大小不一样，需要复制
		return true
	}

	if localFileInfo.ModTime().Unix() != fileInfo.ModTime().Unix() {
		// 文件修改日志不一样，需要复制
		return true
	}

	return false
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%d hours %d mins %d secs", h, m, s)
}
