package main

import (
	"BSick7/go-temporal-signal-poc/workflows"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"
)

func main() {
	temporalClient, err := client.Dial(client.Options{
		HostPort:  client.DefaultHostPort,
		Namespace: "default",
	})
	if err != nil {
		log.Fatalln("Unable to connect to Temporal Cloud.", err)
	}
	defer temporalClient.Close()

	// Create a new Worker
	yourWorker := worker.New(temporalClient, "go-temporal-signal-poc", worker.Options{})
	// Register Workflows
	yourWorker.RegisterWorkflow(workflows.DAGWorkflow)
	yourWorker.RegisterWorkflow(workflows.NodeWorkflow)
	// Register Activities

	// Start the Worker Process
	err = yourWorker.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start the Worker Process", err)
	}
}
