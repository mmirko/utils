package main

import (
	"bufio"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Debug(logline ...interface{}) {
	if *debug {
		log.Println("\033[35m[Debug]\033[0m -", logline)
	}
}

func Info(logline ...interface{}) {
	if *verbose || *debug {
		log.Println("\033[32m[Info]\033[0m  -", logline)
	}
}

func Warning(logline ...interface{}) {
	log.Println("\033[33m[Warn]\033[0m  -", logline)
}

func Alert(logline ...interface{}) {
	log.Println("\033[31m[Alert]\033[0m -", logline)
}

var (
	verbose = flag.Bool("v", false, "Verbose")
	debug   = flag.Bool("d", false, "Debug")
)

func init() {
	flag.Parse()
}

func main() {
	var files []string

	root := "."

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		Alert(err)
		return
	}

	for _, filen := range files {
		filenbase := strings.TrimSuffix(filen, ".go")
		Info("Processing " + filen)

		file, err := os.Open(filen)
		if err != nil {
			Alert(err)
			return
		}

		scanner := bufio.NewScanner(file)

		needprocessing := false
		debugcontent := "// +build debug\n\n"
		nodebugcontent := "// +build !debug\n\n"
		debugstate := false

		for scanner.Scan() {
			text := scanner.Text()
			if strings.HasPrefix(strings.TrimLeft(text, " \t"), "// +build GODEBUG") {
				needprocessing = true
			} else if strings.HasPrefix(strings.TrimLeft(text, " \t"), "// GODEBUGBEGIN") {
				debugstate = true
			} else if strings.HasPrefix(strings.TrimLeft(text, " \t"), "// GODEBUGEND") {
				debugstate = false
			} else {
				if !debugstate {
					nodebugcontent += text + "\n"
				}
				debugcontent += text + "\n"

			}
		}
		file.Close()
		if needprocessing {
			if err := ioutil.WriteFile(filenbase+"_debug.go", []byte(debugcontent), 0600); err != nil {
				Alert(err)
				return
			}
			if err := ioutil.WriteFile(filenbase+"_nodebug.go", []byte(nodebugcontent), 0600); err != nil {
				Alert(err)
				return
			}

		}
	}
}
