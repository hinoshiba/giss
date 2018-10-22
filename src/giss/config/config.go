package config

import (
	"os/user"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Report RepoConfig
	Mail MailConfig
	GitDefault  GitDefaultConfig
	Giss GissConfig
}

type RepoConfig struct {
	Header string `toml:"Header"`
	Futter string `toml:"Futter"`
	TargetRepo []string `toml:"targetRepository"`
}

type MailConfig struct {
	To []string `toml:"To"`
	From string `toml:"From"`
	Subject string `toml:"Subject"`
	Mta string `toml:"MTA"`
	Port int64 `toml:"Port"`

}

type GitDefaultConfig struct {
	Url string `toml:"URL"`
}

type GissConfig struct {
	Editor string `toml:"editor"`
}

var Rc Config
var Fname string = ".gissrc"

func LoadUserConfig() error {
	homeDir, err := getHomeDir()
	if err != nil {
		return err
	}

	cpath := homeDir + "/" + Fname
	return loadConfig(cpath)
}

func loadConfig(path string) error {
	var config Config

	if _, err := toml.DecodeFile(path, &config); err != nil {
		return err
	}

	Rc = config
	return nil
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
