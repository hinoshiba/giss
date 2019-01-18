package gitea

import (
	"time"
	"bytes"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"giss/msg"
	"giss/apicon/httpcl"
	"giss/apicon/issue"
)

type Gitea struct {
	url string
	proxy string
	repository string
	user string
	token string
}

type iIssueE struct {
	Id     int64           `json:"id"`
	Num    int64           `json:"number"`
	Title  string          `json:"title"`
	Body   string          `json:"body"`
	State  string          `json:"state"`
	User   iIUser          `json:"user"`
	Labels []iILabel       `json:"labels, omitempty"`
	Update time.Time       `json:"updated_at"`
}

type iIssue struct {
	Id        int64        `json:"id"`
	Num       int64        `json:"number"`
	Title     string       `json:"title"`
	Body      string       `json:"body"`
	Url       string       `json:"url"`
	State     string       `json:"state"`
	Labels    []iILabel    `json:"labels, omitempty"`
	Milestone iIMilestone  `json:"milestone"`
	Update    time.Time    `json:"updated_at"`
	User      iIUser       `json:"user"`
	Assginees []iIAssgin   `json:"assignees"`
	Comments  []iIComment  `json:"com, omitempty"`
}

type iIComment struct {
	Id     int64      `json:"id"`
	Body   string     `json:"body"`
	Update time.Time  `json:"updated_at"`
	User   iIUser     `json:"user"`
}

type iILabel struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type iIUser struct {
	Id    int64  `json:"id"`
	Name string  `json:"username"`
	Email string `json:"email"`
}

type iIMilestone struct {
	Id     int64  `json:"id"`
	Title  string `json:"title"`
}

type iIAssgin struct {
	Id	int64
	Login	string
}

func (self *Gitea) GetUrl() string {
	return self.url
}

func (self *Gitea) SetUrl(url string) {
	self.setUrl(url)
}

func (self *Gitea) setUrl(url string) {
	self.url = url
}

func (self *Gitea) GetProxy() string {
	return self.proxy
}

func (self *Gitea) SetProxy(proxy string) {
	self.setProxy(proxy)
}

func (self *Gitea) setProxy(proxy string) {
	self.proxy = proxy
}

func (self *Gitea) GetRepositoryName() string {
	return self.repository
}

func (self *Gitea) SetRepositoryName(repo string) {
	self.setRepositoryName(repo)
}

func (self *Gitea) setRepositoryName(repo string) {
	self.repository = repo
}

func (self *Gitea) GetUsername() string {
	return self.user
}

func (self *Gitea) SetUsername(user string) {
	self.setUsername(user)
}

func (self *Gitea) setUsername(user string) {
	self.user = user
}

func (self *Gitea) GetToken() string {
	return self.token
}

func (self *Gitea) SetToken(token string) {
	self.setToken(token)
}

func (self *Gitea) setToken(token string) {
	self.token = token
}

func (self *Gitea) IsLogined() bool {
	return self.isLogined()
}

func (self *Gitea) isLogined() bool {
	if self.token == "" {
		return false
	}
	if self.user == "" {
		return false
	}
	return true
}

func (self *Gitea) GetIssues(com bool, withclose bool) ([]issue.Issue, error) {
	var iss []issue.Issue

	i_iss, err := self.getIssues(com, withclose)
	if err != nil {
		return iss, err
	}

	for _, v := range i_iss {
		iss = append(iss, iIssue2Issue(v))
	}
	return iss, nil
}

func (self *Gitea) getIssues(com bool, withclose bool) ([]iIssue, error) {
	url := self.url + "api/v1/repos/" + self.repository + "/issues?"
	if withclose {
		url = url + "&state=all"
	}
	var p int = 1
	var ret []iIssue
	for {
		u := url + "&page=" + msg.NewStr("%v",p)
		bret, rcode, err := self.reqHttp("GET", u, nil)
		if err != nil {
			return nil, err
		}
		if rcode != 200 {
			err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
			return nil, err
		}

		var iss []iIssue
		if err := json.Unmarshal(bret, &iss); err != nil {
			return nil, err
		}
		if len(iss) < 1 {
			break
		}
		p += 1
		for _, v := range iss {
			if com {
				var err error
				v, err = self.getIssue(msg.NewStr("%v", v.Num))
				if err != nil {
					return ret, err
				}
			}
			ret = append(ret, v)
		}
	}
	return ret, nil
}

func (self *Gitea) GetIssue(num string) (issue.Issue, error) {
	var is issue.Issue

	i_is, err := self.getIssue(num)
	if err != nil {
		return is, err
	}

	is = iIssue2Issue(i_is)

	return is, nil
}

func (self *Gitea) getIssue(num string) (iIssue, error) {
	iurl := self.url + "api/v1/repos/" + self.repository + "/issues/" + num
	curl := iurl + "/comments"

	iret, rcode, err := self.reqHttp("GET", iurl, nil)
	if err != nil {
		return iIssue{}, err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return iIssue{}, err
	}
	cret, rcode, err := self.reqHttp("GET", curl, nil)
	if err != nil {
		return iIssue{}, err
	}
	if rcode != 200 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return iIssue{}, err
	}

	var is iIssue
	if err := json.Unmarshal(iret, &is); err != nil {
		return iIssue{}, err
	}

	var coms []iIComment
	if err := json.Unmarshal(cret, &coms); err != nil {
		return iIssue{}, err
	}
	for _, com := range coms {
		is.Comments = append(is.Comments, com)
	}
	return is, nil
}

func (self *Gitea) CreateIssue(is issue.Issue) error {
	i_is := Issue2iIssue(is)
	i_ise := iIssue2iIssueE(i_is)
	return self.createIssue(i_ise)
}

func (self *Gitea) createIssue(ise iIssueE)  error {
	if err := self.postIssue(&ise); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) AddIssueComment(inum string, comment string) error {
	return self.addIssueComment(inum, comment)
}

func (self *Gitea) addIssueComment(inum string, comment string) error {
	if err := self.httpReqComment("POST", inum, comment); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) ModifyIssue(is issue.Issue) error {
	i_is := Issue2iIssue(is)
	i_ise := iIssue2iIssueE(i_is)
	return self.modifyIssue(i_ise)
}

func (self *Gitea) modifyIssue(ise iIssueE) error {
	if err := self.updatePostIssue(&ise); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) DeleteMilestone(inum string) error {
	return self.deleteMilestone(inum)
}

func (self *Gitea) deleteMilestone(inum string) error {

	ml := iIMilestone{Id:0,Title:""}

	if err := self.httpReqMilestone("PATCH", inum, ml); err != nil {
		return err
	}

	return nil
}

func (self *Gitea) UpdateMilestone(inum string, mlname string) error {
	return self.updateMilestone(inum, mlname)
}

func (self *Gitea) updateMilestone(inum string, mlname string) error {
	mls, err := self.getMilestones(mlname)
	if err != nil {
		return err
	}
	if len(mls) < 1 {
		err := msg.NewErr("undefined milestonename : %s\n", mlname)
		return err
	}

	if err := self.httpReqMilestone("PATCH", inum, mls[0]); err != nil {
		return err
	}

	return nil
}

func (self *Gitea) GetMilestones() ([]issue.Milestone, error) {
	var mls []issue.Milestone
	imls, err := self.getMilestones("")
	if err != nil {
		return mls, nil
	}

	for _, iml := range imls {
		mls = append(mls, iIMilestone2IssueMilestone(iml))
	}
	return mls, nil
}

func (self *Gitea) getMilestones(target string) ([]iIMilestone, error) {
	var mls []iIMilestone

	bret, err := self.httpGetMilestones()
	if err != nil {
		return mls, err
	}
	if err := json.Unmarshal(bret, &mls); err != nil {
		return mls, err
	}

	if target == "" {
		return mls, nil
	}
	for _, ml := range mls {
		if ml.Title == target {
			return []iIMilestone{ml}, nil
		}
	}
	return []iIMilestone{}, nil
}

func (self *Gitea) GetLabels() ([]issue.Label, error) {
	var rlbs []issue.Label

	lbs, err := self.getLabels("")
	if err != nil {
		return rlbs, err
	}

	for _, lb := range lbs {
		rlbs = append(rlbs, iILabel2IssueLabel(lb))
	}
	return rlbs, nil
}

func (self *Gitea) getLabels(target string) ([]iILabel, error) {
	var lbs []iILabel
	bret, err := self.httpGetLabel()
	if err != nil {
		return lbs, err
	}
	if err := json.Unmarshal(bret, &lbs); err != nil {
		return lbs, err
	}

	if target == "" {
		return lbs, nil
	}
	for _, lb := range lbs {
		if lb.Name == target {
			return []iILabel{lb}, nil
		}
	}
	return []iILabel{}, nil

}

func (self *Gitea) AddLabel(inum string, lbname string) error {
	lbs, err := self.getLabels(lbname)
	if err != nil {
		return err
	}
	if len(lbs) != 1 {
		err := msg.NewErr("undefined choose labelname. : %s\n", lbname)
		return err
	}

	if err := self.addLabel(inum, lbs[0]); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) addLabel(inum string, lb iILabel) error {
	if err := self.httpReqLabel("POST", inum, lb); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) DelLabel(inum string, lbname string) error {
	lbs, err := self.getLabels(lbname)
	if err != nil {
		return err
	}
	if len(lbs) != 1 {
		err := msg.NewErr("undefined choose labelname. : %s\n", lbname)
		return err
	}

	if err := self.delLabel(inum, lbs[0]); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) delLabel(inum string, lb iILabel) error {
	if err := self.httpReqLabel("DELETE", inum, lb); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) DoCloseIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doCloseIssue(inum)
}

func (self *Gitea) DoOpenIssue(inum string) error {
	if !self.isLogined() {
		return nil
	}
	return self.doOpenIssue(inum)
}

func (self *Gitea) doCloseIssue(inum string) error {
	if err := self.toggleIssueState(inum, "closed"); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) doOpenIssue(inum string) error {
	if err := self.toggleIssueState(inum, "open"); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) toggleIssueState(inum string, state string) error {
	if state != "open" && state != "closed" {
		err := msg.NewErr("unknown state :%s\n", state)
		return err
	}
	is, err := self.getIssue(inum)
	if err != nil {
		return err
	}
	if is.State == "" {
		err := msg.NewErr("undefined ticket: %s\n", inum)
		return err
	}
	if is.State == state {
		err := msg.NewErr("this issue already state : %s\n", state)
		return err
	}

	old := is.Update
	is.State = state
	eis := iIssue2iIssueE(is)
	if err := self.updatePostIssue(&eis); err != nil {
		return err
	}
	if old == is.Update {
		err := msg.NewErr("not update\n")
		return err
	}

	return nil
}

func (self *Gitea) postIssue(ise *iIssueE) error {
	return self.httpReqIssue("POST", ise)
}

func (self *Gitea) updatePostIssue(ise *iIssueE) error {
	return self.httpReqIssue("PATCH", ise)
}

func (self *Gitea) httpGetMilestones() ([]byte, error) {
	url := self.url + "api/v1/repos/" + self.repository + "/milestones"

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

func (self *Gitea) httpGetLabel() ([]byte, error) {
	url := self.url + "api/v1/repos/" + self.repository + "/labels"

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

func (self *Gitea) httpReqMilestone(method string , inum string, ml iIMilestone) error {
	url := self.url + "api/v1/repos/" + self.repository + "/issues/" + inum

	id := msg.NewStr("%v", ml.Id)
	json_str := `{"milestone":` + id + ` }`

	_, rcode, err := self.reqHttp(method, url, []byte(json_str))
	if err != nil {
		return err
	}
	if rcode != 201 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return err
	}
	return nil
}

func (self *Gitea) httpReqLabel(method string , inum string, lb iILabel) error {
	url := self.url + "api/v1/repos/" + self.repository +
						"/issues/" + inum + "/labels"
	id := msg.NewStr("%v", lb.Id)
	json_str := `{"labels":[`+ id + `]}`

	if method == "DELETE" {
		url += "/" + id
		json_str = ""
	}

	_, rcode, err := self.reqHttp(method, url, []byte(json_str))
	if err != nil {
		return err
	}
	if rcode != 200 && rcode != 204 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return err
	}
	return nil
}

func (self *Gitea) httpReqComment(method string , inum string, body string) error {
	url := self.url + "api/v1/repos/" + self.repository +
						"/issues/" + inum + "/comments"
	json_str := `{"Body":"`+ body + `"}`
	_, rcode, err := self.reqHttp(method, url, []byte(json_str))
	if err != nil {
		return err
	}
	if rcode != 201 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return err
	}
	return nil
}

func (self *Gitea) httpReqIssue(method string, ise *iIssueE) error {
	url := self.url + "api/v1/repos/" + self.repository + "/issues/"
	if ise.Num != 0 {
		url += msg.NewStr("%v", ise.Num)
	}

	ise.Update = time.Now()
	ijson, err := json.Marshal(*ise)
	if err != nil {
		return err
	}
	iret, rcode, err := self.reqHttp(method, url, []byte(ijson))
	if err != nil {
		return err
	}
	if rcode != 201 {
		err := msg.NewErr("detect exceptional response. httpcode:%v\n", rcode)
		return err
	}

	if err := json.Unmarshal(iret, &ise); err != nil {
		return err
	}
	return nil
}

func (self *Gitea) reqHttp(method, url string, param []byte ) ([]byte,
								int, error) {
    	req, err := http.NewRequest(
        	method,
        	url,
        	bytes.NewBuffer(param),
    	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token " + self.token)

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

func iIssue2Issue(is iIssue) issue.Issue {
	var nis issue.Issue

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.Url = is.Url
	nis.State = iState2IssueState(is.State)
	nis.Milestone = iIMilestone2IssueMilestone(is.Milestone)
	nis.Update = is.Update
	nis.User = iIUser2IssueUser(is.User)
	nis.Assginees = iIAssignees2IssueAssgin(is.Assginees)

	for _, label := range is.Labels {
		nis.Labels = append(nis.Labels, iILabel2IssueLabel(label))
	}

	for _, com := range is.Comments {
		nis.Comments = append(nis.Comments, iIComment2IssueComment(com))
	}

	return nis
}

func iState2IssueState(istate string) issue.State {
	var nstate issue.State

	nstate.Name = istate
	return nstate
}

func iIUser2IssueUser(user iIUser) issue.User {
	var nuser issue.User

	nuser.Id = user.Id
	nuser.Name = user.Name
	nuser.Email = user.Email
	return nuser
}

func iILabel2IssueLabel(label iILabel) issue.Label {
	var nlabel issue.Label

	nlabel.Id = label.Id
	nlabel.Name = label.Name
	nlabel.Color = label.Color
	return nlabel
}

func iIMilestone2IssueMilestone(mi iIMilestone) issue.Milestone {
	var nmi issue.Milestone

	nmi.Id = mi.Id
	nmi.Title = mi.Title
	return nmi
}

func iIComment2IssueComment(com iIComment) issue.Comment {
	var ncom issue.Comment

	ncom.Id = com.Id
	ncom.Body = com.Body
	ncom.Update = com.Update
	ncom.User = iIUser2IssueUser(com.User)
	return ncom
}

func iIAssignees2IssueAssgin(ass []iIAssgin) []issue.Assgin {
	var nass []issue.Assgin

	for _, v := range ass {
		var nas issue.Assgin
		nas.Id = v.Id
		nas.Login = v.Login
		nass = append(nass, nas)
	}
	return nass
}

func Issue2iIssue(is issue.Issue) iIssue {
	var nis iIssue

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.Url = is.Url
	nis.State = is.State.Name
	nis.Milestone = IssueMilestone2iIMilestone(is.Milestone)
	nis.Update = is.Update
	nis.User = IssueUser2iIUser(is.User)
	nis.Assginees = IssueAssignees2iIAssignees(is.Assginees)

	for _, label := range is.Labels {
		nis.Labels = append(nis.Labels, IssueLabel2iILabel(label))
	}

	for _, com := range is.Comments {
		nis.Comments = append(nis.Comments, IssueComment2iIComment(com))
	}
	return nis
}

func IssueUser2iIUser(user issue.User) iIUser {
	var nuser iIUser

	nuser.Id = user.Id
	nuser.Name = user.Name
	nuser.Email = user.Email
	return nuser
}

func IssueLabel2iILabel(label issue.Label) iILabel {
	var nlabel iILabel

	nlabel.Id = label.Id
	nlabel.Name = label.Name
	nlabel.Color = label.Color
	return nlabel
}

func IssueMilestone2iIMilestone(mi issue.Milestone) iIMilestone {
	var nmi iIMilestone

	nmi.Id = mi.Id
	nmi.Title = mi.Title
	return nmi
}

func IssueAssignees2iIAssignees(ass []issue.Assgin) []iIAssgin {
	var nass []iIAssgin

	for _, v := range ass {
		var nas iIAssgin
		nas.Id = v.Id
		nas.Login = v.Login
		nass = append(nass, nas)
	}
	return nass
}

func IssueComment2iIComment(com issue.Comment) iIComment {
	var ncom iIComment

	ncom.Id = com.Id
	ncom.Body = com.Body
	ncom.Update = com.Update
	ncom.User = IssueUser2iIUser(com.User)
	return ncom
}

func iIssue2iIssueE(is iIssue) iIssueE {
	var nis iIssueE

	nis.Id = is.Id
	nis.Num = is.Num
	nis.Title = is.Title
	nis.Body = is.Body
	nis.State = is.State
	nis.User  = is.User
	//nis.Milestone = is.Milestone.Id
	return nis
}
