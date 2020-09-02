package github

import (
	"encoding/json"
	"log"
	"net/url"
	"strconv"
)

func IsWorkspaceIn(arr []*WorkflowRun, val *WorkflowRun) bool {
	for _, item := range arr {
		if val.Id == item.Id {
			return true
		}
	}
	return false
}

func ForeachRuns(state StateType, branchName string, cb func(*WorkflowRun, int, int)) (err error) {
	log.Printf("listing %v runs for branch %s in repo %s\n", state, branchName, githubRepo)

	query := make(url.Values)
	query.Set("per_page", strconv.Itoa(requestPerPage))
	if len(branchName) > 0 {
		query.Set("branch", branchName)
	}
	if len(state) > 0 {
		query.Set("status", string(state))
	}

	api := ApiUrl("actions/runs")

	var currentPage = 0
	var processedCount = 0
	dedup := make(map[int64]bool)
	for {
		log.Printf("  * page %v...", currentPage)
		query.Set("page", strconv.Itoa(currentPage))
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

func ListRuns(state StateType, branchName string) (runs []*WorkflowRun, err error) {
	err = ForeachRuns(state, branchName, func(r *WorkflowRun, _, _ int) {
		runs = append(runs, r)
	})
	return
}
