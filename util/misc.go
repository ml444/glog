package util

import (
	"fmt"
	"github.com/ml444/glog/message"
	"regexp"
	"strconv"
	"strings"
)

// parseFormatTemp
func parseFormatTemp(src string) string {
	regexpFieldMap := message.ConstructFieldIndexMap()
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
