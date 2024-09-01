package command

import (
	"io"
)

func goDirection(cmd Command, writer io.Writer) (bool, error) {
	return true, nil
}
