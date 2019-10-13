package cloudbuildnotifier

// Notifier sends messages
type Notifier interface {
	Send(text string, color string) error
}
