package gitea

import (
	"fmt"
	"time"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"giss/apicon/httpcl"
	"giss/apicon/issue"
)

type Gitea struct {
	url string
	repository string
	user string
	token string
}

type IssueEdited struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	State  string     `json:"state"`
	User   IssueUser  `json:"user"`
	Update time.Time  `json:"updated_at"`
	Assgin string     `json:"assignee"`
}

type Issue struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	Url    string     `json:"url"`
	State  string     `json:"state"`
//	Labels IssueLabel `json:"labels"`
	Milestone IssueMilestone `json:"milestone"`
	Update time.Time  `json:"updated_at"`
	User   IssueUser  `json:"user"`
	Assgin string     `json:"assignee"`
}

type IssueComment struct {
	Id     int64      `json:"id"`
	Body   string     `json:"body"`
	Update time.Time  `json:"updated_at"`
	User   IssueUser  `json:"user"`
}

type IssueLabel struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
//	Color string `json:"color"`
}

type IssueUser struct {
	Id    int64  `json:"id"`
	Name string  `json:"username"`
	Email string `json:"email"`
}

type IssueMilestone struct {
	Id     int64  `json:"id"`
	Title  string `json:"title"`
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

func (self *Gitea) GetIssues(withclose bool) ([]issue.Body, error) {
	return self.getIssues(withclose)
}

func (self *Gitea) getIssues(withclose bool) ([]issue.Body, error) {
	url := self.url + "api/v1/repos/" + self.repository + "/issues?"
	if withclose {
		url = url + "&state=all"
	}
	var p int = 1
	var ret []issue.Body
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

		var iss []issue.Body
		if err := json.Unmarshal(bret, &iss); err != nil {
			return nil, err
		}
		if len(iss) < 1 {
			break
		}
		p += 1
		for _, v := range iss {
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (self *Gitea) GetIssue(num string) (issue.Body, []issue.Comment, error) {
	return self.getIssue(num)
}

func (self *Gitea) getIssue(num string) (issue.Body, []issue.Comment, error) {
	var is issue.Body
	var comments []issue.Comment
	iurl := self.url + "api/v1/repos/" + self.repository + "/issues/" + num
	curl := iurl + "/comments"

	iret, rcode, err := self.reqHttp("GET", iurl, nil)
	if err != nil {
		return is, comments, err
	}
	if rcode != 200 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return is, comments, nil
	}
	cret, rcode, err := self.reqHttp("GET", curl, nil)
	if err != nil {
		return is, comments, err
	}
	if rcode != 200 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return is, comments, nil
	}

	if err := json.Unmarshal(iret, &is); err != nil {
		return is, comments, err
	}
	if err := json.Unmarshal(cret, &comments); err != nil {
		return is, comments, err
	}
	return is, comments, nil
}

func (self *Gitea) CreateIssue(ie issue.Edited) error {
	if !self.isLogined() {
		return nil
	}
	return self.createIssue(ie)
}

func (self *Gitea) createIssue(ie issue.Edited) error {
	if err := self.postIssue(&ie); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) AddIssueComment(inum string, comment string) error {
	if !self.isLogined() {
		return nil
	}
	return self.addIssueComment(inum, comment)
}

func (self *Gitea) addIssueComment(inum string, comment string) error {
	if err := self.httpReqComment("POST", inum, comment); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) ModifyIssue(inum string, ie issue.Edited) error {
	if !self.isLogined() {
		return nil
	}
	return self.modifyIssue(inum, ie)
}

func (self *Gitea) modifyIssue(inum string, ie issue.Edited) error {
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
	is, _, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if is.State == "" {
		fmt.Printf("undefined ticket: %s\n", inum)
		return nil
	}
	if is.State == state {
		fmt.Printf("this issue already state : %s\n", state)
		return nil
	}

	old := is.Update
	is.State = state
	eis := ConvIssueEdited(is)
	if err := self.updatePostIssue(inum, &eis); err != nil {
		return err
	}
	if old == is.Update {
		fmt.Printf("not update\n")
		return nil
	}

	fmt.Printf("state updated : %s\n", is.State)
	return nil
}

func (self *Gitea) postIssue(is *issue.Edited) error {
	return self.httpReqIssue("POST", "", is)
}

func (self *Gitea) updatePostIssue(inum string, is *issue.Edited) error {
	return self.httpReqIssue("PATCH", inum, is)
}

func (self *Gitea) httpReqComment(method string , inum string, body string) error {
	url := self.url + "api/v1/repos/" + self.repository +
						"/issues/" + inum + "/comments"
	json_str := `{"Body":"`+ body + `"}`
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

func (self *Gitea) httpReqIssue(method string , inum string, is *issue.Edited) error {
	url := self.url + "api/v1/repos/" + self.repository + "/issues/" + inum

	is.Update = time.Now()
	ijson, err := json.Marshal(*is)
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

	if err := json.Unmarshal(iret, &is); err != nil {
		return err
	}
	fmt.Printf("issue posted : #%v\n",is.Num)
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

	client, err := httpcl.NewClient()
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
var TokenName string = "giss"
type JsonToken struct {
	Name string `json:"name"`
	Sha1 string `json:"sha1"`
}
*/
func ConvIssueEdited(is issue.Body) issue.Edited {
	var nis issue.Edited

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.State = is.State
	nis.User  = is.User
//	nissue.Assgin = issue.Assgin

	return nis
}
