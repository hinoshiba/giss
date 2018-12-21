package issue

import (
	"time"
	"fmt"
	"encoding/xml"
	"encoding/json"
)

type Issue struct {
	Id     int64
	Num    int64
	Title  string
	Body   string
	Url    string
	State  State
	Labels  []Label
	Milestone Milestone
	Update time.Time
	User   User
	Assginees []Assgin
	Comments  []Comment
}

type State struct {
	Id	int64
	Name	string
}

type Comment struct {
	Id     int64
	Body   string
	Update time.Time
	User   User
}

type Label struct {
	Id    int64
	Name  string
	Color string
}

type User struct {
	Id    int64
	Name string
	Email string
}

type Milestone struct {
	Id     int64
	Title  string
}

type Assgin struct {
	Id	int64
	Login	string
}

func ImportJson(iss *[]Issue, j []byte) error {
	if err := json.Unmarshal(j, &iss); err != nil {
		return err
	}
	return nil
}

func ImportXml(iss *[]Issue, x []byte) error {
	if err := xml.Unmarshal(x, &iss); err != nil {
		return err
	}
	return nil
}

func ExportJson(iss *[]Issue) error {
	j, err := json.Marshal(iss)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}

func ExportXml(iss *[]Issue) error {
	j, err := xml.Marshal(iss)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}

func (self *Issue) PrintJson() error {
	j, err := json.Marshal(self)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", j)
	return nil
}

func (self *Issue) PrintXml() error {
	x, err := xml.Marshal(self)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", x)
	return nil
}

func (self *Issue) PrintWiki() error {
	fmt.Printf("h1. %d : %s \n", self.Num, self.Title)
	fmt.Printf("h2. ( %s ) %s %s comments(%d)\n\n",
		self.State.Name, self.User.Name, self.Update, len(self.Comments))
	if len(self.Body) > 0 {
		fmt.Printf("---------------\n\n")
		fmt.Printf("h2. Body \n\n")
		fmt.Printf("%s\n\n",self.Body)
	}

	for _, com := range self.Comments {
		fmt.Printf("---------------\n\n")
		fmt.Printf("h2. Comment #%d %s %s \n\n",
			com.Id, com.User.Name, com.Update)
		fmt.Printf("%s\n\n",com.Body)
	}

	return nil
}

func (self *Issue) PrintMd() error {
	fmt.Printf("# %d : %s \n", self.Num, self.Title)
	fmt.Printf("## ( %s ) %s %s comments(%d)\n\n",
		self.State.Name, self.User.Name, self.Update, len(self.Comments))
	if len(self.Body) > 0 {
		fmt.Printf("## Body #########################\n\n")
		fmt.Printf("%s\n\n",self.Body)
	}

	for _, com := range self.Comments {
		fmt.Printf("## Comment #%d %s %s #########################\n\n",
			com.Id, com.User.Name, com.Update)
		fmt.Printf("%s\n\n",com.Body)
	}

	return nil
}

func (self *Issue) PrintHead() error {
	fmt.Printf(" %4s [ %-10s ] %s\n",
		fmt.Sprintf("#%d", self.Num),
		self.Milestone.Title,
		self.Title,
	)
	return nil
}
