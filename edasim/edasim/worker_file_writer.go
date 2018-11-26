package edasim

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type WorkFileWriter struct {
	JobConfigName string
	StartFileCount int
	CompleteFileCount int
	PaddedString string
}

func InitializeWorkerFileWriter(jobConfigName string, startFileCount int, completeFileCount int) *WorkFileWriter {
	return &WorkFileWriter{
		JobConfigName: jobConfigName,
		StartFileCount: startFileCount,
		CompleteFileCount: completeFileCount,
	}
}

func (w *WorkFileWriter) WriteStartFiles(fullPath string, fileSize int) {
	// read once
	data, err := json.Marshal(w)
	check(err)

	// pad and re-martial to match the bytes
	padLength := (KB * 384)-len(data)
	if padLength > 0 {
		w.PaddedString = RandStringRunes(padLength)
		data, err = json.Marshal(w)
		check(err)
	}
	
	// write the files
	for i :=0 ; i < w.StartFileCount ; i++ {
		filename := w.GetStartFileName(fullPath, i)
		f, err := os.Create(filename)
		check(err)
		defer f.Close()
		_, err = f.Write([]byte(data))
		check(err)
	}
}

func (w *WorkFileWriter) GetStartFileName(fullPath string, index int) string {
	return path.Join(fullPath, fmt.Sprintf("%s.start.%d", w.JobConfigName, index))
}

func (w *WorkFileWriter) FirstStartFile(fullPath string) string {
	return w.GetStartFileName(fullPath, 0)
}