package main

import (
	"BSick7/go-temporal-signal-poc/workflows"
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
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

	nodes := map[string]workflows.Node{
		"A": {ID: "A", Children: []string{"B", "C"}},
		"B": {ID: "B", Children: []string{"D"}, Parents: []string{"A"}},
		"C": {ID: "C", Children: []string{"D"}, Parents: []string{"A"}},
		"D": {ID: "D", Parents: []string{"B", "C"}},
	}

	we, err := temporalClient.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "dag_workflow",
		TaskQueue: "go-temporal-signal-poc",
	}, workflows.DAGWorkflow, nodes)
	if err != nil {
		fmt.Println("Failed to start workflow", err)
	} else {
		fmt.Println("Workflow started", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}
}
