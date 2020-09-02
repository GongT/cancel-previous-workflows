package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Workflow struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type WorkflowsResponse struct {
	Workflows []Workflow `json:"workflows"`
}

type WorkflowRun struct {
	Id         int64  `json:"id"`
	Status     string `json:"status"`
	HeadSha    string `json:"head_sha"`
	HeadBranch string `json:"head_branch"`
	RunNumber  int    `json:"run_number"`
	WorkflowId int64  `json:"workflow_id"`
}

type WorkflowRunsResponse struct {
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

var httpClient http.Client
var githubApiUrl = os.Getenv("GITHUB_API_URL")
var githubRepo = os.Getenv("GITHUB_REPOSITORY")
var githubToken = os.Getenv("GITHUB_TOKEN")
var branchName = strings.Replace(os.Getenv("GITHUB_REF"), "refs/heads/", "", 1)
var currentSha = os.Getenv("GITHUB_SHA")
var workflowName = os.Getenv("GITHUB_WORKFLOW")
var currentRunNumber, _ = strconv.Atoi(os.Getenv("GITHUB_RUN_NUMBER"))
var wg = sync.WaitGroup{}

func githubRequest(request *http.Request) (*http.Response, error) {
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func cancelWorkflow(id int64) error {
	request, err := http.NewRequest("POST", githubApi("repos/%s/actions/runs/%d/cancel", githubRepo, id), nil)
	if err != nil {
		return err
	}
	response, err := githubRequest(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(response.Body)
		return errors.New(fmt.Sprintf("failed to cancel workflow #%d, status code: %d, body: %s", id, response.StatusCode, body))
	}
	return nil
}

// I don't wan't to fail the current workflow if I fail canceling previous workflow's => so I only log errors
func main() {
	if len(githubApiUrl) == 0 {
		githubApiUrl = "https://api.github.com"
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = http.Client{Transport: customTransport, Timeout: time.Minute}

	log.Printf("finding workflow names in repo %s\n", githubRepo)
	workflowId, err := getWorkflowId()
	if err != nil {
		log.Printf("error find workflow id: %v\n", err)
		return
	}

	log.Printf("listing runs for branch %s in repo %s\n", branchName, githubRepo)
	var query url.Values
	query.Set("branch", branchName)
	body, err := doRequest(githubApi("repos/%s/actions/runs", githubRepo), query)
	if err != nil {
		log.Printf("error get action runs: %v\n", err)
		return
	}

	var workflows WorkflowRunsResponse
	if err = json.Unmarshal(body, &workflows); err != nil {
		log.Println(err)
		return
	}
	for _, run := range workflows.WorkflowRuns {
		if run.Status == "completed" {
			continue // not canceling completed jobs
		}
		if run.HeadBranch != branchName {
			continue // should not happen cuz we pre-filter, but better safe than sorry
		}
		if run.HeadSha == currentSha {
			continue // not canceling my own jobs
		}
		if currentRunNumber != 0 && run.RunNumber > currentRunNumber {
			continue // only canceling previous executions, not newer ones
		}
		if run.WorkflowId != workflowId {
			log.Printf(" ! found run %v, number %v, workflow = %v | want = %v", run.Id, run.RunNumber, run.WorkflowId, workflowId)
			continue
		}
		log.Printf("canceling run https://github.com/%s/actions/runs/%d\n", githubRepo, run.Id)
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			if err := cancelWorkflow(id); err != nil {
				log.Printf("error cancel workflow: %v\n", err)
			}
		}(run.Id)
	}
	wg.Wait()
}

func getWorkflowId() (ret int64, err error) {
	var query url.Values
	query.Set("per_page", "100")
	body, err := doRequest(githubApi("repos/%s/actions/workflows", githubRepo), query)
	if err != nil {
		return
	}

	var workflows WorkflowsResponse
	if err = json.Unmarshal(body, &workflows); err != nil {
		return
	}

	for _, item := range workflows.Workflows {
		if item.Name == workflowName {
			return item.Id, nil
		}
	}
	return 0, fmt.Errorf("workflow with name '%v' did not exists", workflowName)
}

func doRequest(url string, query url.Values) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.URL.RawQuery = query.Encode()
	response, err := githubRequest(request)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
func githubApi(format string, a ...interface{}) string {
	return fmt.Sprintf(githubApiUrl+"/"+format, a...)
}
