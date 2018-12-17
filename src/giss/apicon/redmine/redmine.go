package redmine

import (
	"fmt"
	"time"
	"bytes"
	"strings"
	"net/http"
	"io/ioutil"
	"encoding/xml"
	"giss/apicon/httpcl"
	"giss/apicon/issue"
)

type Redmine struct {
	url string
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
	StatusId      string	`xml:"status_id"`
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
	Id	int64	`xml:"id,attr"`
	Name	string	`xml:"name,attr"`
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

func (self *Redmine) GetRepositoryName() string {
	return self.repository
}

func (self *Redmine) SetRepositoryName(repo string) {
	self.setRepositoryName(repo)
}

func (self *Redmine) setRepositoryName(repo string) {
	self.repository = repo
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
		fmt.Printf("not login\n")
		return false
	}
	if self.user == "" {
		fmt.Printf("not login\n")
		return false
	}
	return true
}

func (self *Redmine) GetIssues(withclose bool) ([]issue.Body, error) {
	var iss []issue.Body

	tks, err := self.getIssues(withclose)
	if err != nil {
		return iss, err
	}

	for _, tk := range tks {
		iss = append(iss, ticket2Issue(tk))
	}
	return iss, nil
}

func (self *Redmine) getIssues(withclose bool) ([]Ticket, error) {
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
		u := url + "&offset=" + fmt.Sprintf("%v", offset)
		p++
		limit := p * 100
		u += "&limit=" + fmt.Sprintf("%v", limit)

		bret, rcode, err := self.reqHttp("GET", u, nil)
		if err != nil {
			return nil, err
		}
		if rcode != 200 {
			fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
			return nil, nil
		}

		var tks tickets
		if err := xml.Unmarshal(bret, &tks); err != nil {
			fmt.Printf("%s", err)
			return nil, err
		}
		if len(tks.Tk) < 1 {
			break
		}
		for _, v := range tks.Tk {
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (self *Redmine) GetIssue(num string) (issue.Body, error) {
	var is issue.Body

	i_tk, err := self.getIssue(num)
	if err != nil {
		return is, err
	}

	is = ticket2Issue(i_tk)
	return is, nil
}

func (self *Redmine) getIssue(num string) (Ticket, error) {
	var tk Ticket
	iurl := self.url + "/issues/" +
				num + ".xml?include=attachments,journals"

	tret, rcode, err := self.reqHttp("GET", iurl, nil)
	if err != nil {
		return tk, err
	}
	if rcode != 200 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return tk, nil
	}

	if err := xml.Unmarshal(tret, &tk); err != nil {
		return tk, err
	}
	return tk, nil
}

func (self *Redmine) CreateIssue(is issue.Body) error {
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

func (self *Redmine) ModifyIssue(is issue.Body) error {
	tk := issue2Tikect(is)
	etk := ticket2TicketE(tk)
	etk.StatusId = "2"
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
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return 0, 0, nil
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
		fmt.Printf("undefined ticket: %s\n", inum)
		return nil
	}
	if tk.Status.Id == targetId {
		fmt.Printf("this issue already state : %s\n", state)
		return nil
	}

	etk := ticket2TicketE(tk)
	etk.StatusId = fmt.Sprintf("%v", targetId)
	if err := self.updatePostIssue(&etk); err != nil {
		return err
	}

	ntk, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if tk.Status.Name == ntk.Status.Name {
		fmt.Printf("not update\n")
		return nil
	}

	fmt.Printf("state updated : %s\n", ntk.Status.Name)
	return nil
}

func (self *Redmine) postIssue(etk *TicketE) error {
	return self.httpReqIssue("POST", etk)
}

func (self *Redmine) updatePostIssue(etk *TicketE) error {
	return self.httpReqIssue("PUT", etk)
}

func (self *Redmine) httpReqComment(method string , inum string, body string) error {
	url := self.url + "/issues/" + inum + ".xml"
	cxml := `<issue><notes>`+ body + `</notes></issue>`

	_, rcode, err := self.reqHttp(method, url, []byte(cxml))
	if err != nil {
		return err
	}
	if rcode != 200 {
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return nil
	}

	fmt.Printf("comment added : #%v\n", inum)
	return nil
}

func (self *Redmine) httpReqIssue(method string, etk *TicketE) error {
	url := self.url + "/projects/" + self.repository + "/issues.xml"
	if etk.Id != 0 {
		url = self.url + "/issues/" +
				fmt.Sprintf("%v", etk.Id) + ".xml"
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
		fmt.Printf("detect exceptional response. httpcode:%v\n", rcode)
		return nil
	}

	if rcode == 201 {
		if err := xml.Unmarshal(iret, &etk); err != nil {
			return err
		}
	}
	fmt.Printf("issue posted : #%v\n", etk.Id)
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

	client, err := httpcl.NewClient()
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

func ticket2Issue(tk Ticket) issue.Body {
	var nis issue.Body

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

func issue2Tikect(is issue.Body) Ticket {
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
	return ntk
}

func lf2Desclf(str string) string {
	return strings.NewReplacer(
		"\\n", "\n",
	).Replace(str)
}
