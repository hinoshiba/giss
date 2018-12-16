package mail

import (
	"net/smtp"
	"net/mail"
	"encoding/base64"
	"fmt"
	"giss/values"
)

type Smtp struct {
	server	string
	body	[]byte
	to	[]string
	subject string
	from	string
}

func (self *Smtp) New(mta string, p int64, f string) error {

	self.server = fmt.Sprintf("%s:%v", mta, p)

	address, err := extractionAddress(f)
	if err != nil {
		return err
	}
	self.from = address

	return nil

}

func (self *Smtp) MakeMail ( header, to []string, sub string, b []byte) error {
	self.to = to
	self.subject = sub
	bb64 := slice2mlstr(strSplit(base64.StdEncoding.EncodeToString(b), 76))

	self.body = []byte(
		"To: " + slice2str(self.to) + "\r\n" +
		"Subject: " + encSubject(self.subject) +
		"Reply-To: " + self.from + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"X-Giss-Version: " + values.VersionText + "\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		slice2mlstr(header) + "\r\n" +
		"\r\n" +
		bb64 )
	return nil
}

func slice2str(sl []string) string {
	var str string
	for i, row := range sl {
		if i == 0 {
			str = row
			continue
		}
		str = str + "," + row
	}
	return str
}

func slice2mlstr(sl []string) string {
	var str string
	for i, row := range sl {
		if i == 0 {
			str = row
			continue
		}
		str = str + "\r\n" + row
	}
	return str
}

func (self *Smtp) Send() error {
	if err := smtp.SendMail(self.server, nil,
		self.from, self.to, self.body); err != nil {
			return err
	}
	return nil
}

func extractionAddress(headval string) (string, error){
	e, err  := mail.ParseAddress(headval)
	if err != nil {
		return "", err
	}
	return e.Address, nil
}

func encSubject(subject string) string{
	var ret string
	head := " =?utf-8?B?"
	fut := "?=\r\n"
	for _, r := range strSplit(subject, 13) {
		rb64 := base64.StdEncoding.EncodeToString([]byte(r))
		ret += head + rb64 + fut
	}
	return ret
}

func strSplit(s string, l int) []string {
	if l < 1 {
		return []string{s}
	}
	var ret []string
	rs := []rune(s)
	for i := 0; i < len(rs); i += l {
		if i + l < len(rs) {
			ret = append(ret, string(rs[i:(i+l)]))
			continue
		}
		ret = append(ret, string(rs[i:]))
	}
	return ret
}
