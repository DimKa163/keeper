package cli

import ()

//type ConsoleLine struct {
//	reader *bufio.Reader
//}
//
//func NewConsole() *ConsoleLine {
//	reader := bufio.NewReader(os.Stdin)
//	return &ConsoleLine{reader: reader}
//}
//
//func (c *ConsoleLine) Read() (string, error) {
//	v, err := c.reader.ReadString('\n')
//	if err != nil {
//		return "", err
//	}
//	return strings.TrimSpace(v), nil
//}
//
//func (c *ConsoleLine) ReadRequiredString(text string) (string, error) {
//	fmt.Printf("%s: ", text)
//	v, err := c.reader.ReadString('\n')
//	if err != nil {
//		return "", err
//	}
//	return strings.TrimSpace(v), nil
//}
//
//func (c *ConsoleLine) ReadRequiredInt(text string) (int, error) {
//	fmt.Printf("%s: ", text)
//	v, err := c.reader.ReadString('\n')
//	if err != nil {
//		return -1, err
//	}
//	v = strings.TrimSpace(v)
//	intVal, err := strconv.Atoi(v)
//	if err != nil {
//		return -1, err
//	}
//	return intVal, nil
//}
