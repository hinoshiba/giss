package issue

import (
	"time"
)

type Body struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	Url    string     `json:"url"`
	State  State      `json:"state"`
	Labels  []Label   `json:"labels"`
	Milestone Milestone `json:"milestone"`
	Update time.Time  `json:"updated_at"`
	User   User       `json:"user"`
	Assginees []Assgin   `json:"assignee"`
	Comments  []Comment `json:"comments"`
}

type State struct {
	Id	int64
	Name	string
}

type Comment struct {
	Id     int64      `json:"id"`
	Body   string     `json:"body"`
	Update time.Time  `json:"updated_at"`
	User   User       `json:"user"`
}

type Label struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type User struct {
	Id    int64  `json:"id"`
	Name string  `json:"username"`
	Email string `json:"email"`
}

type Milestone struct {
	Id     int64  `json:"id"`
	Title  string `json:"title"`
}

type Assgin struct {
	Id	int64
	Login	string
}
