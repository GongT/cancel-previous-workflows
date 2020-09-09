package main

import (
	"log"

	"github.com/GongT/cancel-previous-workflows/internal/github"
)

func remove(r *github.WorkflowRun, current, total int) {
	github.RmeoveWorkflowRunVoid(r, current, total)
}

func main() {
	log.Printf("deleting logs in repo %s\n", github.GetCurrentRepo())

	if err := github.ForeachRuns(github.StateTypeAny, remove); err == nil {
	} else {
		log.Printf("error when list runs: %v\n", err)
		return
	}
}
