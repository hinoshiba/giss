package cache

import (
	"os"
	"os/user"
	"io/ioutil"
	"path/filepath"
	"bufio"
	"strings"
)

var Token string
var User string
var CurrentGit string
var CacheDir string
var TmpDir string

func LoadCaches() error {
	homeDir, err := getHomeDir()
	if err != nil {
		return err
	}
	return loadCaches(homeDir)
}

func SaveCurrentGit(giturl string) error {
	return saveCurrentGit(giturl)
}

func SaveCred(username string, token string) error {
	return saveCred(username, token)
}

func saveCurrentGit(giturl string) error {
	path := CacheDir + "/.currentgit"

	if err := writeParam(path, giturl); err != nil {
		return err
	}
	if err := loadCaches(CacheDir); err != nil {
		return err
	}
	return nil
}

func saveCred(username string, token string) error {
	cred := username + "," + token
	path := CacheDir + "/.cred"

	if err := writeParam(path, cred); err != nil {
		return err
	}
	if err := loadCaches(CacheDir); err != nil {
		return err
	}
	return nil
}

func loadCaches(dir string) error {
	fpath, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}

	if CacheDir == "" {
		if err := initCacheDir(fpath); err != nil {
			return err
		}
	}

	cfile := CacheDir + "/.cred"
	fcreds, err := loadParam(cfile)
	if err != nil {
		return err
	}
	creds := strings.Split(fcreds, ",")
	if len(creds) == 2 {
		User = creds[0]
		Token = creds[1]
	}

	cafile := CacheDir + "/.currentgit"
	currentgit, err := loadParam(cafile)
	if err != nil {
		return err
	}
	CurrentGit = currentgit

	t, err := ioutil.TempDir(CacheDir,"giss-cache-")
	if err != nil {
		return err
	}
	TmpDir = t

	return nil
}

func initCacheDir(dir string) error {
	cdir := dir + "/.giss/"

	if _, err := os.Stat(cdir); err != nil {
		if ferr := os.Mkdir(cdir, 0770); ferr != nil {
			return ferr
		}
	}
	CacheDir = cdir
	return nil
}

func loadParam(pfile string) (string, error) {
	fpath, err := filepath.Abs(pfile)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(fpath); err != nil {
		return "", nil
	}

	param, err := read1stLine(fpath)
	if err != nil {
		return "", err
	}
	return param, nil
}

func read1stLine(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var ret string
	s := bufio.NewScanner(f)
	if s.Scan() {
		ret = s.Text()
	}
	return ret, nil
}

func writeParam(pfile string, param string) error {
	f, err := os.Create(pfile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(param)); err != nil {
		return err
	}
	return nil
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
