package channel

import (
	"bufio"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/russross/blackfriday"
)

type FilesystemStore struct {
	Dir       string
	IndexPath string
}

func (c FilesystemStore) Get(postSlug string) (model PostModel, err error) {
	file := filepath.Join("posts", postSlug+".md")
	fi, _ := os.Stat(file)

	source, err := os.Open(file)
	if err == nil {
		parsed := parseMarkdownPost(source)

		model = PostModel{
			PostBlurb: &PostBlurb{
				Title:   firstDefined(parsed.Title, postSlug),
				Blurb:   renderMarkdown(parsed.BlurbMarkdown),
				Created: getCreateTime(fi),
			},

			IndexPath: c.IndexPath,
			Content:   renderMarkdown(parsed.ContentMarkdown),
		}
	}
	return
}

func (c FilesystemStore) Index() (posts []PostBlurb, err error) {
	files, err := lsMarkdownFiles(c.Dir)
	if err != nil {
		return
	}

	posts = make([]PostBlurb, 0, len(files))
	for _, file := range files {
		source, err := os.Open(filepath.Join(c.Dir, file.Name()))
		if err == nil {
			parsed := parseMarkdownPost(source)
			posts = append(posts, PostBlurb{
				Title:   firstDefined(parsed.Title, file.Name()),
				Author:  getOwner(file),
				Created: getCreateTime(file),
				Slug:    strings.Trim(file.Name(), ".md"),
				Blurb:   renderMarkdown(parsed.BlurbMarkdown),
			})
		}
	}
	sort.Sort(sort.Reverse(ByCreated(posts)))
	return
}

func lsMarkdownFiles(dir string) (files []os.FileInfo, err error) {
	listing, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	files = make([]os.FileInfo, 0, len(listing))
	for _, entry := range listing {
		if entry.Mode().IsRegular() && strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, entry)
		}
	}
	return
}

type parserState int

const (
	Start parserState = iota
	HasTitle
	InBlurb
	Content
)

type markdownPost struct {
	Title           string
	BlurbMarkdown   string
	ContentMarkdown string
}

func parseMarkdownPost(markdown io.Reader) (post markdownPost) {
	whitespaceMatcher, _ := regexp.Compile("^\\s*$")
	hruleMatcher, _ := regexp.Compile("^(?:(?:\\* *){3,}|(?:_ *){3,}|(?:- *){3,}) *$")

	state := Start

	goBlurb := func(line string) {
		post.BlurbMarkdown += line + "\n"
		state = InBlurb
	}

	goContent := func(line string) {
		post.ContentMarkdown += line + "\n"
		state = Content
	}

	s := bufio.NewScanner(markdown)
	for s.Scan() {
		line := s.Text()

		switch state {
		case Start:
			if !whitespaceMatcher.MatchString(line) {
				if header, isHeader := matchesAtxHeader(line); isHeader {
					post.Title = header
					state = HasTitle
				} else {
					goContent(line)
				}
			}

		case HasTitle:
			if !whitespaceMatcher.MatchString(line) {
				_, isHeader := matchesAtxHeader(line)
				if hruleMatcher.MatchString(line) || isHeader {
					goContent(line)
				} else {
					goBlurb(line)
				}
			}

		case InBlurb:
			if whitespaceMatcher.MatchString(line) {
				goContent(line)
			} else {
				goBlurb(line)
			}

		case Content:
			goContent(line)
		}
	}

	return
}

func firstDefined(strs ...string) string {
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return ""
}

func matchesAtxHeader(text string) (header string, wasMatch bool) {
	headerMatcher, _ := regexp.Compile("^#{1,6}(?: +|$)(.*)")
	trailingMatcher, _ := regexp.Compile(" +#+ *$")

	match := headerMatcher.FindStringSubmatch(text)
	wasMatch = len(match) >= 1
	if wasMatch {
		header = trailingMatcher.ReplaceAllString(match[1], "")
	}
	return
}

func maybeGetField(struct_ptr interface{}, name string) (field interface{}, hasField bool) {
	value := reflect.ValueOf(struct_ptr).Elem()
	type_ := value.Type()
	_, hasField = type_.FieldByName(name)
	if hasField {
		field = value.FieldByName(name).Interface()
	}
	return
}

func renderMarkdown(source string) template.HTML {
	rendered := blackfriday.MarkdownCommon([]byte(source))
	return template.HTML(rendered)
}
