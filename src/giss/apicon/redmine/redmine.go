package redmine

import (
	"time"
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/xml"
	"giss/msg"
	"giss/apicon/httpcl"
	"giss/apicon/issue"
)

type Redmine struct {
	url string
	proxy string
	repository string
	user string
	token string
}

type TicketE struct {
	XMLName       xml.Name	`xml:"issue"`
	Id            int64	`xml:"id"`
	Subject       string	`xml:"subject"`
	Description   string	`xml:"description"`
	Status        iStatus	`xml:"status"`
	User          iUser	`xml:"author"`
	StatusId      int64	`xml:"status_id"`
	TrackerId     int64     `xml:"tracker_id"`
	CategoryId    int64     `xml:"category_id"`
}

type Ticket struct {
	Id            int64      `xml:"id"`
	Subject       string     `xml:"subject"`
	Description   string     `xml:"description"`
	Status        iStatus    `xml:"status"`
	User          iUser      `xml:"author"`
	Update        time.Time  `xml:"created_on"`
	Category      iCategory  `xml:"category"`
	Tracker       iTracker   `xml:"tracker"`
	Comments      []tComment `xml:"journals>journal"`
}

type tComment struct {
	Id     int64      `xml:"id,attr"`
	Notes  string     `xml:"notes"`
	Update time.Time  `xml:"created_on"`
	User   iUser     `xml:"user"`
}

type iStatus struct {
	Id	int64	`xml:"id,attr"`
	Name	string	`xml:"name,attr"`
}

type iCategory struct {
	Id	int64	`xml:"id,attr"`
	Name	string	`xml:"name,attr"`
}

type iUser struct {
	Id	int64	`xml:"id,attr"`
	Name	string	`xml:"name,attr"`
}

type iTracker struct {
	Id	int64	 `xml:"id,attr"`
	Name	string	 `xml:"name,attr"`
}

func (self *Redmine) GetUrl() string {
	return self.url
}

func (self *Redmine) SetUrl(url string) {
	self.setUrl(url)
}

func (self *Redmine) setUrl(url string) {
	self.url = url
}

func (self *Redmine) GetProxy() string {
	return self.proxy
}

func (self *Redmine) SetProxy(proxy string) {
	self.setProxy(proxy)
}

func (self *Redmine) setProxy(proxy string) {
	self.proxy = proxy
}

func (self *Redmine) GetRepositoryName() string {
	return self.repository
}

func (self *Redmine) SetRepositoryName(repo string) {
	self.setRepositoryName(repo)
}

func (self *Redmine) setRepositoryName(repo string) {
	self.repository = strings.ToLower(repo)
}

func (self *Redmine) GetUsername() string {
	return self.user
}

func (self *Redmine) SetUsername(user string) {
	self.setUsername(user)
}

func (self *Redmine) setUsername(user string) {
	self.user = user
}

func (self *Redmine) GetToken() string {
	return self.token
}

func (self *Redmine) SetToken(token string) {
	self.setToken(token)
}

func (self *Redmine) setToken(token string) {
	self.token = token
}

func (self *Redmine) IsLogined() bool {
	return self.isLogined()
}

func (self *Redmine) isLogined() bool {
	if self.token == "" {
		return false
	}
	if self.user == "" {
		return false
	}
	return true
}

func (self *Redmine) GetIssues(com bool, withclose bool) ([]issue.Issue, error) {
	var iss []issue.Issue

	tks, err := self.getIssues(com, withclose)
	if err != nil {
		return iss, err
	}

	for _, tk := range tks {
		iss = append(iss, ticket2Issue(tk))
	}
	return iss, nil
}

func (self *Redmine) DeleteMilestone(inum string) error {
	return msg.NewErr("can't delete milestone(tracker) at the redmine\n")
}

func (self *Redmine) UpdateMilestone(inum string, mlname string) error {
	return self.updateTracker(inum, mlname)
}

func (self *Redmine) updateTracker(inum string, trname string) error {
	tk, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	etk := ticket2TicketE(tk)

	trs, err := self.getTrackers(trname)
	if err != nil {
		return err
	}
	if len(trs) < 1 {
		err := msg.NewErr("undefined name : %s\n", trname)
		return err
	}
	etk.TrackerId = trs[0].Id
	if err := self.updatePostIssue(&etk); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) GetMilestones() ([]issue.Milestone, error) {
	trs, err := self.getTrackers("")
	if err != nil {
		return []issue.Milestone{}, err
	}

	var mls []issue.Milestone
	for _, tr := range trs {
		mls = append(mls, tTracker2IssueMilestone(tr))
	}
	return mls, nil
}

func (self *Redmine) getTrackers(target string) ([]iTracker, error) {
	type Tracker struct {
		Id    int64     `xml:"id"`
		Name  string    `xml:"name"`
	}
	type Trackers struct {
		Trs  []Tracker `xml:"tracker"`
	}
	var trss Trackers

	bret, err := self.httpReqTrackers()
	if err != nil {
		return []iTracker{}, err
	}
	if err := xml.Unmarshal(bret, &trss); err != nil {
		return []iTracker{}, err
	}

	var trs []iTracker
	if target == "" {
		for _, tr := range trss.Trs {
			trs = append(trs, iTracker{Id:tr.Id, Name:tr.Name})
		}
		return trs, nil
	}
	for _, tr := range trss.Trs {
		if tr.Name == target {
			return []iTracker{iTracker{Id:tr.Id, Name:tr.Name}}, nil
		}
	}
	return []iTracker{}, nil
}

func (self *Redmine) GetLabels() ([]issue.Label, error) {
	cts, err := self.getCategories("")
	if err != nil {
		return []issue.Label{}, err
	}

	var lbs []issue.Label
	for _, ct := range cts {
		lbs = append(lbs, tCategory2IssueLabel(ct)[0])
	}
	return lbs, nil
}

func (self *Redmine) getCategories(target string) ([]iCategory, error){
	type Category struct {
		Id    int64     `xml:"id"`
		Name  string    `xml:"name"`
	}
	type Categories struct {
		Cts  []Category `xml:"issue_category"`
	}
	var ctss Categories

	bret, err := self.httpReqCategories()
	if err != nil {
		return []iCategory{}, err
	}
	if err := xml.Unmarshal(bret, &ctss); err != nil {
		return []iCategory{}, err
	}

	var cts []iCategory
	if target == "" {
		for _, ct := range ctss.Cts {
			cts = append(cts, iCategory{Id:ct.Id, Name:ct.Name})
		}
		return cts, nil
	}
	for _, ct := range ctss.Cts {
		if ct.Name == target {
			return []iCategory{iCategory{Id:ct.Id, Name:ct.Name}}, nil
		}
	}
	return []iCategory{}, err
}

func (self *Redmine) AddLabel(inum string, lb string) error {
	return self.modCategory(inum, lb)
}

func (self *Redmine) modCategory(inum string, ctname string) error {
	tk, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	etk := ticket2TicketE(tk)

	cts, err := self.getCategories(ctname)
	if err != nil {
		return err
	}
	if len(cts) < 1 {
		err := msg.NewErr("undefined name : %s\n", ctname)
		return err
	}
	etk.CategoryId = cts[0].Id
	if err := self.updatePostIssue(&etk); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) DelLabel(inum string, lb string) error {
	return self.delCategory(inum, lb)
}

func (self *Redmine) delCategory(inum string, ctname string) error {
	tk, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if tk.Category.Name != ctname {
		err := msg.NewErr("hasn't category\n")
		return err
	}
	etk := ticket2TicketE(tk)

	etk.CategoryId = 0
	if err := self.updatePostIssue(&etk); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) getIssues(com, withclose bool) ([]Ticket, error) {
	url := self.url + "/projects/" + self.repository +
				"/issues.xml?include=attachments,journals"
	if withclose {
		url = url + "&status_id=*"
	}

	var p int = 0
	var ret []Ticket
	type tickets struct {
		Tk []Ticket `xml:"issue"`
	}
	for {
		offset := p * 100
		u := url + "&offset=" + msg.NewStr("%v", offset)
		p++
		limit := p * 100
		u += "&limit=" + msg.NewStr("%v", limit)

		bret, rcode, err := self.reqHttp("GET", u, nil)
		if err != nil {
			return nil, err
		}
		if rcode != 200 {
			err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
			return nil, err
		}

		var tks tickets
		if err := xml.Unmarshal(bret, &tks); err != nil {
			return nil, err
		}
		if len(tks.Tk) < 1 {
			break
		}
		for _, v := range tks.Tk {
			if com {
				var err error
				v, err = self.getIssue(msg.NewStr("%v", v.Id))
				if err != nil {
					return nil, err
				}
			}
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (self *Redmine) GetIssue(num string) (issue.Issue, error) {
	var is issue.Issue

	i_tk, err := self.getIssue(num)
	if err != nil {
		return is, err
	}

	is = ticket2Issue(i_tk)
	return is, nil
}

func (self *Redmine) getIssue(num string) (Ticket, error) {
	iurl := self.url + "/issues/" +
				num + ".xml?include=attachments,journals"

	tret, rcode, err := self.reqHttp("GET", iurl, nil)
	if err != nil {
		return Ticket{}, err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return Ticket{}, err
	}

	var tk Ticket
	if err := xml.Unmarshal(tret, &tk); err != nil {
		return Ticket{}, err
	}
	return tk, nil
}

func (self *Redmine) CreateIssue(is issue.Issue) error {
	tk := issue2Tikect(is)
	etk := ticket2TicketE(tk)
	return self.createIssue(etk)

}

func (self *Redmine) createIssue(etk TicketE)  error {
	if err := self.postIssue(&etk); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) AddIssueComment(inum string, comment string) error {
	return self.addIssueComment(inum, lf2Desclf(comment))
}

func (self *Redmine) addIssueComment(inum string, comment string) error {
	if err := self.httpReqComment("PUT", inum, comment); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) ModifyIssue(is issue.Issue) error {
	tk := issue2Tikect(is)
	etk := ticket2TicketE(tk)
	return self.modifyIssue(etk)
}

func (self *Redmine) modifyIssue(etk TicketE) error {
	if err := self.updatePostIssue(&etk); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) DoCloseIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doCloseIssue(inum)
}

func (self *Redmine) DoOpenIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doOpenIssue(inum)
}

func (self *Redmine) doCloseIssue(inum string) error {
	if err := self.toggleIssueState(inum, "closed"); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) doOpenIssue(inum string) error {
	if err := self.toggleIssueState(inum, "open"); err != nil {
		return err
	}
	return nil
}

func (self *Redmine) getStateId() (int64, int64, error) {
	url := self.url + "/issue_statuses.xml"
	method := "GET"

	ret, rcode, err := self.reqHttp(method, url, nil)
	if err != nil {
		return 0, 0, err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return 0, 0, err
	}

	type state struct {
		Id	int64	`xml:"id"`
		Closed  bool	`xml:"is_closed"`
	}
	type is_status struct {
		XMLName	xml.Name	`xml:"issue_statuses"`
		Status	[]state		`xml:"issue_status"`
	}

	var sts is_status
	if err := xml.Unmarshal(ret, &sts); err != nil {
		return 0, 0, err
	}

	var open_id, close_id int64
	for _, st := range sts.Status {
		if open_id == 0 && st.Closed != true {
			open_id = st.Id
		}
		if close_id == 0 && st.Closed == true {
			close_id = st.Id
		}
		if open_id != 0 && close_id != 0 {
			break
		}
	}
	return open_id, close_id, nil
}

func (self *Redmine) toggleIssueState(inum string, state string) error {
	var targetId int64
	op, cl, err := self.getStateId()
	if err != nil {
		return err
	}
	if state == "closed" {
		targetId = cl
	}
	if state == "open" {
		targetId = op
	}
	if targetId == 0 {
		return nil
	}

	tk, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if tk.Status.Name == "" {
		err := msg.NewErr("undefined ticket: %s\n", inum)
		return err
	}
	if tk.Status.Id == targetId {
		err := msg.NewErr("this issue already state : %s\n", state)
		return err
	}

	etk := ticket2TicketE(tk)
	etk.StatusId = targetId
	if err := self.updatePostIssue(&etk); err != nil {
		return err
	}

	ntk, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if tk.Status.Name == ntk.Status.Name {
		err := msg.NewErr("not update\n")
		return err
	}

	return nil
}

func (self *Redmine) postIssue(etk *TicketE) error {
	return self.httpReqIssue("POST", etk)
}

func (self *Redmine) updatePostIssue(etk *TicketE) error {
	return self.httpReqIssue("PUT", etk)
}

func (self *Redmine) httpReqCategories() ([]byte, error) {
	url := self.url + "/projects/" + self.repository + "/issue_categories.xml"

	bret, rcode, err := self.reqHttp("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return nil, err
	}
	return bret, nil
}

func (self *Redmine) httpReqTrackers() ([]byte, error) {
	url := self.url + "trackers.xml"

	bret, rcode, err := self.reqHttp("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return nil, err
	}
	return bret, nil
}

func (self *Redmine) httpReqComment(method string , inum string, body string) error {
	url := self.url + "/issues/" + inum + ".xml"
	cxml := `<issue><notes>`+ body + `</notes></issue>`

	_, rcode, err := self.reqHttp(method, url, []byte(cxml))
	if err != nil {
		return err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return err
	}
	return nil
}

func (self *Redmine) httpReqIssue(method string, etk *TicketE) error {
	url := self.url + "/projects/" + self.repository + "/issues.xml"
	if etk.Id != 0 {
		url = self.url + "/issues/" +
				msg.NewStr("%v", etk.Id) + ".xml"
	}

	txml, err := xml.Marshal(*etk)
	if err != nil {
		return err
	}
	iret, rcode, err := self.reqHttp(method, url, []byte(txml))
	if err != nil {
		return err
	}
	if rcode != 200 && rcode != 201 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return err
	}

	if rcode == 201 {
		if err := xml.Unmarshal(iret, &etk); err != nil {
			return err
		}
	}
	return nil
}

func (self *Redmine) reqHttp(method, url string, param []byte ) ([]byte,
								int, error) {
    	req, err := http.NewRequest(
        	method,
        	url,
        	bytes.NewBuffer(param),
    	)
	req.Header.Set("Content-Type", "text/xml")
	req.Header.Set("X-Redmine-API-Key", self.token)

	client, err := httpcl.NewClient(self.proxy)
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
       		return nil, 0, err
    	}
    	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return bytes, resp.StatusCode, nil
}

func ticket2Issue(tk Ticket) issue.Issue {
	var nis issue.Issue

	nis.Id = tk.Id
	nis.Num = tk.Id
	nis.Title = tk.Subject
	nis.Body = tk.Description
	nis.Url = ""
	nis.State = tState2IssueState(tk.Status)
	nis.Labels = tCategory2IssueLabel(tk.Category)
	nis.Milestone = tTracker2IssueMilestone(tk.Tracker)
	nis.Update = tk.Update
	nis.User = tUser2IssueUser(tk.User)
	nis.Assginees = tUser2IssueAssgin(tk.User)
	for _, com := range tk.Comments {
		if com.Notes == "" {
			continue
		}
		nis.Comments = append(nis.Comments, tComment2IssueComment(com))
	}

	return nis
}

func tState2IssueState(ist iStatus) issue.State {
	var nst issue.State

	nst.Name = ist.Name
	nst.Id = ist.Id
	return nst
}

func tCategory2IssueLabel(ct iCategory) []issue.Label {
	var nlabel issue.Label

	nlabel.Id = ct.Id
	nlabel.Name = ct.Name

	return []issue.Label{nlabel}
}

func tUser2IssueUser(user iUser) issue.User {
	var nuser issue.User

	nuser.Id = user.Id
	nuser.Name = user.Name
	return nuser
}

func tUser2IssueAssgin(user iUser) []issue.Assgin {
	var nas issue.Assgin

	nas.Id = user.Id
	nas.Login = user.Name
	return []issue.Assgin{nas}
}

func tTracker2IssueMilestone(mi iTracker) issue.Milestone {
	var nmi issue.Milestone

	nmi.Id = mi.Id
	nmi.Title = mi.Name
	return nmi
}

func tComment2IssueComment(com tComment) issue.Comment {
	var ncom issue.Comment

	ncom.Id = com.Id
	ncom.Body = com.Notes
	ncom.Update = com.Update
	ncom.User = tUser2IssueUser(com.User)
	return ncom
}

func issue2Tikect(is issue.Issue) Ticket {
	var ntk Ticket

	ntk.Id = is.Id
	ntk.Subject = is.Title
	ntk.Description = is.Body
	ntk.Status = issueState2tStatus(is.State)
	ntk.Category = issueLabel2tCategory(is.Labels)
	ntk.Tracker = issueMilestone2tTracker(is.Milestone)
	ntk.Update = is.Update
	ntk.User = issueUser2tUser(is.User)
	for _, com := range is.Comments {
		ntk.Comments = append(ntk.Comments, issueComment2tComment(com))
	}
	return ntk
}

func issueState2tStatus(ist issue.State) iStatus {
	var nst iStatus

	nst.Name = ist.Name
	nst.Id = ist.Id
	return nst
}

func issueLabel2tCategory(labels []issue.Label) iCategory {
	var nct iCategory

	if len(labels) < 1 {
		return nct
	}

	nct.Id = labels[0].Id
	nct.Name = labels[0].Name
	return nct
}

func issueUser2tUser(user issue.User) iUser {
	var nuser iUser

	nuser.Id = user.Id
	nuser.Name = user.Name
	return nuser
}

func issueMilestone2tTracker(mi issue.Milestone) iTracker {
	var ntr iTracker

	ntr.Id = mi.Id
	ntr.Name = mi.Title
	return ntr
}

func issueComment2tComment(com issue.Comment) tComment {
	var ncom tComment

	ncom.Id = com.Id
	ncom.Notes = com.Body
	ncom.Update = com.Update
	ncom.User = issueUser2tUser(com.User)
	return ncom
}

func ticket2TicketE(tk Ticket) TicketE {
	var ntk TicketE

	ntk.Id = tk.Id
	ntk.Subject = tk.Subject
	ntk.Description = tk.Description
	ntk.Status = tk.Status
	ntk.User = tk.User
	ntk.StatusId  = tk.Status.Id
	ntk.TrackerId = tk.Tracker.Id
	ntk.CategoryId = tk.Category.Id
	return ntk
}

func lf2Desclf(str string) string {
	return strings.NewReplacer(
		"\\n", "\n",
	).Replace(str)
}
