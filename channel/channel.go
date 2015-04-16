package channel

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/ww/goautoneg"
	"github.com/gorilla/feeds"
	"github.com/russross/blackfriday"
	"github.com/zenazn/goji/web"
)

type Channel struct {
	Title        string
	Description  string
	BaseUrlPath  string
	TemplatesDir string
	OpenStore    func(indexPath string) ChannelStore
}

func (ch Channel) openStore() ChannelStore {
	return ch.OpenStore(ch.indexPath())
}

func (ch Channel) indexPath() string {
	return ch.BaseUrlPath + "/"
}

func (ch Channel) RegisterRoutes(mux *web.Mux) {
	mux.Get(ch.BaseUrlPath, permanentRedirect(ch.indexPath()))
	mux.Get(ch.indexPath(), indexHandler(ch))
	mux.Get(ch.BaseUrlPath+"/rss", rssHandler(ch))
	mux.Get(ch.BaseUrlPath+"/:slug", postHandler(ch))
}

type PageModel struct {
	Title string
}

type IndexModel struct {
	*PageModel
	Posts []PostBlurb
}

type indexHandler Channel

func (h indexHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.ServeHTTPC(web.C{}, rw, r)
}

func (h indexHandler) ServeHTTPC(c web.C, rw http.ResponseWriter, r *http.Request) {
	posts, err := Channel(h).openStore().Index()
	if err != nil {
		ise(rw)
		return
	}

	model := IndexModel{
		PageModel: &PageModel{
			Title: "Recent Posts",
		},
		Posts: posts,
	}

	switch negotiateAccept(r, jsonMimeType, htmlMimeType) {
	case jsonMimeType:
		renderToJson(model, rw)
	case htmlMimeType:
		renderIndexToHttp(Channel(h), model, rw)
	}
}

const jsonMimeType = "application/json"
const htmlMimeType = "text/html"

type handlerCallback func(ch Channel, model interface{}, rw http.ResponseWriter)

type postHandler Channel

func (h postHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.ServeHTTPC(web.C{}, rw, r)
}

func negotiateAccept(r *http.Request, alternatives ...string) string {
	header := strings.Join(r.Header["Accept"], ", ")
	return goautoneg.Negotiate(header, alternatives)
}

func (h postHandler) ServeHTTPC(c web.C, rw http.ResponseWriter, r *http.Request) {
	model, err := Channel(h).openStore().Get(c.URLParams["slug"])
	if err != nil {
		http.NotFound(rw, r)
		return
	}

	switch negotiateAccept(r, jsonMimeType, htmlMimeType) {
	case jsonMimeType:
		renderToJson(model, rw)
	case htmlMimeType:
		renderPostToHttp(Channel(h), model, rw)
	}
}

type rssHandler Channel

func (h rssHandler) ServeHTTPC(c web.C, rw http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(rw, r)
}

func (h rssHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/rss+xml")

	now := time.Now()
	feed := &feeds.Feed{
		Title:       h.Title,
		Link:        &feeds.Link{Href: getUrl(r, "")},
		Description: h.Description,
		Created:     now,
	}

	posts, err := Channel(h).openStore().Index()
	feed.Items = make([]*feeds.Item, len(posts))
	for i, post := range posts {
		feed.Items[i] = &feeds.Item{
			Title:       post.Title,
			Link:        &feeds.Link{Href: getUrl(r, h.BaseUrlPath+"/"+post.Slug)},
			Description: string(blackfriday.MarkdownCommon([]byte(post.Blurb))),
			Author:      &feeds.Author{post.Author.Name, post.Author.Email},
			Created:     post.Created,
		}
	}

	if err == nil {
		err = feed.WriteRss(rw)
	}
	if err != nil {
		ise(rw)
	}
}

func getUrl(r *http.Request, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if r.Host == "" {
		return path
	} else {
		return "http://" + r.Host + path
	}
}

func renderIndexToHttp(ch Channel, model IndexModel, rw http.ResponseWriter) {
	tmpl, err := ch.loadTemplates("layout.html", "index.html")
	if err != nil {
		ise(rw)
		return
	}

	render(rw, tmpl, model)
}

func renderToJson(model interface{}, rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(rw)
	enc.Encode(model)
}

func renderPostToHttp(ch Channel, model interface{}, rw http.ResponseWriter) {
	tmpl, err := ch.loadTemplates("layout.html", "post.html")
	if err != nil {
		ise(rw)
		return
	}

	render(rw, tmpl, model)
}

func contains(strs []string, target string) bool {
	for _, str := range strs {
		if str == target {
			return true
		}
	}
	return false
}

func render(rw http.ResponseWriter, tmpl *template.Template, model interface{}) (err error) {
	rw.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(rw, model)
	if err != nil {
		ise(rw)
	}
	return
}

func (ch Channel) loadTemplates(names ...string) (*template.Template, error) {
	paths := make([]string, len(names))
	for i, name := range names {
		paths[i] = filepath.Join(ch.TemplatesDir, name)
	}
	return template.ParseFiles(paths...)
}

func ise(rw http.ResponseWriter) {
	http.Error(
		rw,
		"An error occurred processing your request. Please try again later.",
		http.StatusInternalServerError)
}

func permanentRedirect(path string) http.Handler {
	return http.RedirectHandler(path, http.StatusMovedPermanently)
}
