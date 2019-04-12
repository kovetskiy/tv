package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/docopt/docopt-go"
)

var (
	version = "1.0"
	usage   = "tv " + version + `

Usage:
    tv [options]
    tv -h | --help
    tv --version

Options:
    -l <address>  Specify address listen to [default: :80]
    -s <dir>    Specify directory with static [default: static].
    -t <dir>      Specify directory with server templates [default: /srv/tv/templates].
    -h --help     Show this screen.
    --version     Show version.
`
)

func main() {
	args, _ := docopt.Parse(usage, nil, true, "1.0", false, true)

	tpl, err := template.ParseGlob(
		filepath.Join(args["-t"].(string), "*.template"),
	)
	if err != nil {
		log.Fatalln(err)
	}

	handler := &Handler{
		tpl: tpl,
	}

	fs := http.FileServer(http.Dir(args["-s"].(string)))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handler.ServeHTTP)

	err = http.ListenAndServe(
		args["-l"].(string),
		nil,
	)
	if err != nil {
		log.Fatalln(err)
	}
}
