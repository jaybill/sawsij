// Copyright 2012 J. William McCarthy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* Provides a command line tool for running sawsij applications. */

/*Useful for development, it will automatically compile application source and start the app. If the source changes, it will automatically recompile and restart.
makesawsij takes one optional argument, which is the base directory of the sawsij application. If this argument is absent, it will assume the current directory.
Check out http://sawsij.com for more information and documentation.

*/
package main

import (
	"crypto/md5"
	"fmt"
	"github.com/kylelemons/go-gypsy/yaml"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var basedir string

func main() {

	basedir = "."

	if len(os.Args) == 2 {
		basedir = strings.TrimSpace(os.Args[1])
	}

	configFilename := fmt.Sprintf("%v/etc/config.yaml", basedir)
	c, err := yaml.ReadFile(configFilename)
	doe(err)

	watchdir := fmt.Sprintf("%v/src", basedir)

	executable, err := c.Get("app.cmd")
	doe(err)

	fmt.Println("Compiling...")
	err = recompile(executable)
	if err == nil {
		poll(watchdir, executable)
	} else {
		doe(err)
	}

}

func poll(watchdir string, executable string) {
	cmd, p := start(executable)
	for {
		a, err := getFileList(watchdir)
		doe(err)
		time.Sleep(time.Second)
		b, err := getFileList(watchdir)
		if changed, changedFile := changes(a, b); changed {
			fmt.Printf("%v changed.\n", changedFile)

			if cmd != nil {
				stop(cmd, p)
			}

			fmt.Println("Recompiling...")
			err = recompile(executable)
			if err == nil {
				cmd, p = start(executable)
			} else {
				cmd = nil
				fmt.Println("Not restarting.")
			}
		}
	}
}

func start(executable string) (cmd *exec.Cmd, p *chan error) {
	cmd = exec.Command(executable, basedir)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	fmt.Printf("Starting up %v\n", executable)
	err := cmd.Start()
	doe(err)
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	p = &done
	return
}

func stop(cmd *exec.Cmd, p *chan error) {
	done := *p
	fmt.Println("Shutting down.")
	if err := cmd.Process.Kill(); err != nil {
		fmt.Println("Shutdown failed. ", err)
	}
	<-done // allow goroutine to exits
}

func recompile(executable string) (err error) {
	out, err := exec.Command("go", "install", executable).CombinedOutput()
	fmt.Printf("%s", out)
	if err != nil {
		fmt.Println("Compile failed.")
	} else {
		fmt.Println("Compile succeeded.")
	}
	return err
}

func changes(a map[string]string, b map[string]string) (changed bool, changedFile string) {

	changed = false

	// first, see if the number of entries is different
	if len(a) != len(b) {
		changed = true
		return
	}

	// See if any of the checksums have changed
	for filename, checksum := range a {

		if checksum != b[filename] {
			changedFile = filename
			changed = true
			return
		}

	}

	return

}

func getFileList(watchdir string) (filemap map[string]string, err error) {

	filemap = make(map[string]string)

	walk := func(path string, info os.FileInfo, inErr error) (err error) {

		if !info.IsDir() {

			file, err := os.Open(path)
			if err == nil {
				h := md5.New()
				io.Copy(h, file)
				checksum := fmt.Sprintf("%x", h.Sum(nil))
				filemap[path] = checksum
			} else {
				fmt.Println(err.Error)
			}
		}
		return
	}
	filepath.Walk(watchdir, walk)

	return

}

func doe(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func roe(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
