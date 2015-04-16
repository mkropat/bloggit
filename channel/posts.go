package channel

import (
	"html/template"
	"time"
)

type Person struct {
	Name, Email string
}

type PostBlurb struct {
	Title   string
	Author  Person
	Created time.Time
	Slug    string
	Blurb   template.HTML
}

type ByCreated []PostBlurb

func (p ByCreated) Len() int           { return len(p) }
func (p ByCreated) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p ByCreated) Less(i, j int) bool { return p[i].Created.Before(p[j].Created) }

type PostModel struct {
	*PostBlurb
	IndexPath string
	Content   template.HTML
}

type ChannelStore interface {
	Index() ([]PostBlurb, error)
	Get(postSlug string) (PostModel, error)
}
