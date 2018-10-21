package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"strings"
	"giss/config"
	"giss/cache"
	"giss/values"
	"giss/apicon"
	"github.com/hinoshiba/go-editor/editor"
	"golang.org/x/crypto/ssh/terminal"
)

var PrintAll bool
var LineLimit int
var RunMode string
var Options []string
var Git apicon.Gitea

func die(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
	os.Exit(1)
}

func warn(s string, msg ...interface{}) {
	fmt.Fprintf(os.Stderr, s + "\n" , msg...)
}

func giss() error {

	var err error
	switch RunMode {
	case "test":
		err = ComTest()
	case "create":
		err = ComCreate()
	case "close":
		err = ComClose(Options)
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

func ComTest() error {
	ret, err := editor.Call("vim",[]byte("aaaaaaaa"))
	if err != nil {
		return err
	}

	fmt.Printf("%s",string(ret))
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
	fmt.Printf("      CurrentGit  : %s\n",Git.Repo)
	fmt.Printf("      Server      : %s\n",Git.Url)
	fmt.Printf("GitCRED\n")
	fmt.Printf("      User        : %s\n",Git.User)
	fmt.Printf("      Token       : %s****************\n",Git.Token[:10])
	return nil
}

func ComLogin() error {
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
	if err := cache.SaveCred(Git.User, Git.Token); err != nil {
		warn("cache save failed")
		return err
	}
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

func init() {
	var line_limit int
	var print_all bool
	flag.IntVar(&line_limit, "l", 20, "Specify the maximum number of display lines.")
	flag.BoolVar(&print_all, "a", false, "Also displays detail or close")
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

	if err := cache.LoadCaches(); err != nil {
		die("Error : %s\n", err)
	}

	if err := config.LoadUserConfig(); err != nil {
		die("Error : %s\n", err)
	}

	var err error
	Git, err = apicon.NewGiteaCredent(config.Rc.GitDefault.Url)
	if err != nil {
		die("Error : %s\n", err)
	}
	Git.LoadCache(cache.User, cache.Token, cache.CurrentGit)

}

func main() {
	defer os.RemoveAll(cache.TmpDir)
	if err := giss(); err != nil {
		os.RemoveAll(cache.TmpDir)
		die("Error : %s\n", err)
	}
}
