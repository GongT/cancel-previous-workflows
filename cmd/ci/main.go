package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/GongT/cancel-previous-workflows/internal/github"
)

var currentSha = os.Getenv("GITHUB_SHA")
var currentRunNumber, _ = strconv.Atoi(os.Getenv("GITHUB_RUN_NUMBER"))
var isCancelAll = len(os.Getenv("NO_FILTER")) > 0
var needDeleteOld = strings.ToLower(os.Getenv("DELETE")) == "yes"

func main() {
	log.Printf("CurrentRunNumber=%v\n", currentRunNumber)
	log.Printf("CurrentWorkflowName=%v\n", github.GetCurrentWorkflowName())
	log.Printf("BranchName=%v\n", github.GetBranchName())
	log.Printf("GITHUB_SHA=%v\n", currentSha)
	log.Printf("isCancelAll=%v\n", isCancelAll)

	runsList := requestList()
	if runsList == nil {
		return
	}
	log.Printf("  * found %v runs", len(runsList))

	olderThanMe := filterOld(runsList)

	shoudOperate := filterBranch(olderThanMe)

	var shouldCancel []*github.WorkflowRun
	for _, run := range shoudOperate {
		if run.Status != github.StateTypeComplete {
			shouldCancel = append(shouldCancel, run)
		}
	}
	doCancel(shouldCancel)

	doDelete(shoudOperate)
}

func requestList() (runsList []*github.WorkflowRun) {
	log.Printf("finding workflow %v in repo %v\n", github.GetCurrentWorkflowName(), github.GetCurrentRepo())
	var workflowId int64
	if isCancelAll {
		workflowId = 0
	} else {
		id, err := github.GetWorkflowId()
		if err != nil {
			log.Printf("error find workflow id: %v\n", err)
			return
		}
		workflowId = id
	}
	log.Printf("    id is %v\n", workflowId)

	if needDeleteOld {
		if list, err := github.ListWorkflowRuns(workflowId, github.StateTypeAny); err == nil {
			runsList = list
		} else {
			log.Printf("error get action runs: %v\n", err)
		}
	} else {
		if queued, err := github.ListWorkflowRuns(workflowId, github.StateTypeQueue); err == nil {
			runsList = append(runsList, queued...)
		} else {
			log.Printf("error get action runs: %v\n", err)
		}
		if inProgress, err := github.ListWorkflowRuns(workflowId, github.StateTypeInProgress); err == nil {
			runsList = append(runsList, inProgress...)
		} else {
			log.Printf("error get action runs: %v\n", err)
		}
	}
	return
}

func filterOld(runsListDedup []*github.WorkflowRun) []*github.WorkflowRun {
	var olderThanMe []*github.WorkflowRun
	for _, run := range runsListDedup {
		if currentRunNumber != 0 && run.RunNumber >= currentRunNumber {
			continue
		}
		olderThanMe = append(olderThanMe, run)
	}
	return olderThanMe
}

func filterBranch(olderThanMe []*github.WorkflowRun) []*github.WorkflowRun {
	var shoudOperate []*github.WorkflowRun
	if !isCancelAll {
		for _, run := range olderThanMe {
			if run.HeadBranch != github.GetBranchName() {
				log.Printf("      ! [%v] skip other branch: %v != %v", run.Id, run.HeadBranch, github.GetBranchName())
				continue // should not happen cuz we pre-filter, but better safe than sorry
			}
			// if run.HeadSha == currentSha {
			// 	continue // not canceling my own jobs
			// }
			shoudOperate = append(shoudOperate, run)
		}
	}
	return shoudOperate
}

func doCancel(shouldCancel []*github.WorkflowRun) {
	count := len(shouldCancel)

	log.Printf("          %v should cancel", count)

	s := strconv.Itoa(len(strconv.Itoa(count)))

	var okCnt, errCnt int

	for index, run := range shouldCancel {
		id := run.Id
		err := github.CancelWorkflow(id)

		if err == nil {
			okCnt++
			log.Printf("  [%"+s+"d/%"+s+"d] done [%v]\n", index+1, count, id)
		} else {
			errCnt++
			log.Printf("  [%"+s+"d/%"+s+"d] error [%v]: %v\n", index+1, count, id, err)
		}
	}
	log.Printf("All done, %v success, %v error.\n", okCnt, errCnt)
}

func doDelete(shoudDelete []*github.WorkflowRun) {
	if needDeleteOld {
		log.Println("Delete all:")
		time.Sleep(time.Second * 30)
		count := len(shoudDelete)
		log.Printf("          %v should delete", count)
		for index, run := range shoudDelete {
			github.RmeoveWorkflowRunVoid(run, index, count)
		}
		log.Println("Delete all done.")
	} else {
		log.Println("Not delete old runs.")
	}
}
