package edasim

import (
	"time"
)

const (	
	QueueJobReady = "jobready"
	QueueJobComplete = "jobcomplete"
	QueueJobProcess = "jobprocess"
	QueueUploader = "uploader"

	AZURE_STORAGE_ACCOUNT = "AZURE_STORAGE_ACCOUNT"
	AZURE_STORAGE_ACCOUNT_KEY = "AZURE_STORAGE_ACCOUNT_KEY"

	DefaultFileSizeKB = 384
	DefaultJobCount = 10
	DefaultUserCount = 1
	
	DefaultOrchestratorThreads = 16
	DefaultJobStartFiles = 3
	DefaultJobEndFiles = 12

	KB = 1024
	MB = KB * KB
)

const EnqueueVisibilityTimeout time.Duration = time.Second*0
const EnqueueMessageTTL time.Duration = time.Second*-1 // never expire
