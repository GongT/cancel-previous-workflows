package github

type StateType string

const (
	StateTypeQueue      StateType = "queued"
	StateTypeComplete   StateType = "completed"
	StateTypeInProgress StateType = "in_progress"
	StateTypeAny        StateType = ""
)

type ConclusionType string

const (
	ConclusionSuccess        ConclusionType = "success"
	ConclusionFailure        ConclusionType = "failure"
	ConclusionNeutral        ConclusionType = "neutral"
	ConclusionCancelled      ConclusionType = "cancelled"
	ConclusionSkipped        ConclusionType = "skipped"
	ConclusionTimed_out      ConclusionType = "timed_out"
	ConclusionActionRequired ConclusionType = "action_required"
	ConclusionAny            ConclusionType = ""
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
	Id         int64          `json:"id"`
	Status     StateType      `json:"status"`
	Conclusion ConclusionType `json:"conclusion"`
	HeadSha    string         `json:"head_sha"`
	HeadBranch string         `json:"head_branch"`
	RunNumber  int            `json:"run_number"`
	WorkflowId int64          `json:"workflow_id"`
}

type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}
