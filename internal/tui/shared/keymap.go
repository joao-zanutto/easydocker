package shared

func CanOpenShell(state string) bool {
	// Only running containers support shell execution
	return state == "running"
}
