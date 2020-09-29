package mock

// ActivityMonitorImpl is a mock of the ActionMonitor interface.
// It essentially records which activity has taken place (started,
// errored, ended).
//
type ActivityMonitorImpl struct {
	ActionStarted bool
	ErrorOccured  bool
	ActionEnded   bool
	ErrorMessage  string
}

// Message does nothing
func (m *ActivityMonitorImpl) Message(format string, v ...interface{}) {

}

// StartAction records that an action was started.
func (m *ActivityMonitorImpl) StartAction(format string, v ...interface{}) {
	m.ActionStarted = true
}

// CheckError is invoked to inspect for an error.
func (m *ActivityMonitorImpl) CheckError(err error) bool {
	if err != nil {
		m.ErrorOccured = true
		m.ErrorMessage = err.Error()

		return true
	}

	return false
}

// ActionError is what what would be called if we encounter an error.
func (m *ActivityMonitorImpl) ActionError(format string, v ...interface{}) {

}

// EndAction records that an action was ended.
func (m *ActivityMonitorImpl) EndAction(format string, v ...interface{}) {
	m.ActionEnded = true
}
