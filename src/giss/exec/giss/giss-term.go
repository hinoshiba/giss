package main

import (
	"github.com/hinoshiba/goctx"
	"github.com/hinoshiba/termwindow"
	"giss/apicon/issue"
	"giss/values"
	"giss/msg"
	"strings"
	"fmt"
)

func gissterm() error {
	own := goctx.NewOwner()

	if err := termwindow.Init(); err != nil {
		return err
	}
	go termwindow.Input(own.NewWorker())
	go termwindow.Start(own.NewWorker())
	defer termwindow.Close()

	termwindow.SetTitle(values.VersionText)
	termwindow.SetMsg("test init")


	go termMenu(own.NewWorker())

	own.Wait()
	return nil
}
func termMenu(wk goctx.Worker) {
	defer wk.Done()
	issues, err := termLs(false)
	if err != nil {
		termwindow.SetErr(err)
	}
	termwindow.SetMenu(issues.Data)
	var closed_print = false

	for {
		select {
		case <-wk.RecvCancel():
			return
		case  ev := <-termwindow.Key:
			switch ev.Key {
			case termwindow.KeyCtrlN:
				termwindow.SetActiveLine(issues.MvInc())
			case termwindow.KeyCtrlP:
				termwindow.SetActiveLine(issues.MvDec())
			case termwindow.KeySpace:
				if closed_print {
					closed_print = false
					termwindow.SetMsg("print open only.")
					continue
				}
				closed_print = true
				termwindow.SetMsg("print with closed.")
			case termwindow.KeyEnter:
				id, v := issues.GetData(issues.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				is, err := termShow(id)
				if err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("open #%v", id)
				termBody(wk.NewWorker(), is)
				termwindow.SetMsg("closed #%v", id)
				termwindow.ReFlush()
			}
			switch ev.Ch {
			case '$':
				var err error
				issues, err = termLs(closed_print)
				if err != nil {
					termwindow.SetErr(err)
				}
				termwindow.SetMenu(issues.Data)
				termwindow.SetMsg("updated issues data.")
			case 'j':
				termwindow.SetActiveLine(issues.MvInc())
			case 'k':
				termwindow.SetActiveLine(issues.MvDec())
			case 'G':
				termwindow.SetActiveLine(issues.MvBottom())
			case 'g':
				termwindow.SetActiveLine(issues.MvTop())
			case 'n':
				title := inputRecode(wk.NewWorker(), "NewIssueTitle")
				body := inputRecode(wk.NewWorker(), "NewIssueBody")
				if err := termCreate(title, body); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("Created Issue.")
				termwindow.ReFlush()
			case 'c':
				id, v := issues.GetData(issues.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				com := inputRecode(wk.NewWorker(), "comment input")
				if err := termCom(id, com); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("commented.")
				termwindow.ReFlush()
			case 'C':
				id, v := issues.GetData(issues.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				if err := termClose(id); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("update detected")
				termwindow.ReFlush()
			case 'O':
				id, v := issues.GetData(issues.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				if err := termOpen(id); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("update detected")
				termwindow.ReFlush()
			case 'q':
				wk.Cancel()
				return
			}
		default:
		}
	}
}

func termBody(wk goctx.Worker, is termwindow.Lines) {
	defer wk.Done()

	termwindow.SetBody(is.Data)
	defer termwindow.UnsetBody()
	for {
		select {
		case <-wk.RecvCancel():
			return
		case  ev := <-termwindow.Key:
			switch ev.Key {
			case termwindow.KeyCtrlN:
				termwindow.SetActiveLine(is.MvInc())
			case termwindow.KeyCtrlP:
				termwindow.SetActiveLine(is.MvDec())
			}
			switch ev.Ch {
			case 'j':
				termwindow.SetActiveLine(is.MvInc())
			case 'k':
				termwindow.SetActiveLine(is.MvDec())
			case 'G':
				termwindow.SetActiveLine(is.MvBottom())
			case 'g':
				termwindow.SetActiveLine(is.MvTop())
			case 'q':
				termwindow.UnsetBody()
				return
			}
		}
	}
}

func inputRecode(wk goctx.Worker, title string) string {
	defer wk.Done()

	var buf string
	for {
		termwindow.SetMsg(title + " : " + buf)
		select {
		case <-wk.RecvCancel():
			return ""
		case  ev := <-termwindow.Key:
			switch ev.Key {
			case termwindow.KeyEnter:
				return buf
			case termwindow.KeyBackspace2:
				rbuf := []rune(buf)
				if len(rbuf) > 1 {
					buf = string(rbuf[:(len(rbuf) - 1)])
				}
				continue
			case termwindow.KeyBackspace:
				rbuf := []rune(buf)
				if len(rbuf) > 1 {
					buf = string(rbuf[:(len(rbuf) - 1)])
				}
				continue
			default:
			}
			buf += string(ev.Ch)
		}
	}
}

func termShow(id int) (termwindow.Lines, error) {
	sid := fmt.Sprintf("%v", id)
	issue, err := Apicon.GetIssue(sid)
	if err != nil {
		return termwindow.Lines{}, err
	}
	if issue.State.Name == "" {
		return termwindow.Lines{}, msg.NewErr("undefined issue")
	}

	var lines termwindow.Lines
	lines.Data.Title = []byte(msg.NewStr("Issue #%s detail window", sid))
	sis := issue.SprintMd()
	sisml := strings.Split(sis, "\n")
	for _, sisl := range sisml {
		lines.Append(0, []byte(sisl))
	}

	return lines, nil
}

func termClose(id int) error {
	sid := fmt.Sprintf("%v", id)
	if err := Apicon.DoCloseIssue(sid); err != nil {
		return err
	}
	return nil
}

func termOpen(id int) error {
	sid := fmt.Sprintf("%v", id)
	if err := Apicon.DoOpenIssue(sid); err != nil {
		return err
	}
	return nil
}

func termCom(id int, com string) error {
	sid := fmt.Sprintf("%v", id)
	scomment := lf2Esclf(onlyLF(string(com)))
	if err := Apicon.AddIssueComment(sid, scomment); err != nil {
		return err
	}
	return nil
}

func termCreate(title string, body string) error {
	var is issue.Issue
	is.Title = title
	is.Body = body
	if err := Apicon.CreateIssue(is); err != nil {
		return err
	}
	return nil
}

func termLs(closed bool) (termwindow.Lines, error) {
	issues, err := Apicon.GetIssues(false, closed)
	if err != nil {
		return termwindow.Lines{}, err
	}
	if len(issues) < 1 {
		return termwindow.Lines{}, nil
	}

	var lines termwindow.Lines
	for _, is := range issues {
		str, err := is.SprintHead()
		if err != nil {
			return termwindow.Lines{}, err
		}
		lines.Append(int(is.Num), []byte(str))
	}
	return lines, nil
}
