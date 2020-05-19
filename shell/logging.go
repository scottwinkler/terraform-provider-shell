package shell

import (
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/mitchellh/go-linereader"
)

func printStackTrace(actions []Action) {
	log.Printf("-------------------------")
	log.Printf("[DEBUG] Current stack:")
	for _, action := range actions {
		log.Printf("[DEBUG] -- %s", action)
	}
	log.Printf("-------------------------")
}

func logOutput(logCh chan string, secretValues []string) {
	for line := range logCh {
		sanitizedLine := sanitizeString(line, secretValues)
		log.Printf("  %s", sanitizedLine)
	}
}

func sanitizeString(s string, secretValues []string) string {
	newString := s
	for _, secret := range secretValues {
		replacement := strings.Repeat("*", len(s))
		newString = strings.ReplaceAll(newString, secret, replacement)
	}
	return newString
}

func readOutput(r io.Reader, logCh chan<- string, doneCh chan<- string) {
	defer close(doneCh)
	lr := linereader.New(r)
	var output strings.Builder
	for line := range lr.Ch {
		logCh <- line
		output.WriteString(line)
	}
	doneCh <- output.String()
}

func readFile(r io.Reader) string {
	const maxBufSize = 8 * 1024
	buffer := new(bytes.Buffer)
	for {
		tmpdata := make([]byte, maxBufSize)
		bytecount, _ := r.Read(tmpdata)
		if bytecount == 0 {
			break
		}
		buffer.Write(tmpdata)
	}
	return buffer.String()
}
