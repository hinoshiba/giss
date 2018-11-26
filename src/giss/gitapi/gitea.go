package gitapi

import (
	"fmt"
	"time"
	"bytes"
	"strings"
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
var ReportNewTag string = "+"
var ReportTaskTag string = "-"

type Gitea struct {
	Url string
	Repo string
	User string
	Token string
}

func (self *Gitea) GetUrl() string {
	return self.Url
}

func (self *Gitea) GetRepo() string {
	return self.Repo
}

func (self *Gitea) SetRepo(repo string) {
	self.setRepo(repo)
}

func (self *Gitea) setRepo(repo string) {
	self.Repo = repo
}


func (self *Gitea) GetUser() string {
	return self.User
}

func (self *Gitea) GetToken() string {
	return self.Token
}

func (self *Gitea) IsLogined() bool {
	return self.isLogined()
}

func (self *Gitea) isLogined() bool {
	if self.Token == "" {
		fmt.Printf("not login\n")
		return false
	}
	if self.User == "" {
		fmt.Printf("not login\n")
		return false
	}
	return true
}

func (self *Gitea) LoadCache(c cache.Cache) bool {
	return self.loadCache(c)
}

func (self *Gitea) loadCache(c cache.Cache) bool {
	self.User = c.User
	self.Token = c.Token
	self.Repo = c.Repo
	self.Url = c.Url
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

func (self *Gitea) CreateIssue() error {
	if !self.isLogined() {
		return nil
	}
	return self.createIssue()
}

func (self *Gitea) createIssue() error {
	var issue Issue
	if ok, err := editIssue(&issue, true); !ok {
		return err
	}
	if err := self.createPostIssue(&issue); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) ModifyIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.modifyIssue(inum)
}

func (self *Gitea) modifyIssue(inum string) error {
	issue, _, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if issue.State == "" {
		fmt.Printf("undefined ticket: %s\n", inum)
		return nil
	}

	if ok, err := editIssue(&issue, false); !ok {
		return err
	}
	if err := self.updatePostIssue(inum, &issue); err != nil {
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
	if err := self.httpComment("POST", inum, string(comment)); err != nil {
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
	if err := self.updatePostIssue(inum, &issue); err != nil {
		return err
	}
	if old == issue.Update {
		fmt.Printf("not update\n")
		return nil
	}

	fmt.Printf("state updated : %s\n", issue.State)
	return nil
}

func (self *Gitea) GetIssue(num string) (Issue, []IssueComment, error) {
	return self.getIssue(num)
}


func (self *Gitea) updatePostIssue(inum string, issue *Issue) error {
	return self.httpIssue("PATCH", inum, issue)
}

func (self *Gitea) createPostIssue(issue *Issue) error {
	return self.httpIssue("POST", "", issue)
}

func (self *Gitea) httpIssue(method string , inum string, issue *Issue) error {
	url := self.Url + "api/v1/repos/" + self.Repo + "/issues/" + inum

	issue.Update = time.Now()
	issue.Title = lf2space(onlyLF(issue.Title))
	issue.Body = onlyLF(issue.Body)
	ijson, err := json.Marshal(convIssueEdited(*issue))
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

func (self *Gitea) httpComment(method string , inum string, body string) error {
	url := self.Url + "api/v1/repos/" + self.Repo +
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

func (self *Gitea) GetIssues(withclose bool) ([]Issue, error) {
	return self.getIssues(withclose)
}

func (self *Gitea) ReportIssues(now time.Time) (map[string]string, error) {
	iss, err := self.getIssues(true)
	if err != nil {
		return nil, err
	}
	if len(iss) < 1 {
		return nil, nil
	}

	ret := make(map[string]string)
	newtag := dayAgo(now, -6)
	limit := dayAgo(now, -14)
	for _, is := range iss {
		if is.Update.Unix() < limit.Unix() && is.State == "closed" {
			continue
		}

		time.Sleep(50 * time.Millisecond)
		ir, err := self.reportIssue(newtag, &is)
		if err != nil {
			return nil, err
		}
		if is.Milestone.Title == "" {
			ret["none"] += ir
			continue
		}
		ret[is.Milestone.Title] += ir
	}
	return ret, nil
}

func (self *Gitea) reportIssue(newtag time.Time, is *Issue) (string, error) {
	ir := "  - "
	if is.Update.Unix() >= newtag.Unix() {
		ir = "+ - "
	}
	if is.State == "closed" {
		ir += "[closed] "
	}
	ir += fmt.Sprintf("#%v ",is.Num) + lf2space(onlyLF(is.Title)) + "\n"
	for _, row := range strings.Split(onlyLF(is.Body),"\n") {
		ir += makeWithin80c(false,6 ,row)
	}

	_, coms, err := self.getIssue(fmt.Sprintf("%v",is.Num))
	if err != nil {
		return "", err
	}
	for _, com := range coms {
		cr := "    -  "
		if com.Update.Unix() >= newtag.Unix() {
		cr = "+   -  "
		}
		for i, row := range strings.Split(onlyLF(com.Body),"\n") {
		if i == 0 {
			cr += makeWithin80c(true, 5, row)
			continue
		}
		cr += makeWithin80c(false, 5, row)
		}
		ir += cr
	}
	return ir, nil
}

func (self *Gitea) getIssues(withclose bool) ([]Issue, error) {
	url := self.Url + "api/v1/repos/" + self.Repo + "/issues?"
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



func (self *Gitea) reqHttp(method, url string, param []byte ) ([]byte,
								int, error) {
    	req, err := http.NewRequest(
        	method,
        	url,
        	bytes.NewBuffer(param),
    	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token " + self.Token)

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

