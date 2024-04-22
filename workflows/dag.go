package workflows

import (
	"fmt"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"time"
)

type Node struct {
	ID       string   // Node identifier
	Children []string // IDs of child nodes
	Parents  []string // IDs of parent nodes
}

// NodeWorkflow processes individual nodes
func NodeWorkflow(ctx workflow.Context, node Node) error {
	logger := workflow.GetLogger(ctx)

	logger.Info(fmt.Sprintf(" %s: waiting for parents => %s", node.ID, node.Parents))
	// Wait for signals from all parent nodes
	if len(node.Parents) > 0 {
		selector := workflow.NewSelector(ctx)
		receivedSignals := make(map[string]bool)
		for _, parentID := range node.Parents {
			parentID := parentID
			signalName := "completed_" + parentID
			logger.Info(fmt.Sprintf("%s: Waiting for signal %s", signalName))
			selector.AddReceive(workflow.GetSignalChannel(ctx, signalName), func(c workflow.ReceiveChannel, more bool) {
				var signalReceived bool
				c.Receive(ctx, &signalReceived)
				logger.Info(fmt.Sprintf("%s: Received signal %s", node.ID, signalName))
				receivedSignals[parentID] = signalReceived
			})
		}

		// Wait until signals from all parents are received
		for len(receivedSignals) < len(node.Parents) {
			selector.Select(ctx) // This will block until a new signal is received
		}
	}

	// Simulate some work
	logger.Info(fmt.Sprintf("%s: Processing node", node.ID))
	workflow.Sleep(ctx, 2*time.Second)

	return nil
}

type NodeExec struct {
	Node       Node
	Future     workflow.Future
	WorkflowID string
}

// DAGWorkflow starts all workflows and manages signaling
func DAGWorkflow(ctx workflow.Context, nodes map[string]Node) error {
	info := workflow.GetInfo(ctx)

	nws := map[string]NodeExec{}
	for _, node := range nodes {
		ne := NodeExec{
			Node:       node,
			WorkflowID: fmt.Sprintf("workflow/%s/node/%s", info.WorkflowExecution.ID, node.ID),
		}
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
			WorkflowID:        ne.WorkflowID,
			ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE,
		})
		ne.Future = workflow.ExecuteChildWorkflow(childCtx, NodeWorkflow, node)
		nws[node.ID] = ne
	}

	wg := workflow.NewWaitGroup(ctx)
	wg.Add(len(nodes))
	for _, cur := range nws {
		cur := cur
		workflow.Go(ctx, func(ctx workflow.Context) {
			defer wg.Done()
			logger := workflow.GetLogger(ctx)
			err := cur.Future.Get(ctx, nil)
			if err != nil {
				logger.Error(fmt.Sprintf("%s: Node workflow failed: %s", cur.Node.ID, err))
			} else {
				logger.Info(fmt.Sprintf("%s: Node workflow completed", cur.Node.ID))
			}

			// Notify all children that this node has completed
			for _, childID := range cur.Node.Children {
				childExec := nws[childID]
				signalName := "completed_" + cur.Node.ID
				logger.Info(fmt.Sprintf("Signaling %s to %s", signalName, childExec.WorkflowID))
				err := workflow.SignalExternalWorkflow(ctx, childExec.WorkflowID, "", signalName, true).Get(ctx, nil)
				if err != nil {
					logger.Error(fmt.Sprintf("%s: error sending signal %s: %s", cur.Node.ID, signalName, err))
				}
			}
		})
	}
	wg.Wait(ctx)
	return nil
}
