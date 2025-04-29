package main

import (
	"fmt"
	"log"

	"github.com/yourusername/blogparser/internal/parser"
)

func main() {
	p := parser.New()
	blog, err := p.ParseFile("9994362.html")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Title: %s\n", blog.Title)
	fmt.Printf("Date: %s\n", blog.CreatedAt)
	fmt.Printf("Categories: %v\n", blog.Categories)
	fmt.Printf("Tags: %v\n", blog.Tags)
	fmt.Printf("Content length: %d chars\n", len(blog.Content))
	fmt.Printf("FirstImage: %s\n", blog.FirstImage)
}
