# Blogparser

ブログファイル（HTML/Markdown）からメタデータと本文を抽出するGolangパッケージ

## 概要

このパッケージは、与えられたブログ記事ファイルから以下の情報を抽出します：

- タイトル
- 公開日時
- カテゴリ
- タグ
- 本文
- 最初に登場する画像（FirstImage）

HTML形式とMarkdown形式の両方に対応しています。

## ディレクトリ構造

```
blogparser/
├── main.go                # メインエントリーポイント
├── internal/
│   └── parser/
│       ├── parser.go        # パーサーのメインインターフェース
│       ├── title.go         # タイトル抽出ロジック
│       ├── date.go          # 公開日時抽出ロジック
│       ├── category.go      # カテゴリ抽出ロジック
│       ├── tag.go           # タグ抽出ロジック
│       ├── content.go       # 本文抽出ロジック
│       └── image.go         # アイキャッチ画像抽出ロジック
├── pkg/
│   └── models/
│       └── blog.go          # ブログ記事の構造体定義
├── test/
│   ├── testdata/           # テスト用のサンプルファイル
│   │   └── sample_blog.html
│   └── parser_test.go      # パーサーのテスト
├── go.mod                  # Goモジュール定義
├── go.sum                  # 依存関係のチェックサム
└── README.md               # このファイル
```

## 実装計画

1. **モデル定義**: ブログ記事のデータ構造を定義
2. **パーサーインターフェース**: 異なるフォーマットに対応するパーサーの共通インターフェース設計
3. **HTML/Markdownパーサー**: 具体的なファイル形式に応じたパース実装
4. **各要素の抽出ロジック**: タイトル、日付、カテゴリ、タグ、本文、画像それぞれの抽出処理
5. **エラー処理**: errorパッケージを使用したエラーハンドリング
   - パース時のエラー定義
   - ファイル読み込みエラー
   - 必須項目の欠落エラー
   - カスタムエラー型の定義
6. **CLIツール**: コマンドラインから利用可能なインターフェース実装
7. **テスト**: 単体テストとサンプルデータによる統合テスト
   - テーブル駆動テストを基本とし、テストケースの追加や変更を容易に
   - assertパッケージを使用した簡潔で読みやすいアサーション
   - エッジケースを含む豊富なテストケース
   - モック/スタブを活用した依存性の分離
   - テストカバレッジの計測と維持

## 主な依存ライブラリ

- [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html): HTMLパース
- [github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery): HTML操作
- [github.com/yuin/goldmark](https://github.com/yuin/goldmark): Markdownパース
- [github.com/pkg/errors](https://github.com/pkg/errors): エラーハンドリング
- [github.com/stretchr/testify](https://github.com/stretchr/testify): テストアサーションとモック

## モデル定義

ブログ記事を表現するための構造体は以下の通りです。

```go
type BlogPost struct {
    Title       string    // タイトル
    Author      string    // 著者名
    Content     string    // 本文
    Summary     string    // 要約
    Tags        []string  // タグ
    Categories  []string  // カテゴリ
    CreatedAt   time.Time // 作成日時
    UpdatedAt   time.Time // 更新日時
    Published   bool      // 公開フラグ
    Slug        string    // URL用スラッグ
    FirstImage  string    // 記事内で最初に登場する画像のURL
}
```

| フィールド名 | 型         | 説明                             |
| ------------ | ---------- | -------------------------------- |
| Title        | string     | タイトル                         |
| Author       | string     | 著者名                           |
| Content      | string     | 本文                             |
| Summary      | string     | 要約                             |
| Tags         | []string   | タグ                             |
| Categories   | []string   | カテゴリ                         |
| CreatedAt    | time.Time  | 作成日時                         |
| UpdatedAt    | time.Time  | 更新日時                         |
| Published    | bool       | 公開フラグ                       |
| Slug         | string     | URL用スラッグ                    |
| FirstImage   | string     | 記事内で最初に登場する画像のURL  |

## 使用例

### 単一ファイルの処理

```go
package main

import (
	"fmt"
	"log"

	"github.com/yourusername/blogparser/pkg/parser"
)

func main() {
	p := parser.New()
	blog, err := p.ParseFile("path/to/blog.html")
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
```

### 複数ファイルの一括処理

#### CLIツールの場合

```sh
blogparser parse ./blogs/*.html
```

複数のHTMLファイルをまとめて処理できます。

#### Goコードの場合

```go
package main

import (
	"fmt"
	"log"

	"github.com/yourusername/blogparser/pkg/parser"
)

func main() {
	p := parser.New()
	files := []string{"blog1.html", "blog2.html", "blog3.html"}
	blogs, err := p.ParseFiles(files)
	if err != nil {
		log.Fatal(err)
	}

	for _, blog := range blogs {
		fmt.Printf("Title: %s\n", blog.Title)
		fmt.Printf("Date: %s\n", blog.CreatedAt)
		fmt.Printf("Categories: %v\n", blog.Categories)
		fmt.Printf("Tags: %v\n", blog.Tags)
		fmt.Printf("Content length: %d chars\n", len(blog.Content))
		fmt.Printf("FirstImage: %s\n", blog.FirstImage)
		fmt.Println("----------------------")
	}
}
```

## 今後の拡張予定

- JSON/YAML出力対応
- メタデータのカスタムフィールド対応
- ブログプラットフォーム別のパーサー (WordPress, Hugo, etc.)

