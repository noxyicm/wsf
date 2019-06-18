package event

// Event is an event to be logged
type Event struct {
	Timestamp    string
	Message      string
	Priority     int
	PriorityName string
	Info         map[string]string
}
