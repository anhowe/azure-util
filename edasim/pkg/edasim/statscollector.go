package edasim

import (
	"context"
	"log"
	"sync"
	"time"
)

const (
	millisecondsSleep   = 10
	secondsBetweenStats = time.Duration(5) * time.Second
)

// StatsCollector is a go routine that prints the queue statistics on a schedule
func StatsCollector(ctx context.Context, syncWaitGroup *sync.WaitGroup) {
	log.Printf("starting Stats Collector\n")
	defer syncWaitGroup.Done()
	defer log.Printf("StatsCollector complete")
	start := time.Now()

	// for statistics
	jobsProcessedCount := 0
	processFilesWritten := 0
	completedJobsCount := 0
	uploadCount := 0
	errorCount := 0

	statsChannel := GetStatsChannel(ctx)

	ticker := time.NewTicker(time.Duration(millisecondsSleep) * time.Millisecond)
	defer ticker.Stop()

	keepRunning := true
	for keepRunning {
		select {
		case <-statsChannel.ChJobProcessed:
			jobsProcessedCount++
		case <-statsChannel.ChProcessedFilesWritten:
			processFilesWritten++
		case <-statsChannel.ChJobCompleted:
			completedJobsCount++
		case <-statsChannel.ChUpload:
			uploadCount++
		case <-statsChannel.ChError:
			errorCount++
		case <-ctx.Done():
			keepRunning = false
		case <-ticker.C:
		}
		if time.Since(start) > secondsBetweenStats || !keepRunning {
			start = start.Add(secondsBetweenStats)
			log.Printf("jobsProcessedCount: %d", jobsProcessedCount)
			log.Printf("processFilesWritten: %d", processFilesWritten)
			log.Printf("completedJobsCount: %d", completedJobsCount)
			log.Printf("uploadCount: %d", uploadCount)
			log.Printf("errorCount: %d", errorCount)
		}
	}
}
