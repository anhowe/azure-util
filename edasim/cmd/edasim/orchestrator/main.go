package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/anhowe/azure-util/edasim/pkg/azure"
	"github.com/anhowe/azure-util/edasim/pkg/cli"
	"github.com/anhowe/azure-util/edasim/pkg/edasim"
)

func usage(errs ...error) {
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
	}
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       write the job config file and posts to the queue\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "required env vars:\n")
	fmt.Fprintf(os.Stderr, "\t%s - azure storage account\n", azure.AZURE_STORAGE_ACCOUNT)
	fmt.Fprintf(os.Stderr, "\t%s - azure storage account key\n", azure.AZURE_STORAGE_ACCOUNT_KEY)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

func verifyEnvVars() bool {
	available := true
	available = available && cli.VerifyEnvVar(azure.AZURE_STORAGE_ACCOUNT)
	available = available && cli.VerifyEnvVar(azure.AZURE_STORAGE_ACCOUNT_KEY)
	return available
}

func validateQueue(queueName string, queueNameLabel string) {
	if len(queueName) == 0 {
		fmt.Fprintf(os.Stderr, "ERROR: %s is not specified\n", queueNameLabel)
		usage()
		os.Exit(1)
	}
}

func initializeApplicationVariables(ctx context.Context) *edasim.Orchestrator {
	var jobReadyQueueName = flag.String("jobReadyQueueName", edasim.QueueJobReady, "the job read queue name")
	var jobProcessQueueName = flag.String("jobProcessQueueName", edasim.QueueJobProcess, "the job process queue name")
	var jobCompleteQueueName = flag.String("jobCompleteQueueName", edasim.QueueJobComplete, "the job completion queue name")
	var uploaderQueueName = flag.String("uploaderQueueName", edasim.QueueUploader, "the uploader job queue name")

	var jobStartFileConfigSizeKB = flag.Int("jobStartFileConfigSizeKB", edasim.DefaultFileSizeKB, "the job start file size in KB to write at start of job")
	var jobStartFileCount = flag.Int("jobStartFileCount", edasim.DefaultJobStartFiles, "the count of start job files")
	var jobStartFileBasePath = flag.String("jobStartFileBasePath", "", "the job file path")
	var jobCompleteFileSizeKB = flag.Int("jobCompleteFileSizeKB", 384, "the job complete file size in KB to write after job completed")
	var jobCompleteFailedFileSizeKB = flag.Int("jobCompleteFailedFileSizeKB", 1024, "the job start file size in KB to write at start of job")
	var jobFailedProbability = flag.Float64("jobFailedProbability", 0.01, "the probability of a job failure")
	var jobCompleteFileCount = flag.Int("jobCompleteFileCount", 12, "the count of completed job files")

	var orchestratorThreads = flag.Int("orchestratorThreads", edasim.DefaultOrchestratorThreads, "the number of concurrent orechestratorthreads")

	flag.Parse()

	if envVarsAvailable := verifyEnvVars(); !envVarsAvailable {
		usage()
		os.Exit(1)
	}

	storageAccount := cli.GetEnv(azure.AZURE_STORAGE_ACCOUNT)
	storageKey := cli.GetEnv(azure.AZURE_STORAGE_ACCOUNT_KEY)

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

	return edasim.InitializeOrchestrator(
		ctx,
		storageAccount,
		storageKey,
		*jobReadyQueueName,
		*jobProcessQueueName,
		*jobCompleteQueueName,
		*uploaderQueueName,
		*jobStartFileConfigSizeKB,
		*jobStartFileCount,
		*jobStartFileBasePath,
		*jobCompleteFileSizeKB,
		*jobCompleteFailedFileSizeKB,
		*jobFailedProbability,
		*jobCompleteFileCount,
		*orchestratorThreads)
}

func main() {
	// setup the shared context
	ctx, cancel := context.WithCancel(context.Background())
	syncWaitGroup := sync.WaitGroup{}

	// initialize and start the orchestrator
	orchestrator := initializeApplicationVariables(ctx)
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
