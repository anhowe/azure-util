package edasim

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-queue-go/2017-07-29/azqueue"
)

type StorageQueue struct {
	MessagesURL azqueue.MessagesURL
	Context context.Context
}

func InitializeStorageQueue(storageAccount string, storageAccountKey string, queueName string, ctx context.Context) *StorageQueue {
	
	credential := azqueue.NewSharedKeyCredential(storageAccount, storageAccountKey)

	p := azqueue.NewPipeline(credential, azqueue.PipelineOptions{})

	u, _ := url.Parse(fmt.Sprintf("https://%s.queue.core.windows.net", storageAccount))

	serviceURL := azqueue.NewServiceURL(*u, p)

	// Create a URL that references the queue in the Azure Storage account.
	queueURL := serviceURL.NewQueueURL(queueName) // Queue names require lowercase
	
	messagesURL := queueURL.NewMessagesURL()

	return &StorageQueue{
		MessagesURL: messagesURL,
		Context: ctx,
	}
}

func (q *StorageQueue) Enqueue(message string) error {
	_, err := q.MessagesURL.Enqueue(q.Context, message, EnqueueVisibilityTimeout, EnqueueMessageTTL)
	return err
}

func (q *StorageQueue) Dequeue(maxMessages int32, visibilityTimeout time.Duration) (*azqueue.DequeuedMessagesResponse, error) {
	return q.MessagesURL.Dequeue(q.Context, maxMessages, visibilityTimeout)
}

func (q *StorageQueue) DeleteMessage(messageID azqueue.MessageID, popReceipt azqueue.PopReceipt) (*azqueue.MessageIDDeleteResponse, error) {
	msgIDURL := q.MessagesURL.NewMessageIDURL(messageID)
	return msgIDURL.Delete(q.Context, popReceipt)
}