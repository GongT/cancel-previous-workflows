package main

import (
	"encoding/json"
	"log"

	"github.com/GongT/cancel-previous-workflows/internal/github"
)

func main() {
	log.Printf("start failed workflow in repo %v (ref: %v)\n", github.GetCurrentRepo(), github.GetBranchName())

	if err := github.ForeachWorkflow(eachWorkflow); err == nil {
	} else {
		log.Printf("error when list workflows: %v\n", err)
		return
	}
}
func eachWorkflow(r *github.Workflow, current, total int) {
	listRuns := make(map[string]string)

	if len(github.GetBranchName()) != 0 {
		listRuns["branch"] = github.GetBranchName()
	}
	listRuns["per_page"] = "1"

	body, err := github.DoRequest("GET", github.ApiUrl("actions/workflows/%v/runs", r.Id), listRuns)
	if err != nil {
		log.Printf("  [%2d/%2d] error [%v]: [list runs] %v\n", current, total, r.Name, err)
		return
	}

	var runs github.WorkflowRunsResponse
	err = json.Unmarshal(body, &runs)
	if err != nil {
		log.Printf("  [%2d/%2d] error [%v]: [json] %v\n", current, total, r.Name, err)
		return
	}
	if len(runs.WorkflowRuns) == 0 {
		log.Printf("  [%2d/%2d] skip  [%v]: not ever run\n", current, total, r.Name)
		return
	}

	run := runs.WorkflowRuns[0]
	if run.Status != github.StateTypeComplete {
		log.Printf("  [%2d/%2d] skip  [%v]: last state is %v\n", current, total, r.Name, run.Status)
		return
	}

	if run.Conclusion != github.ConclusionFailure &&
		run.Conclusion != github.ConclusionTimed_out &&
		run.Conclusion != github.ConclusionCancelled {
		log.Printf("  [%2d/%2d] skip  [%v]: last conclusion is %v\n", current, total, r.Name, run.Conclusion)
		return
	}
	startWork := make(map[string]string, 1)
	startWork["ref"] = run.HeadBranch

	_, err = github.DoRequest("POST", github.ApiUrl("actions/workflows/%v/dispatches", r.Id), startWork)
	if err == nil {
		log.Printf("  [%2d/%2d] done  [%v]: last conclusion is %v\n", current, total, r.Name, run.Conclusion)
	} else {
		log.Printf("  [%2d/%2d] error [%v]: %v\n", current, total, r.Name, err)
	}
}
