package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/anhowe/azure-util/cli"
	"github.com/anhowe/azure-util/edasim/edasim"
	"github.com/anhowe/azure-util/edasim/edasim/orchestrator"
)

func usage(errs ...error) {
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
	}
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       write the job config file and posts to the queue\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "required env vars:\n")
	fmt.Fprintf(os.Stderr, "\t%s - azure storage account\n", edasim.AZURE_STORAGE_ACCOUNT)
	fmt.Fprintf(os.Stderr, "\t%s - azure storage account key\n", edasim.AZURE_STORAGE_ACCOUNT_KEY)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

func verifyEnvVars() bool {
	available := true
	available = available && cli.VerifyEnvVar(edasim.AZURE_STORAGE_ACCOUNT)
	available = available && cli.VerifyEnvVar(edasim.AZURE_STORAGE_ACCOUNT_KEY)
	return available
}

func validateQueue(queueName string, queueNameLabel string) {
	if len(queueName) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: %s is not specified\n", queueNameLabel)
		usage()
		os.Exit(1)
	}
}

func initializeApplicationVariables() (string, int, int, string, int, string, string, string, string, string) {
	var jobReadyQueueName = flag.String("jobReadyQueueName", edasim.QueueJobReady, "the job read queue name")
	var jobProcessQueueName = flag.String("jobProcessQueueName", edasim.QueueJobProcess, "the job process queue name")
	var jobCompleteQueueName = flag.String("jobCompleteQueueName", edasim.QueueJobComplete, "the job completion queue name")
	var uploaderQueueName = flag.String("uploaderQueueName", edasim.QueueUploader, "the uploader job queue name")

	var jobStartFileConfigSizeKB = flag.Int("jobStartFileConfigSizeKB", edasim.DefaultFileSizeKB, "the job start file size in KB to write at start of job")
	var jobStartFileCount = flag.Int("JobStartFileCount", edasim.DefaultJobStartFiles, "the count of start job files")
	var jobStartFileBasePath = flag.String("jobStartFileBasePath", "", "the job file path")
	var orchestratorThreads = flag.Int("OrchestratorThreads", edasim.DefaultOrchestratorThreads, "the number of concurrent orechestratorthreads")
	
	flag.Parse()

	if envVarsAvailable := verifyEnvVars(); !envVarsAvailable {
		usage()
		os.Exit(1)
	}

	storageAccount := cli.GetEnv(edasim.AZURE_STORAGE_ACCOUNT)
	storageKey := cli.GetEnv(edasim.AZURE_STORAGE_ACCOUNT_KEY)

	if len(*jobStartFileBasePath) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: jobStartFileBasePath is not specified\n")
		usage()
		os.Exit(1)
	}

	if _, err := os.Stat(*jobStartFileBasePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "ERROR: jobStartFileBasePath '%s' does not exist\n", *jobStartFileBasePath)
		usage()
		os.Exit(1)
	}

	validateQueue(*jobReadyQueueName, "jobReadyQueueName")
	validateQueue(*jobProcessQueueName, "jobProcessQueueName")
	validateQueue(*jobCompleteQueueName, "jobCompleteQueueName")
	validateQueue(*uploaderQueueName, "uploaderQueueName")
	
	return *jobReadyQueueName, *jobStartFileConfigSizeKB, *jobStartFileCount, *jobStartFileBasePath, *orchestratorThreads, *jobProcessQueueName, *jobCompleteQueueName, *uploaderQueueName, storageAccount, storageKey
}

func main() {
	jobReadyQueueName, jobStartFileConfigSizeKB, jobStartFileCount, jobStartFileBasePath, orchestratorThreads, jobProcessQueueName, jobCompleteQueueName, uploaderQueueName, storageAccount, storageKey := initializeApplicationVariables()
	
	log.Printf("Starting orchestration jobs\n")
	log.Printf("Orchestrator threads: %d\n", orchestratorThreads)
	log.Printf("File Details:\n")
	log.Printf("\tJob Start File Base Path: %s\n", jobStartFileBasePath)
	log.Printf("\tJob Start FileCount: %d\n", jobStartFileCount)
	log.Printf("\tJob Start Filesize: %d\n", jobStartFileConfigSizeKB)
	log.Printf("\n")
	log.Printf("Storage Details:\n")
	log.Printf("\tstorage account: %s\n", storageAccount)
	//log.Printf("\tstorage account key: %s\n", storageKey)
	log.Printf("\tjob ready queue name: %s\n", jobReadyQueueName)
	log.Printf("\tjob process queue name: %s\n", jobProcessQueueName)
	log.Printf("\tjob completion queue name: %s\n", jobCompleteQueueName)
	log.Printf("\tjob uploader queue name: %s\n", uploaderQueueName)

	// setup the shared context
	ctx, cancel := context.WithCancel(context.Background())
	syncWaitGroup := sync.WaitGroup{}

	// initialize the orchestrator
	orchestrator := orchestrator.InitializeOrchestrator(
		storageAccount,
		storageKey,
		ctx,
		jobReadyQueueName, 
		jobProcessQueueName, 
		jobCompleteQueueName, 
		uploaderQueueName,
		jobStartFileConfigSizeKB,
		jobStartFileCount,
		jobStartFileBasePath,
		orchestratorThreads)
	syncWaitGroup.Add(1)
	go orchestrator.Run(&syncWaitGroup)

	// wait on ctrl-c
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	log.Printf("Received ctrl-c, stopping services...")
	cancel()

	log.Printf("Waiting for all processes to finish")
	syncWaitGroup.Wait()
	log.Printf("finished")
}


