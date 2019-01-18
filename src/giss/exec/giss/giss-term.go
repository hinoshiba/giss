package main

import (
	"github.com/hinoshiba/goctx"
	"github.com/hinoshiba/termwindow"
	"giss/apicon/issue"
	"giss/apicon"
	"giss/values"
	"giss/msg"
	"strings"
)

func gissterm() error {
	own := goctx.NewOwner()

	if err := termwindow.Init(); err != nil {
		return err
	}
	go termwindow.Input(own.NewWorker())
	go termwindow.Start(own.NewWorker())
	defer termwindow.Close()

	termwindow.SetTitle(values.TermTitle)
	termwindow.SetMsg("initializing")
	go termMenu(own.NewWorker())

	own.Wait()
	return nil
}
func termMenu(wk goctx.Worker) {
	defer wk.Done()

	var startmsg termwindow.Window
	startmsg.SetTitle("Message from developer")
	startmsg.Data.Body = str0d0a2bytearr(values.StartTerm)
	termBody(wk.NewWorker(), startmsg)

	termwindow.SetMsg("connect to %s/%s",
		Apicon.GetUrl(), Apicon.GetRepositoryName())
	winiss, err := termLs(false)
	if err != nil {
		termwindow.SetErr(err)
	}
	termwindow.SetMenu(winiss.Data)
	termwindow.SetMsg("pulled issues")

	var closed_print = false

	for {
		select {
		case <-wk.RecvCancel():
			return
		case  ev := <-termwindow.Key:
	//		termwindow.SetMsg("")
			switch ev.Key {
			case termwindow.KeyCtrlN:
				termwindow.SetActiveLine(winiss.MvInc())
				continue
			case termwindow.KeyCtrlP:
				termwindow.SetActiveLine(winiss.MvDec())
				continue
			case termwindow.KeySpace:
				if closed_print {
					closed_print = false
					termwindow.SetMsg("print open only.")
					continue
				}
				closed_print = true
				termwindow.SetMsg("print with closed.")
				continue
			case termwindow.KeyEnter:
				id, v := winiss.GetData(winiss.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				is, err := getIssueWindow(id)
				if err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("open #%v", id)
				termBody(wk.NewWorker(), is)
				termwindow.SetMsg("closed #%v", id)
				termwindow.ReFlush()
				continue
			}
			switch ev.Ch {
			case '?':
				var help termwindow.Window
				help.SetTitle("help window")
				help.Data.Body = str0d0a2bytearr(values.HelpTerm)
				termBody(wk.NewWorker(), help)
				termwindow.ReFlush()
			case '$':
				termwindow.SetMsg("connect to %s/%s",
					Apicon.GetUrl(), Apicon.GetRepositoryName())
				var err error
				winiss, err = termLs(closed_print)
				if err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMenu(winiss.Data)
				termwindow.SetMsg("pulled issues")
			case 'j':
				termwindow.SetActiveLine(winiss.MvInc())
			case 'k':
				termwindow.SetActiveLine(winiss.MvDec())
			case 'G':
				termwindow.SetActiveLine(winiss.MvBottom())
			case 'g':
				termwindow.SetActiveLine(winiss.MvTop())
			case 'n':
				title := inputRecode(wk.NewWorker(), "NewIssueTitle")
				body := inputRecode(wk.NewWorker(), "NewIssueBody")
				if err := termCreate(title, body); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("a possibility that it was updated. Please enter '$'.")
				termwindow.ReFlush()
			case 'c':
				id, v := winiss.GetData(winiss.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				com := inputRecode(wk.NewWorker(), "comment input")
				if err := termCom(id, com); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("commented")
				termwindow.ReFlush()
			case 'C':
				id, v := winiss.GetData(winiss.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				if err := termClose(id); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("a possibility that it was updated. Please enter '$'.")
				termwindow.ReFlush()
			case 'O':
				id, v := winiss.GetData(winiss.Active)
				if len(v) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				if err := termOpen(id); err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("a possibility that it was updated. Please enter '$'.")
				termwindow.ReFlush()
			case 'L':
				id, v := winiss.GetData(winiss.Active)
				if len(v) < 0 {
					termwindow.SetErrStr("target not found")
					continue
				}
				termwindow.SetMsg("called label selecter : #%s", id)
				err := termLabel(wk.NewWorker(), id)
				if err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("a possibility that it was updated. Please enter '$'.")
				termwindow.ReFlush()
			case 'M':
				id, v := winiss.GetData(winiss.Active)
				if len(v) < 0 {
					termwindow.SetErrStr("target not found")
					continue
				}
				termwindow.SetMsg("called milestone selecter : #%s", id)
				err := termMilestone(wk.NewWorker(), id)
				if err != nil {
					termwindow.SetErr(err)
					continue
				}
				termwindow.SetMsg("a possibility that it was updated. Please enter '$'.")
				termwindow.ReFlush()
			case '/':
				napicon, err := termCheckin(wk.NewWorker())
				if err != nil {
					termwindow.SetErr(err)
					termwindow.SetMenu(winiss.Data)
					continue
				}

				termwindow.SetMsg("connect to %s/%s",
					napicon.GetUrl(), napicon.GetRepositoryName())
				buf := Apicon
				Apicon = napicon
				winiss, err = termLs(closed_print)
				if err != nil {
					termwindow.SetErr(err)
					Apicon = buf
					winiss, err = termLs(closed_print)
					if err != nil {
						termwindow.SetErr(err)
						continue
					}
					termwindow.SetMenu(winiss.Data)
					continue
				}
				termwindow.SetMenu(winiss.Data)
				termwindow.SetMsg("checkined %s, pulled issues.", napicon.GetRepositoryName())
			case 'q':
				wk.Cancel()
				return
			default:
				termwindow.SetErrStr("undefined key. please enter to '?'")
			}
		default:
		}
	}
}

func getActiveLabels(id string) ([]string, error) {
	is, err := Apicon.GetIssue(id)
	if err != nil {
		return []string{}, err
	}
	if is.State.Name == "" {
		return []string{}, msg.NewErr("undefined issue")
	}
	var lbsname []string
	for _, lb := range is.Labels {
		lbsname = append(lbsname, lb.Name)
	}
	return lbsname, nil
}

func getActiveMilestone(id string) (string, error) {
	is, err := Apicon.GetIssue(id)
	if err != nil {
		return "", err
	}
	if is.State.Name == "" {
		return "", msg.NewErr("undefined issue")
	}
	return is.Milestone.Title, nil
}

func getIssueWindow(id string) (termwindow.Window, error) {
	issue, err := Apicon.GetIssue(id)
	if err != nil {
		return termwindow.Window{}, err
	}
	if issue.State.Name == "" {
		return termwindow.Window{}, msg.NewErr("undefined issue")
	}

	var win termwindow.Window
	win.Data.Title = []byte(msg.NewStr("Issue #%s detail window", id))
	sis := issue.SprintMd()
	sisml := strings.Split(sis, "\n")
	for _, sisl := range sisml {
		win.Append("", []byte(sisl))
	}

	return win, nil
}

func getLabelWindow(lbsname []string) (termwindow.Window, error) {
	lbs, err := Apicon.GetLabels()
	if err != nil {
		return termwindow.Window{}, err
	}
	if len(lbs) < 1 {
		return termwindow.Window{}, msg.NewErr("undefined labels")
	}

	var win termwindow.Window
	win.SetTitle("label window")
	for _, lb := range lbs {
		head := "  "
		if contains(lbsname, lb.Name) {
			head = "* "
		}
		win.Append(lb.Name, []byte(head + lb.Name))
	}
	return win, nil
}

func getMilestonesWindow(mlname string) (termwindow.Window, error) {
	mls, err := Apicon.GetMilestones()
	if err != nil {
		return termwindow.Window{}, err
	}
	if len(mls) < 1 {
		return termwindow.Window{}, msg.NewErr("undefined milestone")
	}

	var win termwindow.Window
	win.SetTitle("milestone window")
	for _, ml := range mls {
		head := "  "
		if ml.Title == mlname {
			head = "* "
		}
		win.Append(ml.Title, []byte(head + ml.Title))
	}
	return win, nil
}

func termLabel(wk goctx.Worker, id string) error {
	defer wk.Done()

	lbs, err := getActiveLabels(id)
	if err != nil {
		return err
	}
	win, err := getLabelWindow(lbs)
	if err != nil {
		return err
	}
	termwindow.SetBody(win.Data)
	defer termwindow.UnsetBody()

	for {
		select {
		case <-wk.RecvCancel():
			return nil
		case  ev := <-termwindow.Key:
			switch ev.Key {
			case termwindow.KeyCtrlN:
				termwindow.SetActiveLine(win.MvInc())
			case termwindow.KeyCtrlP:
				termwindow.SetActiveLine(win.MvDec())
			case termwindow.KeySpace:
				lb, _ := win.GetData(win.Active)
				if len(lb) < 0 {
					termwindow.SetErrStr("target not found")
					continue
				}
				if contains(lbs, lb) {
					err := Apicon.DelLabel(id, lb)
					if err != nil {
						termwindow.SetErr(err)
						continue
					}
				} else {
					err := Apicon.AddLabel(id, lb)
					if err != nil {
						termwindow.SetErr(err)
						continue
					}
				}

				termwindow.SetMsg("connect to %s/%s",
					Apicon.GetUrl(), Apicon.GetRepositoryName())
				var err error
				lbs, err = getActiveLabels(id)
				if err != nil {
					return err
				}
				win, err = getLabelWindow(lbs)
				if err != nil {
					return err
				}
				termwindow.SetBody(win.Data)
				termwindow.SetMsg("changed the status at label '%s' in this issue", lb)
			}
			switch ev.Ch {
			case 'j':
				termwindow.SetActiveLine(win.MvInc())
			case 'k':
				termwindow.SetActiveLine(win.MvDec())
			case 'G':
				termwindow.SetActiveLine(win.MvBottom())
			case 'g':
				termwindow.SetActiveLine(win.MvTop())
			case 'q':
				return nil
			}
		}
	}
}

func termMilestone(wk goctx.Worker, id string) error {
	defer wk.Done()

	mlname, err := getActiveMilestone(id)
	if err != nil {
		return err
	}
	win, err := getMilestonesWindow(mlname)
	if err != nil {
		return err
	}
	termwindow.SetBody(win.Data)
	defer termwindow.UnsetBody()

	for {
		select {
		case <-wk.RecvCancel():
			return nil
		case  ev := <-termwindow.Key:
			switch ev.Key {
			case termwindow.KeyCtrlN:
				termwindow.SetActiveLine(win.MvInc())
			case termwindow.KeyCtrlP:
				termwindow.SetActiveLine(win.MvDec())
			case termwindow.KeySpace:
				nml, _ := win.GetData(win.Active)
				if len(nml) < 0 {
					termwindow.SetErrStr("target not found")
					continue
				}
				if nml == mlname {
					err := Apicon.DeleteMilestone(id)
					if err != nil {
						termwindow.SetErr(err)
						continue
					}
				} else {
					err := Apicon.UpdateMilestone(id, nml)
					if err != nil {
						termwindow.SetErr(err)
						continue
					}
				}

				termwindow.SetMsg("connect to %s/%s",
					Apicon.GetUrl(), Apicon.GetRepositoryName())
				mlname, err = getActiveMilestone(id)
				var err error
				if err != nil {
					return err
				}
				win, err = getMilestonesWindow(mlname)
				if err != nil {
					return err
				}
				termwindow.SetBody(win.Data)
				termwindow.SetMsg("changed the status at milestone '%s' in this issue", nml)
			}
			switch ev.Ch {
			case 'j':
				termwindow.SetActiveLine(win.MvInc())
			case 'k':
				termwindow.SetActiveLine(win.MvDec())
			case 'G':
				termwindow.SetActiveLine(win.MvBottom())
			case 'g':
				termwindow.SetActiveLine(win.MvTop())
			case 'q':
				return nil
			}
		}
	}
}

func termBody(wk goctx.Worker, is termwindow.Window) {
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

func termCheckin(wk goctx.Worker) (apicon.Apicon, error) {
	defer wk.Done()

	var server termwindow.Window
	server.SetTitle("select and <enter> you want to use server name.")
	for alias, sv := range Conf.Server {
		server.Append(alias, []byte(msg.NewStr("%s : [%s] %s\n", alias, sv.User, sv.Url)))
	}
	alias, err := termSelect(wk.NewWorker(), server)
	if err != nil {
		return nil, err
	}
	if alias == "" {
		return nil, msg.NewErr("empty server alias")
	}

	var repo termwindow.Window
	repo.SetTitle("select and <enter> you want to use repository name.")
	for _, rpname := range Conf.Server[alias].Repos {
		repo.Append(rpname, []byte(rpname))
	}
	rpname, err := termSelect(wk.NewWorker(), repo)
	if err != nil {
		return nil, err
	}
	if rpname == "" {
		return nil, msg.NewErr("empty repository name")
	}

	napicon, err := apicon.NewApicon(Conf, alias)
	if err != nil {
		return nil, err
	}
	napicon.SetUsername(Conf.Server[alias].User)
	napicon.SetToken(Conf.Server[alias].Token)
	napicon.SetRepositoryName(rpname)
	napicon.SetUrl(Conf.Server[alias].Url)
	napicon.SetProxy(Conf.Server[alias].Proxy)
	return napicon, nil
}

func termSelect(wk goctx.Worker, win termwindow.Window) (string, error){
	defer wk.Done()

	termwindow.SetMenu(win.Data)
	defer termwindow.UnsetBody()
	for {
		select {
		case <-wk.RecvCancel():
			return "", nil
		case  ev := <-termwindow.Key:
			switch ev.Key {
			case termwindow.KeyCtrlN:
				termwindow.SetActiveLine(win.MvInc())
			case termwindow.KeyCtrlP:
				termwindow.SetActiveLine(win.MvDec())
			case termwindow.KeyEnter:
				id, _ := win.GetData(win.Active)
				if len(id) < 0 {
					termwindow.SetMsg("target not found")
					continue
				}
				return id, nil
			}
			switch ev.Ch {
			case 'j':
				termwindow.SetActiveLine(win.MvInc())
			case 'k':
				termwindow.SetActiveLine(win.MvDec())
			case 'G':
				termwindow.SetActiveLine(win.MvBottom())
			case 'g':
				termwindow.SetActiveLine(win.MvTop())
			case 'q':
				return "", nil
			}
		}
	}
	return "", nil
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
				if len(rbuf) > 0 {
					buf = string(rbuf[:(len(rbuf) - 1)])
				}
				continue
			case termwindow.KeyBackspace:
				rbuf := []rune(buf)
				if len(rbuf) > 0 {
					buf = string(rbuf[:(len(rbuf) - 1)])
				}
				continue
			case termwindow.KeySpace:
				ev.Ch = ' '
			case 0:
			default:
				continue
			}
			if ev.Ch == 0 {
				continue
			}
			buf += string(ev.Ch)
		}
	}
}



func termClose(id string) error {
	if err := Apicon.DoCloseIssue(id); err != nil {
		return err
	}
	return nil
}

func termOpen(id string) error {
	if err := Apicon.DoOpenIssue(id); err != nil {
		return err
	}
	return nil
}

func termCom(id string, com string) error {
	scomment := lf2Esclf(onlyLF(com))
	if err := Apicon.AddIssueComment(id, scomment); err != nil {
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

func termLs(closed bool) (termwindow.Window, error) {
	iss, err := Apicon.GetIssues(false, closed)
	if err != nil {
		return termwindow.Window{}, err
	}
	if len(iss) < 0 {
		return termwindow.Window{}, msg.NewErr("issue not found")
	}

	var win termwindow.Window
	for _, is := range iss{
		str, err := is.SprintHead()
		if err != nil {
			return termwindow.Window{}, err
		}
		win.Append(msg.NewStr("%v", is.Num), []byte(str))
	}
	return win, nil
}

func contains(arr []string, t string) bool {
	for _, v := range arr {
		if v == t {
			return true
		}
	}
	return false
}

func str0d0a2bytearr(str string) [][]byte {
	ls := strings.Split(onlyLF(str), "\n")

	var ret [][]byte
	for _, l := range ls {
		ret = append(ret, []byte(l))
	}
	return ret
}
