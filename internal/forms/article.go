package forms

import "strings"

type Article struct {
	Title   string
	Content string
	Errors  map[string]string
}

func (a *Article) Validate() bool {
	a.Errors = make(map[string]string)
	if strings.TrimSpace(a.Content) == "" {
		a.Errors["content"] = "Please enter article's content"
	}
	if strings.TrimSpace(a.Title) == "" {
		a.Errors["title"] = "Please enter article's title"
	}

	return len(a.Errors) == 0
}
