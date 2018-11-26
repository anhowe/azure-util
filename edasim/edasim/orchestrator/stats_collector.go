package orchestrator

import (
	"context"
	"log"
	"sync"
	"time"
)

const (
	MillisecondsSleep = 10
	SecondsBetweenStats = time.Duration(5) * time.Second
)

func StatsCollector(ctx context.Context, syncWaitGroup *sync.WaitGroup) {
	log.Printf("starting Stats Collector\n")
	defer syncWaitGroup.Done()
	start := time.Now()
	
	// for statistics
	JobsProcessedCount := 0
	ProcessFilesWritten := 0
	CompletedJobsCount := 0
	UploaderCount := 0

	statsChannel := GetStatsChannel(ctx)

	ticker := time.NewTicker(time.Duration(MillisecondsSleep) * time.Millisecond)
	defer ticker.Stop()

	keepRunning := true
	for keepRunning {
		select {
		case <-statsChannel.ChJobProcessed:
			JobsProcessedCount++
		case <-statsChannel.ChProcessedFilesWritten:
			ProcessFilesWritten++
		case <-statsChannel.ChJobCompleted:
			CompletedJobsCount++
		case <-statsChannel.ChUploadSignaled:
			UploaderCount++
		case <-ctx.Done():
			keepRunning = false
		case <-ticker.C:
		}
		if time.Since(start) > SecondsBetweenStats || !keepRunning {
			start = start.Add(SecondsBetweenStats)
			log.Printf("JobsProcessedCount: %d", JobsProcessedCount)
			log.Printf("ProcessFilesWritten: %d", ProcessFilesWritten)
			log.Printf("CompletedJobsCount: %d", CompletedJobsCount)
			log.Printf("UploaderCount: %d", UploaderCount)
		}
	}
}