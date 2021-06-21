package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"
)

var testPipelineFile = "testInitPipeFile.tmp"

func TestInitialize(t *testing.T) {
	testPipelineFile = fmt.Sprintf("%d", time.Now().Unix()) + testPipelineFile
}

func TestPipefile(t *testing.T) {
	t.Cleanup(func() {
		err := os.Remove(testPipelineFile)
		if err != nil {
			t.Errorf("failed to cleanup pipeline file: %s", err)
		}
	})

	lastTime, err := initPipefile(testPipelineFile)
	if err != nil {
		t.Errorf("testPipefile failed on initPipefile: %s", err)
		return
	}
	if time.Since(lastTime) < time.Duration(5)*time.Second {
		t.Errorf("testPipefile failed because init checkpoint was too young")
	}

	lastTime, err = initPipefile(testPipelineFile)
	if err != nil {
		t.Errorf("testPipefile failed on initPipefile: %s", err)
		return
	}
	if time.Since(lastTime) < time.Duration(5)*time.Second {
		t.Errorf("testPipefile failed because second init checkpoint was too old")
	}

	testTime := time.Now()
	err = writePipefile(testPipelineFile, testTime)
	if err != nil {
		t.Errorf("testPipefile failed because we couldn't write to file: %s", err)
	}

	lastTime, err = readPipefile(testPipelineFile)
	if err != nil {
		t.Errorf("testPipefile failed because we couldn't read from file: %s", err)
		return
	}

	if lastTime.String() != testTime.Format("2006-01-02 15:04:05.9999999 -0700 MST") {
		t.Errorf("testPipefile failed because r/w didn't match: %s != %s", lastTime.String(), testTime.Format("2006-01-02 15:04:05.9999999 -0700 MST"))
	}
}

func TestDoPipe(t *testing.T) {
	in := bytes.NewBufferString("testing")
	var out bytes.Buffer

	err := doPipe(&out, in)
	if err != nil {
		t.Errorf("testDoPipe failed due to err: %s", err)
	}

	if out.String() != "testing" {
		t.Errorf("testDoPipe failed because pipe didn't pass data properly")
	}
}
