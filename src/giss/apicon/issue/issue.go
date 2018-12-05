package issue

import (
	"time"
)

type Edited struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	State  string     `json:"state"`
	User   User       `json:"user"`
	Update time.Time  `json:"updated_at"`
	//Assgin string     `json:"assignee"`
}

type Body struct {
	Id     int64      `json:"id"`
	Num    int64      `json:"number"`
	Title  string     `json:"title"`
	Body   string     `json:"body"`
	Url    string     `json:"url"`
	State  string     `json:"state"`
//	Labels IssueLabel `json:"labels"`
	Milestone Milestone `json:"milestone"`
	Update time.Time  `json:"updated_at"`
	User   User  `json:"user"`
//	Assgin string     `json:"assignee"`
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
//	Color string `json:"color"`
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
