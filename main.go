package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// Read from 'filesChan' until channel closes. Calculate sha256sum
// for each file received.
func calcSha256(wg *sync.WaitGroup, id int, filesChan chan string) {
	defer wg.Done()
	for i := range filesChan {
		f, err := os.Open(i)
		if err != nil {
			logger.Printf("error opening file %s", i)
			continue
		}

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			logger.Printf("error calculating sum of %s", i)
			continue
		}

		logger.Printf("%x  %d  %s\n", h.Sum(nil), id, i)
	}
}

// Walk the path provided recursively. Return all filenames found on the
// 'filesChan' channel
func walkPath(wg *sync.WaitGroup, root string, filesChan chan string) {
	defer wg.Done()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		filesChan <- path
		return nil
	})
	if err != nil {
		panic(err)
	}
	close(filesChan)
}

var logger *log.Logger

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to given path")
	var bufferResults = flag.Bool("buffer", false, "buffer results")
	flag.Parse()

	// Parse optinal root parameter.
	var root = "../"
	if userPath := flag.Arg(0); userPath != "" {
		root = userPath
	}

	// Set up the logger where given -buffer=true, a buffer of bytes is used
	// instead of os.Stderr. The buffer is printed before the program ends.
	var buf bytes.Buffer
	{
		var out io.Writer
		if *bufferResults {
			out = bufio.NewWriter(&buf)
		} else {
			out = os.Stderr
		}

		logger = log.New(out, "[sha256all] ", log.LstdFlags|log.Lshortfile)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
	}

	// New waitgroup
	var wg sync.WaitGroup
	start := time.Now()
	filesChan := make(chan string)

	wg.Add(1)
	go walkPath(&wg, root, filesChan)

	// Use half of the cpus
	numGoRoutines := (runtime.NumCPU() / 2)

	for i := 0; i < numGoRoutines; i++ {
		wg.Add(1)
		go calcSha256(&wg, i, filesChan)
	}

	// Wait until last thread completes.
	wg.Wait()

	if *cpuprofile != "" {
		pprof.StopCPUProfile()
	}

	elapsed := time.Since(start)

	// Print our buffered logging entries.
	if *bufferResults {
		fmt.Println(buf.String())
	}

	logger.SetOutput(os.Stderr)
	logger.Printf("num cpus used=%d; t=%s\n", numGoRoutines+1, elapsed)
}
