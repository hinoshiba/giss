package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"strings"
	"time"
	"giss/mail"
	"giss/config"
	"giss/cache"
	"giss/values"
	"giss/gitapi"
	"golang.org/x/crypto/ssh/terminal"
	"github.com/hinoshiba/go-editor/editor"
)

var PrintAll bool
var RepoAutosend bool
var LineLimit int
var RunMode string
var Options []string
var Git gitapi.Apicon
var Cache cache.Cache

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func warn(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
}

func giss() error {
	defer os.RemoveAll(Cache.TmpDir)

	var err error
	switch RunMode {
	case "checkin":
		err = ComCheckin()
	case "report":
		err = ComReport()
	case "create":
		err = ComCreate()
	case "close":
		err = ComClose(Options)
	case "com":
		err = ComComment(Options)
	case "edit":
		err = ComEdit(Options)
	case "open":
		err = ComOpen(Options)
	case "show":
		err = ComShow(Options)
	case "ls":
		err = ComLs()
	case "login":
		err = ComLogin()
	case "status":
		err = ComStatus()
	case "version":
		err = ComVersion()
	case "help":
		err = ComHelp()
	default:
		warn("invalid argument : %s \nshow 'help' message.", RunMode)
	}
	return err
}

func ComReport() error {
	report_str := config.Rc.Report.Header
	now := time.Now()
	date_now := now.Format("01/02")
	ago := now.AddDate(0, 0, -7)
	date_7ago := ago.Format("2006/01/02")

	for _, v := range config.Rc.Report.TargetRepo {
		Git.SetRepo(v)
		report_str += "----------------- " + Git.GetRepo()
		report_str += " ---------------------------------------------\n"
		report, err := Git.ReportIssues(now)
		if err != nil {
			return err
		}
		for i, v := range report {
			report_str += "■ "+ i + "\n"
			report_str += v
		}
	}
	report_str += config.Rc.Report.Futter
	subject := config.Rc.Mail.Subject + " " + date_7ago + " - " + date_now
	if !RepoAutosend {
		fmt.Printf("Preview, Need -m to sending.\n\n======\n%s",subject)
		fmt.Printf("\n+++++++++++++++++++++++++++++++++\n%s",report_str)
		return nil
	}


	var smtp mail.Smtp
	err := smtp.New(config.Rc.Mail.Mta,
				      config.Rc.Mail.Port, config.Rc.Mail.From)
	if err != nil {
		return err
	}

	err = smtp.MakeMail(config.Rc.Mail.Header, config.Rc.Mail.To,
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
	fmt.Printf("CurServer      : %s\n",Git.GetUrl())
	fmt.Printf("ReportTargetRepository\n")
	for _, v := range config.Rc.Report.TargetRepo {
		fmt.Printf("   - %s\n", v)
	}

	url, err := inputString("enter the server url you want to use.>>")
	if err != nil {
		return nil
	}
	if url == "" {
		url = Git.GetUrl()
	}

	alias := config.GetAlias(url,config.Rc.Server)
	if alias == "" {
		fmt.Printf("undefined config. can't select this url.\n")
		return nil
	}

	fmt.Printf("ReportTargetRepository\n")
	for _, v := range config.Rc.Server[alias].Repos {
		fmt.Printf("   - %s\n", v)
	}
	repo, err := inputString("enter the repository you want to use.>>")
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
	return nil
}

func ComEdit(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	if err := Git.ModifyIssue(options[0]); err != nil {
		return err
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

	comment, err := editor.Call(config.Rc.Giss.Editor, []byte(""))
	if err != nil {
		return nil
	}
	if err := Git.AddIssueComment(options[0], comment); err != nil {
		return err
	}

	return nil
}

func ComCreate() error {
	err := Git.CreateIssue()
	if err != nil {
		return err
	}
	return nil
}

func ComClose(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	if err := Git.DoCloseIssue(options[0]); err != nil {
		return err
	}
	return nil
}

func ComOpen(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	if err := Git.DoOpenIssue(options[0]); err != nil {
		return err
	}
	return nil
}

func ComShow(options []string) error {
	if len(options) < 1 {
		fmt.Printf("can't detect issue number\n")
		return nil
	}
	if err := Git.PrintIssue(options[0], PrintAll); err != nil {
		return err
	}
	return nil
}

func ComLs() error {
	if err := Git.PrintIssues(LineLimit, PrintAll); err != nil {
		return nil
	}
	return nil
}

func ComStatus() error {
	if !Git.IsLogined() {
		fmt.Printf("not login\n")
		return nil
	}
	fmt.Printf("TargetRepo\n")
	fmt.Printf("      CurrentGit  : %s\n",Git.GetRepo())
	fmt.Printf("      Server      : %s\n",Git.GetUrl())
	fmt.Printf("GitCRED\n")
	fmt.Printf("      User        : %s\n",Git.GetUser())
	fmt.Printf("      Token       : %s****************\n",Git.GetToken()[:10])
	return nil
}

func ComLogin() error {
	var user, token string

	if !config.IsDefinedCred(Cache.Alias, config.Rc.Server) {
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
		if err := Git.Login(user, string(passwd)); err != nil {
			warn("login failed")
			return err
		}
		user = Git.GetUser()
		token = Git.GetToken()
	} else {
		user = config.Rc.Server[Cache.Alias].User
		token = config.Rc.Server[Cache.Alias].Token
	}

	if user == "" {
		fmt.Printf("empty username.\n")
		return nil
	}
	if token == "" {
		fmt.Printf("empty token.\n")
		return nil
	}
	if err := Cache.SaveCred(user, token); err != nil {
		warn("cache save failed")
		return err
	}
	fmt.Printf("Login Success. welcome %s !!", user)
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

func init() {
	var line_limit int
	var print_all bool
	var repo_autosend bool
	flag.IntVar(&line_limit, "l", 20, "Specify the maximum number of display lines.")
	flag.BoolVar(&print_all, "a", false, "Also displays detail or close.")
	flag.BoolVar(&repo_autosend, "m", false, "Send the report by e-mail.")
	flag.Usage = func() {
		ComHelp()
		os.Exit(0)
	}
	flag.Parse()

	if flag.NArg() < 1 {
		die("Argument is miss. show 'help' message.\n")
	}
	if flag.Arg(0) == "" {
		die("Argument is miss. show 'help' message.\n")
	}
	RunMode = flag.Arg(0)
	Options = flag.Args()[1:]
	LineLimit = line_limit
	PrintAll = print_all
	RepoAutosend = repo_autosend

	if err := config.LoadUserConfig(); err != nil {
		die("Error : %s\n", err)
	}

	var err error
	c, err := cache.LoadCaches()
	if err != nil {
		die("Error : %s\n", err)
	}

	Git, err = gitapi.NewGiteaCredent(config.Rc, c.Alias)
	if err != nil {
		die("Error : %s\n", err)
	}

	Git.LoadCache(c)
	Cache = c
	values.DebugVersion()
}


func main() {
	if err := giss(); err != nil {
		die("Error : %s\n", err)
	}
}
