package out

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var filename string

func init() {
	p, err := exePath()
	if err != nil {
		log.Fatalf("failed to determine current directory: %v", err)
	}

	filename = filepath.Join(filepath.Dir(p), "output.txt")
	exename := filepath.Base(p)
	LogString(fmt.Sprintf("%s: log file path %s", exename, filename))
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

// LogString will output the string to the logfile, including a timestamp.
// Returns an integer so that it can be used for variable initialization.
func LogString(s string) int {
	logfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		logfile.WriteString(time.Now().Format(time.StampMilli) + ": " + s + "\n")
		logfile.Close()
	}
	return 1
}
