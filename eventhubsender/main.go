package main

import (
	"context"
	"fmt"
	"flag"
	"log"
	"os"
	"time"
	
	"github.com/Azure/azure-amqp-common-go/sas"
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
	ctx = context.Background()
	SecondsToSleepBetweenSend = 5
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
	fmt.Fprintf(os.Stderr, "usage: %s MESSAGE\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       send a message to eventhub")
	fmt.Fprintf(os.Stderr, "\n\n")
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_SENDERKEYNAME)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_SENDERKEY)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_NAMESPACENAME)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_EVENTHUB_HUBNAME)
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

func initializeApplicationVariables() (string, string, string, string, string) {
	if envVarsAvailable := verifyEnvVars(); !envVarsAvailable {
		usage()
		os.Exit(1)
	}

	senderKeyName := getEnv(AZURE_EVENTHUB_SENDERKEYNAME)
	senderKey := getEnv(AZURE_EVENTHUB_SENDERKEY)
	eventHubNamespaceName := getEnv(AZURE_EVENTHUB_NAMESPACENAME)
	eventHubName := getEnv(AZURE_EVENTHUB_HUBNAME)

	flag.Parse()

	if argCount := len(flag.Args()); argCount == 0 {
		usage()
		os.Exit(1)
	}
	message := flag.Arg(0)

	return senderKeyName, senderKey, eventHubNamespaceName, eventHubName, message
}

func main() {
	verifyEnvVars()

	senderKeyName, senderKey, eventHubNamespaceName, eventHubName, message := initializeApplicationVariables()

	provider, err := sas.NewTokenProvider(sas.TokenProviderWithKey(senderKeyName, senderKey))
	if err != nil {
		log.Fatalf("failed to get token provider: %s\n", err)
	}

	// get an existing hub
	hub, err := eventhubs.NewHub(eventHubNamespaceName, eventHubName, provider)
	defer hub.Close(ctx)
	if err != nil {
		log.Fatalf("failed to get hub: %s\n", err)
	}

	// send message to hub.
	// by default the destination partition is selected round-robin by the
	// Event Hubs service
	for {
		log.Printf("sending message '%s'", message)
		err = hub.Send(ctx, eventhubs.NewEventFromString(message))
		if err != nil {
			log.Printf("failed to send message, retrying in %d seconds: %s\n", err, SecondsToSleepBetweenSend)
			time.Sleep(time.Duration(SecondsToSleepBetweenSend) * time.Second)
		}
		break
	}
}