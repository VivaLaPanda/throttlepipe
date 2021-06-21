package main

import (
	"bufio"
	"encoding/gob"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Various runtime flags
var id = flag.String("id", "./templates", "ID for this task/pipe")
var tmpPath = flag.String("tmpPath", "/tmp/", "Where to store temp files")
var sleepTime = flag.Int("time", 1, "How long to 'sleep' in minutes")
var clearPipe = flag.Bool("rm", false, "Set this to true to clear the pipe id for later use")

func readPipefile(pipefile io.Reader) (checkpoint time.Time, err error) {
	decoder := gob.NewDecoder(pipefile)
	err = decoder.Decode(checkpoint)

	return checkpoint, err
}

func writePipefile(pipefile io.Writer, checkpoint time.Time) error {
	encoder := gob.NewEncoder(pipefile)
	err := encoder.Encode(checkpoint)

	return err
}

func initPipefile(pipefilename string) (lastTime time.Time, pipefile io.ReadWriteCloser) {
	// Confirm we can interact with our data file for tracking pipe history
	_, err := os.Stat(pipefilename)
	if err == nil {
		// The file already exists. Open it and store the last checkpoint
		pipefile, err = os.Open(pipefilename)
		if err != nil {
			log.Fatalf("could not open temp file %s: %s", pipefilename, err)
		}

		lastTime, err = readPipefile(pipefile)
	} else if os.IsNotExist(err) {
		// File doesn't exist. Make it, but don't do anything with it yet
		log.Printf("pipefile %s doesn't exist. Creating new pipefile", pipefilename)
		pipefile, err = os.Create(pipefilename)
		if err != nil {
			log.Fatalf("could not create temp file %s: %s", pipefilename, err)
		}

		lastTime = time.Now()
	}
	defer pipefile.Close()
}

// Copy the input from stdin to stdout until EOF or error
// TODO: checkpoint not just on EOF, but if stdin stops producing data at some
// bytes/second
func doPipe(w io.Writer, r io.Reader) error {
	_, err := io.Copy(w, r)
	if err != io.EOF {
		return err
	}
}

func main() {
	lastTime, pipefile := initPipefile(filepath.Join(*tmpPath, "throttlepipe-", *id))

	// Prepare to pipe
	stdinReader := bufio.NewReader(os.Stdin)
	stdoutWriter := bufio.NewWriter(os.Stdout)

	// Check our checkpoint
	if time.Since(lastTime) > (time.Duration(*sleepTime) * time.Minute) {
		// It has been less than sleepTimeDuration since last checkpoint
		// Just terminate for now

		return
	}

	err := doPipe(stdoutWriter, stdinReader)
	if err != nil {
		log.Fatalf("encountered an error while piping data [timer not set]: %s", err)
	}

	// Make checkpoint
	writePipefile(pipefile, time.Now())

	return
}
