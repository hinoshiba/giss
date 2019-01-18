package main

import (
	"os"
	"fmt"
	"flag"
	"giss/conf"
	"giss/cache"
	"giss/apicon"
)

var PrintAll bool
var LineLimit int
var RunMode string
var Options []string

var Apicon apicon.Apicon
var Cache cache.Cache
var Conf conf.Conf

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
	case "term":
		err = ComTerm()
	case "checkin":
		err = ComCheckin()
	case "report":
		if !okApiInit() {
			break
		}
		err = ComReport()
	case "create":
		if !okApiInit() {
			break
		}
		err = ComCreate()
	case "close":
		if !okApiInit() {
			break
		}
		err = ComClose(Options)
	case "com":
		if !okApiInit() {
			break
		}
		err = ComComment(Options)
	case "edit":
		if !okApiInit() {
			break
		}
		err = ComEdit(Options)
	case "open":
		if !okApiInit() {
			break
		}
		err = ComOpen(Options)
	case "show":
		if !okApiInit() {
			break
		}
		err = ComShow(Options)
	case "ls":
		if !okApiInit() {
			break
		}
		err = ComLs()
	case "export":
		if !okApiInit() {
			break
		}
		err = ComExport(Options)
	case "import":
		if !okApiInit() {
			break
		}
		err = ComImport(Options)
	case "mldel":
		if !okApiInit() {
			break
		}
		err = ComMlDel(Options)
	case "mlch":
		if !okApiInit() {
			break
		}
		err = ComMlCh(Options)
	case "mlls":
		if !okApiInit() {
			break
		}
		err = ComMlLs()
	case "lbls":
		if !okApiInit() {
			break
		}
		err = ComLbLs()
	case "lbdel":
		if !okApiInit() {
			break
		}
		err = ComLbDel(Options)
	case "lbadd":
		if !okApiInit() {
			break
		}
		err = ComLbAdd(Options)
	case "login":
		err = ComLogin()
	case "status":
		if !okApiInit() {
			break
		}
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

func okApiInit() bool {
	if Apicon == nil {
		fmt.Printf("empty target repository. you need <giss checkin>.\n")
		return false
	}
	return true
}

func init() {
	var line_limit int
	var print_all bool
	flag.IntVar(&line_limit, "l", 20, "Specify the maximum number of display lines.")
	flag.BoolVar(&print_all, "a", false, "Also displays detail or close.")
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

	co, err := conf.LoadUserConfig()
	if err != nil {
		die("Error : %s\n", err)
	}
	Conf = co

	ca, err := cache.LoadCaches()
	if err != nil {
		die("can't load a caches : %s\n", err)
	}
	Cache = ca

	Apicon, err = apicon.NewApicon(Conf, ca.Alias)
	if err != nil {
		return
	}
	Apicon.SetUsername(ca.User)
	Apicon.SetToken(ca.Token)
	Apicon.SetRepositoryName(ca.Repo)
	Apicon.SetUrl(ca.Url)
	Apicon.SetProxy(Conf.Server[ca.Alias].Proxy)

}

func main() {
	if err := giss(); err != nil {
		die("Error : %s\n", err)
	}
}
