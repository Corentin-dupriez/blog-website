package main

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Post struct {
	Title   string
	Slug    string
	Content template.HTML
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	// Get all md files from posts folder
	files, err := filepath.Glob("content/posts/*.md")

	if err != nil {
		http.Error(w, "Could not read posts", http.StatusInternalServerError)
		return
	}

	var posts []Post
	for _, f := range files {
		// Slug is set as file name without extension
		slug := strings.TrimSuffix(filepath.Base(f), ".md")

		// Create post objects from title of md files
		// Title: slug with first letter capital
		// Slug: slug
		posts = append(posts, Post{Title: strings.Title(slug),
			Slug: slug})
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	//Execute the template with found posts as context
	tmpl.Execute(w, posts)
}

func handlePosts(w http.ResponseWriter, r *http.Request) {

	slug := strings.TrimPrefix(r.URL.Path, "/post/")
	mdPath := filepath.Join("content/posts", slug+".md")

	fmt.Println("mdPath:", mdPath)

	source, err := os.ReadFile(mdPath)

	if err != nil {
		http.Error(w, "Could not read posts", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := goldmark.Convert(source, &buf); err != nil {
		http.Error(w, "Could not convert posts", http.StatusInternalServerError)
		return
	}

	post := Post{
		Title:   strings.Title(slug),
		Slug:    slug,
		Content: template.HTML(buf.String()),
	}

	fmt.Println("post:", post)

	tmpl := template.Must(template.ParseFiles("templates/post.html"))
	tmpl.Execute(w, post)

}

func handleAdmin(w http.ResponseWriter, r *http.Request) {

}

func main() {
	//redirect requests to static
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Index shows list of posts
	http.HandleFunc("/", handleIndex)

	// post shows single post
	http.HandleFunc("/post/", handlePosts)

	http.HandleFunc("/admin/new/", handleAdmin)

	fmt.Println("Server running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
