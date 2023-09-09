package conf

import (
	"encoding/json"
	"github.com/Ericwyn/webdav-backup/log"
	"io/ioutil"
	"os"
)

type WDBackupConfig struct {
	BaseUrl   string
	User      string
	Password  string
	TargetDir string // 完整路径
}

var confNow *WDBackupConfig

func LoadConfig(configJsonFilePath string) *WDBackupConfig {
	file, err := os.ReadFile(configJsonFilePath)
	if err != nil {
		log.E("can't read the default config json file")
		writeDefaultConfig(configJsonFilePath)
		log.I("finish write default config json file in : " + configJsonFilePath)
		log.I("please edit the config file and restart app")
		os.Exit(-1)
	}

	confNow = &WDBackupConfig{}
	err = json.Unmarshal(file, confNow)
	if err != nil {
		log.E("can;t parse the config json file, ", err)
	}
	return confNow
}

func writeDefaultConfig(configJsonFile string) {
	defaultConfig := WDBackupConfig{
		BaseUrl:   "https://test.dav",
		User:      "user-account",
		Password:  "user-password",
		TargetDir: "target-local-backup-dir-full-path",
	}

	file, err := json.MarshalIndent(defaultConfig, "", " ")
	if err != nil {
		log.E("无法编写默认配置: ", err)
	}

	err = ioutil.WriteFile(configJsonFile, file, 0644)
	if err != nil {
		log.E("无法写入默认配置: ", err)
	}
}

func GetTargetBackupRootDir() string {
	return confNow.TargetDir
}
