package sipengine


type ShutdownSignalled struct {
	message string
}

func NewShutDownSignalError(message string) ShutdownSignalled {
	return ShutdownSignalled{message:message}
}

func (s ShutdownSignalled) Error() string {
	return s.message
}