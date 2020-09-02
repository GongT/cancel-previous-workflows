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
	"time"
)

type StateType string

const (
	StateTypeQueue      StateType = "queued"
	StateTypeComplete   StateType = "completed"
	StateTypeInProgress StateType = "in_progress"
)

type Workflow struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type WorkflowsResponse struct {
	TotalCount int        `json:"total_count"`
	Workflows  []Workflow `json:"workflows"`
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
	TotalCount   int           `json:"total_count"`
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
var isCancelAll = len(os.Getenv("CANCEL_ALL")) > 0
var requestPerPage = 100
var ua = "CreatedBy/GongT repo/" + githubRepo + " workflow/" + workflowName + " run/" + os.Getenv("GITHUB_RUN_NUMBER")

func githubRequest(request *http.Request) (*http.Response, error) {
	request.Header.Set("User-Agent", ua)
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Authorization", fmt.Sprintf("token %s", githubToken))
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func cancelWorkflow(id int64) (string, error) {
	request, err := http.NewRequest("POST", githubApi("repos/%s/actions/runs/%d/cancel", githubRepo, id), nil)
	if err != nil {
		return "", err
	}
	response, err := githubRequest(request)
	if err != nil {
		return "", err
	}
	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != http.StatusAccepted {
		return "", errors.New(fmt.Sprintf("failed to cancel workflow #%d, status code: %d, body: %s", id, response.StatusCode, body))
	}
	return string(body), nil
}

func isIn(arr []*WorkflowRun, val *WorkflowRun) bool {
	for _, item := range arr {
		if val.Id == item.Id {
			return true
		}
	}
	return false
}

// I don't wan't to fail the current workflow if I fail canceling previous workflow's => so I only log errors
func main() {
	if len(githubApiUrl) == 0 {
		githubApiUrl = "https://api.github.com"
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient = http.Client{Transport: customTransport, Timeout: time.Minute}

	var runsList []*WorkflowRun
	if queued, err := listRuns(StateTypeQueue); err == nil {
		runsList = append(runsList, queued...)
	} else {
		log.Printf("error get action runs: %v\n", err)
		return
	}
	if inProgress, err := listRuns(StateTypeInProgress); err == nil {
		runsList = append(runsList, inProgress...)
	} else {
		log.Printf("error get action runs: %v\n", err)
		return
	}

	runsListDedup := make([]*WorkflowRun, 0, len(runsList))
	skip := make([]int64, 0, len(runsList))
	for _, r := range runsList {
		if !isIn(runsListDedup, r) {
			runsListDedup = append(runsListDedup, r)
		} else {
			skip = append(skip, r.Id)
		}
	}
	if (len(skip)) > 0 {
		log.Printf("skip ids: %v", skip)
	}

	log.Printf("  * found %v runs", len(runsListDedup))

	var shouldCancel []*WorkflowRun
	if isCancelAll {
		for _, run := range runsListDedup {
			if run.RunNumber == currentRunNumber {
				continue // skip my self anyway
			}
			shouldCancel = append(shouldCancel, run)
		}
	} else {
		log.Printf("finding workflow names in repo %s\n", githubRepo)
		workflowId, err := getWorkflowId()
		if err != nil {
			log.Printf("error find workflow id: %v\n", err)
			return
		}

		for _, run := range runsListDedup {
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

			shouldCancel = append(shouldCancel, run)
		}
	}

	log.Printf("          %v should cancel", len(shouldCancel))

	count := len(shouldCancel)

	s := strconv.Itoa(len(strconv.Itoa(count)))

	var okCnt, errCnt int

	for index, run := range shouldCancel {
		id := run.Id
		_, err := cancelWorkflow(id)

		if err == nil {
			okCnt++
			log.Printf("  [%"+s+"d/%"+s+"d] done [%v]: %v\n", index+1, count, id)
		} else {
			errCnt++
			log.Printf("  [%"+s+"d/%"+s+"d] error [%v]: %v\n", index+1, count, id, err)
		}
	}
	log.Printf("All done, %v success, %v error.\n", okCnt, errCnt)
}

func getWorkflowId() (int64, error) {
	query := make(url.Values)
	query.Set("per_page", strconv.Itoa(requestPerPage))

	api := githubApi("repos/%s/actions/workflows", githubRepo)

	var curr = 0
	for {
		log.Printf("  * page %v...", curr)
		query.Set("page", strconv.Itoa(curr))

		body, err := doRequest(api, query)
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

func listRuns(state StateType) (runs []*WorkflowRun, err error) {
	log.Printf("listing %v runs for branch %s in repo %s\n", state, branchName, githubRepo)

	query := make(url.Values)
	query.Set("per_page", strconv.Itoa(requestPerPage))
	query.Set("branch", branchName)
	query.Set("status", string(state))

	api := githubApi("repos/%s/actions/runs", githubRepo)

	var curr = 0
	for {
		log.Printf("  * page %v...", curr)
		query.Set("page", strconv.Itoa(curr))

		body, err := doRequest(api, query)
		if err != nil {
			log.Printf("    error request api: %v", err)
			continue
		}

		var workflows WorkflowRunsResponse
		err = json.Unmarshal(body, &workflows)
		if err != nil {
			log.Printf("    error parse json: %v", err)
			continue
		}

		for _, item := range workflows.WorkflowRuns {
			r := &WorkflowRun{
				Id:         item.Id,
				Status:     item.Status,
				HeadSha:    item.HeadSha,
				HeadBranch: item.HeadBranch,
				RunNumber:  item.RunNumber,
				WorkflowId: item.WorkflowId,
			}
			if !isIn(runs, r) {
				runs = append(runs, r)
			} else {
				log.Printf("  ! duplicate id: %v", r.Id)
			}
		}
		curr++

		log.Printf("    got=%v | current size=%v | total count=%v", len(workflows.WorkflowRuns), len(runs), workflows.TotalCount)
		if workflows.TotalCount <= len(runs) || len(workflows.WorkflowRuns) == 0 {
			break
		}
	}

	return
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
