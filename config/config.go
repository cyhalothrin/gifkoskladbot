package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var StoragePathFlag string

type Config struct {
	Token        string
	HostUsername string
	ChannelID    int64
	AllowedUsers []string
	StoragePath  string
}

func ReadConfig() (Config, error) {
	var conf Config
	err := viper.Unmarshal(&conf)

	if StoragePathFlag != "" {
		conf.StoragePath = StoragePathFlag
	}
	if conf.StoragePath == "" {
		return conf, errors.New("не указан файл базы данных")
	}

	conf.StoragePath, err = filepath.Abs(conf.StoragePath)
	if err != nil {
		err = fmt.Errorf("get absolute path of storage file: %w")
	}

	return conf, err
}

func GetExecPath() string {
	execPath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	execPath, err = filepath.Abs(filepath.Dir(execPath))
	if err != nil {
		panic(execPath)
	}

	return execPath + string(os.PathSeparator)
}
