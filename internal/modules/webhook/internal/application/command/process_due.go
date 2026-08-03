package command

type ProcessDue struct {
	Limit int
}

type ProcessDueResult struct {
	Processed int
	Sent      int
	Failed    int
	Retried   int
}
