package orchestrator

import (
	"fmt"
	"log"
	"os"
	"sync"
)

type DirectoryManager struct {
	mux sync.Mutex
	directories map[string]bool
}

func InitializeDirectoryManager() *DirectoryManager {
	return &DirectoryManager{
		directories: make(map[string]bool),
	}
}

func (d *DirectoryManager) VerifyDirectory(path string) {
	d.mux.Lock()
	defer d.mux.Unlock()
	if _, ok := d.directories[path]; !ok {
		if e := os.MkdirAll(path, os.ModePerm); e != nil {
			fmt.Fprintf(os.Stderr, "ERROR: unable to create directory '%s': %v\n", path, e)
			log.Fatal(e)
			d.directories[path] = true
		}
	}
}
