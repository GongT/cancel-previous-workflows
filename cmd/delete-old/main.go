package main

import (
	"log"

	"github.com/GongT/cancel-previous-workflows/internal/github"
)

func main() {
	log.Printf("remove old logs of each workflow in repo %v (ref: %v)\n", github.GetCurrentRepo(), github.GetBranchName())

	if err := github.ForeachWorkflow(eachWorkflow); err != nil {
		log.Printf("error when list workflows: %v\n", err)
	}
}
func eachWorkflow(r *github.Workflow, current, total int) {
	log.Printf("workflow [%v/%v] %v:\n", current, total, r.Name)
	if err := github.ForeachWorkflowRuns(r.Id, eachRuns); err != nil {
		log.Printf("error when list workflows: %v\n", err)
	}
}

func eachRuns(r *github.WorkflowRun, current, total int) {
	if current == 1 {
		log.Printf("  [%4d/%4d] skip  [%v]: most current\n", current, total, r.Id)
		return
	}

	github.RmeoveWorkflowRunVoid(r, current, total)
}
