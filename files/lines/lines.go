//Package lines processes lines in documents
package lines

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/google/logger"
)

const wrap = "\\"

//Unwrap allows me to wrap long lines into more readable shorted lines
//Example, instead of:
//1 {{- range key, value := zip (keys			"Rat"		"Pig"					"Monkey"			"Horse")		(values		$.HR		$.TeamLead		$.Marketing		$.Dev)
//2
//3{{- add $team.Members $key $value}}
//4{{- end}}
//I can write
//1 {{- range key, value := zip (keys			"Rat"		"Pig"					"Monkey"			"Horse") \\
//2															(values		$.HR		$.TeamLead		$.Marketing		$.Dev)
//3 {{- add $team.Members $key $value}}
//4 {{- end}}
//and "Unwrap" my templates before parsing them
//*Unrapping keeps line numbers
//Returns: path to a temp file with unwrapped content
//				 function to clean up temp files
//				 error if something went wrong
func Unwrap(filePath string) (newFilePath string, cleanUp func(), err error) {

	cleanUp = func() {} //don't return nul function

	text, err := readFile(filePath)
	if err != nil {
		return "", cleanUp, err
	}

	tmpFile, err := tempFile(filePath)
	if err != nil {
		return "", cleanUp, err
	}

	defer tmpFile.Close()

	cleanUp = func() {
		os.Remove(tmpFile.Name())
	}

	text = unwrapLinesInString(text, wrap)

	_, err = tmpFile.WriteString(text)

	if err != nil {
		message := fmt.Sprintf("Failed to write unwrapped text to: %s", tmpFile.Name())
		log.Warningf(message)
		return tmpFile.Name(), cleanUp, errors.New(message)
	}

	log.Infof("Successfuly unwrapped lines to temp file %s", tmpFile.Name())

	return tmpFile.Name(), cleanUp, nil
}

func readFile(filePath string) (text string, err error) {
	in, error := os.Open(filePath)
	if error != nil {
		message := fmt.Sprintf("Failed to open file: %s", filePath)
		log.Warningf(message)
		return "", errors.New(message)
	}
	defer in.Close()

	b, error := ioutil.ReadAll(in)
	if error != nil {
		message := fmt.Sprintf("Failed to read from file: %s", filePath)
		log.Warningf(message)
		return "", errors.New(message)
	}
	return string(b), nil
}

func tempFile(filePath string) (tmpFile *os.File, err error) {

	ext := filepath.Ext(filePath)

	tmpFilePattern := fmt.Sprintf("%s*%s", strings.TrimSuffix(filepath.Base(filePath), ext), ext)

	tmpFile, err = ioutil.TempFile("", tmpFilePattern)

	if err != nil {
		message := fmt.Sprintf("Failed to created a temp file: %s", tmpFilePattern)
		log.Warningf(message)
		return nil, errors.New(message)
	}
	log.Infof("Successfuly created temp file %s", tmpFile.Name())

	return tmpFile, nil
}

func unwrapLinesInString(text string, connector string) string {
	lines := strings.Split(text, "\n")

	for n := range lines {
		lines[n] = strings.TrimRight(lines[n], " \r\n\t")
		if strings.HasSuffix(lines[n], connector) {

			if n >= len(lines)-1 { //trim connector from last line
				lines[n] = strings.TrimSuffix(lines[n], connector)
				return strings.Join(lines, "\n")
			}

			first, next, last := n, n+1, n
			for current := first; strings.HasSuffix(lines[current], connector); {
				lines[next] = strings.TrimRight(lines[next], " \r\n\t")
				current, next, last = current+1, next+1, last+1
			}

			var lineBuilder strings.Builder
			lineBuilder.WriteString(strings.TrimSuffix(lines[first], connector))
			for i := first + 1; i <= last; i++ {
				lineBuilder.WriteString(strings.TrimLeft(strings.TrimSuffix(lines[i], connector), " \t"))
				lines[i] = ""
			}
			lines[first] = lineBuilder.String()
			lineBuilder.Reset()
		}
	}
	return strings.Join(lines, "\n")
}
