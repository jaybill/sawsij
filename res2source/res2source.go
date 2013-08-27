package main

import (
	"bitbucket.org/jaybill/sawsij/framework"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var resourceTemplate string = `package resources

func GetStaticResources() (r map[string]string) {

	r = map[string]	string{
		{{ range $s := .static }}"{{ $s.Name }}":  "{{ $s.Content }}",
		{{ end }}
	}
	return

}

func GetTemplateResources() (r map[string]string) {

	r = map[string]	string{
		{{ range $s := .templates }}"{{ $s.Name }}":  "{{ $s.Content }}",
		{{ end }}
	}
	return

}
`

func main() {

	log.Println("Converting resources to source code.")
	dir := strings.TrimSpace(os.Args[1])
	out := strings.TrimSpace(os.Args[2])

	type ContentResource struct {
		Name    string
		Content string
	}

	data := make(map[string][]ContentResource, 2)

	scs := make([]ContentResource, 0)
	sdir := fmt.Sprintf("%v/static", dir)
	files, err := ioutil.ReadDir(sdir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		f := file.Name()
		filename := fmt.Sprintf("%v/%v", sdir, f)
		log.Println(filename)
		file, err := framework.ReadFileIntoString(filename)
		if err != nil {
			log.Fatal(err)
		}

		fileb := []byte(file)
		file64 := base64.StdEncoding.EncodeToString(fileb)

		sc := ContentResource{f, file64}

		scs = append(scs, sc)

	}

	data["static"] = scs

	tcs := make([]ContentResource, 0)
	tdir := fmt.Sprintf("%v/templates", dir)
	files, err = ioutil.ReadDir(tdir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		f := file.Name()
		filename := fmt.Sprintf("%v/%v", tdir, f)
		log.Println(filename)
		file, err := framework.ReadFileIntoString(filename)
		if err != nil {
			log.Fatal(err)
		}

		fileb := []byte(file)
		file64 := base64.StdEncoding.EncodeToString(fileb)

		tc := ContentResource{f, file64}

		tcs = append(tcs, tc)

	}

	data["templates"] = tcs

	err = framework.ParseTemplate(resourceTemplate, data, out)
	if err != nil {
		log.Fatal(err)
	}

}
