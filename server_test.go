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
)

var env map[string]string
var configS *yaml.File
var configFileS = flag.String("cf", "./config_test.yaml", "path to config file")
var workDir string = ""
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
	envStrings := os.Environ()

	env = make(map[string]string)

	for _, envString := range envStrings {
		keyval := strings.Split(envString, "=")
		env[keyval[0]] = keyval[1]
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
	
	cy, err := yaml.ReadFile(*configFileS)
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

func TestConfigure(t *testing.T) {
	
	log.Print("Starting test")
	standup(t)
	
	 // Create a new AppSetup type     
    as := new(AppSetup)
    
    // Configure the application
	err := Configure(as,workDir)
	if err != nil {
	    t.Fatal(err)
	}
	
    teardown(t)
}

