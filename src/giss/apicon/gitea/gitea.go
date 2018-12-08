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

type iIssueE struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	State  string     `json:"state"`
	User   iIUser  `json:"user"`
	Update time.Time  `json:"updated_at"`
	Assgin string     `json:"assignee"`
}

type iIssue struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	Url    string     `json:"url"`
	State  string     `json:"state"`
//	Labels IssueLabel `json:"labels"`
	Milestone iIMilestone `json:"milestone"`
	Update time.Time  `json:"updated_at"`
	User   iIUser `json:"user"`
	Assgin string     `json:"assignee"`
}

type iIComment struct {
	Id     int64      `json:"id"`
	Body   string     `json:"body"`
	Update time.Time  `json:"updated_at"`
	User   iIUser     `json:"user"`
}

type iILabel struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type iIUser struct {
	Id    int64  `json:"id"`
	Name string  `json:"username"`
	Email string `json:"email"`
}

type iIMilestone struct {
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
	var iss []issue.Body

	i_iss, err := self.getIssues(withclose)
	if err != nil {
		return iss, err
	}

	for _, v := range i_iss {
		iss = append(iss, iIssue2Issue(v))
	}
	return iss, nil
}

func (self *Gitea) getIssues(withclose bool) ([]iIssue, error) {
	url := self.url + "api/v1/repos/" + self.repository + "/issues?"
	if withclose {
		url = url + "&state=all"
	}
	var p int = 1
	var ret []iIssue
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

		var iss []iIssue
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
	var is issue.Body
	var coms []issue.Comment

	i_is, i_icoms, err := self.getIssue(num)
	if err != nil {
		return is, coms, err
	}

	is = iIssue2Issue(i_is)
	for _, i_com := range i_icoms {
		coms = append(coms, iIComment2IssueComment(i_com))
	}
	return is, coms, nil
}

func (self *Gitea) getIssue(num string) (iIssue, []iIComment, error) {
	var is iIssue
	var comments []iIComment
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

func (self *Gitea) CreateIssue(is issue.Body) error {
	i_is := Issue2iIssue(is)
	i_ise := iIssue2iIssueE(i_is)
	return self.createIssue(i_ise)
}

func (self *Gitea) createIssue(ise iIssueE)  error {
	if err := self.postIssue(&ise); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) AddIssueComment(inum string, comment string) error {
	return self.addIssueComment(inum, comment)
}

func (self *Gitea) addIssueComment(inum string, comment string) error {
	if err := self.httpReqComment("POST", inum, comment); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) ModifyIssue(inum string, is issue.Body) error {
	i_is := Issue2iIssue(is)
	i_ise := iIssue2iIssueE(i_is)
	return self.modifyIssue(inum, i_ise)
}

func (self *Gitea) modifyIssue(inum string, ise iIssueE) error {
	if err := self.updatePostIssue(inum, &ise); err != nil {
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
	eis := iIssue2iIssueE(is)
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

func (self *Gitea) postIssue(ise *iIssueE) error {
	return self.httpReqIssue("POST", "", ise)
}

func (self *Gitea) updatePostIssue(inum string, ise *iIssueE) error {
	return self.httpReqIssue("PATCH", inum, ise)
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

func (self *Gitea) httpReqIssue(method string , inum string, ise *iIssueE) error {
	url := self.url + "api/v1/repos/" + self.repository + "/issues/" + inum

	ise.Update = time.Now()
	ijson, err := json.Marshal(*ise)
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

	if err := json.Unmarshal(iret, &ise); err != nil {
		return err
	}
	fmt.Printf("issue posted : #%v\n",ise.Num)
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
func iIssue2Issue(is iIssue) issue.Body {
	var nis issue.Body

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.Url = is.Url
	nis.State = is.State
//	nis.Label = iILabel2IssueLabel(is.Label)
	nis.Milestone = iIMilestone2IssueMilestone(is.Milestone)
	nis.Update = is.Update
	nis.User = iIUser2IssueUser(is.User)
	nis.Assgin = is.Assgin

	return nis
}

func iIUser2IssueUser(user iIUser) issue.User {
	var nuser issue.User

	nuser.Id = user.Id
	nuser.Name = user.Name
	nuser.Email = user.Email
	return nuser
}

func iILabel2IssueLabel(label iILabel) issue.Label {
	var nlabel issue.Label

	nlabel.Id = label.Id
	nlabel.Name = label.Name
//	nlabel.Color = label.Color
	return nlabel
}

func iIMilestone2IssueMilestone(mi iIMilestone) issue.Milestone {
	var nmi issue.Milestone

	nmi.Id = mi.Id
	nmi.Title = mi.Title
	return nmi
}

func iIComment2IssueComment(com iIComment) issue.Comment {
	var ncom issue.Comment

	ncom.Id = com.Id
	ncom.Body = com.Body
	ncom.Update = com.Update
	ncom.User = iIUser2IssueUser(com.User)
	return ncom
}

func Issue2iIssue(is issue.Body) iIssue {
	var nis iIssue

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.Url = is.Url
	nis.State = is.State
//	nis.Label = IssueLabel2iILabel(is.Label)
	nis.Milestone = IssueMilestone2iIMilestone(is.Milestone)
	nis.Update = is.Update
	nis.User = IssueUser2iIUser(is.User)
	nis.Assgin = is.Assgin

	return nis
}

func IssueUser2iIUser(user issue.User) iIUser {
	var nuser iIUser

	nuser.Id = user.Id
	nuser.Name = user.Name
	nuser.Email = user.Email
	return nuser
}

func IssueLabel2iILabel(label issue.Label) iILabel {
	var nlabel iILabel

	nlabel.Id = label.Id
	nlabel.Name = label.Name
//	nlabel.Color = label.Color
	return nlabel
}

func IssueMilestone2iIMilestone(mi issue.Milestone) iIMilestone {
	var nmi iIMilestone

	nmi.Id = mi.Id
	nmi.Title = mi.Title
	return nmi
}

func iIssue2iIssueE(is iIssue) iIssueE {
	var nis iIssueE

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.State = is.State
	nis.User  = is.User
	return nis
}
