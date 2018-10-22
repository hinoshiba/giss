package mail

import (
	"net/smtp"
	"net/mail"
	"fmt"
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

func (self *Smtp) MakeMail ( to []string, sub string, b string) error {

	self.to = to
	self.subject = sub

	self.body = []byte(
		"To: " + slice2str(self.to) + "\r\n" +
		"Subject: " + self.subject + "\r\n" +
		"\r\n" +
		b )

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

func (self *Smtp) Send () error {
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
