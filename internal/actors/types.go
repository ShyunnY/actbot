package actors

type Actor interface {
	Handler() error

	Capture(event GenericEvent) bool

	Name() string
}

type GenericEvent struct {
	// This represents the actual GitHub events
	Event any
}
