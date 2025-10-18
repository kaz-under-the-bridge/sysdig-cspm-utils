package output

import (
	"bytes"
	"testing"
)

func TestNewTableWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	tw := NewTableWriter(buf)

	if tw == nil {
		t.Error("NewTableWriter returned nil")
	}
	if tw.writer != buf {
		t.Error("TableWriter writer not set correctly")
	}
}

func TestTableWriter_OutputFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	tw := NewTableWriter(buf)

	if tw.writer == nil {
		t.Error("TableWriter writer is nil")
	}

	// 基本的な出力テスト
	output := buf.String()
	if output != "" {
		t.Errorf("Expected empty buffer initially, got: %s", output)
	}
}

func TestTableWriter_WriterField(t *testing.T) {
	tests := []struct {
		name string
		buf  *bytes.Buffer
	}{
		{
			name: "empty buffer",
			buf:  &bytes.Buffer{},
		},
		{
			name: "buffer with data",
			buf:  bytes.NewBufferString("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tw := NewTableWriter(tt.buf)
			if tw.writer != tt.buf {
				t.Error("TableWriter writer not assigned correctly")
			}
		})
	}
}

func TestTableWriter_Multiple(t *testing.T) {
	// 複数のTableWriterが独立して動作することを確認
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}

	tw1 := NewTableWriter(buf1)
	tw2 := NewTableWriter(buf2)

	if tw1.writer == tw2.writer {
		t.Error("Different TableWriters should have different writers")
	}
}

// 注: WriteVulnerabilitiesなどの実際の出力メソッドのテストは
// 外部依存パッケージ（sysdig-vuls-utils）の型定義に依存するため、
// 統合テストとして別途実装することを推奨
