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
	}{
		{
			name:     "HTMLコメントの削除",
			input:    `<div><!-- レスポンシブ -->テスト</div>`,
			expected: `<div>テスト</div>`,
		},
		{
			name:     "複数行HTMLコメントの削除",
			input:    "<div><!-- \nレスポンシブ\n -->テスト</div>",
			expected: "<div>テスト</div>",
		},
		{
			name:     "ネストされたHTMLコメントの削除",
			input:    "<div><!-- outer <!-- inner --> -->テスト</div>",
			expected: "<div>テスト</div>",
		},
	}

	parser := &HTMLParser{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.CleanContent(tt.input)
			result = strings.TrimSpace(result)
			if result != tt.expected {
				t.Errorf("CleanContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}
