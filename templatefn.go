package sawsij

import (
	"time"
	"text/template"

	"github.com/russross/blackfriday"
)

func MarkDown(raw string) (output string) {
	input := []byte(raw)    
	bOutput := blackfriday.MarkdownCommon(input)	
	output = string(bOutput)
	return
}

//func DateFormat(t time.Time,layout string)

func GetFuncMap() (fnm template.FuncMap) {
	fnm = make(template.FuncMap)
    //fnm["dateformat"] = DateFormat
	fnm["markdown"] = MarkDown
	return

}

