package core

type LoginPass struct {
	Name  string `json:"name"`
	Login string `json:"login"`
	Pass  string `json:"pass"`
	URL   string `json:"url"`
}
