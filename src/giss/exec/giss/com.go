package main

import (
	"os"
	"fmt"
	"time"
	"bufio"
	"strings"
	"giss/mail"
	"giss/cache"
	"giss/apicon"
	"giss/values"
	"giss/apicon/issue"
	"github.com/hinoshiba/go-editor/editor"
)

func ComReport() error {
	report_str := Conf.Report.Header
	now := time.Now()
	date_now := now.Format("01/02")
	ago := now.AddDate(0, 0, -7)
	date_7ago := ago.Format("2006/01/02")

	for _, v := range Conf.Report.TargetRepo {
		Apicon.SetRepositoryName(v)
		report_str += "----------------- " + Apicon.GetRepositoryName()
		report_str += " ---------------------------------------------\n"
		report, err := apicon.ReportIssues(Apicon, now)
		if err != nil {
			return err
		}
		for i, v := range report {
			report_str += "■ "+ i + "\n"
			report_str += v
		}
	}
	report_str += Conf.Report.Futter
	subject := Conf.Mail.Subject + " " + date_7ago + " - " + date_now
	if !RepoAutosend {
		fmt.Printf("Preview, Need -m to sending.\n\n======\n%s",subject)
		fmt.Printf("\n+++++++++++++++++++++++++++++++++\n%s",report_str)
		return nil
	}


	var smtp mail.Smtp
	err := smtp.New(Conf.Mail.Mta, Conf.Mail.Port, Conf.Mail.From)
	if err != nil {
		return err
	}

	err = smtp.MakeMail(Conf.Mail.Header, Conf.Mail.To,
						subject, []byte(report_str))
	if err != nil {
		return err
	}

	if err := smtp.Send(); err != nil {
		return err
	}
	return nil
}

func ComCheckin() error {
	var n_alias string
	var n_url string
	if Apicon != nil {
		n_url = Apicon.GetUrl()
		n_alias = Conf.GetAlias(n_url)
	}

	if n_alias != "" {
		fmt.Printf("CurrentServer\n%s : %s\n", n_alias, n_url)
	}
	fmt.Printf("=======================================================\n")
	fmt.Printf("Alias : [Sign User] Server url\n")
	fmt.Printf("-------------------------------------------------------\n")
	for i, v := range Conf.Server {
		fmt.Printf("%s : [%s] %s\n", i, v.User, v.Url)
	}
	var err error
	alias, err := inputString("\nenter the server alias you want to use.>>")
	if err != nil {
		return nil
	}
	if alias == "" {
		if n_alias == "" {
			fmt.Printf("can't select empty alias.\n")
			return nil
		}
		alias = n_alias
	}
	if _, err := apicon.NewApicon(Conf, alias); err != nil {
		return err
	}

	url := Conf.Server[alias].Url
	if url == "" {
		fmt.Printf("undefined alias\n")
		return nil
	}

	fmt.Printf("=======================================================\n")
	fmt.Printf("repository names\n")
	fmt.Printf("-------------------------------------------------------\n")
	for _, v := range Conf.Server[alias].Repos {
		fmt.Printf(" - %s\n", v)
	}
	repo, err := inputString("\nenter the repository you want to use.>>")
	if err != nil {
		return nil
	}
	if repo == "" {
		fmt.Printf("empty repository name.\n")
		return nil
	}

	if err := Cache.SaveCurrentGit(alias, url, repo); err != nil {
		return err
	}

	c, err := cache.LoadCaches()
	if err != nil {
		return err
	}
	if c.Url != url {
		fmt.Printf("checkin failed. empty url.\n")
		return nil

	}
	if c.Repo != repo {
		fmt.Printf("checkin failed. empty repository.\n")
		return nil
	}

	Cache = c
	fmt.Printf("checkin :%s/%s\n", Cache.Url, Cache.Repo)
	if Conf.Server[alias].AutoLogin {
		fmt.Printf("autologin.....\n\n")
		ComLogin()
	}
	return nil
}

func ComComment(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}

	menu, err := inputString("To continue press the enter key....")
	if err != nil {
		return err
	}
	if menu != "" {
		return nil
	}

	comment, err := editor.Call(Conf.Giss.Editor, []byte(""))
	if err != nil {
		return nil
	}
	scomment := lf2Esclf(onlyLF(string(comment)))
	if err := Apicon.AddIssueComment(options[0], scomment); err != nil {
		fmt.Printf("comment failed.\n%s\n", comment)
		return err
	}

	return nil
}

func ComEdit(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	inum := options[0]

	issue, err := Apicon.GetIssue(inum)
	if err != nil {
		return err
	}
	if issue.State.Name == "" {
		fmt.Printf("undefined ticket: %s\n", inum)
		return nil
	}

	if ok, err := apicon.EditIssue(&issue, false); !ok {
		return err
	}
	if err := Apicon.ModifyIssue(issue); err != nil {
		fmt.Printf("update failed\n-------------\n")
		issue.PrintMd()
		return err
	}
	return nil
}

func ComLbAdd(args []string) error {
	if len(args) < 2 {
		fmt.Printf("giss lbadd <issue number> <label name>\n")
		return nil
	}
	if err := Apicon.AddLabel(args[0], args[1]); err != nil {
		return err
	}
	return nil
}

func ComLbDel(args []string) error {
	if len(args) < 2 {
		fmt.Printf("giss lbdel <issue number> <label name>\n")
		return nil
	}
	if err := Apicon.DelLabel(args[0], args[1]); err != nil {
		return err
	}
	return nil
}

func ComMlCh(args []string) error {
	if len(args) < 2 {
		fmt.Printf("giss mlch <issue number> <milestone name>\n")
		return nil
	}
	if err := Apicon.UpdateMilestone(args[0], args[1]); err != nil {
		return err
	}
	return nil
}

func ComMlDel(args []string) error {
	if len(args) < 1 {
		fmt.Printf("giss mldel <issue number>\n")
		return nil
	}
	if err := Apicon.DeleteMilestone(args[0]); err != nil {
		return err
	}
	return nil
}

func ComMlLs() error {
	mls, err := Apicon.GetMilestones()
	if err != nil {
		return err
	}

	for _, ml := range mls {
		fmt.Printf("%s\n", ml.Title)
	}
	return nil
}

func ComLbLs() error {
	lbs, err := Apicon.GetLabels()
	if err != nil {
		return err
	}

	for _, lb := range lbs {
		lstr, err := lb.GetLabelStr()
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", lstr)
	}
	return nil
}

func ComCreate() error {
	var is issue.Issue
	if ok, err := apicon.EditIssue(&is, true); !ok {
		return err
	}

	err := Apicon.CreateIssue(is)
	if err != nil {
		fmt.Printf("create failed\n-------------\n")
		is.PrintMd()
		return err
	}
	return nil
}

func ComClose(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	if err := Apicon.DoCloseIssue(options[0]); err != nil {
		return err
	}
	return nil
}

func ComOpen(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	if err := Apicon.DoOpenIssue(options[0]); err != nil {
		return err
	}
	return nil
}

func ComShow(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}

	issue, err := Apicon.GetIssue(options[0])
	if err != nil {
		return err
	}
	if issue.State.Name == "" {
		fmt.Printf("undefined ticket: %s\n", options[0])
		return nil
	}

	issue.PrintMd()
	return nil
}

func ComLs() error {
	issues, err := Apicon.GetIssues(false, PrintAll)
	if err != nil {
		return err
	}
	if len(issues) < 1 {
		return nil
	}
	apicon.PrintIssues(issues, LineLimit)
	return nil
}

func ComExport(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect export type\n")
		return nil
	}

	if options[0] == "" {
		fmt.Printf("unknown export type : %s\n", options[0])
		return nil
	}
	if options[0] != "json" && options[0] != "xml" {
		fmt.Printf("unknown export type : %s\n", options[0])
		return nil
	}

	issues, err := Apicon.GetIssues(true, PrintAll)
	if err != nil {
		return err
	}
	if len(issues) < 1 {
		return nil
	}

	switch(options[0]) {
	case "json" :
		issue.ExportJson(&issues)
		break
	case "xml" :
		issue.ExportXml(&issues)
		break
	default:
		fmt.Printf("unknown export type : %s\n", options[0])
		return nil
	}

	return nil
}

func ComStatus() error {
	if !Apicon.IsLogined() {
		fmt.Printf("not login\n")
		return nil
	}
	fmt.Printf("TargetRepo\n")
	fmt.Printf("      CurrentApicon  : %s\n",Apicon.GetRepositoryName())
	fmt.Printf("      Server      : %s\n",Apicon.GetUrl())
	fmt.Printf("ApiconCRED\n")
	fmt.Printf("      User        : %s\n",Apicon.GetUsername())
	fmt.Printf("      Token       : %s****************\n",Apicon.GetToken()[:10])
	return nil
}

func ComLogin() error {
	var user, token string

	if Conf.IsDefinedCred(Cache.Alias) {
		user = Conf.Server[Cache.Alias].User
		token = Conf.Server[Cache.Alias].Token
	}
		/*
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Enter Username: ")
	    	cruser, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		user := strings.Trim(cruser, " \n")
		fmt.Printf("Enter password:")
	 	passwd, err := terminal.ReadPassword(0)
		if err != nil {
			return err
	 	}
		fmt.Printf("\n\n")
		if err := Apicon.Login(user, string(passwd)); err != nil {
			warn("login failed")
			return err
		}
		user = Apicon.GetUser()
		token = Apicon.GetToken()
		*/

	if user == "" {
		fmt.Printf("can't autoload username.\n")
		return nil
	}
	if token == "" {
		fmt.Printf("can't autoload token.\n")
		return nil
	}
	if err := Cache.SaveCred(user, token); err != nil {
		warn("cache save failed")
		return err
	}
	fmt.Printf("Login Success. welcome %s !!\n", user)
	return nil
}

func ComVersion() error {
	fmt.Printf("%s\n",values.VersionText)
	return nil
}

func ComHelp() error {
	fmt.Printf("%s\n",values.HelpText)
	return nil
}

func inputString(menu string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(menu)
	istr, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	iline := strings.Trim(onlyLF(istr), " \n")
	return iline, nil
}

var STRINGSLF = strings.NewReplacer("\r\n", "\n", "\r", "\n",)
func onlyLF(str string) string {
	return STRINGSLF.Replace(str)
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
