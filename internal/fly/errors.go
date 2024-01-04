package fly

import "fmt"

type AppNotFoundError struct {
	App string
	Err error
}

func (e AppNotFoundError) Error() string {
	return fmt.Sprintf("app '%s' not found: %v", e.App, e.Err)
}

type MachineNotFoundError struct {
	App     string
	Machine string
	Err     error
}

func (e MachineNotFoundError) Error() string {
	return fmt.Sprintf("app '%s' machine '%s' not found: %v", e.App, e.Machine, e.Err)
}
