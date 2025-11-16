package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Console struct {
	reader *bufio.Reader
}

func NewConsole() *Console {
	reader := bufio.NewReader(os.Stdin)
	return &Console{reader: reader}
}

func (c *Console) Read() (string, error) {
	v, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(v), nil
}

func (c *Console) ReadRequiredString(text string) (string, error) {
	fmt.Printf("%s: ", text)
	v, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(v), nil
}

func (c *Console) ReadRequiredInt(text string) (int, error) {
	fmt.Printf("%s: ", text)
	v, err := c.reader.ReadString('\n')
	if err != nil {
		return -1, err
	}
	v = strings.TrimSpace(v)
	intVal, err := strconv.Atoi(v)
	if err != nil {
		return -1, err
	}
	return intVal, nil
}
