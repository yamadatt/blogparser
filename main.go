package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/yourusername/blogparser/internal/parser"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("ファイルパスを指定してください")
	}

	files := os.Args[1:]
	if len(files) == 0 {
		return errors.New("ファイルが指定されていません")
	}

	p := parser.New()
	posts, err := p.ParseFiles(files)
	if err != nil {
		return errors.Wrap(err, "ファイルの解析に失敗しました")
	}

	for _, post := range posts {
		fmt.Printf("タイトル: %s\n", post.Title)
		fmt.Printf("スラッグ: %s\n", post.Slug)
		fmt.Printf("要約: %s\n", post.Summary)
		fmt.Printf("作成日時: %s\n", post.CreatedAt)
		fmt.Printf("カテゴリ: %v\n", post.Categories)
		fmt.Printf("タグ: %v\n", post.Tags)
		fmt.Printf("本文の長さ: %d文字\n", post.Content)
		fmt.Printf("最初の画像: %s\n", post.FirstImage)
		fmt.Println("----------------------")
	}

	return nil
}
