package util

import "testing"

func f(t *testing.T, s string, expected bool) {
	if IsJapanese(s) != expected {
		t.Fatal(s)
	}
}

func Test(t *testing.T) {
	f(t, "hello", false)
	f(t, "342352", false)
	f(t, "🌙", false)
	f(t, "사랑합니다", false)
	f(t, "０１２３４", false)
	f(t, "ｄｅｆｇｈ", false)

	f(t, "你好", true)
	f(t, "あ", true)
	f(t, "ABCコンテスト", true)
	f(t, "表現", true)
}
