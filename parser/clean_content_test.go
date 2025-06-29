package parser

import (
	"strings"
	"testing"
)

func TestCleanContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "HTMLコメントの削除",
			input:    `<div><!-- コメント -->テスト</div>`,
			expected: `<div>テスト</div>`,
		},
		{
			name:     "複数行HTMLコメントの削除",
			input:    `<div><!-- \nコメント\n -->テスト</div>`,
			expected: `<div>テスト</div>`,
		},
		{
			name:     "scriptタグの削除",
			input:    `<div>本文</div><script>alert('x');</script>`,
			expected: `<div>本文</div>`,
		},
		{
			name:     "styleタグの削除",
			input:    `<div>本文</div><style>body{}</style>`,
			expected: `<div>本文</div>`,
		},
		{
			name:     "iframeタグの削除",
			input:    `<div>本文</div><iframe src='a'></iframe>`,
			expected: `<div>本文</div>`,
		},
		{
			name:     "不要なクラスの削除",
			input:    `<div class='google-auto-placed'>広告</div><div>本文</div>`,
			expected: `<div>本文</div>`,
		},
		{
			name:     "アメブロ特有要素の削除",
			input:    `<div class='skin-entryBody'><div class='adsbygoogle'>広告</div>本文</div>`,
			expected: `<div class="skin-entryBody">本文</div>`,
		},
		{
			name:     "bodyタグがない場合も全体HTML返却",
			input:    `<span>テスト</span>`,
			expected: `<span>テスト</span>`,
		},
		{
			name:     "空文字列はエラー",
			input:    ``,
			expected: ``,
			wantErr:  true,
		},
		{
			name:     "HTMLパースエラー",
			input:    `<div><span>`, // 閉じタグなし
			expected: `<div><span></span></div>`,
			wantErr:  false, // goqueryは自動補完するためエラーにはならない
		},
		{
			name:     "順位表記の削除",
			input:    `<div>１位：テスト</div>`,
			expected: `<div>テスト</div>`,
		},
	}

	parser := &HTMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.CleanContent(tt.input)
			result = strings.TrimSpace(result)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CleanContent() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Errorf("CleanContent() error = %v, want nil", err)
			}
			if result != tt.expected {
				t.Errorf("CleanContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanContentEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "非常に長いHTML",
			input:    `<div>` + strings.Repeat("テスト", 10000) + `</div>`,
			expected: `<div>` + strings.Repeat("テスト", 10000) + `</div>`,
			wantErr:  false,
		},
		{
			name:     "ネストしたscriptタグ",
			input:    `<div><script><script>alert('nested');</script></script>本文</div>`,
			expected: `<div>本文</div>`,
			wantErr:  false,
		},
		{
			name:     "複数のHTMLコメント",
			input:    `<div><!-- コメント1 -->本文<!-- コメント2 --></div>`,
			expected: `<div>本文</div>`,
			wantErr:  false,
		},
		{
			name:     "特殊文字を含むHTML",
			input:    `<div>&lt;&gt;&amp;&quot;&#39;</div>`,
			expected: `<div>&lt;&gt;&amp;&#34;&#39;</div>`,
			wantErr:  false,
		},
		{
			name:     "空白のみのHTML",
			input:    `   <div>   </div>   `,
			expected: `<div>   </div>`,
			wantErr:  false,
		},
		{
			name:     "bodyタグのみ",
			input:    `<body>本文</body>`,
			expected: `本文`,
			wantErr:  false,
		},
		{
			name:     "複数のadsbygoogleクラス",
			input:    `<div class="skin-entryBody"><div class="adsbygoogle">広告1</div><div>本文</div><div class="adsbygoogle">広告2</div></div>`,
			expected: `<div class="skin-entryBody"><div>本文</div></div>`,
			wantErr:  false,
		},
		{
			name:     "順位表記の複数パターン",
			input:    `<div>１位：テスト ３位：テスト</div>`,
			expected: `<div>テスト テスト</div>`,
			wantErr:  false,
		},
	}

	parser := &HTMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.CleanContent(tt.input)
			result = strings.TrimSpace(result)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("CleanContent() error = nil, want error")
				}
				return
			}
			
			if err != nil {
				t.Errorf("CleanContent() error = %v, want nil", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("CleanContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
