package util

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func ConstructFieldIndexMap(entry interface{}) map[string]int {
	dataT := reflect.TypeOf(entry)
	m := map[string]int{}
	for i := 0; i < dataT.NumField(); i++ {
		field := dataT.Field(i)

		key := field.Name
		m[key] = i + 1
	}
	return m
}

// parseFormatTemp
func parseFormatTemp(src string) string {
	regexpFieldMap := ConstructFieldIndexMap(nil)
	regexpReplaceFunc := func(s string) string {
		v, ok := regexpFieldMap[s]
		if !ok {
			panic(fmt.Sprintf("%s in config.Text.PatternStyle,But it isn't in the field of Entry", s))
		}
		return strconv.FormatInt(int64(v), 10)
	}
	var regexpPattern = regexp.MustCompile(`%\[(\w+)?\][sdfwvtq]`)
	var subRegexpPattern = regexp.MustCompile(`(\w+)?`)
	b := regexpPattern.FindAllStringSubmatch(src, -1)
	for _, b2 := range b {
		bb := subRegexpPattern.ReplaceAllStringFunc(b2[1], regexpReplaceFunc)
		ks := strings.Replace(b2[0], b2[1], bb, 1)
		src = strings.ReplaceAll(src, b2[0], ks)
	}
	return src
}

func JoinKV(key string, value interface{}) string {
	v, _ := value.(string)
	builder := strings.Builder{}
	builder.Grow(len(key) + len("=") + len(v))
	builder.WriteString(key + "=" + v)
	return builder.String()
}
