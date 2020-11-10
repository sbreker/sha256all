package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Read from 'filesChan' until channel closes. Calculate sha256sum
// for each file received.
func calcSha256(wg *sync.WaitGroup, id int, filesChan chan string) {
	for i := range filesChan {
		f, err := os.Open(i)
		if err != nil {
			log.Fatal(err)
		}

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%x  %d  %s\n", h.Sum(nil), id, i)
	}
	wg.Done()
}

// Walk the path provided recursively. Return all filenames found on the
// 'filesChan' channel
func walkPath(wg *sync.WaitGroup, root string, filesChan chan string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		filesChan <- path
		return nil
	})
	if err != nil {
		wg.Done()
		panic(err)
	}
	close(filesChan)
	wg.Done()
}

func main() {
	// New waitgroup
	var wg sync.WaitGroup
	root := "../"
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

	elapsed := time.Since(start)
	fmt.Printf("num cpus used=%d; t=%s\n", numGoRoutines+1, elapsed)
}
