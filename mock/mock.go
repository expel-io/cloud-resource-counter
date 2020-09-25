package mock

// MockedActivityMonitor does stuff...
type MockedActivityMonitor struct {
	ErrorOccured bool
}

// StartAction does stuff...
func (m *MockedActivityMonitor) StartAction(format string, v ...interface{}) {

}

// CheckError does more stuff...
func (m *MockedActivityMonitor) CheckError(err error) {
	if err != nil {
		m.ErrorOccured = true
	}
}

// ActionError also does stuff...
func (m *MockedActivityMonitor) ActionError(format string, v ...interface{}) {

}

// EndAction doesn't want to miss out, so it does stuff..
func (m *MockedActivityMonitor) EndAction(format string, v ...interface{}) {

}
