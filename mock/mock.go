/******************************************************************************
Cloud Resource Counter
File: mock.go

Summary: Mocks of common interfaces
******************************************************************************/

package mock

import (
	"fmt"
)

// ActivityMonitorImpl is a mock of the ActionMonitor interface.
// It essentially records which activity has taken place (started,
// errored, ended).
//
type ActivityMonitorImpl struct {
	ActionStarted  bool
	ErrorOccured   bool
	ActionEnded    bool
	ErrorMessage   string
	ProgramExited bool
	ExitCode       int
	Messages       []string
}

// Message does nothing
func (m *ActivityMonitorImpl) Message(format string, v ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, v...))
}

// StartAction records that an action was started.
func (m *ActivityMonitorImpl) StartAction(format string, v ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, v...))
	m.ActionStarted = true
}

// CheckError is invoked to inspect for an error.
func (m *ActivityMonitorImpl) CheckError(err error) bool {
	// Did we encounter an error?
	if err != nil {
		// Record the error message
		m.ErrorMessage = err.Error()

		// Redirect to the ActionError method
		m.ActionError("Error: %s", m.ErrorMessage)

		return true
	}

	return false
}

// ActionError is what what would be called if we encounter an error.
func (m *ActivityMonitorImpl) ActionError(format string, v ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, v...))
	m.ErrorOccured = true
}

// EndAction records that an action was ended.
func (m *ActivityMonitorImpl) EndAction(format string, v ...interface{}) {
	m.Messages = append(m.Messages, fmt.Sprintf(format, v...))
	m.ActionEnded = true
}

// Exit records that the program wishes to exit
func (m *ActivityMonitorImpl) Exit(resultStatus int) {
	m.ProgramExited = true
	m.ExitCode = resultStatus
}
