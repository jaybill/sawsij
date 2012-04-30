package sawsij

import (
	"fmt"
	"github.com/russross/blackfriday"
	"text/template"
	"time"
)
// MarkDown parses a string in MarkDown format and returns HTML. Used primarly by the template parser as "markdown"
func MarkDown(raw string) (output string) {
	input := []byte(raw)
	bOutput := blackfriday.MarkdownCommon(input)
	output = string(bOutput)
	return
}
// DateFormat takes a time and a layout string and returns a string with the formatted date. Used primarily by the template parser as "dateformat"
func DateFormat(t time.Time, layout string) (datestring string) {
	datestring = fmt.Sprintf("%q", t)
	return
}

// GetFuncMap returns a template.FuncMap which will be passed to the template parser. 
func GetFuncMap() (fnm template.FuncMap) {
	fnm = make(template.FuncMap)
	fnm["dateformat"] = DateFormat
	fnm["markdown"] = MarkDown
	return

}
