package parser

import "errors"

// パッケージ共通のエラー定義
var (
	// 一般的なエラー
	ErrEmptyContent = errors.New("コンテンツが空です")
	ErrParseHTML    = errors.New("HTMLのパースに失敗しました")

	// パーサー関連のエラー
	ErrTokenizer = errors.New("形態素解析器の初期化に失敗しました")
	ErrParsing   = errors.New("HTMLコンテンツのパースに失敗しました")
)
