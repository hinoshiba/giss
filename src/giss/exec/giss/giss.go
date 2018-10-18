package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"net/http"
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"giss/config"
	"giss/cache"
	"giss/values"
	"golang.org/x/crypto/ssh/terminal"
)

var LineLimit int64
var RunMode string
var Options []string

func die(msg ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error : %s\n", msg...)
	os.Exit(1)
}

func warn(msg ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error : %s\n", msg...)
}

func giss() error {
	var err error
	switch RunMode {
	case "login":
		err = ComLogin()
	case "version":
		err = ComVersion()
	case "help":
		err = ComHelp()
	default:
		warn("invalid argument : %s %s\n", RunMode, Options)
		err = ComHelp()
	}
	return err
}

func ComLogin() error {

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Enter Username: ")
    	user, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("Enter password: ")
 	pass, err := terminal.ReadPassword(0)
	if err != nil {
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
	var line_limit int64
	flag.Int64Var(&line_limit, "l", 20, "Specify the maximum number of display lines.")
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

	if err := cache.LoadCaches(); err != nil {
		die(err)
	}

	if err := config.LoadUserConfig(); err != nil {
		die(err)
	}
	//fmt.Printf("%s\n",config.Rc.Body.Header)
}

func main() {
	defer os.RemoveAll(cache.TmpDir)
	if err := giss(); err != nil {
		die(err)
	}
}
