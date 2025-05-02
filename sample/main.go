package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/yamadatt/blogparser/parser"
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

	ctx := context.Background()
	p := parser.New()

	for _, file := range files {
		post, err := p.ParseFile(ctx, file)
		if err != nil {
			fmt.Printf("%s の解析に失敗: %v\n", file, err)
			continue
		}
		fmt.Printf("タイトル: %s\n", post.Title)
		fmt.Printf("スラッグ: %s\n", post.Slug)
		fmt.Printf("要約: %s\n", post.Summary)
		fmt.Printf("作成日時: %s\n", post.CreatedAt)
		fmt.Printf("カテゴリ: %v\n", post.Categories)
		fmt.Printf("タグ: %v\n", post.Tags)
		fmt.Printf("本文の長さ: %d文字\n", len(post.Content))
		fmt.Printf("最初の画像: %s\n", post.FirstImage)
		fmt.Println("----------------------")
	}

	return nil
}
