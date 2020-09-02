package main

import (
	"log"

	"github.com/GongT/cancel-previous-workflows/internal/github"
)

func remove(r *github.WorkflowRun, current, total int) {
	body, err := github.DoRequest("DELETE", github.ApiUrl("actions/runs/%d", r.Id), nil)
	if err == nil {
		log.Printf("  [%4d/%4d] done  [%v]: %v\n", current, total, r.Id, string(body))
	} else {
		log.Printf("  [%4d/%4d] error [%v]: %v\n", current, total, r.Id, err)
	}
}

func main() {
	log.Printf("deleting logs in repo %s\n", github.GetCurrentRepo())

	if err := github.ForeachRuns(github.StateTypeAny, remove); err == nil {
	} else {
		log.Printf("error when list runs: %v\n", err)
		return
	}
}
