package main

import (
	"net/http"

	"github.com/mkropat/bloggit/channel"
	"github.com/zenazn/goji"
)

func main() {
	goji.Get("/bower/*", http.StripPrefix("/bower", http.FileServer(http.Dir("bower_components"))))
	goji.Get("/themes/*", http.StripPrefix("/themes", http.FileServer(http.Dir("themes"))))

	ch := channel.Channel{
		Title:        "Posts",
		BaseUrlPath:  "/posts",
		TemplatesDir: "themes/default/templates",
		OpenStore:    func(indexPath string) channel.ChannelStore { return channel.FilesystemStore{"posts", indexPath} },
	}
	ch.RegisterRoutes(goji.DefaultMux)
	goji.Get("/", http.RedirectHandler("/posts/", http.StatusFound))

	goji.Serve()
}
