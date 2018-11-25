package gitapi

import (
	"giss/cache"
	"net/http"
	"strings"
	"io/ioutil"
	"encoding/json"
	"time"
	"fmt"
	"bytes"
)

type Github struct {
	url string
	repo string
	user string
	token string
}

type JsonSignin struct {
	Login string `json:"login"`
}

func (self *Github) GetRepo() string {
	return self.repo
}

func (self *Github) GetUrl() string {
	return self.url
}

func (self *Github) GetUser() string {
	return self.user
}

func (self *Github) GetToken() string {
	return self.token
}

func (self *Github) SetRepo(repo string) {
	self.repo = repo
}

func (self *Github) IsLogined() bool {
	return self.isLogined()
}

func (self *Github) isLogined() bool {
	if self.token == "" {
		fmt.Printf("empty token.\n")
		return false
	}
	if self.user == "" {
		fmt.Printf("empty username.\n")
		return false
	}
	return true
}

func (self *Github) LoadCache(c cache.Cache) bool {
	return self.loadCache(c)
}

func (self *Github) loadCache(c cache.Cache) bool {
	self.user = c.User
	self.token = c.Token
	self.repo = c.Repo
	self.url = c.Url
	return self.isLogined()
}

func (self *Github) Login(username, passwd string) error {
	return self.login(username, passwd)
}

func (self *Github) login(username, passwd string) error {
	ok, err := self.checkUserDefined(username, passwd)
	if err != nil {
		fmt.Printf("bad credentials. username : %s\n", username)
		return err
	}
	if !ok {
		fmt.Printf("bad credentials. username : %s\n", username)
	}

	fmt.Printf("login success !!.\n But can't autoload your credential. So Please get your token from Github web ui& adding ~/.gissrc. \nGithub : [https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/]\n")
	return nil
}

func (self *Github) checkUserDefined(username, passwd string) (bool, error) {
	url := self.url + "/users/" + username + "/tokens"
	url = self.url + "users/" + username
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.SetBasicAuth(username, passwd)

	client, err := newClient()
	if err != nil {
		return false, err
	}
	resp, err := client.Do(req)
	if err != nil {
       		return false, err
    	}
    	defer resp.Body.Close()

	var jsignin JsonSignin
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(bytes, &jsignin); err != nil {
		return false, err
	}
	if jsignin.Login != username {
		return false, nil
	}
	return true, nil
}

func (self *Github) CreateIssue() error {
	if !self.isLogined() {
		return nil
	}
	return self.createIssue()
}

func (self *Github) createIssue() error {
	var issue Issue
	if ok, err := editIssue(&issue, true); !ok {
		return err
	}
	if err := self.createPostIssue(&issue); err != nil {
		return err
	}
	return nil
}

func (self *Github) createPostIssue(issue *Issue) error {
	return self.httpIssue("POST", "", issue)
}

func (self *Github) httpIssue(method string , inum string, issue *Issue) error {
	url := self.url + "repos/" + self.repo + "/issues"
	retcode := 201
	if inum != "" {
		url = url + "/" + inum
		retcode = 200
	}

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
	if rcode != retcode {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return nil
	}

	if err := json.Unmarshal(iret, &issue); err != nil {
		return err
	}
	fmt.Printf("issue posted : #%v\n",issue.Num)
	return nil
}

func (self *Github) ModifyIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.modifyIssue(inum)
}

func (self *Github) modifyIssue(inum string) error {
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

func (self *Github) updatePostIssue(inum string, issue *Issue) error {
	return self.httpIssue("PATCH", inum, issue)
}

func (self *Github) getIssue(num string) (Issue, []IssueComment, error) {
	var issue Issue
	var comments []IssueComment
	iurl := self.url + "repos/" + self.repo + "/issues/" + num
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
	local, _ := time.LoadLocation("Local")
	issue.Update = issue.Update.In(local)
	for i, _ := range comments {
		comments[i].Update = comments[i].Update.In(local)
	}
	return issue, comments, nil
}

func (self *Github) AddIssueComment(inum string, comment []byte) error {
	if !self.isLogined() {
		return nil
	}
	return self.addIssueComment(inum, comment)
}

func (self *Github) addIssueComment(inum string, comment []byte) error {
	if err := self.httpComment("POST", inum, string(comment)); err != nil {
		return err
	}
	return nil
}

func (self *Github) httpComment(method string , inum string, body string) error {
	url := self.url + "repos/" + self.repo +
						"/issues/" + inum + "/comments"
	json_str := `{"body":"`+ lf2Esclf(onlyLF(body)) + `"}`
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

func (self *Github) reqHttp(method, url string, param []byte ) ([]byte,
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

func (self *Github) DoCloseIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doCloseIssue(inum)
}

func (self *Github) DoOpenIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doOpenIssue(inum)
}

func (self *Github) doCloseIssue(inum string) error {
	if err := self.toggleIssueState(inum, "closed"); err != nil {
		return err
	}
	return nil
}

func (self *Github) doOpenIssue(inum string) error {
	if err := self.toggleIssueState(inum, "open"); err != nil {
		return err
	}
	return nil
}

func (self *Github) toggleIssueState(inum string, state string) error {
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

func (self *Github) GetIssue(num string) (Issue, []IssueComment, error) {
	return self.getIssue(num)
}

func (self *Github) GetIssues(withclose bool) ([]Issue, error) {
	return self.getIssues(withclose)
}

func (self *Github) getIssues(withclose bool) ([]Issue, error) {
	url := self.url + "repos/" + self.repo + "/issues?"
	if withclose {
		url = url + "&state=all"
	}
	var p int = 1
	var ret []Issue
	local, _ := time.LoadLocation("Local")
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
			v.Update = v.Update.In(local)
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (self *Github) ReportIssues(now time.Time) (map[string]string, error) {
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

func (self *Github) reportIssue(newtag time.Time, is *Issue) (string, error) {
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
