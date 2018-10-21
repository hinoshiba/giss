package apicon

import (
	"fmt"
	"time"
	"bytes"
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"encoding/json"
)

type JsonToken struct {
	Name string `json:"name"`
	Sha1 string `json:"sha1"`
}

type IssueEdited struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	State  string     `json:"state"`
	User   IssueUser  `json:"user"`
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
	Color string `json:"color"`
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

var TokenName string = "giss"

type Gitea struct {
	Url string
	Repo string
	User string
	Token string
}

func NewGiteaCredent(url string) (Gitea, error) {
	var gitea Gitea
	gitea.Url = url
	return gitea, nil
}

func (self *Gitea) IsLogined() bool {
	return self.isLogined()
}

func (self *Gitea) isLogined() bool {
	if self.Token == "" {
		return false
	}
	if self.User == "" {
		return false
	}
	return true
}

func (self *Gitea) LoadCache(username, token, repo string) bool {
	return self.loadCache(username, token, repo)
}

func (self *Gitea) loadCache(username, token, repo string) bool {
	self.User = username
	self.Token = token
	self.Repo = repo
	return self.isLogined()
}

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

func (self *Gitea) DoCloseIssue(inum string) error {
	return self.doCloseIssue(inum)
}

func (self *Gitea) DoOpenIssue(inum string) error {
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
	if err := self.postIssue(inum, &issue); err != nil {
		return err
	}
	if old == issue.Update {
		fmt.Printf("not update\n")
		return nil
	}

	fmt.Printf("state updated : %s\n", issue.State)
	return nil
}


func (self *Gitea) PrintIssue(inum string, detailprint bool) error {
	if !self.isLogined() {
		return nil
	}
	return self.printIssue(inum, detailprint)
}

func (self *Gitea) printIssue(inum string, detailprint bool) error {
	issue, comments, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if issue.State == "" {
		fmt.Printf("undefined ticket: %s\n", inum)
		return nil
	}

	fmt.Printf(" [#%d] %s ( %s )\n",issue.Num, issue.Title, issue.User.Name)
	fmt.Printf(" Status   : %s\n", issue.State)
	fmt.Printf(" Updateat : %s\n", issue.Update)
	fmt.Printf("= body =================================================\n")
	fmt.Printf("%s\n",issue.Body)
	fmt.Printf("= comments =============================================\n")
	for _, comment := range comments {
		fmt.Printf(" [#%d] %s ( %s )\n",
			comment.Id, comment.Update, comment.User.Name)
		fmt.Printf("------------------------>\n")
		fmt.Printf("%s\n",comment.Body)
		fmt.Printf("------------------------------------------------\n")
	}
	return nil
}



func (self *Gitea) postIssue(inum string, issue *Issue) error {
	url := self.Url + "api/v1/repos/" + self.Repo + "/issues/" + inum

	ijson, err := json.Marshal(convIssueEdited(*issue))
	if err != nil {
		return err
	}
	iret, rcode, err := self.reqHttp("PATCH", url, []byte(ijson))
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
	return nil
}

func (self *Gitea) getIssue(num string) (Issue, []IssueComment, error) {
	var issue Issue
	var comments []IssueComment
	iurl := self.Url + "api/v1/repos/" + self.Repo + "/issues/" + num
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

func (self *Gitea) PrintIssues(limit int, withclose bool) error {
	if !self.isLogined() {
		return nil
	}
	return self.printIssues(limit, withclose)
}

func (self *Gitea) printIssues(limit int, withclose bool) error {

	issues, err := self.getIssues(withclose)
	if err != nil {
		return err
	}
	if len(issues) < 1 {
		return nil
	}

	for index, issue := range issues {
		if index >= limit {
			break
		}
		fmt.Printf(" %04d %s %-012s [ %6s / %-010s ] %s\n",
			issue.Num,
			issue.Update.Format("2006/1/2 15:04:05"),
			issue.User.Name,
			issue.State,
			issue.Milestone.Title,
			issue.Title,
		)
	}
	return nil
}

func (self *Gitea) getIssues(withclose bool) ([]Issue, error) {
	url := self.Url + "api/v1/repos/" + self.Repo + "/issues"
	if withclose {
		url = url + "?state=all"
	}
	bret, rcode, err := self.reqHttp("GET", url, nil)
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
	return issues, nil
}

func (self *Gitea) getDefinedToken(username, passwd string) (string, error) {
	url := self.Url + "api/v1/users/" + username + "/tokens"
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.SetBasicAuth(username, passwd)

	client := newClient()
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
	jsonStr := `{"name":"`+ TokenName + `"}`
    	req, err := http.NewRequest(
        	"POST",
        	url,
        	bytes.NewBuffer([]byte(jsonStr)),
    	)
	req.SetBasicAuth(username, passwd)
	req.Header.Set("Content-Type", "application/json")

	client := newClient()
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

func convIssueEdited(issue Issue) IssueEdited {
	var nissue IssueEdited

	nissue.Id = issue.Id
	nissue.Num = issue.Num
	nissue.Title = issue.Title
	nissue.Body = issue.Body
	nissue.State = issue.State
	nissue.User  = issue.User
	nissue.Assgin = issue.Assgin

	return nissue
}

func newClient() *http.Client {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
    }
    return &http.Client{Transport: tr}
}

func (self *Gitea) reqHttp(method, url string, param []byte ) ([]byte,
								int, error) {
    	req, err := http.NewRequest(
        	method,
        	url,
        	bytes.NewBuffer(param),
    	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token " + self.Token)

	client := newClient()
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

