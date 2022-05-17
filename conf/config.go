package conf

import (
	"github.com/gogoods/x/file"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"sync"

)

type ProxyConfig struct {
	Alias string `yaml:"alias"`
	Enabled bool `yaml:"enabled"`
	Listen string `yaml:"listen"`
	Mysql string `yaml:"mysql"`
}

type GlobalConfig struct {
	GUI string `yaml:"gui"`
	Proxies    []*ProxyConfig    `yaml:"proxies"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		cfg = "./conf/config.yaml"
		//log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = yaml.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	configLock.Lock()
	defer configLock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}

func IsFileExist(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}