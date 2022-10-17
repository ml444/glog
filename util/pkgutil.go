package util

import "strings"

func ParsePackageName(f string) (string, string) {
	slashIndex := strings.LastIndex(f, "/")
	if slashIndex > 0 {
		idx := strings.Index(f[slashIndex:], ".") + slashIndex
		return f[:idx], f[idx+1:]
	} else {
		lastPeriod := strings.LastIndex(f, ".")
		if lastPeriod > 0 {
			return f[:lastPeriod], f[lastPeriod+1:]
		}
	}
	return f, ""
}

/*
var f = "github.com/ml444/glog.Info"

func BenchmarkPkg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		util.ParsePackageName(f)
	}
}
func BenchmarkPkg2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		util.GetPackageName(f)
	}
}
BenchmarkParsePackageName    	141816116	         8.671 ns/op
BenchmarkGetPackageName   		61842654	         18.67 ns/op
func GetPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
*/
