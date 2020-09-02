package github

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
)

var workflowName = os.Getenv("GITHUB_WORKFLOW")

func GetCurrentWorkflowName() string {
	return workflowName
}

func GetWorkflowId() (int64, error) {
	query := make(map[string]string)
	query["per_page"] = strconv.Itoa(requestPerPage)

	api := ApiUrl("actions/workflows")

	var curr = 0
	for {
		log.Printf("  * page %v...", curr)
		query["page"] = strconv.Itoa(curr)

		body, err := DoRequest("GET", api, query)
		if err != nil {
			return 0, err
		}

		var workflows WorkflowsResponse
		if err = json.Unmarshal(body, &workflows); err != nil {
			return 0, err
		}

		for _, item := range workflows.Workflows {
			if item.Name == workflowName {
				return item.Id, nil
			}
		}

		curr++

		if len(workflows.Workflows) == 0 {
			break
		}
	}

	return 0, fmt.Errorf("workflow with name '%v' did not exists", workflowName)
}

func ForeachWorkflow(cb func(*Workflow, int, int)) error {
	query := make(map[string]string)
	query["per_page"] = strconv.Itoa(requestPerPage)

	api := ApiUrl("actions/workflows")

	var currentPage = 0
	var processedCount = 0
	dedup := make(map[int64]bool)
	for {
		log.Printf("  * page %v...", currentPage)
		query["page"] = strconv.Itoa(currentPage)

		body, err := DoRequest("GET", api, query)
		if err != nil {
			return err
		}

		var workflows WorkflowsResponse
		if err = json.Unmarshal(body, &workflows); err != nil {
			return err
		}

		for _, item := range workflows.Workflows {
			processedCount++

			if dedup[item.Id] {
				log.Printf("  ! duplicate id: %v", item.Id)
				continue
			}
			dedup[item.Id] = true

			cb(&Workflow{
				Id:   item.Id,
				Name: item.Name,
			}, processedCount, workflows.TotalCount)
		}

		currentPage++

		if workflows.TotalCount <= processedCount || len(workflows.Workflows) == 0 {
			break
		}
	}
	return nil
}
