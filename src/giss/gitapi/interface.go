package gitapi

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"strings"
	"net/http"
	"crypto/tls"
	"giss/cache"
)

type Apicon interface {
	GetRepo() string
	GetUrl() string
	GetUser() string
	GetToken() string
	SetRepo(string)
	IsLogined() bool
	LoadCache(cache.Cache) bool
	Login(string, string) error
	CreateIssue() error
	ModifyIssue(string) error
	AddIssueComment(string) error
	DoOpenIssue(string) error
	DoCloseIssue(string) error
	PrintIssues(int, bool) error
	PrintIssue(string, bool) error
	ReportIssues(time.Time) (map[string]string, error)
}

func NewGiteaCredent(url string) (Apicon, error) {
	var gitea Gitea
	var con Apicon
	con = &gitea
	gitea.Url = url
	return con, nil
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

func newClient() *http.Client {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{ InsecureSkipVerify: true },
    }
    return &http.Client{Transport: tr}
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
