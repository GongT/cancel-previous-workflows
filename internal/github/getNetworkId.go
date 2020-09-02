package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
)

var workflowName = os.Getenv("GITHUB_WORKFLOW")

func GetCurrentWorkflowName() string {
	return workflowName
}

func GetWorkflowId() (int64, error) {
	query := make(url.Values)
	query.Set("per_page", strconv.Itoa(requestPerPage))

	api := ApiUrl("actions/workflows")

	var curr = 0
	for {
		log.Printf("  * page %v...", curr)
		query.Set("page", strconv.Itoa(curr))

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
