package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"

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

	//path := request.URL.Path

	command := request.URL.Query().Get("command")
	if command != "" {
		err := handler.command(command)
		if err != nil {
			fmt.Fprintln(writer, hierr.Errorf(
				err, "unable to write command: %s", command,
			))
		}
	}

	handler.HandleDir(writer, request)
}

func (handler *Handler) command(cmd string) error {
	_, _, err := executil.Run(exec.Command("xdotool", "key", cmd))
	if err != nil {
		return err
	}

	return nil
}

func (handler *Handler) HandleDir(
	writer http.ResponseWriter,
	request *http.Request,
) {
	err := handler.tpl.ExecuteTemplate(
		writer,
		"directory.template",
		nil,
	)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(writer, hierr.Errorf(
			err, "unable to render template",
		))
		return
	}
}

//func mkfifo() (string, error) {
//    var filename string
//    for {
//        filename = filepath.Join(
//            os.TempDir(),
//            fmt.Sprintf("tv.fifo.%d", rand.Int()),
//        )

//        if !isFileExists(filename) {
//            break
//        }
//    }

//    cmd := exec.Command("mkfifo", filename)

//    _, _, err := executil.Run(cmd)
//    if err != nil {
//        return "", err
//    }

//    return filename, nil
//}

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

	//if handler.fifo != "" && isFileExists(handler.fifo) {
	//    err := os.RemoveAll(handler.fifo)
	//    if err != nil {
	//        return hierr.Errorf(
	//            err, "unable to remove fifo",
	//        )
	//    }

	//    handler.fifo = ""
	//}

	return nil
}

func isFileExists(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}
