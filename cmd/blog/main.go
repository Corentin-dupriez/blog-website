package main

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/yuin/goldmark"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

	caser := cases.Title(language.English)
	var posts []Post
	for _, f := range files {
		// Slug is set as file name without extension
		slug := strings.TrimSuffix(filepath.Base(f), ".md")

		// Create post objects from title of md files
		// Title: slug with first letter capital
		// Slug: slug
		posts = append(posts, Post{Title: caser.String(slug),
			Slug: slug})
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	//Execute the template with found posts as context
	tmpl.Execute(w, posts)
}

func handlePosts(w http.ResponseWriter, r *http.Request) {

	caser := cases.Title(language.English)

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
		Title:   caser.String(slug),
		Slug:    slug,
		Content: template.HTML(buf.String()),
	}

	tmpl := template.Must(template.ParseFiles("templates/post.html"))
	tmpl.Execute(w, post)

}

func handleAdmin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/admin/admin.html"))
	tmpl.Execute(w, nil)
}

func handleNewPost(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/admin/new-post.html"))
	tmpl.Execute(w, nil)
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	adminUser := os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	//redirect requests to static
	http.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("static"))))

	// Index shows list of posts
	http.HandleFunc(
		"/",
		handleIndex)

	// post shows single post
	http.HandleFunc(
		"/post/",
		handlePosts)

	http.HandleFunc(
		"/admin/",
		basicAuth(adminUser, adminPass, handleAdmin),
	)

	http.HandleFunc(
		"/admin/new-post/",
		basicAuth(adminUser, adminPass, handleNewPost),
	)

	fmt.Println("Server running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
