package conf

import (
	"os/user"
	"path/filepath"
	"github.com/BurntSushi/toml"
)

const (
	Fname string = ".gissrc"
)

type Conf struct {
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

func LoadUserConfig() (Conf, error) {
	var conf Conf
	homeDir, err := getHomeDir()
	if err != nil {
		return conf, err
	}

	cpath := filepath.Join(homeDir, Fname)
	return loadConfig(cpath)
}

func loadConfig(path string) (Conf, error) {
	var conf Conf

	if _, err := toml.DecodeFile(path, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func (self *Conf) GetAlias(url string) string {
	for s, v := range self.Server {
		if url == v.Url {
			return s
		}
	}
	return ""
}

func (self *Conf) IsDefinedCred(alias string) bool {
	return self.Server[alias].User != "" && self.Server[alias].Token != ""
}
