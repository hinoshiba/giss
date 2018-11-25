package gitapi

import (
	"giss/cache"
	"time"
)

type Github struct {
	url string
	repo string
	user string
	token string
}

func (self *Github) GetRepo() string {
	return self.repo
}

func (self *Github) GetUrl() string {
	return self.url
}

func (self *Github) GetUser() string {
	return self.user
}

func (self *Github) GetToken() string {
	return self.token
}

func (self *Github) SetRepo(repo string) {
	self.repo = repo
}

//// tba

func (self *Github) IsLogined() bool {
	return false
}
func (self *Github) LoadCache(cache.Cache) bool {
	return false
}
func (self *Github) Login(string, string) error {
	return nil
}
func (self *Github) CreateIssue() error {
	return nil
}
func (self *Github) ModifyIssue(string) error {
	return nil
}
func (self *Github) AddIssueComment(string, []byte) error {
	return nil
}
func (self *Github) DoOpenIssue(string) error {
	return nil
}
func (self *Github) DoCloseIssue(string) error {
	return nil
}
func (self *Github) PrintIssues(int, bool) error {
	return nil
}
func (self *Github) PrintIssue(string, bool) error {
	return nil
}
func (self *Github) ReportIssues(time.Time) (map[string]string, error) {
	var tmp map[string]string
	return tmp, nil
}
