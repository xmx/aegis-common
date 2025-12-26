package jsvm

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/grafana/sobek/parser"
)

type tagMapper string

func (tm tagMapper) FieldName(t reflect.Type, f reflect.StructField) string {
	tag := f.Tag.Get(string(tm))
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		tag = tag[:idx]
	}
	if parser.IsIdentifier(tag) {
		return tag
	}

	return tm.lowerCase(f.Name)
}

func (tm tagMapper) MethodName(_ reflect.Type, m reflect.Method) string {
	return tm.lowerCase(m.Name)
}

// lowerCase 将 Go 可导出变量转为 JS 风格的变量。
//
//	HTTP -> http
//	MyHTTP -> myHTTP
//	CopyN -> copyN
//	N -> n
func (tagMapper) lowerCase(s string) string {
	runes := []rune(s)
	size := len(runes)
	for i, r := range runes {
		if unicode.IsLower(r) {
			break
		}
		next := i + 1
		if i == 0 ||
			next >= size ||
			unicode.IsUpper(runes[next]) {
			runes[i] = unicode.ToLower(r)
		}
	}

	return string(runes)
}
