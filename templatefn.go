package sawsij

import (
	"fmt"
	"github.com/russross/blackfriday"
	"text/template"
	"time"
)

func MarkDown(raw string) (output string) {
	input := []byte(raw)
	bOutput := blackfriday.MarkdownCommon(input)
	output = string(bOutput)
	return
}

func DateFormat(t time.Time, layout string) (datestring string) {
	datestring = fmt.Sprintf("%q", t)
	return
}

func GetFuncMap() (fnm template.FuncMap) {
	fnm = make(template.FuncMap)
	fnm["dateformat"] = DateFormat
	fnm["markdown"] = MarkDown
	return

}

