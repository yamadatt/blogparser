# Blogparser

ブログ記事ファイル（HTML/Markdown）からメタデータ・本文・要約を抽出するGo言語パーサー

[Deepwiki](https://deepwiki.com/yamadatt/blogparser)

## 概要

このパッケージは、与えられたブログ記事ファイルから以下の情報を抽出します：

- タイトル
- 公開日時
- カテゴリ（複数対応・不要なプレフィックス除去）
- タグ（複数対応・重複除去）
- 本文（多様なセレクタ対応・クリーニング）
- 要約（BM25+形態素解析による自動生成）
- 最初に登場する画像（FirstImage）

HTML形式とMarkdown形式の両方に対応予定です（現状はHTML中心）。

## ディレクトリ構成

```
blogparser/
├── main.go                # メインエントリーポイント（サンプルCLI）
├── parser/
│   ├── parser.go          # パーサーのメインインターフェース・統合処理
│   ├── title.go           # タイトル抽出ロジック
│   ├── date.go            # 公開日時抽出ロジック
│   ├── category.go        # カテゴリ抽出ロジック
│   ├── tag.go             # タグ抽出ロジック
│   ├── content.go         # 本文抽出ロジック
│   ├── clean_content.go   # 本文クリーニングロジック
│   ├── image.go           # 画像抽出ロジック
│   ├── summary.go         # 要約生成ロジック
│   └── errors.go          # エラー定義
├── pkg/
│   └── models/
│       └── blog.go        # ブログ記事の構造体定義
├── sample/
│   └── main.go            # CLIサンプル（複数ファイル一括処理）
├── go.mod                 # Goモジュール定義
├── go.sum                 # 依存関係のチェックサム
└── README.md              # このファイル
```

## 主な機能・特徴

- **多様な抽出パターン対応**
  - タイトル: og:title, h1, titleタグ, meta[name=title], ld_blog_vars等
  - 日付: timeタグ, meta, JSON-LD, ld_blog_vars等
  - 本文: article, main, .content, .article, body等の多様なセレクタ
  - カテゴリ・タグ: 多様なセレクタ、ld_blog_vars、meta属性、class属性等
  - 画像: OGP画像、Twitter Card画像、imgタグ等
- **カテゴリ・タグのクリーニング**
  - 不要なプレフィックス（例:「テーマ：」）や重複の除去
  - タグとカテゴリが重複する場合の除外は今後の拡張予定
- **本文クリーニング**
  - script, style, iframe等の不要タグや広告・SNSボタン・コメント欄等の除去
  - 空白行の正規化、HTML整形
- **要約生成**
  - BM25スコア＋形態素解析（kagome）で本文から重要文を自動抽出
  - 300文字以内に要約を整形
- **エラー処理**
  - parser/errors.goで共通エラー定義（空コンテンツ・HTMLパース失敗・形態素解析失敗等）
  - 各抽出関数で詳細なエラー内容を返却
- **テスト方針**
  - assertパッケージ利用、テーブル駆動テストを重視
  - サンプルHTMLや実データでの動作確認

## モデル定義

ブログ記事を表現する構造体は以下の通りです（`pkg/models/blog.go`参照）。

```go
type BlogPost struct {
    Title      string    // タイトル
    Author     string    // 著者名
    Content    string    // 本文
    Summary    string    // 要約
    Tags       []string  // タグ
    Categories []string  // カテゴリ
    CreatedAt  time.Time // 作成日時
    UpdatedAt  time.Time // 更新日時
    Published  bool      // 公開フラグ
    Slug       string    // URL用スラッグ
    FirstImage string    // 記事内で最初に登場する画像のURL
}
```

| フィールド名 | 型         | 説明                             |
| ------------ | ---------- | -------------------------------- |
| Title        | string     | タイトル                         |
| Author       | string     | 著者名                           |
| Content      | string     | 本文                             |
| Summary      | string     | 要約（自動生成）                 |
| Tags         | []string   | タグ                             |
| Categories   | []string   | カテゴリ                         |
| CreatedAt    | time.Time  | 作成日時                         |
| UpdatedAt    | time.Time  | 更新日時                         |
| Published    | bool       | 公開フラグ                       |
| Slug         | string     | URL用スラッグ                    |
| FirstImage   | string     | 記事内で最初に登場する画像のURL  |

## 使用例

### CLIサンプル（複数ファイル一括処理）

```go
package main

import (
	"context"
	"fmt"
	"os"
	"log"

	"github.com/yamadatt/blogparser/parser"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("ファイルパスを指定してください")
	}
	files := os.Args[1:]
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
```

### Goコードからの利用（単一ファイル）

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yamadatt/blogparser/parser"
)

func main() {
	ctx := context.Background()
	p := parser.New()
	post, err := p.ParseFile(ctx, "path/to/blog.html")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("タイトル: %s\n", post.Title)
	fmt.Printf("要約: %s\n", post.Summary)
	fmt.Printf("作成日時: %s\n", post.CreatedAt)
	fmt.Printf("カテゴリ: %v\n", post.Categories)
	fmt.Printf("タグ: %v\n", post.Tags)
	fmt.Printf("本文の長さ: %d文字\n", len(post.Content))
	fmt.Printf("最初の画像: %s\n", post.FirstImage)
}
```

## 今後の拡張予定

- タグとカテゴリが同じ値の場合の重複除去
- JSON/YAML出力対応
- メタデータのカスタムフィールド対応
- ブログプラットフォーム別のパーサー（WordPress等）
- コンテンツクリーニング機能の強化
  - 不要なHTML要素の削除（script, style, iframe, 広告, SNS, コメント欄等）
  - 関連記事・プロモーション・重複コンテンツの除去
  - 空白行の正規化、HTML整形
- サマリ生成アルゴリズムの高度化
- テストカバレッジの向上

## ライセンス

MIT
