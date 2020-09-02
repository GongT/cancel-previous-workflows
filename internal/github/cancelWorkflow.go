package github

func CancelWorkflow(id int64) error {
	_, err := DoRequest("POST", ApiUrl("actions/runs/%d/cancel", id), nil)
	return err
}
