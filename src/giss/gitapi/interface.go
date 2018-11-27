package gitapi

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"errors"
	"strings"
	"net/url"
	"net/http"
	"crypto/tls"
	"giss/cache"
	"giss/config"
	"github.com/hinoshiba/go-editor/editor"
)

var Conf config.Config

type Apicon interface {
	GetRepo() string
	GetUrl() string
	GetUser() string
	GetToken() string
	GetIssue(string) (Issue, []IssueComment, error)
	GetIssues(bool) ([]Issue, error)
	SetRepo(string)
	IsLogined() bool
	LoadCache(cache.Cache) bool
	Login(string, string) error
	CreateIssue() error
	ModifyIssue(string) error
	AddIssueComment(string, []byte) error
	DoOpenIssue(string) error
	DoCloseIssue(string) error
}

func NewGiteaCredent(rc config.Config, alias string) (Apicon, error) {
	apitype := rc.Server[alias].Type
	var ret Apicon
	switch apitype {
		case "Gitea":
			var gitea Gitea
			Conf = rc
			ret = &gitea
			return ret, nil
		case "Github":
			var github Github
			Conf = rc
			ret = &github
			return ret, nil
	}

	err := errors.New(fmt.Sprintf("selected unknown api type : %s",alias))
	return ret, err
}

type IssueEdited struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	State  string     `json:"state"`
	User   IssueUser  `json:"user"`
	//Assgin string     `json:"assignee"`
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
//	Assgin string     `json:"assignee"`
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

func newClient() (*http.Client, error) {
	http_proxy := os.Getenv("http_proxy")
	if http_proxy == "" {
		http_proxy = os.Getenv("https_proxy")
	}
	if http_proxy != "" {
		proxy, err := url.Parse(http_proxy)
		if err != nil {
			return nil, err
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
			Proxy: http.ProxyURL(proxy),
		}
		return &http.Client{Transport: tr}, nil
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
	}
	return &http.Client{Transport: tr}, nil
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

func editIssue(issue *Issue, fastedit bool) (bool, error) {
	if fastedit {
		b, err := editor.Call(Conf.Giss.Editor, []byte(issue.Title))
		if err != nil {
			return false, err
		}
		issue.Title = string(b)
		fmt.Printf("Title : %s\n", issue.Title)
	}
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
			b, err := editor.Call(Conf.Giss.Editor, []byte(issue.Title))
			if err != nil {
				return false, err
			}
			issue.Title = string(b)
			fmt.Printf("title eddited\n")
		case "b":
			b, err := editor.Call(Conf.Giss.Editor, []byte(issue.Body))
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

func PrintIssue(issue Issue, comments []IssueComment) {
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
}
func PrintIssues(issues []Issue, limit int) {
	if len(issues) < 1 {
		return
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
}

func convIssueEdited(issue Issue) IssueEdited {
	var nissue IssueEdited

	nissue.Id = issue.Id
	nissue.Num = issue.Num
	nissue.Title = issue.Title
	nissue.Body = issue.Body
	nissue.State = issue.State
	nissue.User  = issue.User
//	nissue.Assgin = issue.Assgin

	return nissue
}

func ReportIssues(git Apicon, now time.Time) (map[string]string, error) {
	iss, err := git.GetIssues(true)
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
		ir, err := reportIssue(git, newtag, &is)
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

func reportIssue(git Apicon, newtag time.Time, is *Issue) (string, error) {
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

	_, coms, err := git.GetIssue(fmt.Sprintf("%v",is.Num))
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
