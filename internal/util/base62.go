package util

import (
	"math"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func EncodeBase62(id int64) string {
	if id == 0 {
		return string(base62Chars[0])
	}

	var result []byte
	for id > 0 {
		result = append([]byte{base62Chars[id%62]}, result...)
		id /= 62
	}
	return string(result)
}

func DecodeBase62(code string) int64 {
	var id int64
	for i, c := range code {
		pos := strings.IndexRune(base62Chars, c)
		if pos == -1 {
			return -1
		}
		id += int64(pos) * int64(math.Pow(62, float64(len(code)-1-i)))
	}
	return id
}

var reservedWords = map[string]bool{
	"api": true, "login": true, "logout": true, "admin": true,
	"dashboard": true, "links": true, "stats": true, "health": true,
	"shorten": true, "expand": true, "auth": true, "register": true,
	"docs": true, "swagger": true, "static": true, "assets": true,
	"favicon.ico": true, "robots.txt": true, "sitemap.xml": true,
}

func IsReservedWord(code string) bool {
	lower := strings.ToLower(code)
	return reservedWords[lower]
}

func IsValidCustomCode(code string) bool {
	if len(code) < 1 || len(code) > 20 {
		return false
	}
	for _, c := range code {
		if !strings.ContainsRune(base62Chars, c) && c != '-' && c != '_' {
			return false
		}
	}
	if IsReservedWord(code) {
		return false
	}
	return true
}
