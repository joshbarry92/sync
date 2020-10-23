// Package main
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Files struct
type Files struct {
	fileName map[string]fileMetadata
	mux      sync.Mutex
}

type fileMetadata struct {
	filePath string
	size     int64
	sig      []byte
}

var wg sync.WaitGroup

func (fn *Files) walkDir(dir string) {
	defer wg.Done()
	visit := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() && path != dir {
			wg.Add(1)
			go fn.walkDir(path)
			return filepath.SkipDir
		}
		if f.Mode().IsRegular() {
			file := metadata(f, path)
			fn.mux.Lock()
			defer fn.mux.Unlock()
			fn.fileName[f.Name()] = file
		}
		return nil
	}
	filepath.Walk(dir, visit)
}

func nonP(path string) {
	z := make(map[string]fileMetadata)
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			file := metadata(info, path)
			z[path] = file
			return nil
		})
	if err != nil {
		fmt.Println(err)
	}
}

func metadata(f os.FileInfo, path string) fileMetadata {

	m := fileMetadata{
		filePath: path,
		size:     f.Size(),
		sig:      nil,
	}
	return m
}

func main() {
	//paths := [2]string{"/usr/local/google/home/jdbarry/Downloads", "/usr/local/google/home/jdbarry/Documents"}
	paths := [2]string{"/mnt/user/TV Shows", "/mnt/user/Movies"}

	// Validate Path exists
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.Exit(1)
		}
	}

	// Mutex Ops
	start := time.Now()
	// Initialize the map of files
	f := Files{fileName: make(map[string]fileMetadata)}
	for _, path := range paths {
		wg.Add(1)
		f.walkDir(path)
	}
	wg.Wait()
	elapsed := time.Since(start)
	fmt.Printf("Channel mutex took %s\n", elapsed)

	// Singular Exec timing
	start2 := time.Now()
	for _, path := range paths {
		nonP(path)
	}
	elapsed2 := time.Since(start2)
	fmt.Printf("Singular exec took %s\n", elapsed2)

	print(f)
}

func print(f Files) {
	for k, v := range f.fileName {
		fmt.Printf("Filepath: %s, Size: %i\n", k, v.size)
	}
}
