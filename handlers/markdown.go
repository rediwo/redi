package handlers

import (
	"bytes"
	"net/http"

	"github.com/yuin/goldmark"
	"github.com/rediwo/redi/filesystem"
)

type MarkdownHandler struct {
	fs       filesystem.FileSystem
	markdown goldmark.Markdown
}

func NewMarkdownHandler(fs filesystem.FileSystem) *MarkdownHandler {
	return &MarkdownHandler{
		fs:       fs,
		markdown: goldmark.New(),
	}
}

func (mh *MarkdownHandler) Handle(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := mh.fs.ReadFile(route.FilePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		var htmlBuffer bytes.Buffer
		if err := mh.markdown.Convert(content, &htmlBuffer); err != nil {
			http.Error(w, "Error converting markdown", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(htmlBuffer.Bytes())
	}
}