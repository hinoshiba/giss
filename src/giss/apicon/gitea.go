package gitea

import (
	
)

var TokenName string = "giss"

type Gitea struct {
	url string
	username string
	token string
}

func (self *Gitea) Login(username, passwd) error {
}

func (self *Gitea) isLogined(curuser string) bool {
	if self.token == "" {
		return false
	}
	if self.username == "" {
		return false
	}
}

func (self *Gitea) login(username, passwd) error {
	token, err := getDefinedToken(username, passwd)
	if err != nil {
		return err
	}
	if token != "" {
		self.Token = token
		self.Username = username
		return nil
	}

	token. err := createReqToken(username, passwd)
	if err != nil {
		return err
	}
	if token != "" {
		self.Token = token
		self.Username = username
		return nil
	}

	return nil
}

func (self *Gitea) getDefinedToken (username, passwd string) (string, error) {
	url := "https://www.ds.i.hinoshiba.com/gitea/api/v1/users/s.k.noe/tokens"
	//url := self.url
	req, err := http.NewRequest(
		"GET",
		url
	)
	req.SetBasicAuth(username, string(passwd))

	client := newClient()
	resp, err := client.Do(req)
	if err != nil {
       		return err
    	}
    	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(bodyText))
}

func (self *Gitea) createReqToken(username, passwd) (string, error) {
	url := "https://www.ds.i.hinoshiba.com/gitea/api/v1/users/s.k.noe/tokens"
	//url := self.url
	jsonStr := `{"name":"giss"}`
    	req, err := http.NewRequest(
        	"POST",
        	url,
        	bytes.NewBuffer([]byte(jsonStr)),
    	)
	req.SetBasicAuth(user, string(pass))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := newClient()
	resp, err := client.Do(req)
	if err != nil {
       		return err
    	}
    	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	fmt.Printf(string(bodyText))
	//{"name":"giss","sha1":"cec219cd7697ad7ec29d8241490e2afe058eb3e5"}s.k.noe@m-cpu01:~/git/gitea/giss$
}

func newClient() *http.Client {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
    }
    return &http.Client{Transport: tr}
}
