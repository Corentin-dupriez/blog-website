package main

import (
	"bytes"
	"fmt"
	slug2 "github.com/gosimple/slug"
	"github.com/joho/godotenv"
	"github.com/yuin/goldmark"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Post struct {
	Title   string
	Slug    string
	Content template.HTML
}

type PostMetadata struct {
	Title string `yaml:"title"`
	Slug  string `yaml:"slug"`
	Date  string `yaml:"date"`
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

	content := string(source)
	var re = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)
	matches := re.FindStringSubmatch(content)

	if len(matches) < 2 {
		http.Error(w, "Could not find post", http.StatusInternalServerError)
		return
	}

	yamlPart := matches[1]

	var meta PostMetadata
	err = yaml.Unmarshal([]byte(yamlPart), &meta)
	if err != nil {
		http.Error(w, "Could not parse post", http.StatusInternalServerError)
		return
	}

	mdContent := re.ReplaceAllString(content, "")

	mdContentBytes := []byte(mdContent)

	var buf bytes.Buffer
	if err := goldmark.Convert(mdContentBytes, &buf); err != nil {
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

func handlePostUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not parse form", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	slug := slug2.Make(r.FormValue("title"))
	content := r.FormValue("content")
	date := r.FormValue("date")

	filePath := fmt.Sprintf("content/posts/%s.md", slug)
	md := fmt.Sprintf(`---
title: %s
date: %s
slug: %s
---

%s
`, title, date, slug, content)
	err := os.WriteFile(filePath, []byte(md), 0644)
	if err != nil {
		http.Error(w, "Could not upload post", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	http.HandleFunc(
		"/admin/create-post",
		basicAuth(adminUser, adminPass, handlePostUpload),
	)

	fmt.Println("Server running at localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
