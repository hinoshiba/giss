package config

import (
	"os/user"
	"path/filepath"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Report RepoConfig
	Mail MailConfig
	GitDefault  GitDefaultConfig
	Server map[string]GitServerConfig
	Giss GissConfig
}

type GitServerConfig struct {
	Url	string `toml:URL`
	Type	string `toml:Type`
	Repos	[]string `toml:Repos`
	AutoLogin bool `toml:AutoLogin`
	User	string `toml:User`
	Token	string `toml:Token`
}

type RepoConfig struct {
	Header string `toml:"Header"`
	Futter string `toml:"Futter"`
	TargetRepo []string `toml:"targetRepository"`
}

type MailConfig struct {
	To []string `toml:"To"`
	Header []string `toml:"Header"`
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

	cpath := filepath.Join(homeDir, Fname)
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

func GetAlias(url string, conf map[string]GitServerConfig) string {
	for s, v := range conf {
		if url == v.Url {
			return s
		}
	}
	return ""
}

func IsDefinedCred(alias string, conf map[string]GitServerConfig) bool {
	return conf[alias].User != "" && conf[alias].Token != ""
}
