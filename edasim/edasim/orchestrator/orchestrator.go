package orchestrator

import (
	"context"
	"log"
	"path"
	"sync"
	"time"

	"github.com/anhowe/azure-util/edasim/edasim"
	"github.com/Azure/azure-storage-queue-go/2017-07-29/azqueue"
)

const (
	SleepTimeNoWorkers = time.Duration(10) * time.Millisecond // 10ms
	SleepTimeNoQueueMessages = time.Duration(10) * time.Second // 1 second between checking queue
	SleepTimeNoQueueMessagesTick = time.Duration(10) * time.Millisecond // 10 ms between ticks
	VisibilityTimeout = time.Duration(300) * time.Second // 5 minute visibility timeout
)

type Orchestrator struct {
	Context context.Context
	ReadyQueue *edasim.StorageQueue
	ProcessQueue *edasim.StorageQueue
	CompleteQueue *edasim.StorageQueue
	UploaderQueue *edasim.StorageQueue
	JobFileSizeKB int
	JobStartFileCount int
	JobProcessFilesPath string
	OrchestratorThreads int
	ReadyCh chan struct{}
	MsgCh chan *azqueue.DequeuedMessage
	DirManager *DirectoryManager
}

func InitializeOrchestrator(
	storageAccount string,
	storageAccountKey string,
	ctx context.Context,
	readyQueueName string,
	processQueueName string,
	completedQueueName string,
	uploaderQueueName string,
	jobFileSizeKB int,
	jobStartFileCount int,
	jobProcessFilesPath string,
	orchestratorThreads int) *Orchestrator {
	return &Orchestrator {
		Context: ctx,
		ReadyQueue: edasim.InitializeStorageQueue(storageAccount, storageAccountKey, readyQueueName, ctx),
		ProcessQueue: edasim.InitializeStorageQueue(storageAccount, storageAccountKey, processQueueName, ctx),
		CompleteQueue: edasim.InitializeStorageQueue(storageAccount, storageAccountKey, completedQueueName, ctx),
		UploaderQueue: edasim.InitializeStorageQueue(storageAccount, storageAccountKey, uploaderQueueName, ctx),
		JobFileSizeKB: jobFileSizeKB,
		JobStartFileCount: jobStartFileCount,
		JobProcessFilesPath: jobProcessFilesPath,
		OrchestratorThreads: orchestratorThreads,
		ReadyCh: make(chan struct{}),
		MsgCh: make(chan *azqueue.DequeuedMessage, orchestratorThreads),
		DirManager: InitializeDirectoryManager(),
	}
}

func (o *Orchestrator) Run(syncWaitGroup *sync.WaitGroup) {
	log.Printf("started orchestrator.Run()\n")
	defer syncWaitGroup.Done()

	// start the stats collector
	o.Context = SetStatsChannel(o.Context)
	syncWaitGroup.Add(1)
	go StatsCollector(o.Context, syncWaitGroup)

	// start the ready queue listener and its workers
	// this uses the example from here: https://github.com/Azure/azure-storage-queue-go/blob/master/2017-07-29/azqueue/zt_examples_test.go
	for i := 0; i < o.OrchestratorThreads; i++ {
		syncWaitGroup.Add(1)
		go o.StartJobWorker(syncWaitGroup)	
	}
	
	syncWaitGroup.Add(1)
	go o.JobDispatcher(syncWaitGroup)

	// start the completed queue listener

	keepRunning := true
	for keepRunning {
		select {
		case <-o.Context.Done():
			keepRunning = false
		}
	}

	log.Printf("completed orchestrator.Run()\n")
}

func (o *Orchestrator) StartJobWorker(syncWaitGroup *sync.WaitGroup) {
	defer syncWaitGroup.Done()

	statsChannel := GetStatsChannel(o.Context)

	for {
		// signal that the work is ready to receive work
		o.ReadyCh <-struct{}{}
		select {
		case <-o.Context.Done():
			return
		case msg := <-o.MsgCh:
			jobConfig := edasim.ReadJobConfigFile(msg.Text)
			fullPath := o.GetDirectory(jobConfig.BatchName)
			workerFileWriter := edasim.InitializeWorkerFileWriter(jobConfig.Name, o.JobStartFileCount, 0)
			workerFileWriter.WriteStartFiles(fullPath, o.JobFileSizeKB)
			o.ProcessQueue.Enqueue(workerFileWriter.FirstStartFile(fullPath))
			if _, err := o.ReadyQueue.DeleteMessage(msg.ID, msg.PopReceipt); err != nil {
				log.Fatal(err)
			}
			statsChannel.ProcessedFilesWritten()
		}
	}
}

func (o *Orchestrator) GetDirectory(batchName string) string {
	fullPath := path.Join(o.JobProcessFilesPath, batchName)
	o.DirManager.VerifyDirectory(fullPath)
	return fullPath
}

func (o *Orchestrator) JobDispatcher(syncWaitGroup *sync.WaitGroup) {
	log.Printf("starting JobDispatcher\n")
	defer syncWaitGroup.Done()

	count := int32(0)

	statsChannel := GetStatsChannel(o.Context)

	for {
		done := false	
		for !done {
			select {
			case <-o.Context.Done():
				return
			case <-o.ReadyCh:
				count++
			default:
				done = true
			}
		}
		if count == 0 {
			// no workers, wait 1ms
			time.Sleep(SleepTimeNoWorkers)
			continue
		}

		// dequeue the messages, with a count of count
		dequeue, err := o.ReadyQueue.Dequeue(count, VisibilityTimeout)
		if err != nil {
			log.Fatal(err)
		}

		if dequeue.NumMessages() != 0 {
			now := time.Now()
			for m := int32(0); m < dequeue.NumMessages(); m++ {
				msg := dequeue.Message(m)
				if now.After(msg.NextVisibleTime) {
					log.Printf("ERROR: %v is after, ignoring", msg)
					continue
				}
				o.MsgCh <- msg
				statsChannel.JobProcessed()
				count--
			}
		} else {
			// otherwise sleep 10 seconds
			log.Printf("Dispatcher: no messages, sleeping, %d ready workers", count)
			ticker := time.NewTicker(time.Duration(MillisecondsSleep) * time.Millisecond)
			start := time.Now()
			for time.Since(start) < SleepTimeNoQueueMessages {
				select {
				case <-o.Context.Done():
					return
				case <-ticker.C:
				}
			}
			ticker.Stop()
			log.Printf("Dispatcher: awake")
		}
	}
}


