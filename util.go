package sawsij

import (
	"strconv"
	"strings"
	"unicode"
	//"log"
)

const RT_HTML = 0
const RT_XML = 1
const RT_JSON = 2

func MakeDbName(fieldName string) string {
	runes := []rune(fieldName)
	copy := []rune{}
	usrunes := []rune("_")
	us := usrunes[0]
	for i := 0; i < len(runes); i++ {
		if i > 0 && unicode.IsUpper(runes[i]) {
			copy = append(copy, us)
		}
		runes[i] = unicode.ToLower(runes[i])
		copy = append(copy, runes[i])

	}
	return string(copy)
}

func GetIntId(strId string) (intId int64) {
	intId, err := strconv.ParseInt(strId, 0, 0)

	if err != nil {
		intId = -1
	}

	return
}

func MakeFieldName(dbName string) string {

	runes := []rune(dbName)
	copy := []rune{}
	usrunes := []rune("_")
	us := usrunes[0]
	for i := 0; i < len(runes); i++ {
		if runes[i] != us {
			if i == 0 {
				runes[i] = unicode.ToUpper(runes[i])
			}
			copy = append(copy, runes[i])
		} else {
			runes[i+1] = unicode.ToUpper(runes[i+1])
		}
	}
	return string(copy)
}

func GetUrlParams(pattern string, urlPath string) (urlParams map[string]string) {
	rp := strings.NewReplacer(pattern, "")
	restOfUrl := rp.Replace(urlPath)
	//log.Printf("URL rest: %v", restOfUrl)
	urlParams = make(map[string](string))
	if len(restOfUrl) > 0 && strings.Contains(restOfUrl, "/") {
		allUrlParts := strings.Split(restOfUrl, "/")
		//log.Printf("URL vars: %v", allUrlParts)
		if len(allUrlParts)%2 == 0 {
			for i := 0; i < len(allUrlParts); i += 2 {
				urlParams[allUrlParts[i]] = allUrlParts[i+1]
			}
		}
	}
	return
}

func GetReturnType(url string) (rt int, restOfUrl string) {
	jp := "/json"
	if strings.Index(url, jp) == 0 {
		jrp := strings.NewReplacer(jp, "")
		restOfUrl = jrp.Replace(url)
		rt = RT_JSON
	}

	xp := "/xml"
	if strings.Index(url, xp) == 0 {
		xrp := strings.NewReplacer(xp, "")
		restOfUrl = xrp.Replace(url)
		rt = RT_XML
	}

	if len(restOfUrl) == 0 {
		restOfUrl = url
		rt = RT_HTML
	}

	return
}

