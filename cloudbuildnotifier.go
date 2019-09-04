package cloudbuildnotifier

// Subscriber listens for events
type Subscriber interface {
	Receive() (string, error)
}

// Notifier sends messages
type Notifier interface {
	Send(string) error
}
