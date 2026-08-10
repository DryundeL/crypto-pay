package command

type ExpireDue struct {
	Limit int
}

type ExpireDueResult struct {
	Expired int
	Skipped int
}
