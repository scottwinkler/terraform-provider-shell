package shell

import (
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
	for _, secret := range secretValues {
		s = strings.ReplaceAll(s, secret, "******")
	}
	return s
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
