package edasim

import (
	"fmt"
	"log"
	"path"
	"sync"
)

type JobSubmitter struct {
	BatchName string
	Id int
	ReadyQueue *StorageQueue
	JobCount int
	JobPath string
	JobFileSizeKB int
}

func InitializeJobSubmitter(batchName string, id int, readyQueue *StorageQueue, jobCount int, jobPath string, jobFileSizeKB int) *JobSubmitter {
	return &JobSubmitter{
		BatchName: batchName,
		Id: id,
		ReadyQueue: readyQueue,
		JobCount: jobCount,
		JobPath: jobPath,
		JobFileSizeKB: jobFileSizeKB,
	}
}

func (j *JobSubmitter) GetJobName(index int) string {
	return fmt.Sprintf("%d_%d", j.Id, index)
}

func (j *JobSubmitter) Run(userSyncWaitGroup *sync.WaitGroup) {
	defer userSyncWaitGroup.Done()
	log.Printf("user %d: starting to submit %d jobs\n", j.Id, j.JobCount)

	for i :=0; i < j.JobCount; i++ {
		jobConfig := InitializeJobConfig(j.GetJobName(i), j.BatchName)
		fullPath := path.Join(j.JobPath, jobConfig.GetJobConfigName())

		// write the file
		jobConfig.WriteFile(fullPath, j.JobFileSizeKB)

		// queue completion
		j.ReadyQueue.Enqueue(fullPath)
	}

	log.Printf("user %d: completed submitting %d jobs\n", j.Id, j.JobCount)
}