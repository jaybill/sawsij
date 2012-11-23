// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package framework

import (
	//"fmt"
	"fmt"
	"github.com/russross/blackfriday"
	"strings"
	"text/template"
	"time"
)

// MarkDown parses a string in MarkDown format and returns HTML. Used by the template parser as "markdown"
func MarkDown(raw string) (output string) {
	input := []byte(raw)
	bOutput := blackfriday.MarkdownCommon(input)
	output = string(bOutput)
	return
}

// DateFormat takes a time and a layout string and returns a string with the formatted date. Used by the template parser as "dateformat"
func DateFormat(t time.Time, layout string) (datestring string) {
	datestring = t.Format(layout)
	return
}

// Compare is a quick and dirty comparison function. It will convert whatever you give it to strings and see if the two values are equal.
// Whitespace is trimmed. Used by the template parser as "eq"
func Compare(a, b interface{}) (equal bool) {
	equal = false
	if strings.TrimSpace(fmt.Sprintf("%v", a)) == strings.TrimSpace(fmt.Sprintf("%v", b)) {
		equal = true
	}
	return
}

// Returns the first "length" characters of "input". Used by the template parser as "truncate". Not super accurate
// when dealing with non-latin character sets. "cont" will be added to the end of a string if it has been shortened.
// The length of "cont" will be subtracted from the original string.
func Truncate(input string, length int, cont string) (output string) {
	if len(input) > length {
		output = input[0:length-len(cont)] + cont
	} else {
		output = input
	}

	return
}

// GetFuncMap returns a template.FuncMap which will be passed to the template parser. 
func GetFuncMap() (fnm template.FuncMap) {
	fnm = make(template.FuncMap)
	fnm["truncate"] = Truncate
	fnm["dateformat"] = DateFormat
	fnm["markdown"] = MarkDown
	fnm["eq"] = Compare
	return
}
