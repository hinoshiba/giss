package cache

import (
	"os"
	"os/user"
	"io/ioutil"
	"path/filepath"
	"bufio"
)

var Token string
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

func loadCaches(dir string) error {
	fpath, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}

	if err := initCacheDir(fpath); err != nil {
		return err
	}

	tfile := CacheDir + "/.token"
	token, err := loadParam(tfile)
	if err != nil {
		return err
	}
	Token = token

	cfile := CacheDir + "/.currentgit"
	currentgit, err := loadParam(cfile)
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

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
