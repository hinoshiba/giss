package msg

import (
	"fmt"
	"errors"
)

func NewStr(s string, msg ...interface{}) string {
	return fmt.Sprintf(s , msg...)
}

func NewErr(s string, msg ...interface{}) error {
	return errors.New(fmt.Sprintf(s , msg...))
}
