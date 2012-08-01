// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package framework

import (
	"archive/zip"

	"bufio"
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Return type constants, used in the switch for determining what format the response will be returned in.
const (
	RT_HTML = 5 // return HTML
	RT_XML  = 7 // return XML
	RT_JSON = 9 // return JSON
)

// Constants for how to return URL parameters
const (
	PARAMS_MAP   = 13
	PARAMS_ARRAY = 15
)

// GetIntId is a utility function for convertion a string into an int64. Useful for URL params.
func GetIntId(strId string) (intId int64) {
	intId, err := strconv.ParseInt(strId, 0, 0)

	if err != nil {
		intId = -1
	}

	return
}

// GetUrlParams removes the string specified in "pattern" and returns key value pairs as a map of strings.
func GetUrlParamsMap(pattern string, urlPath string) (urlParams map[string]string) {

	rp := strings.NewReplacer(pattern, "")
	restOfUrl := rp.Replace(urlPath)

	if strings.HasPrefix(restOfUrl, "/") {
		restOfUrl = restOfUrl[1:len(restOfUrl)]
	}

	urlParams = make(map[string](string))
	if len(restOfUrl) > 0 && strings.Contains(restOfUrl, "/") {
		allUrlParts := strings.Split(restOfUrl, "/")
		log.Printf("Url part count: %v", len(allUrlParts))
		if len(allUrlParts)%2 == 0 {
			for i := 0; i < len(allUrlParts); i += 2 {
				if i+1 < len(allUrlParts) {
					urlParams[allUrlParts[i]] = allUrlParts[i+1]
				}
			}
		} else {
			log.Printf("Uneven number of params: %+v", allUrlParts)
		}

	}

	return
}

// GetUrlParams removes the string specified in "pattern" and returns key value pairs as a map of strings.
func GetUrlParamsArray(pattern string, urlPath string) (urlParams []string) {
	rp := strings.NewReplacer(pattern, "")
	restOfUrl := rp.Replace(urlPath)

	if strings.HasPrefix(restOfUrl, "/") {
		restOfUrl = restOfUrl[1:len(restOfUrl)]
	}
	urlParams = strings.Split(restOfUrl, "/")
	return
}

func PasswordHash(password string, salt string) (hash string) {
	bHash, _ := bcrypt.GenerateFromPassword([]byte(salt+password), bcrypt.DefaultCost)
	hash = string(bHash)
	return
}

func CompareHashAndPassword(hash string, testPassword string, salt string) (valid bool) {

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(salt+testPassword))
	if err == nil {
		valid = true
	} else {
		valid = false
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
func InArray(needle int, haystack []int) (ret bool) {
	ret = false

	for i := 0; i < len(haystack); i++ {
		if needle == haystack[i] {
			ret = true
			break
		}
	}

	return
}

// Takes a string and a path and writes the string to the file specified by the path.
func WriteStringToFile(input string, filepath string) (err error) {

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

// Takes a string and a path and appends the string to the file specified by the path.
func AppendStringToFile(input string, filepath string) (err error) {

	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0744)
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

// Takes a url and a file path. Downloads the url to the path.
func CopyUrlToFile(url string, filepath string) (err error) {

	f, err := os.Create(filepath)

	if err != nil {
		return
	}
	defer f.Close()

	res, err := http.Get(url)
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	_, err = f.Write(data)
	if err != nil {
		return
	}

	return
}

// Takes a path to a zip file and a directory to extract to, extracts the zip file, creating directories as needed.
func UnzipFileToPath(zipfile string, path string) (err error) {
	// Open a zip archive for reading.
	r, err := zip.OpenReader(zipfile)
	if err != nil {
		return
	}
	defer r.Close()

	// Iterate through the files in the archive
	for _, f := range r.File {
		var filepath string = ""
		pathparts := strings.Split(f.Name, "/")

		filename := pathparts[len(pathparts)-1]

		for i := range pathparts {
			if i < len(pathparts)-1 {
				filepath += "/" + pathparts[i]
			}
		}

		filepath = path + filepath

		err = os.MkdirAll(filepath, os.FileMode(0777))
		if err != nil {
			return err
		}

		rc, err := f.Open()
		defer rc.Close()
		if err != nil {
			return err
		}
		outfile, err := os.Create(filepath + "/" + filename)
		if err != nil {
			return err
		}

		infile, err := ioutil.ReadAll(rc)
		if err != nil {
			return err
		}

		_, err = outfile.Write(infile)
		if err != nil {
			return err
		}

	}

	return

}

// Returns a fixed length random identifier. NOT guaranteed to be globally unique. Useful for generating temporary directory names.
func MakeRandomId() (ident string) {
	rand.Seed(time.Now().UnixNano())
	cr := strconv.FormatInt(int64(rand.Intn(999999999)+111111111), 10)
	h := md5.New()
	io.WriteString(h, cr)
	ident = fmt.Sprintf("%x", h.Sum(nil))
	return
}

// Creates a prompt and waits for input from the user. If a default answer is supplied, it will be returned 
// if the user presses enter without entering a value.
func GetUserInput(prompt string, defaultAnswer string) (answer string, err error) {
	var fmtPrompt string = ""

	if defaultAnswer == "" {
		fmtPrompt = "%v: %v"
	} else {
		fmtPrompt = "%v [%v]: "
	}

	fmt.Printf(fmtPrompt, prompt, defaultAnswer)
	rd := bufio.NewReader(os.Stdin)
	line, _, err := rd.ReadLine()

	if err != nil {
		return
	} else {
		answer = strings.TrimSpace(string(line))
		if answer == "" {
			answer = defaultAnswer
		}
	}

	return

}

// Copies file source to destination dest.
func CopyFile(source string, dest string) (err error) {
	//TODO This ought to be replaced with something that uses filepath.Walk() http://golang.org/pkg/path/filepath/#Walk (issue #2)

	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}

	return
}

// Recursively copies a directory tree, attempting to preserve permissions. Source directory must exist, 
// destination directory must *not* exist. 
func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return &SawsijError{"Source is not a directory"}
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return &SawsijError{"Destination already exists"}
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {

		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		// TODO Check for symlinks (issue #3)		
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy			
			err = CopyFile(sfp, dfp)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

// A struct for returning custom error messages
type SawsijError struct {
	What string
}

// Returns the error message defined in What as a string
func (e *SawsijError) Error() string {
	return e.What
}
