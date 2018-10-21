package apicon

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"bufio"
	"strings"
	"net/http"
	"crypto/tls"
	"io/ioutil"
	"encoding/json"
	"github.com/hinoshiba/go-editor/editor"
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
var ReportNewTag string = "+"
var ReportTaskTag string = "-"

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
		fmt.Printf("not login\n")
		return false
	}
	if self.User == "" {
		fmt.Printf("not login\n")
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

func (self *Gitea) CreateIssue() error {
	if !self.isLogined() {
		return nil
	}
	return self.createIssue()
}

func (self *Gitea) createIssue() error {
	var issue Issue
	if ok, err := editIssue(&issue); !ok {
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

	if ok, err := editIssue(&issue); !ok {
		return err
	}
	if err := self.updatePostIssue(inum, &issue); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) AddIssueComment(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.addIssueComment(inum)
}

func (self *Gitea) addIssueComment(inum string) error {
	menu, err := inputString("To continue press the enter key....")
	if err != nil {
		return err
	}
	if menu != "" {
		return nil
	}

	comment, err := editor.Call("vim", []byte(""))
	if err := self.postComment("POST", inum, string(comment)); err != nil {
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

func (self *Gitea) updatePostIssue(inum string, issue *Issue) error {
	return self.postIssue("PATCH", inum, issue)
}

func (self *Gitea) createPostIssue(issue *Issue) error {
	return self.postIssue("POST", "", issue)
}

func (self *Gitea) postIssue(method string , inum string, issue *Issue) error {
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

func (self *Gitea) postComment(method string , inum string, body string) error {
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

func (self *Gitea) ReportIssues() (map[string]string, error) {
	now := time.Now()
	iss, err := self.getIssues(true)
	if err != nil {
		return nil, err
	}
	if len(iss) < 1 {
		return nil, nil
	}

	ret := make(map[string]string)
	newtag := dayAgo(now, -7)
	limit := dayAgo(now, -14)
	for _, is := range iss {
		if is.Update.Unix() < limit.Unix() {
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
	json_str := `{"name":"`+ TokenName + `"}`
    	req, err := http.NewRequest(
        	"POST",
        	url,
        	bytes.NewBuffer([]byte(json_str)),
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

func editIssue(issue *Issue) (bool, error) {
	for {
		fmt.Printf("edit option  \n\tt: title, b: body\n")
		fmt.Printf("other option \n\tp: issue print, done: edit done\n")
		menu, err := inputString("Please enter the menu (or cancel) >>")
		if err != nil {
			return false, err
		}
		switch menu {
		case "p":
			fmt.Printf("\n\n================ISSUE=============\n")
			fmt.Printf("Title : %s\n", issue.Title)
			fmt.Printf("Body ------->  \n%s\n", issue.Body)
			fmt.Printf("\n================END===============\n\n\n")
		case "t":
			b, err := editor.Call("vim", []byte(issue.Title))
			if err != nil {
				return false, err
			}
			issue.Title = string(b)
			fmt.Printf("title eddited\n")
		case "b":
			b, err := editor.Call("vim", []byte(issue.Body))
			if err != nil {
				return false, err
			}
			issue.Body = string(b)
			fmt.Printf("body eddited\n")
		case "done":
			fmt.Printf("done...\n")
			return true, nil
		case "cancel":
			fmt.Printf("Cancel was pressed.quitting...\n")
			return false, nil
		default:
			fmt.Printf("undefined command\n")
		}
		fmt.Printf("-----------------------Menu--------------------\n")
	}
	return false, nil
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

func  inputString(menu string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(menu)
    	istr, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	iline := strings.Trim(istr, " \n")
	return iline, nil
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

func onlyLF(str string) string {
	return strings.NewReplacer(
		"\r\n", "\n",
		"\r", "\n",
	).Replace(str)
}

func lf2Esclf(str string) string {
	return strings.NewReplacer(
		"\n", "\\n",
	).Replace(str)
}

func lf2space(str string) string {
	return strings.NewReplacer(
		"\n", " ",
	).Replace(str)
}

func dayAgo(t time.Time, ago int) time.Time {
	bw := t.AddDate(0, 0, ago)
	y, m, d := bw.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func makeWithin80c(inithead bool, hs int, str string) string {
	hatspace := strings.Repeat(" ", hs) + "| "
	splen := 80
    	runes := []rune(str)
	var ret string

	for i := 0; i < len(runes); i += splen {
		tmphat := hatspace
		if inithead && i == 0 {
			tmphat = ""
		}
		if i+splen < len(runes) {
			r := tmphat + string(runes[i:(i + splen)])
			ret += lf2space(onlyLF(r)) + "\n"
		} else {
			r := tmphat + string(runes[i:])
			ret += lf2space(onlyLF(r)) + "\n"
        	}
    	}
	return ret
}
