package main

import (
	"log"
	"os"
	"regexp"

	"github.com/GongT/cancel-previous-workflows/internal/github"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	if len(github.GetBranchName()) == 0 {
		log.Println("branch [env:GITHUB_REF] is required.")
		os.Exit(1)
	}

	filter, err := regexp.Compile(os.Getenv("FILTER_REGEX"))
	if err != nil {
		log.Printf("regexp error: %v", err)
	}

	log.Printf("start all workflow [matching /%v/] in repo %v (ref: %v)\n", filter.String(), github.GetCurrentRepo(), github.GetBranchName())

	q := make(map[string]string, 1)
	q["ref"] = github.GetBranchName()

	start := func(r *github.Workflow, current, total int) {
		if !filter.MatchString(r.Name) && !filter.MatchString(r.Path) {
			log.Printf("  [%2d/%2d] skip  [%v]: %v\n", current, total, r.Name, r.Path)
			return
		}

		body, err := github.DoRequest("POST", github.ApiUrl("actions/workflows/%v/dispatches", r.Id), q)
		if err == nil {
			log.Printf("  [%2d/%2d] done  [%v]\n", current, total, r.Name)
		} else {
			log.Printf("  [%2d/%2d] error [%v]: %v\n", current, total, r.Name, err)
			if body != nil {
				spew.Dump(body)
			}
		}
	}

	if err := github.ForeachWorkflow(start); err == nil {
	} else {
		log.Printf("error when list workflows: %v\n", err)
		return
	}
}
