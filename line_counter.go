package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	pathToFiles   string
	maxGoroutines int
	timeStart     time.Time
	supportedExt  = []string{"go", "html", "css", "js"}
	counter       = 0
)

func init() {
	timeStart = time.Now()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&pathToFiles, "path", wd, "path to directory")
	flag.IntVar(&maxGoroutines, "max", 10, "max threads")
}

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	var mu sync.Mutex
	guard := make(chan struct{}, maxGoroutines)

	iterate(&wg, &mu, &guard)
	wg.Wait()

	elapsed := time.Since(timeStart).Seconds()

	fmt.Printf("Total lines: %d\n", counter)
	fmt.Printf("Execution time is %fsec\n", elapsed)

}

func iterate(wg *sync.WaitGroup, mu *sync.Mutex, guard *chan struct{}) {
	filepath.Walk(pathToFiles, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		wg.Add(1)
		*guard <- struct{}{}
		go func(path string, name string, mu *sync.Mutex, guard *chan struct{}) {
			defer wg.Done()
			if validate(name) {
				lines, err := countLines(path)
				if err != nil {
					log.Println(err)
				}
				mu.Lock()
				counter += lines
				mu.Unlock()
			}
			<-*guard
		}(path, info.Name(), mu, guard)
		return nil
	})
}

func validate(filename string) bool {
	if len(supportedExt) == 0 {
		return true
	}
	for _, ext := range supportedExt {
		if strings.HasSuffix(filename, "."+ext) {
			return true
		}
	}
	return false
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	lineCounter := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineCounter++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return lineCounter, nil
}
