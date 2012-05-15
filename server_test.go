// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sawsij

import (
	"os"
	"strings"
	"testing"
	"flag"
	"github.com/kylelemons/go-gypsy/yaml"
)

var env map[string]string
var config *yaml.File
var configFile = flag.String("c", "./config_test.yaml", "path to config file")
var tmpdir string = ""
var workdir string = ""

func standup(t *testing.T) {
	envStrings := os.Environ()

	env = make(map[string]string)

	for _, envString := range envStrings {
		keyval := strings.Split(envString, "=")
		env[keyval[0]] = keyval[1]
	}

	c, err := yaml.ReadFile(*configFile)
	if err != nil {
		t.Fatal(err)
	}

	tmpdir, err := c.Get("server.tmpdir")
	if err != nil {
		t.Fatal(err)
	}

    //workdir = tmpdir + "/sawsijtestapp"

t.Log(tmpdir)

	t.Log(os.TempDir())

}

func Test(t *testing.T) {
	standup(t)

}

