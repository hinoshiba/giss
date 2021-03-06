package apicon

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"strings"
	"giss/conf"
	"giss/msg"
	"giss/apicon/issue"
	"giss/apicon/gitea"
	"giss/apicon/github"
	"giss/apicon/redmine"
	"github.com/hinoshiba/go-editor/editor"
)

var Conf conf.Conf

type Apicon interface {
	GetRepositoryName() string
	SetRepositoryName(string)
	GetUrl() string
	SetUrl(string)
	GetUsername() string
	SetUsername(string)
	GetToken() string
	SetToken(string)
	GetProxy() string
	SetProxy(string)
	GetIssue(string) (issue.Issue, error)
	GetIssues(bool, bool) ([]issue.Issue, error)
	CreateIssue(issue.Issue) error
	ModifyIssue(issue.Issue) error
	AddIssueComment(string, string) error
	DoOpenIssue(string) error
	DoCloseIssue(string) error
	IsLogined() bool
	GetMilestones() ([]issue.Milestone, error)
	UpdateMilestone(string, string) error
	DeleteMilestone(string) error
	GetLabels() ([]issue.Label, error)
	AddLabel(string, string) error
	DelLabel(string, string) error
}

func NewApicon(rc conf.Conf, alias string) (Apicon, error) {
	apitype := rc.Server[alias].Type
	var ret Apicon
	switch apitype {
		case "Gitea":
			var obj gitea.Gitea
			Conf = rc
			ret = &obj
		case "Github":
			var obj github.Github
			Conf = rc
			ret = &obj
		case "Redmine":
			var obj redmine.Redmine
			Conf = rc
			ret = &obj
	}
	if ret == nil {
		err := msg.NewErr("selected unknown api type : %s",alias)
		return ret, err
	}
	return ret, nil


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

func  inputString(menu string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(menu)
    	istr, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	iline := strings.Trim(onlyLF(istr), " \n")
	return iline, nil
}

func EditIssue(is *issue.Issue, fastedit bool) (bool, error) {
	if fastedit {
		b, err := editor.Call(Conf.Giss.Editor, []byte(is.Title))
		if err != nil {
			return false, err
		}
		is.Title = lf2space(onlyLF(string(b)))
		fmt.Printf("Title : %s\n", is.Title)
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
			fmt.Printf("Title : %s\n", is.Title)
			fmt.Printf("Body ------->  \n%s\n", is.Body)
			fmt.Printf("\n================END===============\n\n\n")
		case "t":
			b, err := editor.Call(Conf.Giss.Editor, []byte(is.Title))
			if err != nil {
				return false, err
			}
			is.Title = lf2space(onlyLF(string(b)))
			fmt.Printf("title eddited\n")
		case "b":
			b, err := editor.Call(Conf.Giss.Editor, []byte(is.Body))
			if err != nil {
				return false, err
			}
			is.Body = onlyLF(string(b))
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

func PrintIssues(iss []issue.Issue, limit int) {
	if len(iss) < 1 {
		return
	}

	fmt.Printf("   Id [ Milestone  ] Title\n----------------------------\n")
	for index, is := range iss {
		if index >= limit {
			break
		}
		is.PrintHead()
	}
}

func ReportIssues(git Apicon, now time.Time) (map[string]string, error) {
	iss, err := git.GetIssues(false, true)
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
		if is.Update.Unix() < limit.Unix() && is.State.Name == "closed" {
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

func reportIssue(git Apicon, newtag time.Time, is *issue.Issue) (string, error) {
	ir := "  - "
	if is.Update.Unix() >= newtag.Unix() {
		ir = "+ - "
	}
	if is.State.Name == "closed" {
		ir += "[closed] "
	}
	ir += msg.NewStr("#%v ",is.Num) + lf2space(onlyLF(is.Title)) + "\n"
	for _, row := range strings.Split(onlyLF(is.Body),"\n") {
		ir += makeWithin80c(false,6 ,row)
	}

	nis, err := git.GetIssue(msg.NewStr("%v",is.Num))
	if err != nil {
		return "", err
	}
	for _, com := range nis.Comments {
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
