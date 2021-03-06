package issue

import (
	"time"
	"fmt"
	"encoding/xml"
	"encoding/json"
	"encoding/hex"
	"github.com/aybabtme/rgbterm"
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
	labelstr, err := self.getLabelsStr()
	if err != nil {
		return err
	}
	fmt.Printf("# %d : %s %s\n", self.Num, self.Title, labelstr)
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

func (self *Issue) SprintMd() string {
	var ret string

	labelstr := ""
	for _, lb := range self.Labels {
		if lb.Name == "" {
			continue
		}
		labelstr += " " + lb.Name
	}

	ret += fmt.Sprintf("# %d : %s %s\n", self.Num, self.Title, labelstr)
	ret += fmt.Sprintf("## ( %s ) %s %s comments(%d)\n\n",
		self.State.Name, self.User.Name, self.Update, len(self.Comments))
	if len(self.Body) > 0 {
		ret += fmt.Sprintf("## Body #########################\n\n")
		ret += fmt.Sprintf("%s\n\n",self.Body)
	}
	for _, com := range self.Comments {
		ret += fmt.Sprintf("## Comment #%d %s %s #########################\n\n",
			com.Id, com.User.Name, com.Update)
		ret += fmt.Sprintf("%s\n\n",com.Body)
	}

	return ret
}

func (self *Label) GetLabelStr() (string, error) {
	return self.getLabelStr()
}

func (self *Label) getLabelStr() (string, error) {
		c, err := hex.DecodeString(self.Color)
		if err != nil {
			return "", err
		}
		if len(c) < 3 {
			c = []uint8{255,255,255}
		}
		l := rgbterm.String(self.Name,
			uint8(-c[0]), uint8(-c[1]), uint8(-c[2]),
			uint8(c[0]), uint8(c[1]), uint8(c[2]))
		return l, nil
}

func (self *Issue) getLabelsStr() (string, error) {
	var ret string

	for _, lb := range self.Labels {
		if lb.Name == "" {
			continue
		}
		l, err := lb.getLabelStr()
		if err != nil {
			return "", err
		}

		ret += " " + l + "\x1b[0m"
	}
	return ret, nil
}

func (self *Issue) SprintHead() (string, error) {
	labelstr, err := self.getLabelsStr()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(" %4s [ %-10s ] [ %-10s ] %s %s\n",
		fmt.Sprintf("#%d", self.Num),
		self.State.Name,
		self.Milestone.Title,
		self.Title,
		labelstr,
	), nil
}

func (self *Issue) PrintHead() error {
	labelstr, err := self.getLabelsStr()
	if err != nil {
		return err
	}

	fmt.Printf(" %4s [ %-10s ] %s %s\n",
		fmt.Sprintf("#%d", self.Num),
		self.Milestone.Title,
		self.Title,
		labelstr,
	)
	return nil
}
