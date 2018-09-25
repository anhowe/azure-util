package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	
	"github.com/Azure/azure-amqp-common-go/sas"
	"github.com/Azure/azure-amqp-common-go/persist"
	eventhubs "github.com/Azure/azure-event-hubs-go"
)

// 
// To setup, instructions from: https://docs.microsoft.com/en-us/azure/event-hubs/event-hubs-go-get-started-send
//
// go get -u github.com/Azure/azure-event-hubs-go
// go get -u github.com/Azure/azure-amqp-common-go/...
//
// some auth info: https://github.com/Azure/azure-event-hubs-go
//

var (
	ctx        = context.Background()
)

const (
	AZURE_EVENTHUB_SENDERKEYNAME = "AZURE_EVENTHUB_SENDERKEYNAME"
	AZURE_EVENTHUB_SENDERKEY = "AZURE_EVENTHUB_SENDERKEY"
	AZURE_EVENTHUB_NAMESPACENAME = "AZURE_EVENTHUB_NAMESPACENAME"
	AZURE_EVENTHUB_HUBNAME = "AZURE_EVENTHUB_HUBNAME"
)

func usage(errs ...error) {
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
	}
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       receive messages from eventhub")
	fmt.Fprintf(os.Stderr, "\n\n")
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_SENDERKEYNAME)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_SENDERKEY)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_NAMESPACENAME)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_HUBNAME)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

func verifyEnvVar(envvar string) bool {
	if _, available := os.LookupEnv(envvar); !available {
		fmt.Fprintf(os.Stderr, "ERROR: Missing Environment Variable %s\n", envvar)
		return false
	}
	return true
}

func verifyEnvVars() bool {
	available := true
	available = available && verifyEnvVar(AZURE_EVENTHUB_SENDERKEYNAME)
	available = available && verifyEnvVar(AZURE_EVENTHUB_SENDERKEY)
	available = available && verifyEnvVar(AZURE_EVENTHUB_NAMESPACENAME)
	available = available && verifyEnvVar(AZURE_EVENTHUB_HUBNAME)
	return available
}

func getEnv(envVarName string) string {
	s := os.Getenv(envVarName)
	
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	return s
}

func initializeApplicationVariables() (string, string, string, string, bool) {
	if envVarsAvailable := verifyEnvVars(); !envVarsAvailable {
		usage()
		os.Exit(1)
	}

	var receiveWithLatestOffset = flag.Bool("receiveWithLatestOffset", true, "receive with latest offset, otherwise start from beginning")

	flag.Parse()

	senderKeyName := getEnv(AZURE_EVENTHUB_SENDERKEYNAME)
	senderKey := getEnv(AZURE_EVENTHUB_SENDERKEY)
	eventHubNamespaceName := getEnv(AZURE_EVENTHUB_NAMESPACENAME)
	eventHubName := getEnv(AZURE_EVENTHUB_HUBNAME)

	return senderKeyName, senderKey, eventHubNamespaceName, eventHubName, *receiveWithLatestOffset
}

func main() {
	senderKeyName, senderKey, nsName, hubName, receiveWithLatestOffset := initializeApplicationVariables()
	
	provider, err := sas.NewTokenProvider(sas.TokenProviderWithKey(senderKeyName,senderKey))
	if err != nil {
		log.Fatalf("failed to get token provider: %s\n", err)
	}

	// get an existing hub
	hub, err := eventhubs.NewHub(nsName, hubName, provider)
	defer hub.Close(ctx)
	if err != nil {
		log.Fatalf("failed to get hub: %s\n", err)
	}

	// get info about partitions in hub
	info, err := hub.GetRuntimeInformation(ctx)
	if err != nil {
		log.Fatalf("failed to get runtime info: %s\n", err)
	}
	log.Printf("partition IDs: %s\n", info.PartitionIDs)

	// set up wait group to wait for expected message
	eventReceived := make(chan struct{})

	// declare handler for incoming events
	handler := func(ctx context.Context, event *eventhubs.Event) error {
		log.Printf("received: %s\n", string(event.Data))
		// notify channel that event was received
		eventReceived <- struct{}{}
		return nil
	}

	var receiveOption eventhubs.ReceiveOption
	if receiveWithLatestOffset {
		receiveOption = eventhubs.ReceiveWithLatestOffset()
	} else {
		receiveOption = eventhubs.ReceiveWithStartingOffset(persist.StartOfStream)
	}

	for _, partitionID := range info.PartitionIDs {
		_, err := hub.Receive(
			ctx,
			partitionID,
			handler,
			receiveOption,
		)
		if err != nil {
			log.Fatalf("failed to receive for partition ID %s: %s\n", partitionID, err)
		}
	}

	for {
		select {
		case <-eventReceived:
		}
	}
}


