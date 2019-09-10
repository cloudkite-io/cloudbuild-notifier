package cloudbuildnotifier

// Notifier sends messages
type Notifier interface {
	Send(string) error
}
