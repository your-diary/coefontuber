package util

import "unicode"

func IsJapanese(s string) bool {
	for _, c := range s {
		if unicode.In(c, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return true
		}

	}
	return false
}
