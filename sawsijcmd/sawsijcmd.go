package main

import (
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {

	listen := "localhost:8889"
	if len(os.Args) == 2 {
		listen = os.Args[1]
	}

	http.HandleFunc("/fmt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		ssrc := r.PostFormValue("src")
		fsrc, err := format.Source([]byte(ssrc))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		fmt.Fprint(w, string(fsrc))

	})

	http.HandleFunc("/suggest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		src := r.PostFormValue("src")
		location := r.PostFormValue("location")
		filename := r.PostFormValue("filename")
		log.Printf("Filename: %v - Location: %v", filename, location)
		cmd := exec.Command("gocode", "-f=json", "autocomplete", "-in="+filename)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Print(err)
			return
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Print(err)
			return
		}
		if err := cmd.Start(); err != nil {
			log.Print(err)
			return
		}
		go func() {
			_, err := io.WriteString(stdin, src)
			if err != nil {
				log.Print(err)
				return
			}
			stdin.Close()
		}()
		done := make(chan bool)
		go func() {
			_, err := io.Copy(os.Stdout, stdout)
			if err != nil {
				log.Print(err)
				return
			}
			stdout.Close()
			done <- true
		}()
		<-done
		if err := cmd.Wait(); err != nil {
			log.Print(err)
			return
		}
	})

	log.Print("Starting!")
	log.Fatal(http.ListenAndServe(listen, nil))

}
