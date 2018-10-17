package config

import (
	"os/user"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Body BodyConfig
	Mail MailConfig
}

type BodyConfig struct {
	Header string `toml:"header"`
	Futter string `toml:"futter"`
}

type MailConfig struct {
	To string `toml:"To"`
	Cc string `toml:"Cc"`
	Bcc string `toml:"Bcc"`
	From string `toml:"From"`

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
