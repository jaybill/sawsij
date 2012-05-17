// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sawsij

import (
	"flag"
	"github.com/kylelemons/go-gypsy/yaml"
	"os"
	"strings"
	"testing"
	"log"
	"net/http"

)

var workDir string = ""
var sConfigFile = flag.String("cf", "./config_test.yaml", "path to config file")


func writeStringToFile(input string,filepath string) (err error){

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

func standup(t *testing.T) {

    var env map[string]string        
    var staticDir string = "/static"
    var templateDir string = "/templates"
    var etcDir string = "/etc"
    var dummyTemplate = "<% .val %>"
    var dummyConfigFile = `
    server:
      port: 8066
      cacheTemplates: false
        
    database:
      driver: @@DB_DRIVER@@
      connect: @@DB_CONNECT@@
        
    encryption:
      salt: 213asjdhaskjh213
      key: sakjdhuh23i123123
    `
    log.Print("Attempting to stand environment up...")

	envStrings := os.Environ()

	env = make(map[string]string)

	for _, envString := range envStrings {
		keyval := strings.Split(envString, "=")
		env[keyval[0]] = keyval[1]
	}

    if ! flag.Parsed(){
        flag.Parse()
    }

	workDir = os.TempDir() + "/sawsijtestapp"
	t.Log(workDir)

	err := os.RemoveAll(workDir)
	if err != nil {
		t.Fatal(err)
	}

	staticDir = workDir + staticDir
	templateDir = workDir + templateDir
	etcDir = workDir + etcDir

	err = os.MkdirAll(staticDir, os.FileMode(0777))
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.MkdirAll(templateDir, os.FileMode(0777))
	if err != nil {
		t.Fatal(err)
	}
	
	err = os.MkdirAll(etcDir, os.FileMode(0777))
	if err != nil {
		t.Fatal(err)
	}
	
	cy, err := yaml.ReadFile(*configFile)
	if err != nil {
		t.Fatal(err)
	}

	dbDriver, err := cy.Get("database.driver")
	if err != nil {
		t.Fatal(err)
	}
	
	dbConnect, err := cy.Get("database.connect")
	if err != nil {
		t.Fatal(err)
	}
    
    dummyConfigFile = strings.Replace(dummyConfigFile, "@@DB_DRIVER@@", dbDriver, -1)
    dummyConfigFile = strings.Replace(dummyConfigFile, "@@DB_CONNECT@@", dbConnect, -1)
	
	//t.Log(dummyConfigFile)
	
	err = writeStringToFile(dummyConfigFile,etcDir + "/config.yaml")
	if err != nil{
    	t.Fatal(err)
	}
		
	err = writeStringToFile(dummyTemplate,templateDir + "/dummy.html")
	if err != nil{
    	t.Fatal(err)
	}			
    
}

func teardown(t *testing.T) {
	err := os.RemoveAll(workDir)
	if err != nil {
		t.Fatal(err)
	}
}

func testHandler(r *http.Request, a *AppScope, rs *RequestScope) (h HandlerResponse, err error) {


    return
}

func TestRouteAndConfigure(t *testing.T) {
	
	
	standup(t)	
    as := new(AppSetup)
	err := Configure(as,workDir)
	if err != nil {
	    t.Fatal(err)
	}
	Route(RouteConfig{Pattern: "/", Handler: testHandler, Roles: make([]int,0)})
    teardown(t)
}
