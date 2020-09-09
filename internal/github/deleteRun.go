package github

import "log"

func RmeoveWorkflowRun(r *WorkflowRun, current, total int) error {
	_, err := DoRequest("DELETE", ApiUrl("actions/runs/%d", r.Id), nil)
	return err
}

func RmeoveWorkflowRunVoid(r *WorkflowRun, current, total int) {
	body, err := DoRequest("DELETE", ApiUrl("actions/runs/%d", r.Id), nil)
	if err == nil {
		log.Printf("  [%4d/%4d] done  [%v]: %v\n", current, total, r.Id, string(body))
	} else {
		log.Printf("  [%4d/%4d] error [%v]: %v\n", current, total, r.Id, err)
	}
}
