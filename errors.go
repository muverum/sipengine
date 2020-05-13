package sipengine


type ShutdownSignalledError struct {
	message string
}

func NewShutDownSignalError(message string) ShutdownSignalledError {
	return ShutdownSignalledError{message: message}
}

func (s ShutdownSignalledError) Error() string {
	return s.message
}


type MessageTerminationError struct {
	message string
}

func NewMessageTerminationError(input string) MessageTerminationError {
	return MessageTerminationError{message:input}
}

func (m MessageTerminationError) Error() string {
	return m.message
}