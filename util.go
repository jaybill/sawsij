// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package sawsij

import (
	"strconv"
	"strings"
	"unicode"
	"os"
)

// Return type constants, used in the switch for determining what format the response will be returned in.
const (
	RT_HTML = 0 // return HTML
	RT_XML  = 1 // return XML
	RT_JSON = 2 // return JSON
)

// MakeDbName converts a struct field name into a database field name. 
// The string will be converted to lowercase, and any capital letters after the first one will be prepended with an underscore.
// "FirstName" will become "first_name" and so on.
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

// GetIntId is a utility function for convertion a string into an int64. Useful for URL params.
func GetIntId(strId string) (intId int64) {
	intId, err := strconv.ParseInt(strId, 0, 0)

	if err != nil {
		intId = -1
	}

	return
}

// MakeDbName converts a database column name into a struct field name. 
// The first letter will be made capital, and underscores will be removed and the following letter made capital.
// "first_name" will become "FirstName", etc.
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

// GetUrlParams removes the string specified in "pattern" and returns key value pairs as a map of strings.
func GetUrlParams(pattern string, urlPath string) (urlParams map[string]string) {
	rp := strings.NewReplacer(pattern, "")
	restOfUrl := rp.Replace(urlPath)
	
	urlParams = make(map[string](string))
	if len(restOfUrl) > 0 && strings.Contains(restOfUrl, "/") {
		allUrlParts := strings.Split(restOfUrl, "/")
		
		if len(allUrlParts)%2 == 0 {
			for i := 0; i < len(allUrlParts); i += 2 {
				urlParams[allUrlParts[i]] = allUrlParts[i+1]
			}
		}
	}
	return
}

// GetReturnType takes a pattern and determines the type of response being requested. Currently, "/json" is the only one implemented.
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

// GetTemplateName takes a URL pattern and returns a template Id as a string.
func GetTemplateName(pattern string) (templateId string) {

	patternParts := strings.Split(pattern, "/")
	maxParts := len(patternParts)
	

	if strings.LastIndex(pattern, "/") == len(pattern)-1 && len(pattern) > 1 {
		maxParts = maxParts - 1
	}

	templateParts := make([]string, 0)
	for i := 0; i < maxParts; i++ {
		if i > 0 {
			if patternParts[i] != "" {
				templateParts = append(templateParts, patternParts[i])
			} else {
				templateParts = append(templateParts, "index")
			}
		}

	}
	templateId = strings.Join(templateParts, "-")

	return
}

// Determines if int "needle" is in the array "haystack". Returns true if it is, false if it isn't.
func InArray(needle int,haystack []int)(ret bool){
    ret = false
    
    for i := 0; i < len(haystack); i++ {
        if needle == haystack[i]{
            ret = true
            break
        }
    }    
    
    return
}

// Takes a string and a path and writes the string to the file specified by the path.
func WriteStringToFile(input string,filepath string) (err error){

 f, err := os.Create(filepath)
    if err != nil {
        return
    }

    defer f.Close()

    _, err = f.Write([]byte(input))
    if err != nil {
        return 
    }

    return 
}

