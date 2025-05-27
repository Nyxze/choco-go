package choco

import "fmt"

func NewError(message string, args ...any) error {
	msg := fmt.Sprintf("[choco]:%s", message)
	return fmt.Errorf(msg, args)
}
