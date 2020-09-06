package github

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
)

var branchName = strings.Replace(os.Getenv("GITHUB_REF"), "refs/heads/", "", 1)

func GetBranchName() string {
	return branchName
}

func IsWorkspaceIn(arr []*WorkflowRun, val *WorkflowRun) bool {
	for _, item := range arr {
		if val.Id == item.Id {
			return true
		}
	}
	return false
}

func ForeachRuns(state StateType, cb func(*WorkflowRun, int, int)) (err error) {
	query := make(map[string]string)
	if len(state) > 0 {
		query["status"] = string(state)
	}
	if len(branchName) > 0 {
		query["branch"] = branchName
	}
	return foreachApi(ApiUrl("actions/runs"), query, cb)
}

func ForeachWorkflowRuns(workflow int64, cb func(*WorkflowRun, int, int)) (err error) {
	query := make(map[string]string)

	return foreachApi(ApiUrl("actions/workflows/%v/runs", workflow), query, cb)
}

func foreachApi(api string, query map[string]string, cb func(*WorkflowRun, int, int)) (err error) {
	log.Printf("listing runs for branch %s in repo %s\n", branchName, githubRepo)

	query["per_page"] = strconv.Itoa(requestPerPage)

	var currentPage = 0
	var processedCount = 0
	dedup := make(map[int64]bool)
	for {
		log.Printf("  * page %v...", currentPage)
		query["page"] = strconv.Itoa(currentPage)
		currentPage++

		body, err := DoRequest("GET", api, query)
		if err != nil {
			log.Printf("    error request api: %v", err)
			break
		}

		var workflows WorkflowRunsResponse
		err = json.Unmarshal(body, &workflows)
		if err != nil {
			log.Printf("    error parse json: %v", err)
			continue
		}

		for _, item := range workflows.WorkflowRuns {
			processedCount++

			if dedup[item.Id] {
				log.Printf("  ! duplicate id: %v", item.Id)
				continue
			}
			dedup[item.Id] = true

			cb(&WorkflowRun{
				Id:         item.Id,
				Status:     item.Status,
				HeadSha:    item.HeadSha,
				HeadBranch: item.HeadBranch,
				RunNumber:  item.RunNumber,
				WorkflowId: item.WorkflowId,
			}, processedCount, workflows.TotalCount)
		}

		log.Printf("    got=%v | current size=%v | total count=%v", len(workflows.WorkflowRuns), processedCount, workflows.TotalCount)
		if workflows.TotalCount <= processedCount || len(workflows.WorkflowRuns) == 0 {
			break
		}
	}

	return
}

func ListRuns(state StateType) (runs []*WorkflowRun, err error) {
	err = ForeachRuns(state, func(r *WorkflowRun, _, _ int) {
		runs = append(runs, r)
	})
	return
}
