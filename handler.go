package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/reconquest/executil-go"
	"github.com/reconquest/hierr-go"
)

type Handler struct {
	root string
	tpl  *template.Template
	cmd  *exec.Cmd
	fifo string
}

func (handler *Handler) ServeHTTP(
	writer http.ResponseWriter,
	request *http.Request,
) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")

	path := request.URL.Path

	command := request.URL.Query().Get("command")
	if command != "" {
		err := handler.command(command)
		if err != nil {
			fmt.Fprintln(writer, hierr.Errorf(
				err, "unable to write command: %s", command,
			))
		}
	}

	switch {
	case strings.HasSuffix(path, "/"):
		handler.HandleDir(writer, request, path)

	default:
		handler.HandleStart(writer, request, path)
	}
}

func (handler *Handler) command(cmd string) error {
	if handler.fifo == "" {
		return nil
	}

	if !isFileExists(handler.fifo) {
		return errors.New("fifo file does not exists")
	}

	err := ioutil.WriteFile(handler.fifo, []byte(cmd+"\n"), 0644)
	if err != nil {
		return hierr.Errorf(
			err, "unable to write command to file",
		)
	}

	return nil
}

func (handler *Handler) HandleDir(
	writer http.ResponseWriter,
	request *http.Request,
	dir string,
) {
	fullpath := filepath.Join(handler.root, dir)

	children, err := ioutil.ReadDir(fullpath)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(writer, hierr.Errorf(
			err, "unable to readdir: %s", fullpath,
		))
		return
	}

	err = handler.tpl.ExecuteTemplate(
		writer,
		"directory.template",
		map[string]interface{}{
			"dir":   strings.TrimSuffix(dir, "/") + "/",
			"files": children,
		},
	)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(writer, hierr.Errorf(
			err, "unable to render template",
		))
		return
	}
}

func (handler *Handler) HandleStart(
	writer http.ResponseWriter,
	request *http.Request,
	path string,
) {
	if !isFileExists(filepath.Join(handler.root, path)) {
		writer.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(writer, "file not found\n")
		return
	}

	err := handler.stop()
	if err != nil {
		fmt.Fprintln(writer, hierr.Errorf(
			err, "unable to stop player",
		))
	}

	fifo, err := mkfifo()
	if err != nil {
		fmt.Fprintln(writer, hierr.Errorf(
			err, "unable to create fifo",
		))

		return
	}

	handler.fifo = fifo

	cmd := exec.Command(
		"mplayer",
		"-quiet",
		"-slave",
		"-noconfig", "all",
		"-input", "file="+handler.fifo,
		filepath.Join(handler.root, path),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(writer, hierr.Errorf(
			err, "unable to start %s", path,
		))
		return
	}

	handler.cmd = cmd

}

func mkfifo() (string, error) {
	var filename string
	for {
		filename = filepath.Join(
			os.TempDir(),
			fmt.Sprintf("tv.fifo.%d", rand.Int()),
		)

		if !isFileExists(filename) {
			break
		}
	}

	cmd := exec.Command("mkfifo", filename)

	_, _, err := executil.Run(cmd)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (handler *Handler) stop() error {
	if handler.cmd != nil {
		if handler.cmd.Process != nil {
			err := handler.cmd.Process.Kill()
			if err != nil {
				return hierr.Errorf(
					err, "unable to stop process",
				)
			}

			handler.cmd = nil
		}
	}

	if handler.fifo != "" && isFileExists(handler.fifo) {
		err := os.RemoveAll(handler.fifo)
		if err != nil {
			return hierr.Errorf(
				err, "unable to remove fifo",
			)
		}

		handler.fifo = ""
	}

	return nil
}

func isFileExists(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}
