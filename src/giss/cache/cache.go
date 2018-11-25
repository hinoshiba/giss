package cache

import (
	"os"
	"os/user"
	"io/ioutil"
	"path/filepath"
	"bufio"
	"strings"
)

type Cache struct {
	Token string
	User string
	Repo string
	Url string
	Alias string
	CacheDir string
	TmpDir string
	HomeDir string
}

func LoadCaches() (Cache, error) {
	var c Cache
	homeDir, err := getHomeDir()
	if err != nil {
		return c, err
	}
	return loadCaches(homeDir)
}

func (self *Cache)SaveCurrentGit(alias string, url string, repo string) error {
	return self.saveCurrentGit(alias, url, repo)
}

func (self *Cache)SaveCred(username string, token string) error {
	return self.saveCred(username, token)
}

func (self *Cache)saveCurrentGit(alias string, url string, repo string) error {
	target := alias + "," + url + "," + repo
	path := self.CacheDir + "/.currentgit"

	if err := writeParam(path, target); err != nil {
		return err
	}
	return nil
}

func (self *Cache)saveCred(username string, token string) error {
	cred := username + "," + token
	path := self.CacheDir + "/.cred"

	if err := writeParam(path, cred); err != nil {
		return err
	}
	return nil
}

func loadCaches(dir string) (Cache, error) {
	var cache Cache
	fpath, err := filepath.Abs(dir)
	if err != nil {
		return cache, err
	}
	cache.HomeDir = fpath

	cdir := fpath + "/.giss/"
	if err := checkCacheDir(fpath); err != nil {
		return cache, err
	}
	cache.CacheDir = cdir

	cfile := cdir + "/.cred"
	fcreds, err := loadParam(cfile)
	if err != nil {
		return cache, err
	}
	creds := strings.Split(fcreds, ",")
	if len(creds) == 2 {
		cache.User = creds[0]
		cache.Token = creds[1]
	}

	cafile := cdir + "/.currentgit"
	scurgits, err := loadParam(cafile)
	if err != nil {
		return cache, err
	}
	curgits := strings.Split(scurgits, ",")
	if len(curgits) == 3 {
		cache.Alias = curgits[0]
		cache.Url = curgits[1]
		cache.Repo = curgits[2]
	}

	t, err := ioutil.TempDir(cdir,"giss-cache-")
	if err != nil {
		return cache, err
	}
	os.RemoveAll(cache.TmpDir)
	cache.TmpDir = t

	return cache, nil
}

func checkCacheDir(cdir string) error {

	if _, err := os.Stat(cdir); err != nil {
		if ferr := os.Mkdir(cdir, 0770); ferr != nil {
			return ferr
		}
	}
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
