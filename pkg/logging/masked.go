package logging

import "log/slog"

// MaskedString はログ出力時にアスタリスクにマスキングされる文字列型です。
type MaskedString string

// LogValue は slog.LogValuer インターフェースを実装し、ログ出力時のマスク値を返します。
func (m MaskedString) LogValue() slog.Value {
	s := string(m)
	if len(s) == 0 {
		return slog.StringValue("")
	}
	if len(s) <= 8 {
		return slog.AnyValue("****")
	}

	// 先頭4文字と末尾3文字だけを残してマスク (例: "agf_****123")
	prefix := s[:4]
	suffix := s[len(s)-3:]
	return slog.StringValue(prefix + "****" + suffix)
}
