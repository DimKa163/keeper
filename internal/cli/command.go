package cli

//type CommandHandler func(ctx *CLI) error
//
//type Command interface {
//	Determinate() error
//	Name() string
//	Get(argName string) (string, bool)
//}
//
//type command struct {
//	Raw  string
//	name string
//	args map[string]string
//}
//
//func Parse(raw string) Command {
//	return &command{
//		Raw:  raw,
//		args: make(map[string]string),
//	}
//}
//
//func (c *command) Determinate() error {
//	strs := strings.Split(c.Raw, " ")
//	c.name = strs[0]
//	for i := 1; i < len(strs)-1; i++ {
//		argName := strings.ReplaceAll(strs[i], "-", "")
//		if _, ok := c.args[argName]; ok {
//			return errors.New("argument already exists")
//		}
//		c.args[argName] = strs[i+1]
//	}
//	return nil
//}
//func (c *command) Name() string {
//	return c.name
//}
//
//func (c *command) Get(argName string) (string, bool) {
//	v, ok := c.args[argName]
//	return v, ok
//}
