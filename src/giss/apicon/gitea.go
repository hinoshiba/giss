package apicon

import (
	"fmt"
	"time"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"giss/cache"
)

type JsonToken struct {
	Name string `json:"name"`
	Sha1 string `json:"sha1"`
}

var TokenName string = "giss"

type Gitea struct {
	url string
	repository string
	user string
	token string
}

func (self *Gitea) LoadCache(c cache.Cache) bool {
	return self.loadCache(c)
}

func (self *Gitea) loadCache(c cache.Cache) bool {
	self.user = c.User
	self.token = c.Token
	self.repository = c.Repo
	self.url = c.Url
	return self.isLogined()
}

func (self *Gitea) GetUrl() string {
	return self.url
}

func (self *Gitea) SetUrl(url string) {
	self.setUrl(url)
}

func (self *Gitea) setUrl(url string) {
	self.url = url
}

func (self *Gitea) GetRepositoryName() string {
	return self.repository
}

func (self *Gitea) SetRepositoryName(repo string) {
	self.setRepositoryName(repo)
}

func (self *Gitea) setRepositoryName(repo string) {
	self.repository = repo
}

func (self *Gitea) GetUsername() string {
	return self.user
}

func (self *Gitea) SetUsername(user string) {
	self.setUsername(user)
}

func (self *Gitea) setUsername(user string) {
	self.user = user
}

func (self *Gitea) GetToken() string {
	return self.token
}

func (self *Gitea) SetToken(token string) {
	self.setToken(token)
}

func (self *Gitea) setToken(token string) {
	self.token = token
}

func (self *Gitea) IsLogined() bool {
	return self.isLogined()
}

func (self *Gitea) isLogined() bool {
	if self.token == "" {
		fmt.Printf("not login\n")
		return false
	}
	if self.user == "" {
		fmt.Printf("not login\n")
		return false
	}
	return true
}

func (self *Gitea) GetIssues(withclose bool) ([]Issue, error) {
	return self.getIssues(withclose)
}

func (self *Gitea) getIssues(withclose bool) ([]Issue, error) {
	url := self.url + "api/v1/repos/" + self.repository + "/issues?"
	if withclose {
		url = url + "&state=all"
	}
	var p int = 1
	var ret []Issue
	for {
		u := url + "&page=" + fmt.Sprintf("%v",p)
		bret, rcode, err := self.reqHttp("GET", u, nil)
		if err != nil {
			return nil, err
		}
		if rcode != 200 {
			fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
			return nil, nil
		}

		var issues []Issue
		if err := json.Unmarshal(bret, &issues); err != nil {
			return nil, err
		}
		if len(issues) < 1 {
			break
		}
		p += 1
		for _, v := range issues {
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (self *Gitea) GetIssue(num string) (Issue, []IssueComment, error) {
	return self.getIssue(num)
}

func (self *Gitea) getIssue(num string) (Issue, []IssueComment, error) {
	var issue Issue
	var comments []IssueComment
	iurl := self.url + "api/v1/repos/" + self.repository + "/issues/" + num
	curl := iurl + "/comments"

	iret, rcode, err := self.reqHttp("GET", iurl, nil)
	if err != nil {
		return issue, comments, err
	}
	if rcode != 200 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return issue, comments, nil
	}
	cret, rcode, err := self.reqHttp("GET", curl, nil)
	if err != nil {
		return issue, comments, err
	}
	if rcode != 200 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return issue, comments, nil
	}

	if err := json.Unmarshal(iret, &issue); err != nil {
		return issue, comments, err
	}
	if err := json.Unmarshal(cret, &comments); err != nil {
		return issue, comments, err
	}
	return issue, comments, nil
}

func (self *Gitea) CreateIssue(ie IssueEdited) error {
	if !self.isLogined() {
		return nil
	}
	return self.createIssue(ie)
}

func (self *Gitea) createIssue(ie IssueEdited) error {
	if err := self.postIssue(&ie); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) AddIssueComment(inum string, comment []byte) error {
	if !self.isLogined() {
		return nil
	}
	return self.addIssueComment(inum, comment)
}

func (self *Gitea) addIssueComment(inum string, comment []byte) error {
	if err := self.httpReqComment("POST", inum, string(comment)); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) ModifyIssue(inum string, ie IssueEdited) error {
	if !self.isLogined() {
		return nil
	}
	return self.modifyIssue(inum, ie)
}

func (self *Gitea) modifyIssue(inum string, ie IssueEdited) error {
	if err := self.updatePostIssue(inum, &ie); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) DoCloseIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doCloseIssue(inum)
}

func (self *Gitea) DoOpenIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doOpenIssue(inum)
}

func (self *Gitea) doCloseIssue(inum string) error {
	if err := self.toggleIssueState(inum, "closed"); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) doOpenIssue(inum string) error {
	if err := self.toggleIssueState(inum, "open"); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) toggleIssueState(inum string, state string) error {
	if state != "open" && state != "closed" {
		fmt.Printf("unknown state :%s\n", state)
		return nil
	}
	issue, _, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if issue.State == "" {
		fmt.Printf("undefined ticket: %s\n", inum)
		return nil
	}
	if issue.State == state {
		fmt.Printf("this issue already state : %s\n", state)
		return nil
	}

	old := issue.Update
	issue.State = state
	eissue := ConvIssueEdited(issue)
	if err := self.updatePostIssue(inum, &eissue); err != nil {
		return err
	}
	if old == issue.Update {
		fmt.Printf("not update\n")
		return nil
	}

	fmt.Printf("state updated : %s\n", issue.State)
	return nil
}

func (self *Gitea) postIssue(issue *IssueEdited) error {
	return self.httpReqIssue("POST", "", issue)
}

func (self *Gitea) updatePostIssue(inum string, issue *IssueEdited) error {
	return self.httpReqIssue("PATCH", inum, issue)
}

func (self *Gitea) httpReqComment(method string , inum string, body string) error {
	url := self.url + "api/v1/repos/" + self.repository +
						"/issues/" + inum + "/comments"
	json_str := `{"Body":"`+ lf2Esclf(onlyLF(body)) + `"}`
	_, rcode, err := self.reqHttp(method, url, []byte(json_str))
	if err != nil {
		return err
	}
	if rcode != 201 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return nil
	}
	fmt.Printf("comment added : #%v\n", inum)
	return nil
}

func (self *Gitea) httpReqIssue(method string , inum string, issue *IssueEdited) error {
	url := self.url + "api/v1/repos/" + self.repository + "/issues/" + inum

	issue.Update = time.Now()
	issue.Title = lf2space(onlyLF(issue.Title))
	issue.Body = onlyLF(issue.Body)
	ijson, err := json.Marshal(*issue)
	if err != nil {
		return err
	}
	iret, rcode, err := self.reqHttp(method, url, []byte(ijson))
	if err != nil {
		return err
	}
	if rcode != 201 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return nil
	}

	if err := json.Unmarshal(iret, &issue); err != nil {
		return err
	}
	fmt.Printf("issue posted : #%v\n",issue.Num)
	return nil
}

func (self *Gitea) reqHttp(method, url string, param []byte ) ([]byte,
								int, error) {
    	req, err := http.NewRequest(
        	method,
        	url,
        	bytes.NewBuffer(param),
    	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token " + self.token)

	client, err := newClient()
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
       		return nil, 0, err
    	}
    	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return bytes, resp.StatusCode, nil
}
/*
func (self *Gitea) Login(username, passwd string) error {
	return self.login(username, passwd)
}

func (self *Gitea) login(username, passwd string) error {
	curtoken, err := self.getDefinedToken(username, passwd)
	if err != nil {
		return err
	}
	if curtoken != "" {
		self.Token = curtoken
		self.User = username
		fmt.Printf("Login success !!\n")
		return nil
	}

	newtoken, err := self.createReqToken(username, passwd)
	if err != nil {
		return err
	}
	if newtoken != "" {
		self.Token = newtoken
		self.User = username
		fmt.Printf("Login success !!\n")
		return nil
	}

	if !self.isLogined() {
		fmt.Printf("Login Failed...\n")
	}
	return nil
}

func (self *Gitea) getDefinedToken(username, passwd string) (string, error) {
	url := self.Url + "api/v1/users/" + username + "/tokens"
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.SetBasicAuth(username, passwd)

	client, err := newClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
       		return "", err
    	}
    	defer resp.Body.Close()

	var token string
	var jtokens []JsonToken
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(bytes, &jtokens); err != nil {
		return "", err
	}
	for _, t := range jtokens {
		if TokenName == t.Name {
			token = t.Sha1
			break
		}
	}
	return token, nil
}

func (self *Gitea) createReqToken(username, passwd string) (string, error) {
	url := self.Url + "api/v1/users/" + username + "/tokens"
	json_str := `{"name":"`+ TokenName + `"}`
    	req, err := http.NewRequest(
        	"POST",
        	url,
        	bytes.NewBuffer([]byte(json_str)),
    	)
	req.SetBasicAuth(username, passwd)
	req.Header.Set("Content-Type", "application/json")

	client, err := newClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
       		return "", err
    	}
    	defer resp.Body.Close()

	var token string
	var jtoken JsonToken
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(bytes, &jtoken); err != nil {
		return "", err
	}
	if TokenName == jtoken.Name {
		token = jtoken.Sha1
	}
	return token, nil
}
*/
