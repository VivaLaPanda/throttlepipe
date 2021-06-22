package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Various runtime flags
var id = flag.String("id", "defaultPipe", "ID for this task/pipe")
var tmpPath = flag.String("tmpPath", "/tmp/", "Where to store temp files")
var sleepTime = flag.Int("time", 1, "How long to 'sleep' in minutes")
var clearPipe = flag.Bool("rm", false, "Set this to true to clear the pipe id for later use")

func readPipefile(pipefilename string) (checkpoint time.Time, err error) {
	// The file already exists. Open it and store the last checkpoint
	pipefile, err := os.Open(pipefilename)
	if err != nil {
		return checkpoint, fmt.Errorf("could not open temp file %s: %s", pipefilename, err)
	}
	defer pipefile.Close()

	decoder := json.NewDecoder(pipefile)
	err = decoder.Decode(&checkpoint)

	return checkpoint, err
}

func writePipefile(pipefilename string, checkpoint time.Time) error {
	// The file already exists. Open it and store the last checkpoint
	pipefile, err := os.OpenFile(pipefilename, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("could not open temp file %s: %s", pipefilename, err)
	}
	defer pipefile.Close()

	encoder := json.NewEncoder(pipefile)
	err = encoder.Encode(checkpoint)

	return err
}

func initPipefile(pipefilename string) (lastTime time.Time, err error) {
	// Confirm we can interact with our data file for tracking pipe history
	_, err = os.Stat(pipefilename)
	if err == nil {
		// The file already exists. Open it and store the last checkpoint
		lastTime, err = readPipefile(pipefilename)
		if err != nil {
			err = fmt.Errorf("could not read from pipefile %s: %s", pipefilename, err)
			return
		}
	} else if os.IsNotExist(err) {
		// File doesn't exist. Make it, but don't do anything with it yet
		log.Printf("pipefile %s doesn't exist. Creating new pipefile", pipefilename)

		lastTime = time.Time{}
		err = writePipefile(pipefilename, lastTime)
	}

	return
}

// Copy the input from stdin to stdout until EOF or error
// TODO: checkpoint not just on EOF, but if stdin stops producing data at some
// bytes/second
func doPipe(w io.Writer, r io.Reader) error {
	_, err := io.Copy(w, r)
	if err != io.EOF {
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	pipefilename := filepath.Join(*tmpPath, "throttlepipe-"+*id)
	lastTime, err := initPipefile(pipefilename)
	if err != nil {
		log.Fatalf("couldn't init pipefile: %s", err)
	}

	// Prepare to pipe
	stdinReader := bufio.NewReader(os.Stdin)
	stdoutWriter := bufio.NewWriter(os.Stdout)

	// Check our checkpoint
	if time.Since(lastTime) < (time.Duration(*sleepTime) * time.Minute) {
		// It has been less than sleepTimeDuration since last checkpoint
		// Just terminate for now

		return
	}

	err = doPipe(stdoutWriter, stdinReader)
	if err != nil {
		log.Fatalf("encountered an error while piping data [timer not set]: %s", err)
	}

	// Make checkpoint
	writePipefile(pipefilename, time.Now())
}
