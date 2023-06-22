package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Theme          string          `yaml:"theme"`
	Username       *Username       `yaml:"username"`
	Mobile         *Mobile         `yaml:"mobile"`
	TrackingNumber *TrackingNumber `yaml:"tracking_number"`
	AutoRefresh    *AutoRefresh    `yaml:"auto_refresh"`
}
type Username struct {
	Text string `yaml:"text"`
	Show bool   `yaml:"show"`
}
type Mobile struct {
	Text string `yaml:"text"`
	Show bool   `yaml:"show"`
}
type TrackingNumber struct {
	ShipId int  `yaml:"ship_id"`
	Show   bool `yaml:"show"`
}
type AutoRefresh struct {
	Interval int  `yaml:"interval"`
	Enable   bool `yaml:"enable"`
}

func InitConfig(path string) *Config {
	var _config *Config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("read:", err.Error())
		_config = &Config{}
	}
	err = yaml.Unmarshal(yamlFile, &_config)
	if err != nil {
		fmt.Println("Unmarshal:", err.Error())
		_config = &Config{}
	}
	if _config == nil {
		_config = &Config{}
	}
	if _config.Theme == "" {
		_config.Theme = "auto"
	}
	if _config.Username == nil {
		_config.Username = &Username{"李明", true}
	}
	if _config.Mobile == nil {
		_config.Mobile = &Mobile{"15612345678", true}
	}
	if _config.TrackingNumber == nil {
		_config.TrackingNumber = &TrackingNumber{1, true}
	}
	if _config.AutoRefresh == nil {
		_config.AutoRefresh = &AutoRefresh{3, false}
	}
	return _config
}

func fileIsExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetConfigFile() string {
	if xdg_config_home, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		config_file := filepath.Join(xdg_config_home, "/barcode/barcode.yaml")
		if isExist, _ := fileIsExist(config_file); isExist {
			return config_file
		}
	}
	if home, ok := user.Current(); ok == nil {
		config_file := filepath.Join(home.HomeDir, "/.config/barcode/barcode.yaml")
		if isExist, _ := fileIsExist(config_file); isExist {
			return config_file
		}
		config_file = filepath.Join(home.HomeDir, "/.barcode.yaml")
		if isExist, _ := fileIsExist(config_file); isExist {
			return config_file
		}
	}
	return ""
}
