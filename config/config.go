package config

import (
	"errors"
	"os"
	"strings"

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

	if StoragePathFlag == "" {
		conf.StoragePath = StoragePathFlag
	}
	if conf.StoragePath == "" {
		return conf, errors.New("не указан файл базы данных")
	}

	return conf, err
}

func GetExecPath() string {
	pathSep := string(os.PathSeparator)
	folders := strings.Split(os.Args[0], pathSep)

	return strings.Join(folders[:len(folders)-1], pathSep) + pathSep
}
